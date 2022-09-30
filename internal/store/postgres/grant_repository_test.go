package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/odpf/guardian/core/grant"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type GrantRepositoryTestSuite struct {
	suite.Suite
	store      *postgres.Store
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.GrantRepository

	dummyProvider *domain.Provider
	dummyPolicy   *domain.Policy
	dummyResource *domain.Resource
	dummyAppeal   *domain.Appeal
}

func TestGrantRepository(t *testing.T) {
	suite.Run(t, new(GrantRepositoryTestSuite))
}

func (s *GrantRepositoryTestSuite) SetupSuite() {
	var err error
	logger := log.NewLogrus(log.LogrusWithLevel("debug"))
	s.store, s.pool, s.resource, err = newTestStore(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.repository = postgres.NewGrantRepository(s.store.DB())

	s.dummyPolicy = &domain.Policy{
		ID:      "policy_test",
		Version: 1,
	}
	policyRepository := postgres.NewPolicyRepository(s.store.DB())
	err = policyRepository.Create(s.dummyPolicy)
	s.Require().NoError(err)

	s.dummyProvider = &domain.Provider{
		Type: "provider_test",
		URN:  "provider_urn_test",
		Config: &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: "resource_type_test",
					Policy: &domain.PolicyConfig{
						ID:      s.dummyPolicy.ID,
						Version: int(s.dummyPolicy.Version),
					},
				},
			},
		},
	}
	providerRepository := postgres.NewProviderRepository(s.store.DB())
	err = providerRepository.Create(s.dummyProvider)
	s.Require().NoError(err)

	s.dummyResource = &domain.Resource{
		ProviderType: s.dummyProvider.Type,
		ProviderURN:  s.dummyProvider.URN,
		Type:         "resource_type_test",
		URN:          "resource_urn_test",
		Name:         "resource_name_test",
	}
	resourceRepository := postgres.NewResourceRepository(s.store.DB())
	err = resourceRepository.BulkUpsert([]*domain.Resource{s.dummyResource})
	s.Require().NoError(err)

	s.dummyAppeal = &domain.Appeal{
		ResourceID:    s.dummyResource.ID,
		PolicyID:      s.dummyPolicy.ID,
		PolicyVersion: s.dummyPolicy.Version,
		AccountID:     "user@example.com",
		AccountType:   domain.DefaultAppealAccountType,
		Role:          "role_test",
		Permissions:   []string{"permission_test"},
		CreatedBy:     "user@example.com",
	}
	appealRepository := postgres.NewAppealRepository(s.store.DB())
	err = appealRepository.BulkUpsert([]*domain.Appeal{s.dummyAppeal})
	s.Require().NoError(err)
}

func (s *GrantRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	db, err := s.store.DB().DB()
	if err != nil {
		s.T().Fatal(err)
	}
	err = db.Close()
	if err != nil {
		s.T().Fatal(err)
	}

	err = purgeTestDocker(s.pool, s.resource)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *GrantRepositoryTestSuite) TestList() {
	expDate := time.Now().Truncate(time.Nanosecond)
	dummyGrants := []*domain.Grant{
		{
			Status:         domain.GrantStatusActive,
			AppealID:       s.dummyAppeal.ID,
			AccountID:      s.dummyAppeal.AccountID,
			AccountType:    s.dummyAppeal.AccountType,
			ResourceID:     s.dummyAppeal.ResourceID,
			Role:           s.dummyAppeal.Role,
			Permissions:    s.dummyAppeal.Permissions,
			CreatedBy:      s.dummyAppeal.CreatedBy,
			ExpirationDate: &expDate,
		},
	}
	err := s.repository.BulkInsert(context.Background(), dummyGrants)
	s.Require().NoError(err)

	s.Run("should return list of grant on success", func() {
		expectedGrant := &domain.Grant{}
		*expectedGrant = *dummyGrants[0]
		expectedGrant.Resource = s.dummyResource
		expectedGrant.Appeal = s.dummyAppeal

		grants, err := s.repository.List(context.Background(), domain.ListGrantsFilter{
			Statuses:                  []string{string(domain.GrantStatusActive)},
			AccountIDs:                []string{s.dummyAppeal.AccountID},
			AccountTypes:              []string{s.dummyAppeal.AccountType},
			ResourceIDs:               []string{s.dummyAppeal.ResourceID},
			Roles:                     []string{s.dummyAppeal.Role},
			Permissions:               s.dummyAppeal.Permissions,
			ProviderTypes:             []string{s.dummyResource.ProviderType},
			ProviderURNs:              []string{s.dummyResource.ProviderURN},
			ResourceTypes:             []string{s.dummyResource.Type},
			ResourceURNs:              []string{s.dummyResource.URN},
			CreatedBy:                 s.dummyAppeal.CreatedBy,
			OrderBy:                   []string{"status"},
			ExpirationDateLessThan:    time.Now(),
			ExpirationDateGreaterThan: time.Now().Add(-24 * time.Hour),
		})

		s.NoError(err)
		s.Len(grants, 1)
		if diff := cmp.Diff(*expectedGrant, grants[0], cmpopts.IgnoreFields(domain.Grant{}, "CreatedAt", "UpdatedAt")); diff != "" {
			s.T().Errorf("result not match, diff: %v", diff)
		}
	})

	s.Run("could return error if db returns an error", func() {
		grants, err := s.repository.List(context.Background(), domain.ListGrantsFilter{
			ResourceIDs: []string{"invalid uuid"},
		})

		s.Error(err)
		s.Nil(grants)
	})
}

