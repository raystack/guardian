package approval_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/approval"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository      *mocks.ApprovalRepository
	mockPolicyService   *mocks.PolicyService

	service domain.ApprovalService
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(mocks.ApprovalRepository)
	s.mockPolicyService = new(mocks.PolicyService)

	s.service = approval.NewService(s.mockRepository, s.mockPolicyService)
}

func (s *ServiceTestSuite) TestBulkInsert() {
	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("repository error")
		s.mockRepository.On("BulkInsert", mock.Anything).Return(expectedError).Once()

		actualError := s.service.BulkInsert([]*domain.Approval{})

		s.EqualError(actualError, expectedError.Error())
	})
}

func (s *ServiceTestSuite) TestAdvanceApproval() {
	s.Run("should return error if got error on finding policies", func() {
		testappeal := domain.Appeal{
			PolicyID:      "test-id",
			PolicyVersion: 1,
		}
		expectedError := errors.New("policy error")
		s.mockPolicyService.On("GetOne", mock.Anything, mock.Anything).Return(nil, expectedError).Once()
		actualError := s.service.AdvanceApproval(&testappeal)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should resolve multiple automatic approval steps", func() {
		testappeal := domain.Appeal{
			PolicyID:      "test-id",
			PolicyVersion: 1,
			Resource: &domain.Resource{
				Name: "grafana",
				Details: map[string]interface{}{
					"owner": "test-owner",
				},
			},
			Policy: &domain.Policy{
				ID:      "test-id",
				Version: 1,
				Steps: []*domain.Step{
					{
						Name: "step-1",
						Conditions: []*domain.Condition{
							{
								Field: "$resource.details.owner",
								Match: &domain.MatchCondition{
									Eq: "test-owner",
								},
							},
						},
						Dependencies: []string{},
					},
					{
						Name: "step-2",
						Conditions: []*domain.Condition{
							{
								Field: "$resource.details.owner",
								Match: &domain.MatchCondition{
									Eq: "test-owner",
								},
							},
						},
						Dependencies: []string{},
					},
					{
						Name: "step-3",
						Conditions: []*domain.Condition{
							{
								Field: "$resource.details.owner",
								Match: &domain.MatchCondition{
									Eq: "test-owner",
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			Approvals: []*domain.Approval{
				{
					Status: "pending",
					Index:  0,
				},
				{
					Status: "blocked",
					Index:  1,
				},
				{
					Status: "blocked",
					Index:  2,
				},
			},
		}

		actualError := s.service.AdvanceApproval(&testappeal)
		s.Nil(actualError)
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
