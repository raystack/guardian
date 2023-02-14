package appeal_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/odpf/guardian/core/appeal"
	appealmocks "github.com/odpf/guardian/core/appeal/mocks"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository      *appealmocks.Repository
	mockApprovalService *appealmocks.ApprovalService
	mockResourceService *appealmocks.ResourceService
	mockProviderService *appealmocks.ProviderService
	mockPolicyService   *appealmocks.PolicyService
	mockGrantService    *appealmocks.GrantService
	mockIAMManager      *appealmocks.IamManager
	mockIAMClient       *mocks.IAMClient
	mockNotifier        *appealmocks.Notifier
	mockAuditLogger     *appealmocks.AuditLogger

	service *appeal.Service
	now     time.Time
}

func (s *ServiceTestSuite) setup() {
	s.mockRepository = new(appealmocks.Repository)
	s.mockApprovalService = new(appealmocks.ApprovalService)
	s.mockResourceService = new(appealmocks.ResourceService)
	s.mockProviderService = new(appealmocks.ProviderService)
	s.mockPolicyService = new(appealmocks.PolicyService)
	s.mockGrantService = new(appealmocks.GrantService)
	s.mockIAMManager = new(appealmocks.IamManager)
	s.mockIAMClient = new(mocks.IAMClient)
	s.mockNotifier = new(appealmocks.Notifier)
	s.mockAuditLogger = new(appealmocks.AuditLogger)
	s.now = time.Now()

	service := appeal.NewService(appeal.ServiceDeps{
		s.mockRepository,
		s.mockApprovalService,
		s.mockResourceService,
		s.mockProviderService,
		s.mockPolicyService,
		s.mockGrantService,
		s.mockIAMManager,
		s.mockNotifier,
		validator.New(),
		log.NewNoop(),
		s.mockAuditLogger,
	})
	service.TimeNow = func() time.Time {
		return s.now
	}

	s.service = service
}

func (s *ServiceTestSuite) SetupTest() {
	s.setup()
}

