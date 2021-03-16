package resource_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
	mockResourceService *mocks.ResourceService
	handler             *resource.Handler
	res                 *httptest.ResponseRecorder
}

func (s *HandlerTestSuite) Setup() {
	s.mockResourceService = new(mocks.ResourceService)
	s.handler = &resource.Handler{s.mockResourceService}
	s.res = httptest.NewRecorder()
}

func (s *HandlerTestSuite) SetupTest() {
	s.Setup()
}

func (s *HandlerTestSuite) AfterTest() {
	s.mockResourceService.AssertExpectations(s.T())
}

func (s *HandlerTestSuite) TestUpdate() {
	s.Run("should return error if got invalid id param", func() {
		testCases := []struct {
			params             map[string]string
			expectedStatusCode int
		}{
			{
				params:             map[string]string{},
				expectedStatusCode: http.StatusBadRequest,
			},
			{
				params: map[string]string{
					"id": "",
				},
				expectedStatusCode: http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			s.Setup()
			req, _ := http.NewRequest(http.MethodPut, "/", nil)
			req = mux.SetURLVars(req, tc.params)

			expectedStatusCode := tc.expectedStatusCode

			s.handler.Update(s.res, req)
			actualStatusCode := s.res.Result().StatusCode

			s.Equal(expectedStatusCode, actualStatusCode)
		}
	})

	s.Run("should return bad request if the payload is invalid", func() {
		testCases := []struct {
			name               string
			payload            string
			expectedStatusCode int
		}{
			{
				name:               "malformed json",
				payload:            `invalid json format...`,
				expectedStatusCode: http.StatusBadRequest,
			},
			{
				name: "invalid json update payload validation",
				payload: `{
	"labels": true,
	"details": true,
}`,
				expectedStatusCode: http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.Setup()
				req, _ := http.NewRequest(http.MethodPut, "/", strings.NewReader(tc.payload))
				req = mux.SetURLVars(req, map[string]string{
					"id": "1",
				})

				expectedStatusCode := tc.expectedStatusCode

				s.handler.Update(s.res, req)
				actualStatusCode := s.res.Result().StatusCode

				s.Equal(expectedStatusCode, actualStatusCode)
			})
		}
	})

	validPayload := `{
	"labels": {
		"key": "value"
	},
	"details": {
		"key": "value"
	}
}`

	s.Run("should return error based on the service error", func() {
		testCases := []struct {
			name                 string
			expectedServiceError error
			expectedStatusCode   int
		}{
			{
				name:                 "any unexpected error from the resource service",
				expectedServiceError: errors.New("any service error"),
				expectedStatusCode:   http.StatusInternalServerError,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.Setup()
				req, _ := http.NewRequest(http.MethodPut, "/", strings.NewReader(validPayload))
				req = mux.SetURLVars(req, map[string]string{
					"id": "1",
				})

				expectedStatusCode := tc.expectedStatusCode
				s.mockResourceService.On("Update", mock.Anything).Return(tc.expectedServiceError).Once()

				s.handler.Update(s.res, req)
				actualStatusCode := s.res.Result().StatusCode

				s.Equal(expectedStatusCode, actualStatusCode)
			})
		}
	})

	s.Run("should return the updated values on success", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodPut, "/", strings.NewReader(validPayload))

		expectedID := uint(1)
		req = mux.SetURLVars(req, map[string]string{
			"id": fmt.Sprintf("%d", expectedID),
		})
		expectedStatusCode := http.StatusOK
		expectedResource := &domain.Resource{
			ID: expectedID,
			Labels: map[string]interface{}{
				"key": "value",
			},
			Details: map[string]interface{}{
				"key": "value",
			},
		}
		expectedResponseBody := expectedResource
		s.mockResourceService.On("Update", expectedResource).Return(nil).Once()

		s.handler.Update(s.res, req)
		actualStatusCode := s.res.Result().StatusCode
		actualResponseBody := &domain.Resource{}
		err := json.NewDecoder(s.res.Body).Decode(&actualResponseBody)
		s.NoError(err)

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
