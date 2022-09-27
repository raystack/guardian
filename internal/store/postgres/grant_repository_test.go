package postgres_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"
	"time"

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

	timeNow     time.Time
	columnNames []string
}

func TestGrantRepository(t *testing.T) {
	suite.Run(t, new(GrantRepositoryTestSuite))
}

func (s *GrantRepositoryTestSuite) SetupSuite() {
	s.setup()
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

func (s *GrantRepositoryTestSuite) setup() {
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
	}
	appealRepository := postgres.NewAppealRepository(s.store.DB())
	err = appealRepository.BulkUpsert([]*domain.Appeal{s.dummyAppeal})
	s.Require().NoError(err)

	s.timeNow = time.Now()
	s.columnNames = []string{
		"id", "status", "account_id", "account_type", "resource_id", "role", "permissions",
		"expiration_date", "appeal_id", "revoked_by", "revoked_at", "revoke_reason",
		"created_by", "created_at", "updated_at",
	}
}

func (s *GrantRepositoryTestSuite) toRow(a domain.Grant) []driver.Value {
	permissions := fmt.Sprintf("{%s}", strings.Join(a.Permissions, ","))
	return []driver.Value{
		a.ID, a.Status, a.AccountID, a.AccountType, a.ResourceID, a.Role, permissions,
		a.ExpirationDate, a.AppealID, a.RevokedBy, a.RevokedAt, a.RevokeReason,
		a.CreatedBy, a.CreatedAt, a.UpdatedAt,
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
		s.Equal(*expectedGrant, grants[0])
	})

	s.Run("sould return error if db returns an error", func() {
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
		s.Equal(expectedGrant, grant)
	})

	s.Run("should return not found error if record not found", func() {
		newID := uuid.NewString()
		actualGrant, err := s.repository.GetByID(context.Background(), newID)

		s.ErrorIs(err, grant.ErrGrantNotFound)
		s.Nil(actualGrant)
	})
}