func (s *ServiceTestSuite) TestGetByID() {
	s.Run("should return error if id is empty/0", func() {
		expectedError := appeal.ErrAppealIDEmptyParam

		actualResult, actualError := s.service.GetByID(context.Background(), "")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got any from repository", func() {
		expectedError := errors.New("repository error")
		s.mockRepository.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.GetByID(context.Background(), "1")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return record on success", func() {
		expectedID := "1"
		expectedResult := &domain.Appeal{
			ID: expectedID,
		}
		s.mockRepository.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), expectedID).Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.GetByID(context.Background(), expectedID)

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
	})
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return error if got any from repository", func() {
		expectedError := errors.New("unexpected repository error")
		s.mockRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find(context.Background(), &domain.ListAppealsFilter{})

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return records on success", func() {
		expectedFilters := &domain.ListAppealsFilter{
			AccountID: "user@email.com",
		}
		expectedResult := []*domain.Appeal{
			{
				ID:         "1",
				ResourceID: "1",
				AccountID:  "user@email.com",
				Role:       "viewer",
			},
		}
		s.mockRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx"), expectedFilters).Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.Find(context.Background(), expectedFilters)

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
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return(nil, expectedError).Once()

		actualError := s.service.Create(context.Background(), []*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from provider service", func() {
		expectedResources := []*domain.Resource{}
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return(expectedResources, nil).Once()
		expectedError := errors.New("provider service error")
		s.mockProviderService.On("Find", mock.Anything).Return(nil, expectedError).Once()

		actualError := s.service.Create(context.Background(), []*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from policy service", func() {
		expectedResources := []*domain.Resource{}
		expectedProviders := []*domain.Provider{}
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return(expectedResources, nil).Once()
		s.mockProviderService.On("Find", mock.Anything).Return(expectedProviders, nil).Once()
		expectedError := errors.New("policy service error")
		s.mockPolicyService.On("Find", mock.Anything).Return(nil, expectedError).Once()

		actualError := s.service.Create(context.Background(), []*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from appeal repository", func() {
		expectedResources := []*domain.Resource{}
		expectedProviders := []*domain.Provider{}
		expectedPolicies := []*domain.Policy{}
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return(expectedResources, nil).Once()
		s.mockProviderService.On("Find", mock.Anything).Return(expectedProviders, nil).Once()
		s.mockPolicyService.On("Find", mock.Anything).Return(expectedPolicies, nil).Once()
		expectedError := errors.New("appeal repository error")
		s.mockRepository.EXPECT().
			Find(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(nil, expectedError).Once()

		actualError := s.service.Create(context.Background(), []*domain.Appeal{})

		s.ErrorIs(actualError, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error for invalid appeals", func() {
		testProvider := &domain.Provider{
			ID:   "1",
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

		testPolicies := []*domain.Policy{{ID: "policy_id", Version: 1}}

		testCases := []struct {
			name                          string
			resources                     []*domain.Resource
			providers                     []*domain.Provider
			policies                      []*domain.Policy
			existingAppeals               []*domain.Appeal
			activeGrants                  []domain.Grant
			callMockValidateAppeal        bool
			expectedAppealValidationError error
			callMockGetPermissions        bool
			appeals                       []*domain.Appeal
			expectedError                 error
		}{
			{
				name: "creating appeal for other normal user with allow_on_behalf=false",
				appeals: []*domain.Appeal{{
					CreatedBy:  "addOnBehalfApprovedNotification-user",
					AccountID:  "addOnBehalfApprovedNotification-user-2",
					ResourceID: "1",
					Role:       "addOnBehalfApprovedNotification-role",
				}},
				resources: []*domain.Resource{{
					ID:           "1",
					ProviderType: testProvider.Type,
					ProviderURN:  testProvider.URN,
					Type:         "resource_type",
				}},
				providers:              []*domain.Provider{testProvider},
				policies:               []*domain.Policy{{ID: "policy_id", Version: 1, AppealConfig: &domain.PolicyAppealConfig{AllowOnBehalf: false}}},
				callMockValidateAppeal: true,
				callMockGetPermissions: true,
				expectedError:          appeal.ErrCannotCreateAppealForOtherUser,
			},
			{
				name: "duplicate appeal",
				existingAppeals: []*domain.Appeal{{
					CreatedBy:  "test-user",
					AccountID:  "test-user",
					ResourceID: "1",
					Role:       "test-role",
					Status:     domain.AppealStatusPending,
				}},
				appeals: []*domain.Appeal{{
					CreatedBy:  "test-user",
					AccountID:  "test-user",
					ResourceID: "1",
					Role:       "test-role",
				}},
				expectedError: appeal.ErrAppealDuplicate,
			},
			{
				name: "resource not found",
				resources: []*domain.Resource{{
					ID: "1",
				}},
				appeals: []*domain.Appeal{{
					CreatedBy:  "test-user",
					AccountID:  "test-user",
					ResourceID: "2",
					Role:       "test-role",
				}},
				expectedError: appeal.ErrResourceNotFound,
			},
			{
				name: "provider type not found",
				resources: []*domain.Resource{{
					ID:           "1",
					ProviderType: "invalid_provider_type",
					ProviderURN:  "provider_urn",
				}},
				providers:     []*domain.Provider{testProvider},
				appeals:       []*domain.Appeal{{ResourceID: "1"}},
				expectedError: appeal.ErrProviderTypeNotFound,
			},
			{
				name: "user still have active grant",
				resources: []*domain.Resource{{
					ID:           "1",
					Type:         "resource_type",
					ProviderType: testProvider.Type,
					ProviderURN:  testProvider.URN,
				}},
				activeGrants: []domain.Grant{{
					AccountID:  "test-user",
					ResourceID: "1",
					Role:       "test-role",
					Status:     domain.GrantStatusActive,
				}},
				policies: testPolicies,
				appeals: []*domain.Appeal{{
					CreatedBy:  "test-user",
					AccountID:  "test-user",
					ResourceID: "1",
					Role:       "test-role",
				}},
				providers:     []*domain.Provider{testProvider},
				expectedError: appeal.ErrAppealFoundActiveGrant,
			},
			{
				name: "invalid extension duration",
				resources: []*domain.Resource{{
					ID:           "1",
					Type:         "resource_type",
					ProviderType: testProvider.Type,
					ProviderURN:  testProvider.URN,
				}},
				activeGrants: []domain.Grant{{
					AccountID:  "test-user",
					ResourceID: "1",
					Role:       "test-role",
					Status:     domain.GrantStatusActive,
				}},
				appeals: []*domain.Appeal{{
					CreatedBy:  "test-user",
					AccountID:  "test-user",
					ResourceID: "1",
					Role:       "test-role",
				}},
				policies: testPolicies,
				providers: []*domain.Provider{{
					ID:   "1",
					Type: testProvider.Type,
					URN:  testProvider.URN,
					Config: &domain.ProviderConfig{
						Appeal: &domain.AppealConfig{
							AllowActiveAccessExtensionIn: "invalid",
						},
						Resources: testProvider.Config.Resources,
					},
				}},
				expectedError: appeal.ErrAppealInvalidExtensionDuration,
			},
			{
				name: "extension not eligible",
				resources: []*domain.Resource{{
					ID:           "1",
					Type:         "resource_type",
					ProviderType: testProvider.Type,
					ProviderURN:  testProvider.URN,
				}},
				activeGrants: []domain.Grant{{
					AccountID:      "test-user",
					ResourceID:     "1",
					Role:           "test-role",
					Status:         domain.GrantStatusActive,
					ExpirationDate: &expDate,
				}},
				appeals: []*domain.Appeal{{
					CreatedBy:  "test-user",
					AccountID:  "test-user",
					ResourceID: "1",
					Role:       "test-role",
				}},
				policies: testPolicies,
				providers: []*domain.Provider{{
					ID:   "1",
					Type: testProvider.Type,
					URN:  testProvider.URN,
					Config: &domain.ProviderConfig{
						Appeal: &domain.AppealConfig{
							AllowActiveAccessExtensionIn: "23h",
						},
						Resources: testProvider.Config.Resources,
					},
				}},
				expectedError: appeal.ErrGrantNotEligibleForExtension,
			},
			{
				name: "provider urn not found",
				resources: []*domain.Resource{{
					ID:           "1",
					ProviderType: "provider_type",
					ProviderURN:  "invalid_provider_urn",
				}},
				providers:     []*domain.Provider{testProvider},
				appeals:       []*domain.Appeal{{ResourceID: "1"}},
				expectedError: appeal.ErrProviderURNNotFound,
			},
			{
				name: "duration not found when the appeal config prevents permanent access",
				resources: []*domain.Resource{{
					ID:           "1",
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
				}},
				policies:                      []*domain.Policy{{ID: "policy_id", Version: 1}},
				providers:                     []*domain.Provider{testProvider},
				callMockValidateAppeal:        true,
				expectedAppealValidationError: provider.ErrOptionsDurationNotFound,
				appeals: []*domain.Appeal{{
					ResourceID: "1",
				}},
				expectedError: appeal.ErrOptionsDurationNotFound,
			},
			{
				name: "empty duration option",
				resources: []*domain.Resource{{
					ID:           "1",
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
				}},
				policies:                      testPolicies,
				providers:                     []*domain.Provider{testProvider},
				callMockValidateAppeal:        true,
				expectedAppealValidationError: provider.ErrDurationIsRequired,
				appeals: []*domain.Appeal{{
					ResourceID: "1",
					Options: &domain.AppealOptions{
						Duration: "",
					},
				}},
				expectedError: appeal.ErrDurationIsRequired,
			},
			{
				name: "invalid role",
				resources: []*domain.Resource{{
					ID:           "1",
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
				}},
				policies:                      testPolicies,
				providers:                     []*domain.Provider{testProvider},
				callMockValidateAppeal:        true,
				expectedAppealValidationError: provider.ErrInvalidRole,
				appeals: []*domain.Appeal{{
					ResourceID: "1",
					Role:       "invalid_role",
					Options: &domain.AppealOptions{
						ExpirationDate: &timeNow,
					},
				}},
				expectedError: appeal.ErrInvalidRole,
			},
			{
				name: "resource type not found",
				resources: []*domain.Resource{{
					ID:           "1",
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "invalid_resource_type",
				}},
				policies:      testPolicies,
				providers:     []*domain.Provider{testProvider},
				appeals:       []*domain.Appeal{{ResourceID: "1"}},
				expectedError: appeal.ErrResourceTypeNotFound,
			},
			{
				name: "policy id not found",
				resources: []*domain.Resource{{
					ID:           "1",
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
				}},
				providers: []*domain.Provider{testProvider},
				appeals: []*domain.Appeal{{
					ResourceID: "1",
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
					ID:           "1",
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
				}},
				providers: []*domain.Provider{testProvider},
				policies: []*domain.Policy{{
					ID: "policy_id",
				}},
				appeals: []*domain.Appeal{{
					ResourceID: "1",
					Role:       "role_1",
					Options: &domain.AppealOptions{
						ExpirationDate: &timeNow,
					},
				}},
				expectedError: appeal.ErrPolicyVersionNotFound,
			},
			{
				name: "appeal duration not found in policy appeal config",
				resources: []*domain.Resource{{
					ID:           "1",
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
				}},
				providers: []*domain.Provider{testProvider},
				policies: []*domain.Policy{{
					ID:      "policy_id",
					Version: uint(1),
					AppealConfig: &domain.PolicyAppealConfig{
						DurationOptions: []domain.AppealDurationOption{
							{Name: "1 Day", Value: "24h"},
							{Name: "3 Days", Value: "72h"},
							{Name: "90 Days", Value: "2160h"},
						},
					},
				}},
				callMockValidateAppeal: true,
				callMockGetPermissions: true,
				appeals: []*domain.Appeal{{
					ResourceID:    "1",
					Role:          "role_1",
					PolicyID:      "policy_id",
					PolicyVersion: uint(1),
					Options: &domain.AppealOptions{
						Duration: "100h",
					},
				}},
				expectedError: appeal.ErrOptionsDurationNotFound,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return(tc.resources, nil).Once()
				s.mockProviderService.On("Find", mock.Anything).Return(tc.providers, nil).Once()
				s.mockPolicyService.On("Find", mock.Anything).Return(tc.policies, nil).Once()
				s.mockRepository.EXPECT().
					Find(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
					Return(tc.existingAppeals, nil).Once()
				s.mockGrantService.EXPECT().
					List(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("domain.ListGrantsFilter")).
					Return(tc.activeGrants, nil).Once()
				if tc.callMockValidateAppeal {
					s.mockProviderService.On("ValidateAppeal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.expectedAppealValidationError).Once()
				}
				if tc.callMockGetPermissions {
					s.mockProviderService.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return([]interface{}{}, nil).Once()
				}

				actualError := s.service.Create(context.Background(), tc.appeals)

				s.Contains(actualError.Error(), tc.expectedError.Error())
				s.mockProviderService.AssertExpectations(s.T())
				s.mockRepository.AssertExpectations(s.T())
			})
		}
	})

	s.Run("should return error if got error from repository on bulk upsert", func() {
		expectedResources := []*domain.Resource{}
		expectedProviders := []*domain.Provider{}
		expectedPolicies := []*domain.Policy{}
		expectedPendingAppeals := []*domain.Appeal{}
		expectedActiveGrants := []domain.Grant{}
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return(expectedResources, nil).Once()
		s.mockProviderService.On("Find", mock.Anything).Return(expectedProviders, nil).Once()
		s.mockPolicyService.On("Find", mock.Anything).Return(expectedPolicies, nil).Once()
		s.mockRepository.EXPECT().
			Find(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedPendingAppeals, nil).Once()
		s.mockGrantService.EXPECT().
			List(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("domain.ListGrantsFilter")).
			Return(expectedActiveGrants, nil).Once()
		expectedError := errors.New("repository error")
		s.mockRepository.EXPECT().
			BulkUpsert(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedError).Once()

		actualError := s.service.Create(context.Background(), []*domain.Appeal{})

		s.ErrorIs(actualError, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return appeals on success", func() {
		accountID := "test@email.com"
		resourceIDs := []string{"1", "2"}
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
				ID:   "1",
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
									ID:          "role_id",
									Permissions: []interface{}{"test-permission-1"},
								},
							},
						},
					},
				},
			},
		}
		expDate := timeNow.Add(23 * time.Hour)
		expectedExistingAppeals := []*domain.Appeal{}
		expectedActiveGrants := []domain.Grant{
			{
				ID:         "99",
				AccountID:  accountID,
				ResourceID: "2",
				Resource: &domain.Resource{
					ID:  "2",
					URN: "urn",
				},
				Role:           "role_id",
				Status:         domain.GrantStatusActive,
				ExpirationDate: &expDate,
			},
		}
		policies := []*domain.Policy{
			{
				ID:      "policy_1",
				Version: 1,
				Steps: []*domain.Step{
					{
						Name:     "step_1",
						Strategy: "manual",
						Approvers: []string{
							"$appeal.resource.details.owner",
						},
					},
					{
						Name:     "step_2",
						Strategy: "manual",
						Approvers: []string{
							"$appeal.creator.managers",
							"$appeal.creator.managers", // test duplicate approvers
						},
					},
				},
				IAM: &domain.IAMConfig{
					Provider: "http",
					Config: map[string]interface{}{
						"url": "http://localhost",
					},
					Schema: map[string]string{
						"managers": `managers`,
						"name":     "name",
						"role":     `$response.roles[0].name`,
						"roles":    `map($response.roles, {#.name})`,
					},
				},
				AppealConfig: &domain.PolicyAppealConfig{AllowOnBehalf: true},
			},
		}
		expectedCreatorUser := map[string]interface{}{
			"managers": []interface{}{"user.approver@email.com"},
			"name":     "test-name",
			"role":     "test-role-1",
			"roles":    []interface{}{"test-role-1", "test-role-2"},
		}
		expectedAppealsInsertionParam := []*domain.Appeal{}
		for i, r := range resourceIDs {
			appeal := &domain.Appeal{
				ResourceID:    r,
				Resource:      resources[i],
				PolicyID:      "policy_1",
				PolicyVersion: 1,
				Status:        domain.AppealStatusPending,
				AccountID:     accountID,
				AccountType:   domain.DefaultAppealAccountType,
				CreatedBy:     accountID,
				Creator:       expectedCreatorUser,
				Role:          "role_id",
				Permissions:   []string{"test-permission-1"},
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
				Description: "The answer is 42",
			}
			if r == "2" {
				appeal.AccountID = "addOnBehalfApprovedNotification-user"
			}
			expectedAppealsInsertionParam = append(expectedAppealsInsertionParam, appeal)
		}
		expectedResult := []*domain.Appeal{
			{
				ID:            "1",
				ResourceID:    "1",
				Resource:      resources[0],
				PolicyID:      "policy_1",
				PolicyVersion: 1,
				Status:        domain.AppealStatusPending,
				AccountID:     accountID,
				AccountType:   domain.DefaultAppealAccountType,
				CreatedBy:     accountID,
				Creator:       expectedCreatorUser,
				Role:          "role_id",
				Permissions:   []string{"test-permission-1"},
				Approvals: []*domain.Approval{
					{
						ID:            "1",
						Name:          "step_1",
						Index:         0,
						Status:        domain.ApprovalStatusPending,
						PolicyID:      "policy_1",
						PolicyVersion: 1,
						Approvers:     []string{"resource.owner@email.com"},
					},
					{
						ID:            "2",
						Name:          "step_2",
						Index:         1,
						Status:        domain.ApprovalStatusBlocked,
						PolicyID:      "policy_1",
						PolicyVersion: 1,
						Approvers:     []string{"user.approver@email.com"},
					},
				},
				Description: "The answer is 42",
			},
			{
				ID:            "2",
				ResourceID:    "2",
				Resource:      resources[1],
				PolicyID:      "policy_1",
				PolicyVersion: 1,
				Status:        domain.AppealStatusPending,
				AccountID:     "addOnBehalfApprovedNotification-user",
				AccountType:   domain.DefaultAppealAccountType,
				CreatedBy:     accountID,
				Creator:       expectedCreatorUser,
				Role:          "role_id",
				Permissions:   []string{"test-permission-1"},
				Approvals: []*domain.Approval{
					{
						ID:            "1",
						Name:          "step_1",
						Index:         0,
						Status:        domain.ApprovalStatusPending,
						PolicyID:      "policy_1",
						PolicyVersion: 1,
						Approvers:     []string{"resource.owner@email.com"},
					},
					{
						ID:            "2",
						Name:          "step_2",
						Index:         1,
						Status:        domain.ApprovalStatusBlocked,
						PolicyID:      "policy_1",
						PolicyVersion: 1,
						Approvers:     []string{"user.approver@email.com"},
					},
				},
				Description: "The answer is 42",
			},
		}

		expectedResourceFilters := domain.ListResourcesFilter{IDs: resourceIDs}
		s.mockResourceService.On("Find", mock.Anything, expectedResourceFilters).Return(resources, nil).Once()
		s.mockProviderService.On("Find", mock.Anything).Return(providers, nil).Once()
		s.mockPolicyService.On("Find", mock.Anything).Return(policies, nil).Once()
		expectedExistingAppealsFilters := &domain.ListAppealsFilter{
			Statuses: []string{domain.AppealStatusPending},
		}
		s.mockRepository.EXPECT().
			Find(mock.AnythingOfType("*context.emptyCtx"), expectedExistingAppealsFilters).
			Return(expectedExistingAppeals, nil).Once()
		s.mockGrantService.EXPECT().
			List(mock.AnythingOfType("*context.emptyCtx"), domain.ListGrantsFilter{
				Statuses: []string{string(domain.GrantStatusActive)},
			}).
			Return(expectedActiveGrants, nil).Once()
		s.mockProviderService.On("ValidateAppeal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		s.mockProviderService.On("GetPermissions", mock.Anything, mock.Anything, "resource_type_1", "role_id").
			Return([]interface{}{"test-permission-1"}, nil)
		s.mockIAMManager.On("ParseConfig", mock.Anything, mock.Anything).Return(nil, nil)
		s.mockIAMManager.On("GetClient", mock.Anything, mock.Anything).Return(s.mockIAMClient, nil)
		expectedCreatorResponse := map[string]interface{}{
			"managers": []interface{}{"user.approver@email.com"},
			"name":     "test-name",
			"roles": []map[string]interface{}{
				{"name": "test-role-1"},
				{"name": "test-role-2"},
			},
		}
		s.mockIAMClient.On("GetUser", accountID).Return(expectedCreatorResponse, nil)
		s.mockRepository.EXPECT().
			BulkUpsert(mock.AnythingOfType("*context.emptyCtx"), expectedAppealsInsertionParam).
			Return(nil).
			Run(func(_a0 context.Context, appeals []*domain.Appeal) {
				for i, a := range appeals {
					a.ID = expectedResult[i].ID
					for j, approval := range a.Approvals {
						approval.ID = expectedResult[i].Approvals[j].ID
					}
				}
			}).
			Once()
		s.mockNotifier.On("Notify", mock.Anything).Return(nil).Once()
		s.mockAuditLogger.On("Log", mock.Anything, appeal.AuditKeyBulkInsert, mock.Anything).Return(nil).Once()

		appeals := []*domain.Appeal{
			{
				CreatedBy:  accountID,
				AccountID:  accountID,
				ResourceID: "1",
				Resource: &domain.Resource{
					ID:  "1",
					URN: "urn",
				},
				Role:        "role_id",
				Description: "The answer is 42",
			},
			{
				CreatedBy:  accountID,
				AccountID:  "addOnBehalfApprovedNotification-user",
				ResourceID: "2",
				Resource: &domain.Resource{
					ID:  "2",
					URN: "urn",
				},
				Role:        "role_id",
				Description: "The answer is 42",
			},
		}
		actualError := s.service.Create(context.Background(), appeals)

		s.Nil(actualError)
		s.Equal(expectedResult, appeals)
		s.mockProviderService.AssertExpectations(s.T())
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("additional appeal creation", func() {
		s.Run("should use the overridding policy", func() {
			input := &domain.Appeal{
				ResourceID:    uuid.New().String(),
				AccountID:     "user@example.com",
				AccountType:   domain.DefaultAppealAccountType,
				CreatedBy:     "user@example.com",
				Role:          "test-role",
				PolicyID:      "test-policy",
				PolicyVersion: 99,
			}
			dummyResource := &domain.Resource{
				ID:           input.ResourceID,
				ProviderType: "test-provider-type",
				ProviderURN:  "test-provider-urn",
				Type:         "test-type",
				URN:          "test-urn",
			}
			expectedPermissions := []string{
				"test-permission-1",
				"test-permission-2",
			}
			dummyProvider := &domain.Provider{
				Type: dummyResource.ProviderType,
				URN:  dummyResource.ProviderURN,
				Config: &domain.ProviderConfig{
					Type: dummyResource.ProviderType,
					URN:  dummyResource.ProviderURN,
					Resources: []*domain.ResourceConfig{
						{
							Type: dummyResource.Type,
							Policy: &domain.PolicyConfig{
								ID:      "test-dummy-policy",
								Version: 1,
							},
							Roles: []*domain.Role{
								{
									ID: input.Role,
									Permissions: []interface{}{
										expectedPermissions[0],
										expectedPermissions[1],
									},
								},
							},
						},
					},
				},
			}
			dummyPolicy := &domain.Policy{
				ID:      "test-dummy-policy",
				Version: 1,
			}
			overriddingPolicy := &domain.Policy{
				ID:      input.PolicyID,
				Version: input.PolicyVersion,
				Steps: []*domain.Step{
					{
						Name:      "test-approval",
						Strategy:  "auto",
						ApproveIf: "true",
					},
				},
			}

			s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return([]*domain.Resource{dummyResource}, nil).Once()
			s.mockProviderService.On("Find", mock.Anything).Return([]*domain.Provider{dummyProvider}, nil).Once()
			s.mockPolicyService.On("Find", mock.Anything).Return([]*domain.Policy{dummyPolicy, overriddingPolicy}, nil).Once()
			s.mockRepository.EXPECT().
				Find(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
				Return([]*domain.Appeal{}, nil).Once()
			s.mockGrantService.EXPECT().
				List(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("domain.ListGrantsFilter")).
				Return([]domain.Grant{}, nil).Once()
			s.mockProviderService.On("ValidateAppeal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			s.mockProviderService.On("GetPermissions", mock.Anything, dummyProvider.Config, dummyResource.Type, input.Role).
				Return(dummyProvider.Config.Resources[0].Roles[0].Permissions, nil)
			s.mockIAMManager.On("ParseConfig", mock.Anything, mock.Anything).Return(nil, nil)
			s.mockIAMManager.On("GetClient", mock.Anything, mock.Anything).Return(s.mockIAMClient, nil)
			s.mockIAMClient.On("GetUser", input.AccountID).Return(map[string]interface{}{}, nil)
			s.mockRepository.EXPECT().
				BulkUpsert(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
				Return(nil).Once()
			s.mockNotifier.On("Notify", mock.Anything).Return(nil).Once()
			s.mockAuditLogger.On("Log", mock.Anything, appeal.AuditKeyBulkInsert, mock.Anything).Return(nil).Once()
			s.mockGrantService.On("List", mock.Anything, mock.Anything).Return([]domain.Grant{}, nil).Once()
			s.mockGrantService.On("Prepare", mock.Anything, mock.Anything).Return(&domain.Grant{}, nil).Once()
			s.mockPolicyService.On("GetOne", mock.Anything, mock.Anything, mock.Anything).Return(overriddingPolicy, nil).Once()
			s.mockProviderService.On("GrantAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

			err := s.service.Create(context.Background(), []*domain.Appeal{input}, appeal.CreateWithAdditionalAppeal())

			s.NoError(err)
			s.Equal("test-approval", input.Approvals[0].Name)
			s.Equal(expectedPermissions, input.Permissions)
		})
	})
}

func (s *ServiceTestSuite) TestCreateAppeal__WithExistingAppealAndWithAutoApprovalSteps() {
	s.setup()

	timeNow := time.Now()
	appeal.TimeNow = func() time.Time {
		return timeNow
	}

	accountID := "test@email.com"
	resourceIDs := []string{"1"}
	var resources []*domain.Resource
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
			ID:   "1",
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
								ID:          "role_id",
								Permissions: []interface{}{"test-permission"},
							},
						},
					},
				},
			},
		},
	}

	expectedExistingAppeals := []*domain.Appeal{}
	currentActiveGrant := domain.Grant{
		ID:         "99",
		AccountID:  accountID,
		ResourceID: "1",
		Resource: &domain.Resource{
			ID:  "1",
			URN: "urn",
		},
		Role:   "role_id",
		Status: domain.AppealStatusApproved,
	}
	expectedExistingGrants := []domain.Grant{currentActiveGrant}

	policies := []*domain.Policy{
		{
			ID:      "policy_1",
			Version: 1,
			Steps: []*domain.Step{
				{
					Name:        "step_1",
					Strategy:    "auto",
					AllowFailed: false,
					ApproveIf:   "1==1",
				},
			},
			IAM: &domain.IAMConfig{
				Provider: "http",
				Config: map[string]interface{}{
					"url": "http://localhost",
				},
			},
		},
	}

	expectedCreatorUser := map[string]interface{}{
		"managers": []interface{}{"user.approver@email.com"},
	}
	var expectedAppealsInsertionParam []*domain.Appeal

	for i, r := range resourceIDs {
		expectedAppealsInsertionParam = append(expectedAppealsInsertionParam, &domain.Appeal{
			ResourceID:    r,
			Resource:      resources[i],
			PolicyID:      "policy_1",
			PolicyVersion: 1,
			Status:        domain.AppealStatusApproved,
			AccountID:     accountID,
			AccountType:   domain.DefaultAppealAccountType,
			CreatedBy:     accountID,
			Creator:       expectedCreatorUser,
			Role:          "role_id",
			Permissions:   []string{"test-permission"},
			Approvals: []*domain.Approval{
				{
					Name:          "step_1",
					Index:         0,
					Status:        domain.ApprovalStatusApproved,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
				},
			},
			Grant: &domain.Grant{
				ResourceID:  r,
				Status:      domain.GrantStatusActive,
				AccountID:   accountID,
				AccountType: domain.DefaultAppealAccountType,
				Role:        "role_id",
				Permissions: []string{"test-permission"},
				Resource:    resources[i],
			},
		})
	}

	expectedResult := []*domain.Appeal{
		{
			ID:            "1",
			ResourceID:    "1",
			Resource:      resources[0],
			PolicyID:      "policy_1",
			PolicyVersion: 1,
			Status:        domain.AppealStatusApproved,
			AccountID:     accountID,
			AccountType:   domain.DefaultAppealAccountType,
			CreatedBy:     accountID,
			Creator:       expectedCreatorUser,
			Role:          "role_id",
			Permissions:   []string{"test-permission"},
			Approvals: []*domain.Approval{
				{
					ID:            "1",
					Name:          "step_1",
					Index:         0,
					Status:        domain.ApprovalStatusApproved,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
				},
			},
			Grant: &domain.Grant{
				ResourceID:  "1",
				Status:      domain.GrantStatusActive,
				AccountID:   accountID,
				AccountType: domain.DefaultAppealAccountType,
				Role:        "role_id",
				Permissions: []string{"test-permission"},
				Resource:    resources[0],
			},
		},
	}

	appeals := []*domain.Appeal{
		{
			CreatedBy:  accountID,
			AccountID:  accountID,
			ResourceID: "1",
			Resource: &domain.Resource{
				ID:  "1",
				URN: "urn",
			},
			Role: "role_id",
		},
	}

	expectedResourceFilters := domain.ListResourcesFilter{IDs: resourceIDs}
	s.mockResourceService.On("Find", mock.Anything, expectedResourceFilters).Return(resources, nil).Once()
	s.mockProviderService.On("Find", mock.Anything).Return(providers, nil).Once()
	s.mockPolicyService.On("Find", mock.Anything).Return(policies, nil).Once()
	expectedExistingAppealsFilters := &domain.ListAppealsFilter{
		Statuses: []string{domain.AppealStatusPending},
	}
	s.mockRepository.EXPECT().
		Find(mock.AnythingOfType("*context.emptyCtx"), expectedExistingAppealsFilters).
		Return(expectedExistingAppeals, nil).Once()
	s.mockGrantService.EXPECT().
		List(mock.AnythingOfType("*context.emptyCtx"), domain.ListGrantsFilter{
			Statuses: []string{string(domain.GrantStatusActive)},
		}).
		Return(expectedExistingGrants, nil).Once()
	s.mockProviderService.On("ValidateAppeal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.mockProviderService.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]interface{}{"test-permission"}, nil)
	s.mockIAMManager.On("ParseConfig", mock.Anything, mock.Anything).Return(nil, nil)
	s.mockIAMManager.On("GetClient", mock.Anything).Return(s.mockIAMClient, nil)
	s.mockIAMClient.On("GetUser", accountID).Return(expectedCreatorUser, nil)

	s.mockGrantService.EXPECT().
		List(mock.AnythingOfType("*context.emptyCtx"), domain.ListGrantsFilter{
			AccountIDs:  []string{accountID},
			ResourceIDs: []string{"1"},
			Statuses:    []string{string(domain.GrantStatusActive)},
			Permissions: []string{"test-permission"},
		}).Return(expectedExistingGrants, nil).Once()
	preparedGrant := &domain.Grant{
		Status:      domain.GrantStatusActive,
		AccountID:   accountID,
		AccountType: domain.DefaultAppealAccountType,
		ResourceID:  "1",
		Role:        "role_id",
		Permissions: []string{"test-permission"},
	}
	s.mockGrantService.EXPECT().
		Prepare(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("domain.Appeal")).
		Return(preparedGrant, nil).Once()
	s.mockGrantService.EXPECT().
		Revoke(mock.AnythingOfType("*context.emptyCtx"), currentActiveGrant.ID, domain.SystemActorName, appeal.RevokeReasonForExtension,
			mock.AnythingOfType("grant.Option"), mock.AnythingOfType("grant.Option"),
		).
		Return(preparedGrant, nil).Once()

	s.mockPolicyService.On("GetOne", mock.Anything, "policy_1", uint(1)).Return(policies[0], nil).Once()

	s.mockResourceService.On("Get", mock.Anything, appeals[0].Resource).Return(resources[0], nil).Once()
	s.mockProviderService.On("GrantAccess", mock.Anything, appeals[0]).Return(nil).Once()

	s.mockRepository.EXPECT().
		BulkUpsert(mock.AnythingOfType("*context.emptyCtx"), expectedAppealsInsertionParam).
		Return(nil).
		Run(func(_a0 context.Context, appeals []*domain.Appeal) {
			for i, a := range appeals {
				a.ID = expectedResult[i].ID
				for j, approval := range a.Approvals {
					approval.ID = expectedResult[i].Approvals[j].ID
				}
			}
		}).Once()
	s.mockNotifier.On("Notify", mock.Anything).Return(nil).Once()
	s.mockAuditLogger.On("Log", mock.Anything, appeal.AuditKeyBulkInsert, mock.Anything).Return(nil).Once()

	actualError := s.service.Create(context.Background(), appeals)

	s.Nil(actualError)
	s.Equal(expectedResult, appeals)
}

