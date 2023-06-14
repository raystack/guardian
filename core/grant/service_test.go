package grant_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/raystack/guardian/core/grant"
	"github.com/raystack/guardian/core/grant/mocks"
	"github.com/raystack/guardian/domain"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository      *mocks.Repository
	mockProviderService *mocks.ProviderService
	mockResourceService *mocks.ResourceService
	mockAuditLogger     *mocks.AuditLogger
	mockNotifier        *mocks.Notifier
	service             *grant.Service
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) setup() {
	s.mockRepository = new(mocks.Repository)
	s.mockProviderService = new(mocks.ProviderService)
	s.mockResourceService = new(mocks.ResourceService)
	s.mockAuditLogger = new(mocks.AuditLogger)
	s.mockNotifier = new(mocks.Notifier)
	s.service = grant.NewService(grant.ServiceDeps{
		Repository:      s.mockRepository,
		Logger:          log.NewNoop(),
		Validator:       validator.New(),
		ProviderService: s.mockProviderService,
		ResourceService: s.mockResourceService,
		Notifier:        s.mockNotifier,
		AuditLogger:     s.mockAuditLogger,
	})
}

func (s *ServiceTestSuite) TestList() {
	s.Run("should return list of grant on success", func() {
		s.setup()

		filter := domain.ListGrantsFilter{}
		expectedGrants := []domain.Grant{}
		s.mockRepository.EXPECT().
			List(mock.AnythingOfType("*context.emptyCtx"), filter).
			Return(expectedGrants, nil).Once()

		grants, err := s.service.List(context.Background(), filter)

		s.NoError(err)
		s.Equal(expectedGrants, grants)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if repository returns an error", func() {
		s.setup()

		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().
			List(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("domain.ListGrantsFilter")).
			Return(nil, expectedError).Once()

		grants, err := s.service.List(context.Background(), domain.ListGrantsFilter{})

		s.ErrorIs(err, expectedError)
		s.Nil(grants)
		s.mockRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestGetByID() {
	s.Run("should return grant details on success", func() {
		s.setup()

		id := uuid.New().String()
		expectedGrant := &domain.Grant{}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), id).
			Return(expectedGrant, nil).
			Once()

		grant, err := s.service.GetByID(context.Background(), id)

		s.NoError(err)
		s.Equal(expectedGrant, grant)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if id param is empty", func() {
		s.setup()

		expectedError := grant.ErrEmptyIDParam

		grant, err := s.service.GetByID(context.Background(), "")

		s.ErrorIs(err, expectedError)
		s.Nil(grant)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if repository returns an error", func() {
		s.setup()

		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string")).
			Return(nil, expectedError).Once()

		grant, err := s.service.GetByID(context.Background(), "test-id")

		s.ErrorIs(err, expectedError)
		s.Nil(grant)
		s.mockRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestUpdate() {
	s.Run("should only update allowed fields", func() {
		s.setup()

		yesterday := time.Now().Add(-24 * time.Hour)
		id := uuid.New().String()
		existingGrant := &domain.Grant{
			ID:          id,
			AccountID:   "test-account-id",
			AccountType: "user",
			Owner:       "owner@example.com",
			Role:        "test-role",
			CreatedAt:   yesterday,
			UpdatedAt:   yesterday,
		}

		updatePayload := &domain.Grant{
			ID:          id,
			Owner:       "new-owner@example.com",
			IsPermanent: true,  // should be ignored
			Role:        "xyz", // should be ignored
		}

		expectedUpdateParam := &domain.Grant{
			ID:    id,
			Owner: "new-owner@example.com",
		}

		expectedUpdatedGrant := &domain.Grant{}
		*expectedUpdatedGrant = *existingGrant
		expectedUpdatedGrant.Owner = updatePayload.Owner
		expectedUpdatedGrant.UpdatedAt = time.Now()

		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), id).
			Return(existingGrant, nil).Once()
		s.mockRepository.EXPECT().
			Update(mock.AnythingOfType("*context.emptyCtx"), expectedUpdateParam).
			Return(nil).Run(func(_a0 context.Context, g *domain.Grant) {
			g.UpdatedAt = time.Now()
		}).Once()
		s.mockAuditLogger.EXPECT().
			Log(mock.AnythingOfType("*context.emptyCtx"), grant.AuditKeyUpdate, mock.AnythingOfType("map[string]interface {}")).Return(nil).Once()
		notificationMessage := domain.NotificationMessage{
			Type: domain.NotificationTypeGrantOwnerChanged,
			Variables: map[string]interface{}{
				"grant_id":       id,
				"previous_owner": existingGrant.Owner,
				"new_owner":      expectedUpdatedGrant.Owner,
			},
		}
		expectedNotifications := []domain.Notification{{
			User: updatePayload.Owner,
			Labels: map[string]string{
				"appeal_id": existingGrant.AppealID,
				"grant_id":  existingGrant.ID,
			},
			Message: notificationMessage,
		}, {
			User: existingGrant.Owner,
			Labels: map[string]string{
				"appeal_id": existingGrant.AppealID,
				"grant_id":  existingGrant.ID,
			},
			Message: notificationMessage,
		}}
		s.mockNotifier.EXPECT().
			Notify(expectedNotifications).Return(nil).Once()

		actualError := s.service.Update(context.Background(), updatePayload)
		s.NoError(actualError)
		s.Empty(cmp.Diff(expectedUpdatedGrant, updatePayload, cmpopts.EquateApproxTime(time.Second)))
	})

	s.Run("should return error if owner is updated to empty", func() {
		s.setup()

		id := uuid.New().String()
		existingGrant := &domain.Grant{
			ID:          id,
			AccountID:   "test-account-id",
			AccountType: "user",
			Owner:       "user@example.com",
			Role:        "test-role",
		}

		updatePayload := &domain.Grant{
			ID:    id,
			Owner: "",
		}

		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), id).
			Return(existingGrant, nil).Once()

		actualError := s.service.Update(context.Background(), updatePayload)
		s.ErrorIs(actualError, grant.ErrEmptyOwner)
	})
}

