package appeal_test

import (
	"errors"
	"testing"
	"time"

	"github.com/odpf/guardian/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/provider"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository      *mocks.AppealRepository
	mockApprovalService *mocks.ApprovalService
	mockResourceService *mocks.ResourceService
	mockProviderService *mocks.ProviderService
	mockPolicyService   *mocks.PolicyService
	mockIAMService      *mocks.IAMService
	mockNotifier        *mocks.Notifier

	service *appeal.Service
	now     time.Time
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(mocks.AppealRepository)
	s.mockApprovalService = new(mocks.ApprovalService)
	s.mockResourceService = new(mocks.ResourceService)
	s.mockProviderService = new(mocks.ProviderService)
	s.mockPolicyService = new(mocks.PolicyService)
	s.mockIAMService = new(mocks.IAMService)
	s.mockNotifier = new(mocks.Notifier)
	s.now = time.Now()

	service := appeal.NewService(
		s.mockRepository,
		s.mockApprovalService,
		s.mockResourceService,
		s.mockProviderService,
		s.mockPolicyService,
		s.mockIAMService,
		s.mockNotifier,
		log.NewNoop(),
	)
	service.TimeNow = func() time.Time {
		return s.now
	}

	s.service = service
}

func (s *ServiceTestSuite) TestGetByID() {
	s.Run("should return error if id is empty/0", func() {
		expectedError := appeal.ErrAppealIDEmptyParam

		actualResult, actualError := s.service.GetByID(0)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got any from repository", func() {
		expectedError := errors.New("repository error")
		s.mockRepository.On("GetByID", mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.GetByID(1)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return record on success", func() {
		expectedID := uint(1)
		expectedResult := &domain.Appeal{
			ID: expectedID,
		}
		s.mockRepository.On("GetByID", expectedID).Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.GetByID(expectedID)

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
	})
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return error if got any from repository", func() {
		expectedError := errors.New("unexpected repository error")
		s.mockRepository.On("Find", mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find(map[string]interface{}{})

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return records on success", func() {
		expectedFilters := map[string]interface{}{
			"user": "user@email.com",
		}
		expectedResult := []*domain.Appeal{
			{
				ID:         1,
				ResourceID: 1,
				AccountID:  "user@email.com",
				Role:       "viewer",
			},
		}
		s.mockRepository.On("Find", expectedFilters).Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.Find(expectedFilters)

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
	})
}