func (s *GrantRepositoryTestSuite) TestGetByID() {
	dummyGrants := []*domain.Grant{
		{
			Status:      domain.GrantStatusActive,
			AppealID:    s.dummyAppeal.ID,
			AccountID:   s.dummyAppeal.AccountID,
			AccountType: s.dummyAppeal.AccountType,
			ResourceID:  s.dummyAppeal.ResourceID,
			Role:        s.dummyAppeal.Role,
			Permissions: s.dummyAppeal.Permissions,
			CreatedBy:   s.dummyAppeal.CreatedBy,
		},
	}
	err := s.repository.BulkInsert(context.Background(), dummyGrants)
	s.Require().NoError(err)

	s.Run("should return grant details on success", func() {
		expectedID := dummyGrants[0].ID
		expectedGrant := &domain.Grant{}
		*expectedGrant = *dummyGrants[0]
		expectedGrant.Resource = s.dummyResource
		expectedGrant.Appeal = s.dummyAppeal

		grant, err := s.repository.GetByID(context.Background(), expectedID)

		s.NoError(err)
		if diff := cmp.Diff(expectedGrant, grant, cmpopts.IgnoreFields(domain.Grant{}, "CreatedAt", "UpdatedAt")); diff != "" {
			s.T().Errorf("result not match, diff: %v", diff)
		}
	})

	s.Run("should return not found error if record not found", func() {
		newID := uuid.NewString()
		actualGrant, err := s.repository.GetByID(context.Background(), newID)

		s.ErrorIs(err, grant.ErrGrantNotFound)
		s.Nil(actualGrant)
	})
}

func (s *GrantRepositoryTestSuite) TestUpdate() {
	dummyGrants := []*domain.Grant{
		{
			Status:      domain.GrantStatusActive,
			AppealID:    s.dummyAppeal.ID,
			AccountID:   s.dummyAppeal.AccountID,
			AccountType: s.dummyAppeal.AccountType,
			ResourceID:  s.dummyAppeal.ResourceID,
			Role:        s.dummyAppeal.Role,
			Permissions: s.dummyAppeal.Permissions,
			CreatedBy:   s.dummyAppeal.CreatedBy,
		},
	}
	err := s.repository.BulkInsert(context.Background(), dummyGrants)
	s.Require().NoError(err)

	s.Run("should return nil error on success", func() {
		expectedID := dummyGrants[0].ID
		payload := &domain.Grant{
			ID:     expectedID,
			Status: domain.GrantStatusInactive,
		}

		err := s.repository.Update(context.Background(), payload)
		s.NoError(err)

		updatedGrant, err := s.repository.GetByID(context.Background(), expectedID)
		s.Require().NoError(err)

		s.Equal(payload.Status, updatedGrant.Status)
		s.Greater(updatedGrant.UpdatedAt, dummyGrants[0].UpdatedAt)
	})

	s.Run("should return error if id param is empty", func() {
		payload := &domain.Grant{
			ID:     "",
			Status: domain.GrantStatusInactive,
		}

		err := s.repository.Update(context.Background(), payload)

		s.ErrorIs(err, grant.ErrEmptyIDParam)
	})

	s.Run("should return error if db execution returns an error", func() {
		payload := &domain.Grant{
			ID:     "invalid-uuid",
			Status: domain.GrantStatusInactive,
		}

		err := s.repository.Update(context.Background(), payload)

		s.Error(err)
	})
}