func (s *ServiceTestSuite) TestRevoke() {
	id := uuid.New().String()
	actor := "user@example.com"
	reason := "test reason"
	expectedGrantDetails := &domain.Grant{
		ID:          id,
		AccountID:   "test-account-id",
		AccountType: "user",
		Resource: &domain.Resource{
			ID: "test-resource-id",
		},
	}

	s.Run("should revoke grant on success", func() {
		s.setup()

		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), id).
			Return(expectedGrantDetails, nil).Once()
		s.mockRepository.EXPECT().
			Update(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*domain.Grant")).
			Run(func(_a0 context.Context, _a1 *domain.Grant) {
				s.Equal(id, _a1.ID)
				s.Equal(actor, _a1.RevokedBy)
				s.Equal(reason, _a1.RevokeReason)
				s.NotNil(_a1.RevokedAt)
			}).
			Return(nil).Once()
		s.mockProviderService.EXPECT().
			RevokeAccess(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("domain.Grant")).
			Run(func(_a0 context.Context, _a1 domain.Grant) {
				s.Equal(id, _a1.ID)
				s.Equal(expectedGrantDetails.AccountID, _a1.AccountID)
				s.Equal(expectedGrantDetails.AccountType, _a1.AccountType)
				s.Equal(expectedGrantDetails.Resource.ID, _a1.Resource.ID)
			}).
			Return(nil).Once()

		s.mockNotifier.EXPECT().
			Notify([]domain.Notification{{
				User: expectedGrantDetails.CreatedBy,
				Labels: map[string]string{
					"appeal_id": expectedGrantDetails.AppealID,
					"grant_id":  expectedGrantDetails.ID,
				},
				Message: domain.NotificationMessage{
					Type: domain.NotificationTypeAccessRevoked,
					Variables: map[string]interface{}{
						"resource_name": fmt.Sprintf("%s (%s: %s)", expectedGrantDetails.Resource.Name, expectedGrantDetails.Resource.ProviderType, expectedGrantDetails.Resource.URN),
						"role":          expectedGrantDetails.Role,
						"account_type":  expectedGrantDetails.AccountType,
						"account_id":    expectedGrantDetails.AccountID,
						"requestor":     expectedGrantDetails.Owner,
					},
				},
			}}).
			Return(nil).Once()
		s.mockAuditLogger.EXPECT().
			Log(mock.AnythingOfType("*context.emptyCtx"), grant.AuditKeyRevoke, map[string]interface{}{
				"grant_id": id,
				"reason":   reason,
			}).
			Return(nil).Once()

		expectedGrant, err := s.service.Revoke(context.Background(), id, actor, reason)

		s.NoError(err)
		s.Equal(id, expectedGrant.ID)
		s.Equal(actor, expectedGrant.RevokedBy)
		s.Equal(reason, expectedGrant.RevokeReason)
		s.NotNil(expectedGrant.RevokedAt)
		s.Less(*expectedGrant.RevokedAt, time.Now())
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should skip revoke in provider and notifications as configured", func() {
		s.setup()

		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), id).
			Return(expectedGrantDetails, nil).Once()
		s.mockRepository.EXPECT().
			Update(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*domain.Grant")).
			Run(func(_a0 context.Context, _a1 *domain.Grant) {
				s.Equal(id, _a1.ID)
				s.Equal(actor, _a1.RevokedBy)
				s.Equal(reason, _a1.RevokeReason)
				s.NotNil(_a1.RevokedAt)
			}).
			Return(nil).Once()

		s.mockAuditLogger.EXPECT().
			Log(mock.AnythingOfType("*context.emptyCtx"), grant.AuditKeyRevoke, map[string]interface{}{
				"grant_id": id,
				"reason":   reason,
			}).
			Return(nil).Once()

		expectedGrant, err := s.service.Revoke(context.Background(), id, actor, reason, grant.SkipRevokeAccessInProvider(), grant.SkipNotifications())

		s.NoError(err)
		s.Equal(id, expectedGrant.ID)
		s.Equal(actor, expectedGrant.RevokedBy)
		s.Equal(reason, expectedGrant.RevokeReason)
		s.NotNil(expectedGrant.RevokedAt)
		s.Less(*expectedGrant.RevokedAt, time.Now())
		s.mockRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestBulkRevoke() {
	actor := "test-actor@example.com"
	reason := "test reason"
	filter := domain.RevokeGrantsFilter{
		AccountIDs: []string{"test-account-id"},
	}
	expectedGrants := []domain.Grant{
		{
			ID:          "id1",
			AccountID:   "test-account-id",
			AccountType: "user",
			Resource: &domain.Resource{
				ID: "test-resource-id",
			},
		},
		{
			ID:          "id2",
			AccountID:   "test-account-id",
			AccountType: "user",
			Resource: &domain.Resource{
				ID: "test-resource-id",
			},
		},
	}

	s.Run("should return revoked grants on success", func() {
		s.setup()

		expectedListGrantsFilter := domain.ListGrantsFilter{
			Statuses:      []string{string(domain.GrantStatusActive)},
			AccountIDs:    filter.AccountIDs,
			ProviderTypes: filter.ProviderTypes,
			ProviderURNs:  filter.ProviderURNs,
			ResourceTypes: filter.ResourceTypes,
			ResourceURNs:  filter.ResourceURNs,
		}

		s.mockRepository.EXPECT().
			List(mock.AnythingOfType("*context.emptyCtx"), expectedListGrantsFilter).
			Return(expectedGrants, nil).Once()
		for _, g := range expectedGrants {
			grant := g
			s.mockProviderService.EXPECT().
				RevokeAccess(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("domain.Grant")).
				Run(func(_a0 context.Context, _a1 domain.Grant) {
					s.Equal(grant.ID, _a1.ID)
					s.Equal(grant.AccountID, _a1.AccountID)
					s.Equal(grant.AccountType, _a1.AccountType)
					s.Equal(grant.Resource.ID, _a1.Resource.ID)
				}).
				Return(nil).Once()

			s.mockRepository.EXPECT().
				Update(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*domain.Grant")).
				Run(func(_a0 context.Context, _a1 *domain.Grant) {
					s.Equal(grant.ID, _a1.ID)
					s.Equal(actor, _a1.RevokedBy)
					s.Equal(reason, _a1.RevokeReason)
					s.NotNil(_a1.RevokedAt)
				}).
				Return(nil).Once()
		}

		revokedGrants, actualError := s.service.BulkRevoke(context.Background(), filter, actor, reason)

		s.NoError(actualError)
		for i, g := range revokedGrants {
			revokedGrant := g
			expectedGrant := expectedGrants[i]
			s.Equal(expectedGrant.ID, revokedGrant.ID)
			s.Equal(actor, revokedGrant.RevokedBy)
			s.Equal(reason, revokedGrant.RevokeReason)
			s.NotNil(revokedGrant.RevokedAt)
			s.Less(*revokedGrant.RevokedAt, time.Now())
		}
	})
}

func (s *ServiceTestSuite) TestPrepare() {
	s.Run("should return error if appeal is invalid", func() {
		testCases := []struct {
			name   string
			appeal domain.Appeal
		}{
			{
				"appeal status is not approved",
				domain.Appeal{
					Status:      domain.AppealStatusPending,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-resource-id",
				},
			},
			{
				"account id is empty",
				domain.Appeal{
					Status:      domain.AppealStatusApproved,
					AccountID:   "",
					AccountType: "user",
					ResourceID:  "test-resource-id",
				},
			},
			{
				"account type is empty",
				domain.Appeal{
					Status:      domain.AppealStatusApproved,
					AccountID:   "user@example.com",
					AccountType: "",
					ResourceID:  "test-resource-id",
				},
			},
			{
				"resource id is empty",
				domain.Appeal{
					Status:      domain.AppealStatusApproved,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "",
				},
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.setup()
				actualGrant, actualError := s.service.Prepare(context.Background(), tc.appeal)

				s.Error(actualError)
				s.Nil(actualGrant)
			})
		}
	})

	s.Run("should return valid grant", func() {
		expDate := time.Now().Add(24 * time.Hour)
		testCases := []struct {
			name          string
			appeal        domain.Appeal
			expectedGrant *domain.Grant
		}{
			{
				name: "appeal with empty permanent duration option",
				appeal: domain.Appeal{
					ID:          "test-appeal-id",
					Status:      domain.AppealStatusApproved,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-user-id",
					Role:        "test-role",
					Permissions: []string{"test-permissions"},
					CreatedBy:   "user@example.com",
				},
				expectedGrant: &domain.Grant{
					Status:      domain.GrantStatusActive,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-user-id",
					Role:        "test-role",
					Permissions: []string{"test-permissions"},
					AppealID:    "test-appeal-id",
					CreatedBy:   "user@example.com",
					IsPermanent: true,
				},
			}, {
				name: "appeal with 0h as permanent duration option",
				appeal: domain.Appeal{
					ID:          "test-appeal-id",
					Status:      domain.AppealStatusApproved,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-user-id",
					Role:        "test-role",
					Permissions: []string{"test-permissions"},
					CreatedBy:   "user@example.com",
					Options: &domain.AppealOptions{
						Duration: "0h",
					},
				},
				expectedGrant: &domain.Grant{
					Status:      domain.GrantStatusActive,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-user-id",
					Role:        "test-role",
					Permissions: []string{"test-permissions"},
					AppealID:    "test-appeal-id",
					CreatedBy:   "user@example.com",
					IsPermanent: true,
				},
			},
			{
				name: "appeal with duration option",
				appeal: domain.Appeal{
					ID:          "test-appeal-id",
					Status:      domain.AppealStatusApproved,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-user-id",
					Role:        "test-role",
					Permissions: []string{"test-permissions"},
					CreatedBy:   "user@example.com",
					Options: &domain.AppealOptions{
						Duration: "24h",
					},
				},
				expectedGrant: &domain.Grant{
					Status:                  domain.GrantStatusActive,
					AccountID:               "user@example.com",
					AccountType:             "user",
					ResourceID:              "test-user-id",
					Role:                    "test-role",
					Permissions:             []string{"test-permissions"},
					AppealID:                "test-appeal-id",
					CreatedBy:               "user@example.com",
					IsPermanent:             false,
					ExpirationDate:          &expDate,
					RequestedExpirationDate: &expDate,
					ExpirationDateReason:    domain.ExpirationDateReasonFromAppeal,
				},
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.setup()
				actualGrant, actualError := s.service.Prepare(context.Background(), tc.appeal)

				s.NoError(actualError)
				if diff := cmp.Diff(tc.expectedGrant, actualGrant, cmpopts.EquateApproxTime(time.Second)); diff != "" {
					s.T().Errorf("result not match, diff: %v", diff)
				}
			})
		}
	})
}

