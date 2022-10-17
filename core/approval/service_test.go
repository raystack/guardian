package approval_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/odpf/guardian/core/approval"
	approvalmocks "github.com/odpf/guardian/core/approval/mocks"
	"github.com/odpf/guardian/domain"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository    *approvalmocks.Repository
	mockPolicyService *approvalmocks.PolicyService

	service *approval.Service
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(approvalmocks.Repository)
	s.mockPolicyService = new(approvalmocks.PolicyService)

	s.service = approval.NewService(approval.ServiceDeps{
		s.mockRepository,
		s.mockPolicyService,
	})
}

func (s *ServiceTestSuite) TestBulkInsert() {
	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("repository error")
		s.mockRepository.EXPECT().
			BulkInsert(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedError).Once()

		actualError := s.service.BulkInsert(context.Background(), []*domain.Approval{})

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
		s.mockPolicyService.On("GetOne", mock.Anything, mock.Anything, mock.Anything).Return(nil, expectedError).Once()
		actualError := s.service.AdvanceApproval(context.Background(), &testappeal)
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
						Name:      "step-1",
						ApproveIf: `$appeal.resource.details.owner == "test-owner"`,
					},
					{
						Name:      "step-2",
						ApproveIf: `$appeal.resource.details.owner == "test-owner"`,
					},
					{
						Name:      "step-3",
						ApproveIf: `$appeal.resource.details.owner == "test-owner"`,
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

		actualError := s.service.AdvanceApproval(context.Background(), &testappeal)
		s.Nil(actualError)
	})

	s.Run("should autofill rejection reason on auto-reject", func() {
		rejectionReason := "test rejection reason"
		testAppeal := &domain.Appeal{
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
						Name:            "step-1",
						Strategy:        "auto",
						RejectionReason: rejectionReason,
						ApproveIf:       `false`, // hard reject for testing purpose
					},
				},
			},
			Approvals: []*domain.Approval{
				{
					Status: domain.ApprovalStatusPending,
					Index:  0,
				},
			},
		}
		expectedApprovals := []*domain.Approval{
			{
				Status: domain.ApprovalStatusRejected,
				Index:  0,
				Reason: rejectionReason,
			},
		}

		actualError := s.service.AdvanceApproval(context.Background(), testAppeal)

		s.Nil(actualError)
		s.Equal(expectedApprovals, testAppeal.Approvals)
	})

	s.Run("should update approval statuses", func() {
		resourceFlagStep := &domain.Step{
			Name: "resourceFlagStep",
			When: "$appeal.resource.details.flag == true",
			Approvers: []string{
				"user@email.com",
			},
		}
		humanApprovalStep := &domain.Step{
			Name: "humanApprovalStep",
			Approvers: []string{
				"human@email.com",
			},
		}

		testCases := []struct {
			name                     string
			appeal                   *domain.Appeal
			steps                    []*domain.Step
			existingApprovalStatuses []string
			expectedApprovalStatuses []string
			expectedErrorStr         string
		}{
			{
				name: "initial process, When on the first step",
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Details: map[string]interface{}{
							"flag": false,
						},
					},
				},
				steps: []*domain.Step{
					resourceFlagStep,
					humanApprovalStep,
				},
				existingApprovalStatuses: []string{
					domain.ApprovalStatusPending,
					domain.ApprovalStatusBlocked,
				},
				expectedApprovalStatuses: []string{
					domain.ApprovalStatusSkipped,
					domain.ApprovalStatusPending,
				},
			},
			{
				name: "When expression fulfilled",
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Details: map[string]interface{}{
							"flag": true,
						},
					},
				},
				steps: []*domain.Step{
					humanApprovalStep,
					resourceFlagStep,
					humanApprovalStep,
				},
				existingApprovalStatuses: []string{
					domain.ApprovalStatusApproved,
					domain.ApprovalStatusPending,
					domain.ApprovalStatusBlocked,
				},
				expectedApprovalStatuses: []string{
					domain.ApprovalStatusApproved,
					domain.ApprovalStatusPending,
					domain.ApprovalStatusBlocked,
				},
			},
			{
				name: "should access nested fields properly in expression",
				appeal: &domain.Appeal{
					Resource: &domain.Resource{},
				},
				steps: []*domain.Step{
					{
						Strategy:  "manual",
						When:      `$appeal.details != nil && $appeal.details.foo != nil && $appeal.details.bar != nil && ($appeal.details.foo.foo contains "foo" || $appeal.details.foo.bar contains "bar")`,
						Approvers: []string{"approver1@email.com"},
					},
					{
						Strategy:  "manual",
						Approvers: []string{"approver2@email.com"},
					},
				},
				existingApprovalStatuses: []string{
					domain.ApprovalStatusPending,
					domain.ApprovalStatusBlocked,
				},
				expectedApprovalStatuses: []string{
					domain.ApprovalStatusSkipped,
					domain.ApprovalStatusPending,
				},
			},
			{
				name: "should return error if failed when evaluating expression",
				appeal: &domain.Appeal{
					Resource: &domain.Resource{},
				},
				steps: []*domain.Step{
					{
						Strategy:  "manual",
						When:      `$appeal.details != nil && $appeal.details.foo != nil && $appeal.details.bar != nil && $appeal.details.foo.foo contains "foo" || $appeal.details.foo.bar contains "bar"`,
						Approvers: []string{"approver1@email.com"},
					},
					{
						Strategy:  "manual",
						Approvers: []string{"approver2@email.com"},
					},
				},
				existingApprovalStatuses: []string{
					domain.ApprovalStatusPending,
					domain.ApprovalStatusPending,
				},
				expectedErrorStr: "evaluating expression ",
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				appeal := *tc.appeal
				for i, s := range tc.existingApprovalStatuses {
					appeal.Approvals = append(appeal.Approvals, &domain.Approval{
						Status: s,
						Index:  i,
					})
				}
				appeal.Policy = &domain.Policy{
					Steps: tc.steps,
				}
				actualError := s.service.AdvanceApproval(context.Background(), &appeal)
				if tc.expectedErrorStr == "" {
					s.Nil(actualError)
					for i, a := range appeal.Approvals {
						s.Equal(a.Status, tc.expectedApprovalStatuses[i])
					}
				} else {
					s.Contains(actualError.Error(), tc.expectedErrorStr)
				}
			})
		}
	})
}

func (s *ServiceTestSuite) TestAddApprover() {
	s.Run("should return nil error on success", func() {
		expectedApprover := &domain.Approver{
			ApprovalID: uuid.New().String(),
			Email:      "user@example.com",
		}
		s.mockRepository.EXPECT().AddApprover(mock.AnythingOfType("*context.emptyCtx"), expectedApprover).Return(nil)

		err := s.service.AddApprover(context.Background(), expectedApprover.ApprovalID, expectedApprover.Email)

		s.NoError(err)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if repository returns an error", func() {
		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().AddApprover(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(expectedError)

		err := s.service.AddApprover(context.Background(), "", "")

		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestDeleteApprover() {
	s.Run("should return nil error on success", func() {
		approvalID := uuid.New().String()
		approverEmail := "user@example.com"

		s.mockRepository.EXPECT().DeleteApprover(mock.AnythingOfType("*context.emptyCtx"), approvalID, approverEmail).Return(nil)

		err := s.service.DeleteApprover(context.Background(), approvalID, approverEmail)

		s.NoError(err)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if repository returns an error", func() {
		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().DeleteApprover(mock.AnythingOfType("*context.emptyCtx"), mock.Anything, mock.Anything).Return(expectedError)

		err := s.service.DeleteApprover(context.Background(), "", "")

		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})
}