func (s *ServiceTestSuite) TestCreate() {
	timeNow := time.Now()
	appeal.TimeNow = func() time.Time {
		return timeNow
	}
	s.Run("should return error if got error from resource service", func() {
		expectedError := errors.New("resource service error")
		s.mockResourceService.On("Find", mock.Anything).Return(nil, expectedError).Once()

		actualError := s.service.Create([]*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from provider service", func() {
		expectedResources := []*domain.Resource{}
		s.mockResourceService.On("Find", mock.Anything).Return(expectedResources, nil).Once()
		expectedError := errors.New("provider service error")
		s.mockProviderService.On("Find").Return(nil, expectedError).Once()

		actualError := s.service.Create([]*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from policy service", func() {
		expectedResources := []*domain.Resource{}
		expectedProviders := []*domain.Provider{}
		s.mockResourceService.On("Find", mock.Anything).Return(expectedResources, nil).Once()
		s.mockProviderService.On("Find").Return(expectedProviders, nil).Once()
		expectedError := errors.New("policy service error")
		s.mockPolicyService.On("Find").Return(nil, expectedError).Once()

		actualError := s.service.Create([]*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from appeal repository", func() {
		expectedResources := []*domain.Resource{}
		expectedProviders := []*domain.Provider{}
		expectedPolicies := []*domain.Policy{}
		s.mockResourceService.On("Find", mock.Anything).Return(expectedResources, nil).Once()
		s.mockProviderService.On("Find").Return(expectedProviders, nil).Once()
		s.mockPolicyService.On("Find").Return(expectedPolicies, nil).Once()
		expectedError := errors.New("appeal repository error")
		s.mockRepository.On("Find", mock.Anything).Return(nil, expectedError).Once()

		actualError := s.service.Create([]*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error for invalid appeals", func() {
		testProvider := &domain.Provider{
			ID:   1,
			Type: "provider_type",
			URN:  "provider_urn",
			Config: &domain.ProviderConfig{
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess: false,
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: "resource_type",
						Policy: &domain.PolicyConfig{
							ID:      "policy_id",
							Version: 1,
						},
						Roles: []*domain.Role{
							{
								ID: "role_1",
							},
						},
					},
				},
			},
		}
		timeNow := time.Now()
		expDate := timeNow.Add(24 * time.Hour)
		testCases := []struct {
			name                          string
			resources                     []*domain.Resource
			providers                     []*domain.Provider
			policies                      []*domain.Policy
			existingAppeals               []*domain.Appeal
			callMockValidateAppeal        bool
			expectedAppealValidationError error
			appeals                       []*domain.Appeal
			expectedError                 error
		}{
			{
				name: "duplicate appeal",
				existingAppeals: []*domain.Appeal{{
					AccountID:  "test-user",
					ResourceID: 1,
					Role:       "test-role",
					Status:     domain.AppealStatusPending,
				}},
				appeals: []*domain.Appeal{{
					AccountID:  "test-user",
					ResourceID: 1,
					Role:       "test-role",
				}},
				expectedError: appeal.ErrAppealDuplicate,
			},
			{
				name: "resource not found",
				resources: []*domain.Resource{{
					ID: 1,
				}},
				appeals: []*domain.Appeal{{
					AccountID:  "test-user",
					ResourceID: 2,
					Role:       "test-role",
				}},
				expectedError: appeal.ErrResourceNotFound,
			},
			{
				name: "provider type not found",
				resources: []*domain.Resource{{
					ID:           1,
					ProviderType: "invalid_provider_type",
					ProviderURN:  "provider_urn",
				}},
				providers:     []*domain.Provider{testProvider},
				appeals:       []*domain.Appeal{{ResourceID: 1}},
				expectedError: appeal.ErrProviderTypeNotFound,
			},
			{
				name: "user still have active access",
				resources: []*domain.Resource{{
					ID:           1,
					ProviderType: testProvider.Type,
					ProviderURN:  testProvider.URN,
				}},
				existingAppeals: []*domain.Appeal{{
					AccountID:  "test-user",
					ResourceID: 1,
					Role:       "test-role",
					Status:     domain.AppealStatusActive,
				}},
				appeals: []*domain.Appeal{{
					AccountID:  "test-user",
					ResourceID: 1,
					Role:       "test-role",
				}},
				providers:     []*domain.Provider{testProvider},
				expectedError: appeal.ErrAppealFoundActiveAccess,
			},
			{
				name: "invalid extension duration",
				resources: []*domain.Resource{{
					ID:           1,
					ProviderType: testProvider.Type,
					ProviderURN:  testProvider.URN,
				}},
				existingAppeals: []*domain.Appeal{{
					AccountID:  "test-user",
					ResourceID: 1,
					Role:       "test-role",
					Status:     domain.AppealStatusActive,
				}},
				appeals: []*domain.Appeal{{
					AccountID:  "test-user",
					ResourceID: 1,
					Role:       "test-role",
				}},
				providers: []*domain.Provider{{
					ID:   1,
					Type: testProvider.Type,
					URN:  testProvider.URN,
					Config: &domain.ProviderConfig{
						Appeal: &domain.AppealConfig{
							AllowActiveAccessExtensionIn: "invalid",
						},
					},
				}},
				expectedError: appeal.ErrAppealInvalidExtensionDuration,
			},
			{
				name: "extension not eligible",
				resources: []*domain.Resource{{
					ID:           1,
					ProviderType: testProvider.Type,
					ProviderURN:  testProvider.URN,
				}},
				existingAppeals: []*domain.Appeal{{
					AccountID:  "test-user",
					ResourceID: 1,
					Role:       "test-role",
					Status:     domain.AppealStatusActive,
					Options: &domain.AppealOptions{
						ExpirationDate: &expDate,
					},
				}},
				appeals: []*domain.Appeal{{
					AccountID:  "test-user",
					ResourceID: 1,
					Role:       "test-role",
				}},
				providers: []*domain.Provider{{
					ID:   1,
					Type: testProvider.Type,
					URN:  testProvider.URN,
					Config: &domain.ProviderConfig{
						Appeal: &domain.AppealConfig{
							AllowActiveAccessExtensionIn: "23h",
						},
					},
				}},
				expectedError: appeal.ErrAppealNotEligibleForExtension,
			},
			{
				name: "provider urn not found",
				resources: []*domain.Resource{{
					ID:           1,
					ProviderType: "provider_type",
					ProviderURN:  "invalid_provider_urn",
				}},
				providers:     []*domain.Provider{testProvider},
				appeals:       []*domain.Appeal{{ResourceID: 1}},
				expectedError: appeal.ErrProviderURNNotFound,
			},
			{
				name: "resource type not found",
				resources: []*domain.Resource{{
					ID:           1,
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "invalid_resource_type",
				}},
				providers:     []*domain.Provider{testProvider},
				appeals:       []*domain.Appeal{{ResourceID: 1}},
				expectedError: appeal.ErrResourceTypeNotFound,
			},
			{
				name: "duration not found when the appeal config prevents permanent access",
				resources: []*domain.Resource{{
					ID:           1,
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
				}},
				providers:                     []*domain.Provider{testProvider},
				callMockValidateAppeal:        true,
				expectedAppealValidationError: provider.ErrOptionsDurationNotFound,
				appeals: []*domain.Appeal{{
					ResourceID: 1,
				}},
				expectedError: appeal.ErrOptionsDurationNotFound,
			},
			{
				name: "empty duration option",
				resources: []*domain.Resource{{
					ID:           1,
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
				}},
				providers:                     []*domain.Provider{testProvider},
				callMockValidateAppeal:        true,
				expectedAppealValidationError: provider.ErrDurationIsRequired,
				appeals: []*domain.Appeal{{
					ResourceID: 1,
					Options: &domain.AppealOptions{
						Duration: "",
					},
				}},
				expectedError: appeal.ErrDurationIsRequired,
			},
			{
				name: "invalid role",
				resources: []*domain.Resource{{
					ID:           1,
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
				}},
				providers:                     []*domain.Provider{testProvider},
				callMockValidateAppeal:        true,
				expectedAppealValidationError: provider.ErrInvalidRole,
				appeals: []*domain.Appeal{{
					ResourceID: 1,
					Role:       "invalid_role",
					Options: &domain.AppealOptions{
						ExpirationDate: &timeNow,
					},
				}},
				expectedError: appeal.ErrInvalidRole,
			},
			{
				name: "policy id not found",
				resources: []*domain.Resource{{
					ID:           1,
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
				}},
				providers:              []*domain.Provider{testProvider},
				callMockValidateAppeal: true,
				appeals: []*domain.Appeal{{
					ResourceID: 1,
					Role:       "role_1",
					Options: &domain.AppealOptions{
						ExpirationDate: &timeNow,
					},
				}},
				expectedError: appeal.ErrPolicyIDNotFound,
			},
			{
				name: "policy version not found",
				resources: []*domain.Resource{{
					ID:           1,
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
				}},
				callMockValidateAppeal: true,
				providers:              []*domain.Provider{testProvider},
				policies: []*domain.Policy{{
					ID: "policy_id",
				}},
				appeals: []*domain.Appeal{{
					ResourceID: 1,
					Role:       "role_1",
					Options: &domain.AppealOptions{
						ExpirationDate: &timeNow,
					},
				}},
				expectedError: appeal.ErrPolicyVersionNotFound,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.mockResourceService.On("Find", mock.Anything).Return(tc.resources, nil).Once()
				s.mockProviderService.On("Find").Return(tc.providers, nil).Once()
				s.mockPolicyService.On("Find").Return(tc.policies, nil).Once()
				s.mockRepository.On("Find", mock.Anything).Return(tc.existingAppeals, nil).Once()
				if tc.callMockValidateAppeal {
					s.mockProviderService.On("ValidateAppeal", mock.Anything, mock.Anything).Return(tc.expectedAppealValidationError).Once()
				}

				actualError := s.service.Create(tc.appeals)

				s.Contains(actualError.Error(), tc.expectedError.Error())
			})
		}
	})

	s.Run("should return error if got error from repository on bulk upsert", func() {
		expectedResources := []*domain.Resource{}
		expectedProviders := []*domain.Provider{}
		expectedPolicies := []*domain.Policy{}
		expectedPendingAppeals := []*domain.Appeal{}
		s.mockResourceService.On("Find", mock.Anything).Return(expectedResources, nil).Once()
		s.mockProviderService.On("Find").Return(expectedProviders, nil).Once()
		s.mockPolicyService.On("Find").Return(expectedPolicies, nil).Once()
		s.mockRepository.On("Find", mock.Anything).Return(expectedPendingAppeals, nil).Once()
		expectedError := errors.New("repository error")
		s.mockRepository.On("BulkUpsert", mock.Anything).Return(expectedError).Once()

		actualError := s.service.Create([]*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	accountID := "test@email.com"
	resourceIDs := []uint{1, 2}
	resources := []*domain.Resource{}
	for _, id := range resourceIDs {
		resources = append(resources, &domain.Resource{
			ID:           id,
			Type:         "resource_type_1",
			ProviderType: "provider_type",
			ProviderURN:  "provider1",
			Details: map[string]interface{}{
				"owner": []string{"resource.owner@email.com"},
			},
		})
	}
	providers := []*domain.Provider{
		{
			ID:   1,
			Type: "provider_type",
			URN:  "provider1",
			Config: &domain.ProviderConfig{
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess:         true,
					AllowActiveAccessExtensionIn: "24h",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: "resource_type_1",
						Policy: &domain.PolicyConfig{
							ID:      "policy_1",
							Version: 1,
						},
						Roles: []*domain.Role{
							{
								ID: "role_id",
							},
						},
					},
				},
			},
		},
	}
	expDate := timeNow.Add(23 * time.Hour)
	currentActiveAppeal := &domain.Appeal{
		ID:         99,
		AccountID:  accountID,
		ResourceID: 2,
		Resource: &domain.Resource{
			ID:  2,
			URN: "urn",
		},
		Role:   "role_id",
		Status: domain.AppealStatusActive,
		Options: &domain.AppealOptions{
			ExpirationDate: &expDate,
		},
	}
	expectedExistingAppeals := []*domain.Appeal{currentActiveAppeal}
	policies := []*domain.Policy{
		{
			ID:      "policy_1",
			Version: 1,
			Steps: []*domain.Step{
				{
					Name:      "step_1",
					Approvers: "$resource.details.owner",
				},
				{
					Name:      "step_2",
					Approvers: "$user_approvers",
				},
			},
		},
	}
	expectedAppealsInsertionParam := []*domain.Appeal{}
	for i, r := range resourceIDs {
		expectedAppealsInsertionParam = append(expectedAppealsInsertionParam, &domain.Appeal{
			ResourceID:    r,
			Resource:      resources[i],
			PolicyID:      "policy_1",
			PolicyVersion: 1,
			Status:        domain.AppealStatusPending,
			AccountID:     accountID,
			AccountType:   domain.DefaultAppealAccountType,
			Role:          "role_id",
			Approvals: []*domain.Approval{
				{
					Name:          "step_1",
					Index:         0,
					Status:        domain.ApprovalStatusPending,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
					Approvers:     []string{"resource.owner@email.com"},
				},
				{
					Name:          "step_2",
					Index:         1,
					Status:        domain.ApprovalStatusBlocked,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
					Approvers:     []string{"user.approver@email.com"},
				},
			},
		})
	}
	insertionParamExpiredAppeal := &domain.Appeal{}
	*insertionParamExpiredAppeal = *currentActiveAppeal
	insertionParamExpiredAppeal.Status = domain.AppealStatusTerminated
	expectedAppealsInsertionParam = append(expectedAppealsInsertionParam, insertionParamExpiredAppeal)
	expectedResult := []*domain.Appeal{
		{
			ID:            1,
			ResourceID:    1,
			Resource:      resources[0],
			PolicyID:      "policy_1",
			PolicyVersion: 1,
			Status:        domain.AppealStatusPending,
			AccountID:     accountID,
			AccountType:   domain.DefaultAppealAccountType,
			Role:          "role_id",
			Approvals: []*domain.Approval{
				{
					ID:            1,
					Name:          "step_1",
					Index:         0,
					Status:        domain.ApprovalStatusPending,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
					Approvers:     []string{"resource.owner@email.com"},
				},
				{
					ID:            2,
					Name:          "step_2",
					Index:         1,
					Status:        domain.ApprovalStatusBlocked,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
					Approvers:     []string{"user.approver@email.com"},
				},
			},
		},
		{
			ID:            2,
			ResourceID:    2,
			Resource:      resources[1],
			PolicyID:      "policy_1",
			PolicyVersion: 1,
			Status:        domain.AppealStatusPending,
			AccountID:     accountID,
			AccountType:   domain.DefaultAppealAccountType,
			Role:          "role_id",
			Approvals: []*domain.Approval{
				{
					ID:            1,
					Name:          "step_1",
					Index:         0,
					Status:        domain.ApprovalStatusPending,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
					Approvers:     []string{"resource.owner@email.com"},
				},
				{
					ID:            2,
					Name:          "step_2",
					Index:         1,
					Status:        domain.ApprovalStatusBlocked,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
					Approvers:     []string{"user.approver@email.com"},
				},
			},
		},
	}

	s.Run("should return appeals on success", func() {
		expectedResourceFilters := map[string]interface{}{"ids": resourceIDs}
		s.mockResourceService.On("Find", expectedResourceFilters).Return(resources, nil).Once()
		s.mockProviderService.On("Find").Return(providers, nil).Once()
		s.mockPolicyService.On("Find").Return(policies, nil).Once()
		expectedExistingAppealsFilters := map[string]interface{}{
			"statuses": []string{
				domain.AppealStatusPending,
				domain.AppealStatusActive,
			},
		}
		s.mockRepository.On("Find", expectedExistingAppealsFilters).Return(expectedExistingAppeals, nil).Once()
		s.mockProviderService.On("ValidateAppeal", mock.Anything, mock.Anything).Return(nil)
		expectedUserApprovers := []string{"user.approver@email.com"}
		s.mockIAMService.On("GetUserApproverEmails", accountID).Return(expectedUserApprovers, nil)
		s.mockApprovalService.On("AdvanceApproval", mock.Anything).Return(nil)
		s.mockRepository.
			On("BulkUpsert", expectedAppealsInsertionParam).
			Return(nil).
			Run(func(args mock.Arguments) {
				appeals := args.Get(0).([]*domain.Appeal)
				for i, a := range appeals[0 : len(appeals)-1] {
					a.ID = expectedResult[i].ID
					for j, approval := range a.Approvals {
						approval.ID = expectedResult[i].Approvals[j].ID
					}
				}
			}).
			Once()
		s.mockNotifier.On("Notify", mock.Anything).Return(nil).Once()

		appeals := []*domain.Appeal{
			{
				AccountID:  accountID,
				ResourceID: 1,
				Resource: &domain.Resource{
					ID:  1,
					URN: "urn",
				},
				Role: "role_id",
			},
			{
				AccountID:  accountID,
				ResourceID: 2,
				Resource: &domain.Resource{
					ID:  2,
					URN: "urn",
				},
				Role: "role_id",
			},
		}
		actualError := s.service.Create(appeals)

		s.Equal(expectedResult, appeals)
		s.Nil(actualError)
	})
}

func (s *ServiceTestSuite) TestMakeAction() {
	timeNow := time.Now()
	appeal.TimeNow = func() time.Time {
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
				AppealID: 1,
				Actor:    "user@email.com",
				Action:   "name",
			},
			{
				AppealID:     1,
				ApprovalName: "approval_1",
				Actor:        "invalidemail",
				Action:       "name",
			},
			{
				AppealID:     1,
				ApprovalName: "approval_1",
				Action:       "name",
			},
			{
				AppealID:     1,
				ApprovalName: "approval_1",
				Actor:        "user@email.com",
			},
		}

		for _, param := range invalidApprovalActionParameters {
			actualResult, actualError := s.service.MakeAction(param)

			s.Nil(actualResult)
			s.Error(actualError)
		}
	})

	validApprovalActionParam := domain.ApprovalAction{
		AppealID:     1,
		ApprovalName: "approval_1",
		Actor:        "user@email.com",
		Action:       "approve",
	}

	s.Run("should return error if got any from repository while getting appeal details", func() {
		expectedError := errors.New("repository error")
		s.mockRepository.On("GetByID", mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.MakeAction(validApprovalActionParam)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if appeal not found", func() {
		s.mockRepository.On("GetByID", validApprovalActionParam.AppealID).Return(nil, appeal.ErrAppealNotFound).Once()

		actualResult, actualError := s.service.MakeAction(validApprovalActionParam)

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
				name:          "appeal not eligible, status: active",
				appealStatus:  domain.AppealStatusActive,
				expectedError: appeal.ErrAppealStatusApproved,
			},
			{
				name:          "appeal not eligible, status: rejected",
				appealStatus:  domain.AppealStatusRejected,
				expectedError: appeal.ErrAppealStatusRejected,
			},
			{
				name:          "appeal not eligible, status: terminated",
				appealStatus:  domain.AppealStatusTerminated,
				expectedError: appeal.ErrAppealStatusTerminated,
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
				expectedError: appeal.ErrApprovalNameNotFound,
			},
		}

		for _, tc := range testCases {
			expectedAppeal := &domain.Appeal{
				ID:        validApprovalActionParam.AppealID,
				Status:    tc.appealStatus,
				Approvals: tc.approvals,
			}
			s.mockRepository.On("GetByID", validApprovalActionParam.AppealID).Return(expectedAppeal, nil).Once()

			actualResult, actualError := s.service.MakeAction(validApprovalActionParam)

			s.Nil(actualResult)
			s.EqualError(actualError, tc.expectedError.Error())
		}
	})

	s.Run("should return error if got any from repository while updating appeal", func() {
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
			},
		}
		s.mockRepository.On("GetByID", validApprovalActionParam.AppealID).Return(expectedAppeal, nil).Once()
		expectedError := errors.New("repository error")
		s.mockApprovalService.On("AdvanceApproval", expectedAppeal).Return(nil).Once()
		s.mockProviderService.On("GrantAccess", expectedAppeal).Return(nil).Once()
		s.mockRepository.On("Update", mock.Anything).Return(expectedError).Once()
		s.mockProviderService.On("RevokeAccess", expectedAppeal).Return(nil).Once()

		actualResult, actualError := s.service.MakeAction(validApprovalActionParam)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return updated appeal on success", func() {
		user := "user@email.com"
		testCases := []struct {
			name                   string
			expectedApprovalAction domain.ApprovalAction
			expectedAppealDetails  *domain.Appeal
			expectedResult         *domain.Appeal
			expectedNotifications  []domain.Notification
		}{
			{
				name:                   "approve",
				expectedApprovalAction: validApprovalActionParam,
				expectedAppealDetails: &domain.Appeal{
					ID:         validApprovalActionParam.AppealID,
					AccountID:  "user@email.com",
					ResourceID: 1,
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           1,
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
					ResourceID: 1,
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           1,
						URN:          "urn",
						Name:         "test-resource-name",
						ProviderType: "test-provider",
					},
					Status: domain.AppealStatusActive,
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
				},
				expectedNotifications: []domain.Notification{
					{
						User: "user@email.com",
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
					AppealID:     1,
					ApprovalName: "approval_1",
					Actor:        "user@email.com",
					Action:       domain.AppealActionNameReject,
				},
				expectedAppealDetails: &domain.Appeal{
					ID:         validApprovalActionParam.AppealID,
					AccountID:  "user@email.com",
					ResourceID: 1,
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           1,
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
					ResourceID: 1,
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           1,
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
					},
				},
				expectedNotifications: []domain.Notification{
					{
						User: "user@email.com",
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
					AppealID:     1,
					ApprovalName: "approval_1",
					Actor:        user,
					Action:       domain.AppealActionNameReject,
				},
				expectedAppealDetails: &domain.Appeal{
					ID:         validApprovalActionParam.AppealID,
					AccountID:  "user@email.com",
					ResourceID: 1,
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           1,
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
					ResourceID: 1,
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           1,
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
						User: "user@email.com",
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
					ResourceID: 1,
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           1,
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
					ResourceID: 1,
					Role:       "test-role",
					Resource: &domain.Resource{
						ID:           1,
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
				expectedNotifications: []domain.Notification{
					{
						User: "nextapprover1@email.com",
						Message: domain.NotificationMessage{
							Type: domain.NotificationTypeApproverNotification,
							Variables: map[string]interface{}{
								"resource_name": "test-resource-name (test-provider: urn)",
								"role":          "test-role",
								"requestor":     "user@email.com",
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
								"requestor":     "user@email.com",
								"appeal_id":     validApprovalActionParam.AppealID,
							},
						},
					},
				},
			},
		}
		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.mockRepository.On("GetByID", validApprovalActionParam.AppealID).
					Return(tc.expectedAppealDetails, nil).
					Once()
				s.mockApprovalService.On("AdvanceApproval", tc.expectedAppealDetails).
					Return(nil).Once()
				s.mockProviderService.On("GrantAccess", tc.expectedAppealDetails).
					Return(nil).
					Once()
				s.mockRepository.On("Update", mock.Anything).
					Return(nil).
					Once()
				s.mockNotifier.On("Notify", tc.expectedNotifications).Return(nil).Once()

				actualResult, actualError := s.service.MakeAction(tc.expectedApprovalAction)

				s.Equal(tc.expectedResult, actualResult)
				s.Nil(actualError)
			})
		}
	})
}