func (s *ServiceTestSuite) TestImportFromProvider() {
	s.Run("should insert or update grants accordingly", func() {
		dummyProvider := &domain.Provider{
			ID:   "test-provider-id",
			Type: "test-provider-type",
			URN:  "test-provider-urn",
			Config: &domain.ProviderConfig{
				Type: "test-provider-type",
				URN:  "test-provider-urn",
				Resources: []*domain.ResourceConfig{
					{
						Type: "test-resource-type",
						Roles: []*domain.Role{
							{
								ID: "test-role-id",
								Permissions: []interface{}{
									"test-permission",
								},
							},
						},
					},
				},
			},
		}
		dummyResources := []*domain.Resource{
			{
				ID:           "test-resource-id",
				URN:          "test-resource-urn",
				Type:         "test-resource-type",
				ProviderType: "test-provider-type",
				ProviderURN:  "test-provider-urn",
			},
		}

		testCases := []struct {
			name                        string
			provider                    domain.Provider
			importedGrants              domain.MapResourceAccess
			existingGrants              []domain.Grant
			expectedDeactivatedGrants   []*domain.Grant
			expectedNewAndUpdatedGrants []*domain.Grant
		}{
			{
				name:                        "should return empty grants if no grants are imported",
				provider:                    *dummyProvider,
				importedGrants:              nil,
				expectedNewAndUpdatedGrants: nil,
			},
			{
				name:     "should insert imported grants",
				provider: *dummyProvider,
				importedGrants: domain.MapResourceAccess{
					"test-resource-urn": []domain.AccessEntry{
						{
							AccountID:   "test-account-id",
							AccountType: "user",
							Permission:  "test-permission",
						},
						{
							AccountID:   "test-account-id-2",
							AccountType: "serviceAccount",
							Permission:  "test-permission-2",
						},
					},
				},
				existingGrants:            []domain.Grant{},
				expectedDeactivatedGrants: nil,
				expectedNewAndUpdatedGrants: []*domain.Grant{
					{
						ResourceID:       "test-resource-id",
						AccountID:        "test-account-id",
						AccountType:      "user",
						Role:             "test-role-id",
						Permissions:      []string{"test-permission"},
						IsPermanent:      true,
						Status:           domain.GrantStatusActive,
						StatusInProvider: domain.GrantStatusActive,
						Source:           domain.GrantSourceImport,
						Owner:            "test-account-id",
					},
					{
						ResourceID:       "test-resource-id",
						AccountID:        "test-account-id-2",
						AccountType:      "serviceAccount",
						Role:             "test-permission-2",
						Permissions:      []string{"test-permission-2"},
						IsPermanent:      true,
						Status:           domain.GrantStatusActive,
						StatusInProvider: domain.GrantStatusActive,
						Source:           domain.GrantSourceImport,
					},
				},
			},
			{
				name:     "should deactivate status_in_provider of grants that are not in the imported grants",
				provider: *dummyProvider,
				importedGrants: domain.MapResourceAccess{
					"test-resource-urn": []domain.AccessEntry{
						{
							AccountID:   "test-account-id",
							AccountType: "user",
							Permission:  "test-permission",
						},
					},
				},
				existingGrants: []domain.Grant{
					{
						ID:               "test-grant-id",
						Status:           domain.GrantStatusActive,
						StatusInProvider: domain.GrantStatusActive,
						ResourceID:       "test-resource-id",
						AccountID:        "test-account-id",
						AccountType:      "user",
						Role:             "test-role-id",
						Permissions:      []string{"test-permission"},
						Resource:         dummyResources[0],
						Owner:            "test-account-id",
					},
					{
						ID:               "test-grant-id-2",
						Status:           domain.GrantStatusActive,
						StatusInProvider: domain.GrantStatusActive,
						ResourceID:       "test-resource-id",
						AccountID:        "test-account-id-2",
						AccountType:      "user",
						Role:             "test-role-id",
						Permissions:      []string{"test-permission"},
						Resource:         dummyResources[0],
						Owner:            "test-account-id-2",
					},
				},
				expectedDeactivatedGrants: []*domain.Grant{
					{
						ID:               "test-grant-id-2",
						Status:           domain.GrantStatusActive,
						StatusInProvider: domain.GrantStatusInactive,
						ResourceID:       "test-resource-id",
						AccountID:        "test-account-id-2",
						AccountType:      "user",
						Role:             "test-role-id",
						Permissions:      []string{"test-permission"},
						Resource:         dummyResources[0],
						Owner:            "test-account-id-2",
					},
				},
				expectedNewAndUpdatedGrants: []*domain.Grant{
					{
						ID:               "test-grant-id",
						Status:           domain.GrantStatusActive,
						StatusInProvider: domain.GrantStatusActive,
						ResourceID:       "test-resource-id",
						AccountID:        "test-account-id",
						AccountType:      "user",
						Role:             "test-role-id",
						Permissions:      []string{"test-permission"},
						Resource:         dummyResources[0],
						Owner:            "test-account-id",
					},
				},
			},
			{
				name: "imported access contain multiple permissions configured in one role id",
				provider: domain.Provider{
					ID:   "test-provider-id",
					Type: "test-provider-type",
					URN:  "test-provider-urn",
					Config: &domain.ProviderConfig{
						Type: "test-provider-type",
						URN:  "test-provider-urn",
						Resources: []*domain.ResourceConfig{
							{
								Type: "test-resource-type",
								Roles: []*domain.Role{
									{
										ID: "test-role-id",
										Permissions: []interface{}{
											"test-permission", "test-permission-2",
										},
									},
								},
							},
						},
					},
				},
				importedGrants: domain.MapResourceAccess{
					"test-resource-urn": []domain.AccessEntry{
						{
							AccountID:   "test-account-id", // existing
							AccountType: "user",
							Permission:  "test-permission",
						},
						{
							AccountID:   "test-account-id", // existing
							AccountType: "user",
							Permission:  "test-permission-2",
						},
						{
							AccountID:   "test-account-id", // new
							AccountType: "user",
							Permission:  "test-permission-3",
						},
						{
							AccountID:   "test-account-id-2", // new
							AccountType: "user",
							Permission:  "test-permission",
						},
						{
							AccountID:   "test-account-id-2", // new
							AccountType: "user",
							Permission:  "test-permission-2",
						},
					},
				},
				existingGrants: []domain.Grant{
					{
						ID:               "test-grant-id",
						Status:           domain.GrantStatusActive,
						StatusInProvider: domain.GrantStatusActive,
						ResourceID:       "test-resource-id",
						AccountID:        "test-account-id",
						AccountType:      "user",
						Role:             "test-role-id",
						Permissions:      []string{"test-permission", "test-permission-2"},
						Resource:         dummyResources[0],
						Owner:            "test-account-id",
					},
				},
				expectedNewAndUpdatedGrants: []*domain.Grant{
					{ // existing
						ID:               "test-grant-id",
						Status:           domain.GrantStatusActive,
						StatusInProvider: domain.GrantStatusActive,
						ResourceID:       "test-resource-id",
						AccountID:        "test-account-id",
						AccountType:      "user",
						Role:             "test-role-id",
						Permissions:      []string{"test-permission", "test-permission-2"},
						Resource:         dummyResources[0],
						Owner:            "test-account-id",
					},
					{ // new
						Status:           domain.GrantStatusActive,
						StatusInProvider: domain.GrantStatusActive,
						ResourceID:       "test-resource-id",
						AccountID:        "test-account-id",
						AccountType:      "user",
						Role:             "test-permission-3",
						Permissions:      []string{"test-permission-3"},
						Owner:            "test-account-id",
						Source:           "import",
						IsPermanent:      true,
					},
					{ // new
						Status:           domain.GrantStatusActive,
						StatusInProvider: domain.GrantStatusActive,
						ResourceID:       "test-resource-id",
						AccountID:        "test-account-id-2",
						AccountType:      "user",
						Role:             "test-role-id",
						Permissions:      []string{"test-permission", "test-permission-2"},
						Owner:            "test-account-id-2",
						Source:           "import",
						IsPermanent:      true,
					},
				},
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.setup()

				s.mockProviderService.EXPECT().
					GetByID(mock.AnythingOfType("*context.emptyCtx"), "test-provider-id").
					Return(&tc.provider, nil).Once()
				expectedListResourcesFilter := domain.ListResourcesFilter{
					ProviderType: "test-provider-type",
					ProviderURN:  "test-provider-urn",
				}
				s.mockResourceService.EXPECT().
					Find(mock.AnythingOfType("*context.emptyCtx"), expectedListResourcesFilter).
					Return(dummyResources, nil).Once()
				s.mockProviderService.EXPECT().
					ListAccess(mock.AnythingOfType("*context.emptyCtx"), tc.provider, dummyResources).
					Return(tc.importedGrants, nil).Once()
				expectedListGrantsFilter := domain.ListGrantsFilter{
					ProviderTypes: []string{"test-provider-type"},
					ProviderURNs:  []string{"test-provider-urn"},
					Statuses:      []string{string(domain.GrantStatusActive)},
				}
				s.mockRepository.EXPECT().
					List(mock.AnythingOfType("*context.emptyCtx"), expectedListGrantsFilter).
					Return(tc.existingGrants, nil).Once()

				s.mockRepository.EXPECT().
					BulkUpsert(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("[]*domain.Grant")).
					Return(nil).Once()

				s.mockRepository.EXPECT().
					BulkUpsert(mock.AnythingOfType("*context.emptyCtx"), tc.expectedDeactivatedGrants).
					Return(nil).Once()

				newGrants, err := s.service.ImportFromProvider(context.Background(), grant.ImportFromProviderCriteria{
					ProviderID: "test-provider-id",
				})

				s.NoError(err)
				s.Empty(cmp.Diff(tc.expectedNewAndUpdatedGrants, newGrants, cmpopts.SortSlices(func(a, b *domain.Grant) bool {
					if a.AccountID != b.AccountID {
						return a.AccountID < b.AccountID
					}
					return a.Role < b.Role
				})))
			})
		}
	})
}