func (s *ServiceTestSuite) TestUpdateApproval() {
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
		s.mockRepository.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.UpdateApproval(context.Background(), validApprovalActionParam)

		s.mockRepository.AssertExpectations(s.T())
		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if appeal not found", func() {
		expectedError := appeal.ErrAppealNotFound
		s.mockRepository.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(nil, expectedError).Once()

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
			s.mockRepository.EXPECT().
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
		s.mockRepository.EXPECT().
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

		s.mockRepository.EXPECT().
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
		s.mockRepository.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), appealDetails).Return(nil).Once()
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
				s.setup()

				s.mockRepository.EXPECT().
					GetByID(mock.AnythingOfType("*context.emptyCtx"), validApprovalActionParam.AppealID).
					Return(tc.expectedAppealDetails, nil).Once()

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
					s.mockProviderService.EXPECT().GrantAccess(mock.Anything, *tc.expectedGrant).Return(nil).Once()

					tc.expectedResult.Policy = mockPolicy
				}

				s.mockRepository.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), tc.expectedResult).Return(nil).Once()
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

func (s *ServiceTestSuite) TestGrantAccessToProvider() {
	s.setup()

	s.Run("should return error when policy is not found", func() {
		expectedError := errors.New("retrieving policy: not found")

		s.mockPolicyService.On("GetOne", mock.Anything, "policy_1", uint(1)).Return(nil, errors.New("not found")).Once()

		actualError := s.service.GrantAccessToProvider(context.Background(), &domain.Appeal{
			PolicyID:      "policy_1",
			PolicyVersion: 1,
		})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("handle appeal requirements", func() {
		s.Run("matching error", func() {
			expectedError := errors.New("handling appeal requirements: evaluating requirements[1]: error parsing regexp: missing closing ]: `[InvalidRegex`")

			s.mockPolicyService.
				On("GetOne", mock.Anything, "policy_1", uint(1)).
				Return(&domain.Policy{
					ID:      "policy_1",
					Version: 1,
					Requirements: []*domain.Requirement{
						{
							On: &domain.RequirementTrigger{
								ProviderType: "not-matching",
							},
						},
						{
							On: &domain.RequirementTrigger{
								ProviderType: "[InvalidRegex",
							},
						},
					},
				}, nil).Once()

			actualError := s.service.GrantAccessToProvider(context.Background(), &domain.Appeal{
				PolicyID:      "policy_1",
				PolicyVersion: 1,
				Resource: &domain.Resource{
					ProviderType: "example-provider",
				},
			})
			s.EqualError(actualError, expectedError.Error())
		})
	})

	s.Run("should return error when grant access to provider fails", func() {
		expectedError := errors.New("granting access: error")

		s.mockPolicyService.
			On("GetOne", mock.Anything, "policy_1", uint(1)).
			Return(&domain.Policy{
				ID:      "policy_1",
				Version: 1,
			}, nil).Once()

		s.mockProviderService.
			On("GrantAccess", mock.Anything, mock.Anything).
			Return(fmt.Errorf("error")).Once()

		actualError := s.service.GrantAccessToProvider(context.Background(), &domain.Appeal{
			PolicyID:      "policy_1",
			PolicyVersion: 1,
			Grant:         &domain.Grant{},
		})
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should be able to grant access", func() {
		s.mockPolicyService.
			On("GetOne", mock.Anything, "policy_1", uint(1)).
			Return(&domain.Policy{
				ID:      "policy_1",
				Version: 1,
			}, nil).Once()

		s.mockProviderService.
			On("GrantAccess", mock.Anything, mock.Anything).
			Return(nil).Once()

		actualError := s.service.GrantAccessToProvider(context.Background(), &domain.Appeal{
			PolicyID:      "policy_1",
			PolicyVersion: 1,
			Grant:         &domain.Grant{},
		})
		s.Nil(actualError, actualError)
	})
}

// func (s *ServiceTestSuite) TestCancel() {
// 	s.Run("should return error from")
// }

func (s *ServiceTestSuite) TestAddApprover() {
	s.Run("should return appeal on success", func() {
		appealID := uuid.New().String()
		approvalID := uuid.New().String()
		approvalName := "test-approval-name"
		newApprover := "user@example.com"

		testCases := []struct {
			name, appealID, approvalID, newApprover string
		}{
			{
				name:     "with approval ID",
				appealID: appealID, approvalID: approvalID, newApprover: newApprover,
			},
			{
				name:     "with approval name",
				appealID: appealID, approvalID: approvalName, newApprover: newApprover,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				expectedAppeal := &domain.Appeal{
					ID:     appealID,
					Status: domain.AppealStatusPending,
					Approvals: []*domain.Approval{
						{
							ID:     approvalID,
							Name:   approvalName,
							Status: domain.ApprovalStatusPending,
							Approvers: []string{
								"existing.approver@example.com",
							},
						},
					},
					Resource: &domain.Resource{},
				}
				expectedApproval := &domain.Approval{
					ID:     approvalID,
					Name:   approvalName,
					Status: domain.ApprovalStatusPending,
					Approvers: []string{
						"existing.approver@example.com",
						tc.newApprover,
					},
				}
				s.mockRepository.EXPECT().
					GetByID(mock.AnythingOfType("*context.emptyCtx"), appealID).
					Return(expectedAppeal, nil).Once()
				s.mockApprovalService.EXPECT().
					AddApprover(mock.AnythingOfType("*context.emptyCtx"), approvalID, newApprover).
					Return(nil).Once()
				s.mockAuditLogger.EXPECT().
					Log(mock.AnythingOfType("*context.emptyCtx"), appeal.AuditKeyAddApprover, expectedApproval).Return(nil).Once()
				s.mockNotifier.EXPECT().Notify(mock.Anything).
					Run(func(notifications []domain.Notification) {
						s.Len(notifications, 1)
						n := notifications[0]
						s.Equal(tc.newApprover, n.User)
						s.Equal(domain.NotificationTypeApproverNotification, n.Message.Type)
					}).
					Return(nil).Once()

				actualAppeal, actualError := s.service.AddApprover(context.Background(), appealID, approvalID, newApprover)

				s.NoError(actualError)
				s.Equal(expectedApproval, actualAppeal.Approvals[0])
				s.mockRepository.AssertExpectations(s.T())
				s.mockApprovalService.AssertExpectations(s.T())
			})
		}
	})

	s.Run("params validation", func() {
		testCases := []struct {
			name, appealID, approvalID, email string
		}{
			{
				name:       "empty appealID",
				approvalID: uuid.New().String(),
				email:      "user@example.com",
			},
			{
				name:     "empty approvalID",
				appealID: uuid.New().String(),
				email:    "user@example.com",
			},
			{
				name:       "empty email",
				appealID:   uuid.New().String(),
				approvalID: uuid.New().String(),
			},
			{
				name:       "invalid email",
				appealID:   uuid.New().String(),
				approvalID: uuid.New().String(),
				email:      "invalid email",
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				appeal, err := s.service.AddApprover(context.Background(), tc.appealID, tc.approvalID, tc.email)

				s.Nil(appeal)
				s.Error(err)
			})
		}
	})

	s.Run("should return error if getting appeal details returns an error", func() {
		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(nil, expectedError).Once()

		appeal, err := s.service.AddApprover(context.Background(), uuid.New().String(), uuid.New().String(), "user@example.com")

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if appeal status is not pending", func() {
		approvalID := uuid.New().String()
		expectedError := appeal.ErrUnableToAddApprover
		expectedAppeal := &domain.Appeal{
			Status: domain.AppealStatusApproved,
			Approvals: []*domain.Approval{
				{
					ID: approvalID,
				},
			},
		}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedAppeal, nil).Once()

		appeal, err := s.service.AddApprover(context.Background(), uuid.New().String(), approvalID, "user@example.com")

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if approval not found", func() {
		expectedError := appeal.ErrApprovalNotFound
		expectedAppeal := &domain.Appeal{
			Status: domain.AppealStatusPending,
			Approvals: []*domain.Approval{
				{
					ID: "foobar",
				},
			},
		}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedAppeal, nil).Once()

		appeal, err := s.service.AddApprover(context.Background(), uuid.New().String(), uuid.New().String(), "user@example.com")

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if approval status is not pending or blocked", func() {
		expectedError := appeal.ErrUnableToAddApprover
		approvalID := uuid.New().String()
		expectedAppeal := &domain.Appeal{
			Status: domain.AppealStatusPending,
			Approvals: []*domain.Approval{
				{
					ID:     approvalID,
					Status: domain.ApprovalStatusApproved,
				},
			},
		}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedAppeal, nil).Once()

		appeal, err := s.service.AddApprover(context.Background(), uuid.New().String(), approvalID, "user@example.com")

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if approval is a manual step", func() {
		expectedError := appeal.ErrUnableToAddApprover
		approvalID := uuid.New().String()
		expectedAppeal := &domain.Appeal{
			Status: domain.AppealStatusPending,
			Approvals: []*domain.Approval{
				{
					ID:        approvalID,
					Status:    domain.ApprovalStatusBlocked,
					Approvers: nil,
				},
			},
		}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedAppeal, nil).Once()

		appeal, err := s.service.AddApprover(context.Background(), uuid.New().String(), approvalID, "user@example.com")

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if approval service returns an error when adding the new approver", func() {
		expectedError := errors.New("unexpected error")
		approvalID := uuid.New().String()
		expectedAppeal := &domain.Appeal{
			Status: domain.AppealStatusPending,
			Approvals: []*domain.Approval{
				{
					ID:        approvalID,
					Status:    domain.ApprovalStatusPending,
					Approvers: []string{"approver1@example.com"},
				},
			},
		}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedAppeal, nil).Once()
		s.mockApprovalService.EXPECT().AddApprover(mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

		appeal, err := s.service.AddApprover(context.Background(), uuid.New().String(), approvalID, "user@example.com")

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestDeleteApprover() {
	s.Run("should return nil error on success", func() {
		appealID := uuid.New().String()
		approvalID := uuid.New().String()
		approvalName := "test-approval-name"
		approverEmail := "user@example.com"

		testCases := []struct {
			name, appealID, approvalID, newApprover string
		}{
			{
				name:     "with approval ID",
				appealID: appealID, approvalID: approvalID, newApprover: approverEmail,
			},
			{
				name:     "with approval name",
				appealID: appealID, approvalID: approvalName, newApprover: approverEmail,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				expectedAppeal := &domain.Appeal{
					ID:     appealID,
					Status: domain.AppealStatusPending,
					Approvals: []*domain.Approval{
						{
							ID:     approvalID,
							Name:   approvalName,
							Status: domain.ApprovalStatusPending,
							Approvers: []string{
								"approver1@example.com",
								tc.newApprover,
							},
						},
					},
					Resource: &domain.Resource{},
				}
				expectedApproval := &domain.Approval{
					ID:     approvalID,
					Name:   approvalName,
					Status: domain.ApprovalStatusPending,
					Approvers: []string{
						"approver1@example.com",
					},
				}
				s.mockRepository.EXPECT().
					GetByID(mock.AnythingOfType("*context.emptyCtx"), appealID).
					Return(expectedAppeal, nil).Once()
				s.mockApprovalService.EXPECT().
					DeleteApprover(mock.AnythingOfType("*context.emptyCtx"), approvalID, approverEmail).
					Return(nil).Once()
				s.mockAuditLogger.EXPECT().
					Log(mock.AnythingOfType("*context.emptyCtx"), appeal.AuditKeyDeleteApprover, expectedApproval).Return(nil).Once()

				actualAppeal, actualError := s.service.DeleteApprover(context.Background(), appealID, approvalID, approverEmail)

				s.NoError(actualError)
				s.Equal(expectedApproval, actualAppeal.Approvals[0])
				s.mockRepository.AssertExpectations(s.T())
				s.mockApprovalService.AssertExpectations(s.T())
			})
		}
	})

	s.Run("params validation", func() {
		testCases := []struct {
			name, appealID, approvalID, email string
		}{
			{
				name:       "empty appealID",
				approvalID: uuid.New().String(),
				email:      "user@example.com",
			},
			{
				name:     "empty approvalID",
				appealID: uuid.New().String(),
				email:    "user@example.com",
			},
			{
				name:       "empty email",
				appealID:   uuid.New().String(),
				approvalID: uuid.New().String(),
			},
			{
				name:       "invalid email",
				appealID:   uuid.New().String(),
				approvalID: uuid.New().String(),
				email:      "invalid email",
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				appeal, err := s.service.DeleteApprover(context.Background(), tc.appealID, tc.approvalID, tc.email)

				s.Nil(appeal)
				s.Error(err)
			})
		}
	})

	s.Run("should return error if getting appeal details returns an error", func() {
		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(nil, expectedError).Once()

		appeal, err := s.service.DeleteApprover(context.Background(), uuid.New().String(), uuid.New().String(), "user@example.com")

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if appeal status is not pending", func() {
		approvalID := uuid.New().String()
		expectedError := appeal.ErrUnableToDeleteApprover
		expectedAppeal := &domain.Appeal{
			Status: domain.AppealStatusApproved,
			Approvals: []*domain.Approval{
				{
					ID: approvalID,
				},
			},
		}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedAppeal, nil).Once()

		appeal, err := s.service.DeleteApprover(context.Background(), uuid.New().String(), approvalID, "user@example.com")

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if approval status is not pending or blocked", func() {
		expectedError := appeal.ErrUnableToDeleteApprover
		approvalID := uuid.New().String()
		expectedAppeal := &domain.Appeal{
			Status: domain.AppealStatusPending,
			Approvals: []*domain.Approval{
				{
					ID:     approvalID,
					Status: domain.ApprovalStatusApproved,
				},
			},
		}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedAppeal, nil).Once()

		appeal, err := s.service.DeleteApprover(context.Background(), uuid.New().String(), approvalID, "user@example.com")

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if approval is a manual step", func() {
		expectedError := appeal.ErrUnableToDeleteApprover
		approvalID := uuid.New().String()
		expectedAppeal := &domain.Appeal{
			Status: domain.AppealStatusPending,
			Approvals: []*domain.Approval{
				{
					ID:        approvalID,
					Status:    domain.ApprovalStatusBlocked,
					Approvers: nil,
				},
			},
		}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedAppeal, nil).Once()

		appeal, err := s.service.DeleteApprover(context.Background(), uuid.New().String(), approvalID, "user@example.com")

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if there's only one approver", func() {
		expectedError := appeal.ErrUnableToDeleteApprover
		approvalID := uuid.New().String()
		expectedAppeal := &domain.Appeal{
			Status: domain.AppealStatusPending,
			Approvals: []*domain.Approval{
				{
					ID:        approvalID,
					Status:    domain.ApprovalStatusBlocked,
					Approvers: []string{"approver1@example.com"},
				},
			},
		}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedAppeal, nil).Once()

		appeal, err := s.service.DeleteApprover(context.Background(), uuid.New().String(), approvalID, "user@example.com")

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if approval service returns an error when deleting the new approver", func() {
		expectedError := appeal.ErrUnableToDeleteApprover
		approvalID := uuid.New().String()
		approverEmail := "user@example.com"
		expectedAppeal := &domain.Appeal{
			Status: domain.AppealStatusPending,
			Approvals: []*domain.Approval{
				{
					ID:        approvalID,
					Status:    domain.ApprovalStatusPending,
					Approvers: []string{approverEmail},
				},
			},
		}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedAppeal, nil).Once()
		s.mockApprovalService.EXPECT().DeleteApprover(mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

		appeal, err := s.service.DeleteApprover(context.Background(), uuid.New().String(), approvalID, approverEmail)

		s.Nil(appeal)
		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