// func (s *ServiceTestSuite) TestCancel() {
// 	s.Run("should return error from")
// }

func (s *ServiceTestSuite) TestRevoke() {
	s.Run("should return error if got any while getting appeal details", func() {
		expectedError := errors.New("repository error")
		s.mockRepository.On("GetByID", mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Revoke(0, "", "")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if appeal not found", func() {
		s.mockRepository.On("GetByID", mock.Anything).Return(nil, appeal.ErrAppealNotFound).Once()
		expectedError := appeal.ErrAppealNotFound

		actualResult, actualError := s.service.Revoke(0, "", "")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	appealID := uint(1)
	actor := "user@email.com"
	reason := "test-reason"

	appealDetails := &domain.Appeal{
		ID:         appealID,
		ResourceID: 1,
		Resource: &domain.Resource{
			ID:  1,
			URN: "urn",
		},
	}

	s.Run("should return error if got any while updating appeal", func() {
		s.mockRepository.On("GetByID", appealID).Return(appealDetails, nil).Once()
		expectedError := errors.New("repository error")
		s.mockRepository.On("Update", mock.Anything).Return(expectedError).Once()

		actualResult, actualError := s.service.Revoke(appealID, actor, reason)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error and rollback updated appeal if failed granting the access to the provider", func() {
		s.mockRepository.On("GetByID", appealID).Return(appealDetails, nil).Once()
		s.mockRepository.On("Update", mock.Anything).Return(nil).Once()
		expectedError := errors.New("provider service error")
		s.mockProviderService.On("RevokeAccess", mock.Anything).Return(expectedError).Once()
		s.mockRepository.On("Update", appealDetails).Return(nil).Once()

		actualResult, actualError := s.service.Revoke(appealID, actor, reason)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return appeal and nil error on success", func() {
		s.mockRepository.On("GetByID", appealID).Return(appealDetails, nil).Once()
		expectedAppeal := &domain.Appeal{}
		*expectedAppeal = *appealDetails
		expectedAppeal.Status = domain.AppealStatusTerminated
		expectedAppeal.RevokedAt = s.now
		expectedAppeal.RevokedBy = actor
		expectedAppeal.RevokeReason = reason
		s.mockRepository.On("Update", expectedAppeal).Return(nil).Once()
		s.mockProviderService.On("RevokeAccess", appealDetails).Return(nil).Once()
		s.mockNotifier.On("Notify", mock.Anything).Return(nil).Once()

		actualResult, actualError := s.service.Revoke(appealID, actor, reason)

		s.Equal(expectedAppeal, actualResult)
		s.Nil(actualError)
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
