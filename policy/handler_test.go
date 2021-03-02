package policy_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/policy"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
	mockPolicyService *mocks.PolicyService
	handler           *policy.Handler
	res               *httptest.ResponseRecorder
}

func (s *HandlerTestSuite) Setup() {
	s.mockPolicyService = new(mocks.PolicyService)
	s.handler = &policy.Handler{s.mockPolicyService}
	s.res = httptest.NewRecorder()
}

func (s *HandlerTestSuite) SetupTest() {
	s.Setup()
}

func (s *HandlerTestSuite) AfterTest() {
	s.mockPolicyService.AssertExpectations(s.T())
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
		testCases := []struct {
			name           string
			invalidPayload string
		}{
			{
				name: "incomplete config",
				invalidPayload: `
id: provider_type_test
`,
			},
			{
				name: "a step contains both approvers and conditions",
				invalidPayload: `
id: bq_dataset
steps:
  - name: step_name
    conditions:
    - field: field_name
      match:
        eq: true
		approvers: approver_names
`,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.Setup()
				req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.invalidPayload))

				expectedStatusCode := http.StatusBadRequest

				s.handler.Create(s.res, req)
				actualStatusCode := s.res.Result().StatusCode

				s.Equal(expectedStatusCode, actualStatusCode)
			})
		}
	})

	validPayload := `
id: bq_dataset
steps:
  - name: check_if_dataset_is_pii
    description: pii dataset needs additional approval from the team lead
    conditions:
    - field: $resource.details.is_pii
      match:
        eq: true
    allow_failed: true
  - name: supervisor_approval
    description: 'only will get evaluated if check_if_dataset_is_pii return true'
    dependencies: [check_if_dataset_is_pii]
    approvers: $user.profile.team_leads.[].email
`
	s.Run("should return internal server error if policy service returns error", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(validPayload))

		expectedStatusCode := http.StatusInternalServerError
		expectedError := errors.New("service error")
		s.mockPolicyService.On("Create", mock.Anything).Return(expectedError)

		s.handler.Create(s.res, req)
		actualStatusCode := s.res.Result().StatusCode

		s.Equal(expectedStatusCode, actualStatusCode)
	})

	policy := &domain.Policy{
		ID: "bq_dataset",
		Steps: []*domain.Step{
			{
				Name:        "check_if_dataset_is_pii",
				Description: "pii dataset needs additional approval from the team lead",
				Conditions: []*domain.Condition{
					{
						Field: "$resource.details.is_pii",
						Match: &domain.MatchCondition{
							Eq: true,
						},
					},
				},
				AllowFailed: true,
			},
			{
				Name:         "supervisor_approval",
				Description:  "only will get evaluated if check_if_dataset_is_pii return true",
				Dependencies: []string{"check_if_dataset_is_pii"},
				Approvers:    "$user.profile.team_leads.[].email",
			},
		},
	}

	s.Run("should return ok and the newly created policy data on success", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(validPayload))

		expectedStatusCode := http.StatusOK
		expectedResponseBody := &domain.Policy{}
		*expectedResponseBody = *policy
		expectedResponseBody.Version = 1
		s.mockPolicyService.On("Create", policy).Return(nil).Run(func(args mock.Arguments) {
			p := args.Get(0).(*domain.Policy)
			p.Version = 1
		})

		s.handler.Create(s.res, req)
		actualStatusCode := s.res.Result().StatusCode
		actualResponseBody := &domain.Policy{}
		err := json.NewDecoder(s.res.Body).Decode(&actualResponseBody)
		s.NoError(err)

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func (s *HandlerTestSuite) TestFind() {
	s.Run("should return internal server error if policy service returns error", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodGet, "/", nil)

		expectedStatusCode := http.StatusInternalServerError
		expectedError := errors.New("service error")
		s.mockPolicyService.On("Find").Return(nil, expectedError)

		s.handler.Find(s.res, req)
		actualStatusCode := s.res.Result().StatusCode

		s.Equal(expectedStatusCode, actualStatusCode)
	})

	s.Run("should return ok and the policy records on success", func() {
		s.Setup()
		req, _ := http.NewRequest(http.MethodGet, "/", nil)

		expectedStatusCode := http.StatusOK
		expectedResponseBody := []*domain.Policy{}
		s.mockPolicyService.On("Find").Return(expectedResponseBody, nil)

		s.handler.Find(s.res, req)
		actualStatusCode := s.res.Result().StatusCode
		actualResponseBody := []*domain.Policy{}
		err := json.NewDecoder(s.res.Body).Decode(&actualResponseBody)
		s.NoError(err)

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
