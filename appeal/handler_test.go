package appeal_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/odpf/guardian/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
	mockAppealService *mocks.AppealService
	handler           *appeal.Handler
	res               *httptest.ResponseRecorder
}

func (s *HandlerTestSuite) Setup() {
	s.mockAppealService = new(mocks.AppealService)
	s.handler = &appeal.Handler{s.mockAppealService}
	s.res = httptest.NewRecorder()
}

func (s *HandlerTestSuite) SetupTest() {
	s.Setup()
}

func (s *HandlerTestSuite) AfterTest() {
	s.mockAppealService.AssertExpectations(s.T())
}

func (s *HandlerTestSuite) TestGetByID() {
	s.Run("should return bad request if param ID is invalid", func() {
		s.Setup()

		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "invalid"})

		expectedStatusCode := http.StatusBadRequest

		s.handler.GetByID(s.res, req)
		actualStatusCode := s.res.Result().StatusCode

		s.Equal(expectedStatusCode, actualStatusCode)
	})

	s.Run("should return error if got error from appeal service", func() {
		testCases := []struct {
			expectedAppealServiceError error
			expectedStatusCode         int
		}{
			{
				expectedAppealServiceError: errors.New("unexpected service error"),
				expectedStatusCode:         http.StatusInternalServerError,
			},
		}

		for _, tc := range testCases {
			s.Setup()
			req, _ := http.NewRequest(http.MethodGet, "/", nil)

			expectedID := uint(1)
			req = mux.SetURLVars(req, map[string]string{"id": "1"})
			s.mockAppealService.
				On("GetByID", expectedID).
				Return(nil, tc.expectedAppealServiceError).
				Once()
			expectedStatusCode := tc.expectedStatusCode

			s.handler.GetByID(s.res, req)
			actualStatusCode := s.res.Result().StatusCode

			s.Equal(expectedStatusCode, actualStatusCode)
		}
	})

	s.Run("should return 404 not found if record not found", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodGet, "/", nil)

		expectedID := uint(1)
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		s.mockAppealService.
			On("GetByID", expectedID).
			Return(nil, nil).
			Once()
		expectedStatusCode := http.StatusNotFound

		s.handler.GetByID(s.res, req)
		actualStatusCode := s.res.Result().StatusCode

		s.Equal(expectedStatusCode, actualStatusCode)
	})

	s.Run("should return appeal on success", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodGet, "/", nil)

		expectedID := uint(1)
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		expectedResponseBody := &domain.Appeal{
			ID: expectedID,
		}
		s.mockAppealService.
			On("GetByID", expectedID).
			Return(expectedResponseBody, nil).
			Once()
		expectedStatusCode := http.StatusOK

		s.handler.GetByID(s.res, req)
		actualStatusCode := s.res.Result().StatusCode
		actualResponseBody := &domain.Appeal{}
		err := json.NewDecoder(s.res.Body).Decode(actualResponseBody)
		s.NoError(err)

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func (s *HandlerTestSuite) TestCreate() {
	s.Run("should return bad request error if received malformed payload", func() {
		s.Setup()
		malformedPayload := "invalid json format..."
		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(malformedPayload))

		expectedStatusCode := http.StatusBadRequest

		s.handler.Create(s.res, req)
		actualStatusCode := s.res.Result().StatusCode

		s.Equal(expectedStatusCode, actualStatusCode)
	})

	s.Run("should return bad request if payload validation returns error", func() {
		testCases := []struct {
			name           string
			invalidPayload string
		}{
			{
				name: "missing user",
				invalidPayload: `{
	"resources": [
		{
			"id": 1
		},
		{
			"id": 2
		}
	]
}`,
			},
			{
				name: "missing resources",
				invalidPayload: `{
	"user": "test@domain.com"
}`,
			},
			{
				name: "empty resources",
				invalidPayload: `{
	"user": "test@domain.com",
	"resources": []
}`,
			},
		}
		for _, tc := range testCases {
			s.Setup()
			req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.invalidPayload))

			expectedStatusCode := http.StatusBadRequest

			s.handler.Create(s.res, req)
			actualStatusCode := s.res.Result().StatusCode

			s.Equal(expectedStatusCode, actualStatusCode)
		}
	})

	validPayload := `{
	"user": "test@email.com",
	"resources": [
		{
			"id": 1,
			"options": {
				"role": "viewer"
			}
		},
		{
			"id": 2,
			"options": {
				"role": "editor"
			}
		}
	]
}`

	s.Run("should return error based on the error thrown by appeal service", func() {
		testCases := []struct {
			expectedServiceError error
			expectedStatusCode   int
		}{
			{
				expectedServiceError: errors.New("appeal service error"),
				expectedStatusCode:   http.StatusInternalServerError,
			},
		}

		for _, tc := range testCases {
			s.Setup()
			req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(validPayload))

			s.mockAppealService.On("Create", mock.Anything).Return(tc.expectedServiceError).Once()

			s.handler.Create(s.res, req)
			actualStatusCode := s.res.Result().StatusCode

			s.Equal(tc.expectedStatusCode, actualStatusCode)
		}
	})

	s.Run("should return newly created appeals on success", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(validPayload))

		expectedUser := "test@email.com"
		expectedResponseBody := []*domain.Appeal{
			{
				ID:         1,
				User:       expectedUser,
				ResourceID: 1,
				Role:       "viewer",
			},
			{
				ID:         2,
				User:       expectedUser,
				ResourceID: 2,
				Role:       "editor",
			},
		}
		expectedAppeals := []*domain.Appeal{
			{
				User:       expectedUser,
				ResourceID: 1,
				Role:       "viewer",
			},
			{
				User:       expectedUser,
				ResourceID: 2,
				Role:       "editor",
			},
		}
		s.mockAppealService.
			On("Create", expectedAppeals).
			Return(nil).
			Run(func(args mock.Arguments) {
				appeals := args.Get(0).([]*domain.Appeal)
				for i, a := range appeals {
					a.ID = expectedResponseBody[i].ID
				}
			}).
			Once()
		expectedStatusCode := http.StatusOK

		s.handler.Create(s.res, req)
		actualStatusCode := s.res.Result().StatusCode
		actualResponseBody := []*domain.Appeal{}
		err := json.NewDecoder(s.res.Body).Decode(&actualResponseBody)
		s.NoError(err)

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
