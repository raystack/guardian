package provider_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
auth: service-account-key.json
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
			Type: "google_bigquery",
			URN:  "gcp-project-id",
			Auth: "service-account-key.json",
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
		expectedResponseBody := &domain.Provider{
			ID:     expectedID,
			Type:   provider.Type,
			URN:    provider.URN,
			Config: provider.Config,
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

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
