package provider_test

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
	"github.com/odpf/guardian/provider"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
	mockProviderService *mocks.ProviderService
	handler             *provider.Handler
	res                 *httptest.ResponseRecorder
}

func (s *HandlerTestSuite) Setup() {
	s.mockProviderService = new(mocks.ProviderService)
	s.handler = &provider.Handler{s.mockProviderService}
	s.res = httptest.NewRecorder()
}

func (s *HandlerTestSuite) SetupTest() {
	s.Setup()
}

func (s *HandlerTestSuite) AfterTest() {
	s.mockProviderService.AssertExpectations(s.T())
}

func (s *HandlerTestSuite) TestCreate() {
	s.Run("should return bad request error if received malformed payload", func() {
		s.Setup()
		malformedPayload := `invalid yaml format...`
		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(malformedPayload))

		expectedStatusCode := http.StatusBadRequest

		s.handler.Create(s.res, req)
		actualStatusCode := s.res.Result().StatusCode

		s.Equal(expectedStatusCode, actualStatusCode)
	})

	s.Run("should return bad request if payload validation returns error", func() {
		s.Setup()
		invalidPayload := `
type: provider_type_test
urn: provider_name
`
		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(invalidPayload))

		expectedStatusCode := http.StatusBadRequest

		s.handler.Create(s.res, req)
		actualStatusCode := s.res.Result().StatusCode

		s.Equal(expectedStatusCode, actualStatusCode)
	})

	validPayload := `
type: google_bigquery
urn: gcp-project-id
credentials: service-account-key.json
appeal:
  allow_active_access_extension_in: 7d
resources:
  - type: dataset
    policy:
      id: policy_x
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: roles/bigQuery.dataViewer
          - name: roles/customRole
            target: other-gcp-project-id
      - id: editor
        name: Editor
        permissions:
          - name: roles/bigQuery.dataEditor
`
	s.Run("should return internal server error if provider service returns error", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(validPayload))

		expectedStatusCode := http.StatusInternalServerError
		expectedError := errors.New("service error")
		s.mockProviderService.On("Create", mock.Anything).Return(expectedError)

		s.handler.Create(s.res, req)
		actualStatusCode := s.res.Result().StatusCode

		s.Equal(expectedStatusCode, actualStatusCode)
	})

	provider := &domain.Provider{
		Type: "google_bigquery",
		URN:  "gcp-project-id",
		Config: &domain.ProviderConfig{
			Type:        "google_bigquery",
			URN:         "gcp-project-id",
			Credentials: "service-account-key.json",
			Appeal: &domain.AppealConfig{
				AllowActiveAccessExtensionIn: "7d",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "dataset",
					Policy: &domain.PolicyConfig{
						ID:      "policy_x",
						Version: 1,
					},
					Roles: []*domain.RoleConfig{
						{
							ID:   "viewer",
							Name: "Viewer",
							Permissions: []interface{}{
								map[string]interface{}{"name": "roles/bigQuery.dataViewer"},
								map[string]interface{}{"name": "roles/customRole", "target": "other-gcp-project-id"},
							},
						},
						{
							ID:   "editor",
							Name: "Editor",
							Permissions: []interface{}{
								map[string]interface{}{"name": "roles/bigQuery.dataEditor"},
							},
						},
					},
				},
			},
		},
	}
	s.Run("should return ok and the newly created provider data on success", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(validPayload))

		expectedStatusCode := http.StatusOK
		expectedID := uint(1)
		expectedConfig := &domain.ProviderConfig{}
		*expectedConfig = *provider.Config
		expectedConfig.Credentials = nil
		expectedResponseBody := &domain.Provider{
			ID:     expectedID,
			Type:   provider.Type,
			URN:    provider.URN,
			Config: expectedConfig,
		}
		s.mockProviderService.On("Create", provider).Return(nil).Run(func(args mock.Arguments) {
			p := args.Get(0).(*domain.Provider)
			p.ID = expectedID
		})

		s.handler.Create(s.res, req)
		actualStatusCode := s.res.Result().StatusCode
		actualResponseBody := &domain.Provider{}
		err := json.NewDecoder(s.res.Body).Decode(&actualResponseBody)
		s.NoError(err)

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func (s *HandlerTestSuite) TestFind() {
	s.Run("should return internal server error if provider service returns error", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodGet, "/", nil)

		expectedStatusCode := http.StatusInternalServerError
		expectedError := errors.New("service error")
		s.mockProviderService.On("Find").Return(nil, expectedError)

		s.handler.Find(s.res, req)
		actualStatusCode := s.res.Result().StatusCode

		s.Equal(expectedStatusCode, actualStatusCode)
	})

	s.Run("should return ok and the provider records on success", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodGet, "/", nil)

		expectedStatusCode := http.StatusOK
		expectedResponseBody := []*domain.Provider{}
		s.mockProviderService.On("Find").Return(expectedResponseBody, nil)

		s.handler.Find(s.res, req)
		actualStatusCode := s.res.Result().StatusCode
		actualResponseBody := []*domain.Provider{}
		err := json.NewDecoder(s.res.Body).Decode(&actualResponseBody)
		s.NoError(err)

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
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
				name:               "malformed yaml",
				payload:            `invalid yaml format...`,
				expectedStatusCode: http.StatusBadRequest,
			},
			{
				name: "invalid yaml update payload validation",
				payload: `
appeal:
  - test
	- test2
`,
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

	validPayload := `
appeal:
  allow_active_access_extension_in: 7d
resources:
  - type: type
    policy:
      id: policy_x
      version: 1
`
	s.Run("should return error based on the service error", func() {
		testCases := []struct {
			name                 string
			expectedServiceError error
			expectedStatusCode   int
		}{
			{
				name:                 "provider with the specified id doesn't exists",
				expectedServiceError: provider.ErrRecordNotFound,
				expectedStatusCode:   http.StatusNotFound,
			},
			{
				name:                 "any unexpected error from the provider service",
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
				s.mockProviderService.On("Update", mock.Anything).Return(tc.expectedServiceError).Once()

				s.handler.Update(s.res, req)
				actualStatusCode := s.res.Result().StatusCode

				s.Equal(expectedStatusCode, actualStatusCode)
			})
		}
	})

	s.Run("should return the new version of the provider on success", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodPut, "/", strings.NewReader(validPayload))

		expectedProviderID := uint(1)
		req = mux.SetURLVars(req, map[string]string{
			"id": fmt.Sprintf("%d", expectedProviderID),
		})
		expectedStatusCode := http.StatusOK
		expectedProvider := &domain.Provider{
			ID: expectedProviderID,
			Config: &domain.ProviderConfig{
				Appeal: &domain.AppealConfig{
					AllowActiveAccessExtensionIn: "7d",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: "type",
						Policy: &domain.PolicyConfig{
							ID:      "policy_x",
							Version: 1,
						},
					},
				},
			},
		}
		expectedResponseBody := expectedProvider
		s.mockProviderService.On("Update", expectedProvider).Return(nil).Once()

		s.handler.Update(s.res, req)
		actualStatusCode := s.res.Result().StatusCode
		actualResponseBody := &domain.Provider{}
		err := json.NewDecoder(s.res.Body).Decode(&actualResponseBody)
		s.NoError(err)

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
