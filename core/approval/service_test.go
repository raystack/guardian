package approval_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/core/approval"
	approvalmocks "github.com/odpf/guardian/core/approval/mocks"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository      *approvalmocks.Repository
	mockPolicyService   *approvalmocks.PolicyService
	mockGrantService    *approvalmocks.GrantService
	mockProviderService *approvalmocks.ProviderService
	mockAppealService   *approvalmocks.AppealService
	mockNotifier        *approvalmocks.Notifier
	mockAuditLogger     *approvalmocks.AuditLogger

	service *approval.Service
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(approvalmocks.Repository)
	s.mockPolicyService = new(approvalmocks.PolicyService)
	s.mockGrantService = new(approvalmocks.GrantService)
	s.mockProviderService = new(approvalmocks.ProviderService)
	s.mockAppealService = new(approvalmocks.AppealService)
	s.mockNotifier = new(approvalmocks.Notifier)
	s.mockAuditLogger = new(approvalmocks.AuditLogger)

	s.service = approval.NewService(approval.ServiceDeps{
		s.mockRepository,
		s.mockPolicyService,
		s.mockGrantService,
		s.mockProviderService,
		s.mockNotifier,
		log.NewNoop(),
		s.mockAuditLogger,
	})
	s.service.SetAppealService(s.mockAppealService)
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

func (s *ServiceTestSuite) TestUpdateApproval() {
	timeNow := time.Now()
	approval.TimeNow = func() time.Time {
		return timeNow
	}
	s.Run("should return error if approval action parameter is invalid", func() {
		invalidApprovalActionParameters := []domain.ApprovalAction{
			{
				ApprovalName: "approval_1",
				Actor:        "user@email.com",
				Action:       "name",
			},
			{
				AppealID: "1",
				Actor:    "user@email.com",
				Action:   "name",
			},
			{
				AppealID:     "1",
				ApprovalName: "approval_1",
				Actor:        "invalidemail",
				Action:       "name",
			},
			{
				AppealID:     "1",
				ApprovalName: "approval_1",
				Action:       "name",
			},
			{
				AppealID:     "1",
				ApprovalName: "approval_1",
				Actor:        "user@email.com",
			},
		}

		for _, param := range invalidApprovalActionParameters {
			actualResult, actualError := s.service.UpdateApproval(context.Background(), param)

			s.Nil(actualResult)
			s.Error(actualError)
		}
	})

	validApprovalActionParam := domain.ApprovalAction{
		AppealID:     "1",
		ApprovalName: "approval_1",
		Actor:        "user@email.com",
		Action:       "approve",
	}

	s.Run("should return error if got any from repository while getting appeal details", func() {
		expectedError := errors.New("repository error")
		s.mockAppealService.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.UpdateApproval(context.Background(), validApprovalActionParam)

		s.mockRepository.AssertExpectations(s.T())
		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if appeal not found", func() {
		expectedError := appeal.ErrAppealNotFound
		s.mockAppealService.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.UpdateApproval(context.Background(), validApprovalActionParam)

		s.mockRepository.AssertExpectations(s.T())
		s.Nil(actualResult)
		s.EqualError(actualError, appeal.ErrAppealNotFound.Error())
	})

	s.Run("should return error based on statuses conditions", func() {
		testCases := []struct {
			name          string
			appealStatus  string
			approvals     []*domain.Approval
			expectedError error
		}{
			{
				name:          "appeal not eligible, status: approved",
				appealStatus:  domain.AppealStatusApproved,
				expectedError: appeal.ErrAppealStatusApproved,
			},
			{
				name:          "appeal not eligible, status: rejected",
				appealStatus:  domain.AppealStatusRejected,
				expectedError: appeal.ErrAppealStatusRejected,
			},
			{
				name:          "invalid appeal status",
				appealStatus:  "invalidstatus",
				expectedError: appeal.ErrAppealStatusUnrecognized,
			},
			{
				name:         "previous approval step still on pending",
				appealStatus: domain.AppealStatusPending,
				approvals: []*domain.Approval{
					{
						Name:   "approval_0",
						Status: domain.ApprovalStatusPending,
					},
					{
						Name:   "approval_1",
						Status: domain.ApprovalStatusPending,
					},
				},
				expectedError: appeal.ErrApprovalDependencyIsPending,
			},
			{
				name:         "found one previous approval is reject",
				appealStatus: domain.AppealStatusPending,
				approvals: []*domain.Approval{
					{
						Name:   "approval_0",
						Status: domain.ApprovalStatusRejected,
					},
					{
						Name:   "approval_1",
						Status: domain.ApprovalStatusPending,
					},
				},
				expectedError: appeal.ErrAppealStatusRejected,
			},
			{
				name:         "invalid approval status",
				appealStatus: domain.AppealStatusPending,
				approvals: []*domain.Approval{
					{
						Name:   "approval_0",
						Status: "invalidstatus",
					},
					{
						Name:   "approval_1",
						Status: domain.ApprovalStatusPending,
					},
				},
				expectedError: appeal.ErrApprovalStatusUnrecognized,
			},
			{
				name:         "approval step already approved",
				appealStatus: domain.AppealStatusPending,
				approvals: []*domain.Approval{
					{
						Name:   "approval_0",
						Status: domain.ApprovalStatusApproved,
					},
					{
						Name:   "approval_1",
						Status: domain.ApprovalStatusApproved,
					},
				},
				expectedError: appeal.ErrApprovalStatusApproved,
			},
			{
				name:         "approval step already rejected",
				appealStatus: domain.AppealStatusPending,
				approvals: []*domain.Approval{
					{
						Name:   "approval_0",
						Status: domain.ApprovalStatusApproved,
					},
					{
						Name:   "approval_1",
						Status: domain.ApprovalStatusRejected,
					},
				},
				expectedError: appeal.ErrApprovalStatusRejected,
			},
			{
				name:         "approval step already skipped",
				appealStatus: domain.AppealStatusPending,
				approvals: []*domain.Approval{
					{
						Name:   "approval_0",
						Status: domain.ApprovalStatusApproved,
					},
					{
						Name:   "approval_1",
						Status: domain.ApprovalStatusSkipped,
					},
				},
				expectedError: appeal.ErrApprovalStatusSkipped,
			},
			{
				name:         "invalid approval status",
				appealStatus: domain.AppealStatusPending,
				approvals: []*domain.Approval{
					{
						Name:   "approval_0",
						Status: domain.ApprovalStatusApproved,
					},
					{
						Name:   "approval_1",
						Status: "invalidstatus",
					},
				},
				expectedError: appeal.ErrApprovalStatusUnrecognized,
			},
			{
				name:         "user doesn't have permission",
				appealStatus: domain.AppealStatusPending,
				approvals: []*domain.Approval{
					{
						Name:   "approval_0",
						Status: domain.ApprovalStatusApproved,
					},
					{
						Name:      "approval_1",
						Status:    domain.ApprovalStatusPending,
						Approvers: []string{"another.user@email.com"},
					},
				},
				expectedError: appeal.ErrActionForbidden,
			},
			{
				name:         "approval step not found",
				appealStatus: domain.AppealStatusPending,
				approvals: []*domain.Approval{
					{
						Name:   "approval_0",
						Status: domain.ApprovalStatusApproved,
					},
					{
						Name:   "approval_x",
						Status: domain.ApprovalStatusApproved,
					},
				},
				expectedError: appeal.ErrApprovalNotFound,
			},
		}

		for _, tc := range testCases {
			expectedAppeal := &domain.Appeal{
				ID:        validApprovalActionParam.AppealID,
				Status:    tc.appealStatus,
				Approvals: tc.approvals,
			}
			s.mockAppealService.EXPECT().
				GetByID(mock.AnythingOfType("*context.emptyCtx"), validApprovalActionParam.AppealID).
				Return(expectedAppeal, nil).Once()

			actualResult, actualError := s.service.UpdateApproval(context.Background(), validApprovalActionParam)

			s.Nil(actualResult)
			s.EqualError(actualError, tc.expectedError.Error())
		}
	})

	expectedAppeal := &domain.Appeal{
		ID:     validApprovalActionParam.AppealID,
		Status: domain.AppealStatusPending,
		Approvals: []*domain.Approval{
			{
				Name:   "approval_0",
				Status: domain.ApprovalStatusApproved,
			},
			{
				Name:      "approval_1",
				Status:    domain.ApprovalStatusPending,
				Approvers: []string{"user@email.com"},
			},
			{
				Name:      "approval_2",
				Status:    domain.ApprovalStatusBlocked,
				Approvers: []string{"user@email.com"},
			},
		},
		PolicyID:      "policy-test",
		PolicyVersion: 1,
	}

	s.Run("should return error if got any from approvalService.AdvanceApproval", func() {
		s.mockAppealService.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedAppeal, nil).Once()
		expectedError := errors.New("unexpected error")

		s.mockPolicyService.EXPECT().GetOne(mock.Anything, mock.Anything, mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.UpdateApproval(context.Background(), validApprovalActionParam)

		s.mockRepository.AssertExpectations(s.T())
		s.mockPolicyService.AssertExpectations(s.T())
		s.ErrorIs(actualError, expectedError)
		s.Nil(actualResult)
	})

	s.Run("should terminate existing active grant if present", func() {
		action := domain.ApprovalAction{
			AppealID:     "1",
			ApprovalName: "test-approval-step",
			Action:       "approve",
			Actor:        "approver@example.com",
		}
		appealDetails := &domain.Appeal{
			ID:         "1",
			AccountID:  "user@example.com",
			ResourceID: "1",
			Role:       "test-role",
			Status:     domain.AppealStatusPending,
			Approvals: []*domain.Approval{
				{
					Name:      "test-approval-step",
					Status:    domain.ApprovalStatusPending,
					Approvers: []string{"approver@example.com"},
				},
			},
			Resource: &domain.Resource{
				ID: "1",
			},
		}
		existingGrants := []domain.Grant{
			{
				ID:         "2",
				Status:     domain.GrantStatusActive,
				AccountID:  "user@example.com",
				ResourceID: "1",
				Role:       "test-role",
			},
		}
		expectedRevokedGrant := &domain.Grant{}
		*expectedRevokedGrant = existingGrants[0]
		expectedRevokedGrant.Status = domain.GrantStatusInactive

		s.mockAppealService.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(appealDetails, nil).Once()

		s.mockPolicyService.EXPECT().GetOne(mock.Anything, mock.Anything, mock.Anything).Return(&domain.Policy{}, nil).Once()
		s.mockGrantService.EXPECT().
			List(mock.Anything, mock.Anything).Return(existingGrants, nil).Once()
		expectedNewGrant := &domain.Grant{
			Status:     domain.GrantStatusActive,
			AccountID:  appealDetails.AccountID,
			ResourceID: appealDetails.ResourceID,
		}
		s.mockGrantService.EXPECT().
			Prepare(mock.Anything, mock.Anything).Return(expectedNewGrant, nil).Once()
		s.mockGrantService.EXPECT().
			Revoke(mock.Anything, expectedRevokedGrant.ID, domain.SystemActorName,
				appeal.RevokeReasonForExtension, mock.Anything, mock.Anything).
			Return(expectedNewGrant, nil).Once()
		s.mockAppealService.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), appealDetails).Return(nil).Once()
		s.mockNotifier.EXPECT().Notify(mock.Anything).Return(nil).Once()
		s.mockAuditLogger.EXPECT().Log(mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Once()

		_, actualError := s.service.UpdateApproval(context.Background(), action)

		s.mockRepository.AssertExpectations(s.T())
		s.mockPolicyService.AssertExpectations(s.T())
		s.mockGrantService.AssertExpectations(s.T())
		s.mockNotifier.AssertExpectations(s.T())
		s.mockAuditLogger.AssertExpectations(s.T())
		s.Nil(actualError)
	})

	s.Run("should return updated appeal on success", func() {
		creator := "creator@email.com"
		user := "user@email.com"
		dummyResource := &domain.Resource{
			ID:           "1",
			URN:          "urn",
			Name:         "test-resource-name",
			ProviderType: "test-provider",
		}
		testCases := []struct {
			name                   string
			expectedApprovalAction domain.ApprovalAction
			expectedAppealDetails  *domain.Appeal
			expectedResult         *domain.Appeal
			expectedNotifications  []domain.Notification
			expectedGrant          *domain.Grant
		}{
			{
				name:                   "approve",
				expectedApprovalAction: validApprovalActionParam,
				expectedAppealDetails: &domain.Appeal{
					ID:         validApprovalActionParam.AppealID,
					AccountID:  "user@email.com",
					CreatedBy:  creator,
					ResourceID: "1",
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           "1",
						URN:          "urn",
						Name:         "test-resource-name",
						ProviderType: "test-provider",
					},
					Status: domain.AppealStatusPending,
					Approvals: []*domain.Approval{
						{
							Name:   "approval_0",
							Status: domain.ApprovalStatusApproved,
						},
						{
							Name:      "approval_1",
							Status:    domain.ApprovalStatusPending,
							Approvers: []string{"user@email.com"},
						},
					},
				},
				expectedResult: &domain.Appeal{
					ID:         validApprovalActionParam.AppealID,
					AccountID:  "user@email.com",
					CreatedBy:  creator,
					ResourceID: "1",
					Role:       "test-role",
					Resource:   dummyResource,
					Status:     domain.AppealStatusApproved,
					Approvals: []*domain.Approval{
						{
							Name:   "approval_0",
							Status: domain.ApprovalStatusApproved,
						},
						{
							Name:      "approval_1",
							Status:    domain.ApprovalStatusApproved,
							Approvers: []string{"user@email.com"},
							Actor:     &user,
							UpdatedAt: timeNow,
						},
					},
					Grant: &domain.Grant{
						Status:      domain.GrantStatusActive,
						AccountID:   "user@email.com",
						AccountType: domain.DefaultAppealAccountType,
						ResourceID:  "1",
						Resource:    dummyResource,
						Role:        "test-role",
						IsPermanent: true,
					},
				},
				expectedGrant: &domain.Grant{
					Status:      domain.GrantStatusActive,
					AccountID:   "user@email.com",
					AccountType: domain.DefaultAppealAccountType,
					ResourceID:  "1",
					Resource:    dummyResource,
					Role:        "test-role",
					IsPermanent: true,
				},
				expectedNotifications: []domain.Notification{
					{
						User: creator,
						Message: domain.NotificationMessage{
							Type: domain.NotificationTypeAppealApproved,
							Variables: map[string]interface{}{
								"resource_name": "test-resource-name (test-provider: urn)",
								"role":          "test-role",
							},
						},
					},
				},
			},
			{
				name: "reject",
				expectedApprovalAction: domain.ApprovalAction{
					AppealID:     "1",
					ApprovalName: "approval_1",
					Actor:        "user@email.com",
					Action:       domain.AppealActionNameReject,
					Reason:       "test-reason",
				},
				expectedAppealDetails: &domain.Appeal{
					ID:         validApprovalActionParam.AppealID,
					AccountID:  "user@email.com",
					CreatedBy:  creator,
					ResourceID: "1",
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           "1",
						URN:          "urn",
						Name:         "test-resource-name",
						ProviderType: "test-provider",
					},
					Status: domain.AppealStatusPending,
					Approvals: []*domain.Approval{
						{
							Name:   "approval_0",
							Status: domain.ApprovalStatusApproved,
						},
						{
							Name:      "approval_1",
							Status:    domain.ApprovalStatusPending,
							Approvers: []string{"user@email.com"},
						},
					},
				},
				expectedResult: &domain.Appeal{
					ID:         validApprovalActionParam.AppealID,
					AccountID:  "user@email.com",
					CreatedBy:  creator,
					ResourceID: "1",
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           "1",
						URN:          "urn",
						Name:         "test-resource-name",
						ProviderType: "test-provider",
					},
					Status: domain.AppealStatusRejected,
					Approvals: []*domain.Approval{
						{
							Name:   "approval_0",
							Status: domain.ApprovalStatusApproved,
						},
						{
							Name:      "approval_1",
							Status:    domain.ApprovalStatusRejected,
							Approvers: []string{"user@email.com"},
							Actor:     &user,
							Reason:    "test-reason",
							UpdatedAt: timeNow,
						},
					},
				},
				expectedNotifications: []domain.Notification{
					{
						User: creator,
						Message: domain.NotificationMessage{
							Type: domain.NotificationTypeAppealRejected,
							Variables: map[string]interface{}{
								"resource_name": "test-resource-name (test-provider: urn)",
								"role":          "test-role",
							},
						},
					},
				},
			},
			{
				name: "reject in the middle step",
				expectedApprovalAction: domain.ApprovalAction{
					AppealID:     "1",
					ApprovalName: "approval_1",
					Actor:        user,
					Action:       domain.AppealActionNameReject,
				},
				expectedAppealDetails: &domain.Appeal{
					ID:         validApprovalActionParam.AppealID,
					AccountID:  "user@email.com",
					CreatedBy:  creator,
					ResourceID: "1",
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           "1",
						URN:          "urn",
						Name:         "test-resource-name",
						ProviderType: "test-provider",
					},
					Status: domain.AppealStatusPending,
					Approvals: []*domain.Approval{
						{
							Name:   "approval_0",
							Status: domain.ApprovalStatusApproved,
						},
						{
							Name:      "approval_1",
							Status:    domain.ApprovalStatusPending,
							Approvers: []string{"user@email.com"},
						},
						{
							Name:   "approval_2",
							Status: domain.ApprovalStatusPending,
						},
					},
				},
				expectedResult: &domain.Appeal{
					ID:         validApprovalActionParam.AppealID,
					AccountID:  "user@email.com",
					CreatedBy:  creator,
					ResourceID: "1",
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           "1",
						URN:          "urn",
						Name:         "test-resource-name",
						ProviderType: "test-provider",
					},
					Status: domain.AppealStatusRejected,
					Approvals: []*domain.Approval{
						{
							Name:   "approval_0",
							Status: domain.ApprovalStatusApproved,
						},
						{
							Name:      "approval_1",
							Status:    domain.ApprovalStatusRejected,
							Approvers: []string{"user@email.com"},
							Actor:     &user,
							UpdatedAt: timeNow,
						},
						{
							Name:      "approval_2",
							Status:    domain.ApprovalStatusSkipped,
							UpdatedAt: timeNow,
						},
					},
				},
				expectedNotifications: []domain.Notification{
					{
						User: creator,
						Message: domain.NotificationMessage{
							Type: domain.NotificationTypeAppealRejected,
							Variables: map[string]interface{}{
								"resource_name": "test-resource-name (test-provider: urn)",
								"role":          "test-role",
							},
						},
					},
				},
			},
			{
				name: "should notify the next approvers if there's still manual approvals remaining ahead after approved",
				expectedApprovalAction: domain.ApprovalAction{
					AppealID:     validApprovalActionParam.AppealID,
					ApprovalName: "approval_0",
					Actor:        user,
					Action:       domain.AppealActionNameApprove,
				},
				expectedAppealDetails: &domain.Appeal{
					ID:         validApprovalActionParam.AppealID,
					AccountID:  "user@email.com",
					CreatedBy:  creator,
					ResourceID: "1",
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           "1",
						URN:          "urn",
						Name:         "test-resource-name",
						ProviderType: "test-provider",
					},
					Status: domain.AppealStatusPending,
					Approvals: []*domain.Approval{
						{
							Name:      "approval_0",
							Status:    domain.ApprovalStatusPending,
							Approvers: []string{user},
						},
						{
							Name:   "approval_1",
							Status: domain.ApprovalStatusPending,
							Approvers: []string{
								"nextapprover1@email.com",
								"nextapprover2@email.com",
							},
						},
					},
				},
				expectedResult: &domain.Appeal{
					ID:         validApprovalActionParam.AppealID,
					AccountID:  "user@email.com",
					CreatedBy:  creator,
					ResourceID: "1",
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           "1",
						URN:          "urn",
						Name:         "test-resource-name",
						ProviderType: "test-provider",
					},
					Status: domain.AppealStatusPending,
					Approvals: []*domain.Approval{
						{
							Name:      "approval_0",
							Status:    domain.ApprovalStatusApproved,
							Approvers: []string{user},
							Actor:     &user,
							UpdatedAt: timeNow,
						},
						{
							Name:   "approval_1",
							Status: domain.ApprovalStatusPending,
							Approvers: []string{
								"nextapprover1@email.com",
								"nextapprover2@email.com",
							},
						},
					},
				},
				expectedGrant: &domain.Grant{
					Status:      domain.GrantStatusActive,
					AccountID:   "user@email.com",
					AccountType: domain.DefaultAppealAccountType,
					ResourceID:  "1",
					Resource:    dummyResource,
					Role:        "test-role",
				},
				expectedNotifications: []domain.Notification{
					{
						User: "nextapprover1@email.com",
						Message: domain.NotificationMessage{
							Type: domain.NotificationTypeApproverNotification,
							Variables: map[string]interface{}{
								"resource_name": "test-resource-name (test-provider: urn)",
								"role":          "test-role",
								"requestor":     creator,
								"appeal_id":     validApprovalActionParam.AppealID,
							},
						},
					},
					{
						User: "nextapprover2@email.com",
						Message: domain.NotificationMessage{
							Type: domain.NotificationTypeApproverNotification,
							Variables: map[string]interface{}{
								"resource_name": "test-resource-name (test-provider: urn)",
								"role":          "test-role",
								"requestor":     creator,
								"appeal_id":     validApprovalActionParam.AppealID,
							},
						},
					},
				},
			},
		}
		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.mockAppealService.EXPECT().
					GetByID(mock.AnythingOfType("*context.emptyCtx"), validApprovalActionParam.AppealID).
					Return(tc.expectedAppealDetails, nil).Once()
				s.mockAppealService.EXPECT().GrantAccessToProvider(mock.Anything, mock.Anything).Return(nil).Once()
				if tc.expectedApprovalAction.Action == "approve" {
					s.mockGrantService.EXPECT().
						List(mock.Anything, domain.ListGrantsFilter{
							AccountIDs:  []string{tc.expectedAppealDetails.AccountID},
							ResourceIDs: []string{tc.expectedAppealDetails.ResourceID},
							Statuses:    []string{string(domain.GrantStatusActive)},
							Permissions: tc.expectedAppealDetails.Permissions,
						}).Return([]domain.Grant{}, nil).Once()
					s.mockGrantService.EXPECT().
						Prepare(mock.Anything, mock.Anything).Return(tc.expectedGrant, nil).Once()
					mockPolicy := &domain.Policy{
						Steps: []*domain.Step{
							{
								Name: "step-1",
							},
							{
								Name: "step-2",
							},
						},
					}
					s.mockPolicyService.EXPECT().
						GetOne(mock.Anything, tc.expectedAppealDetails.PolicyID, tc.expectedAppealDetails.PolicyVersion).
						Return(mockPolicy, nil).Once()

					tc.expectedResult.Policy = mockPolicy
					s.mockProviderService.EXPECT().GrantAccess(mock.Anything, *tc.expectedGrant).Return(nil).Once()
				}

				s.mockAppealService.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), tc.expectedResult).Return(nil).Once()
				s.mockNotifier.EXPECT().Notify(mock.Anything).Return(nil).Once()
				s.mockAuditLogger.EXPECT().Log(mock.Anything, mock.Anything, mock.Anything).
					Return(nil).Once()

				actualResult, actualError := s.service.UpdateApproval(context.Background(), tc.expectedApprovalAction)

				s.NoError(actualError)
				tc.expectedResult.Policy = actualResult.Policy
				s.Equal(tc.expectedResult, actualResult)
			})
		}
	})
}
