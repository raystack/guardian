package postgres_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/goto/guardian/core/appeal"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/internal/store/postgres"
	"github.com/goto/guardian/pkg/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type AppealRepositoryTestSuite struct {
	suite.Suite
	store      *postgres.Store
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.AppealRepository

	dummyProvider *domain.Provider
	dummyPolicy   *domain.Policy
	dummyResource *domain.Resource
}

func (s *AppealRepositoryTestSuite) SetupSuite() {
	var err error
	logger := log.NewCtxLogger("debug", []string{"test"})
	s.store, s.pool, s.resource, err = newTestStore(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	ctx := context.Background()

	s.repository = postgres.NewAppealRepository(s.store.DB())

	s.dummyPolicy = &domain.Policy{
		ID:      "policy_test",
		Version: 1,
	}
	policyRepository := postgres.NewPolicyRepository(s.store.DB())
	err = policyRepository.Create(ctx, s.dummyPolicy)
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
	err = providerRepository.Create(ctx, s.dummyProvider)
	s.Require().NoError(err)

	s.dummyResource = &domain.Resource{
		ProviderType: s.dummyProvider.Type,
		ProviderURN:  s.dummyProvider.URN,
		Type:         "resource_type_test",
		URN:          "resource_urn_test",
		Name:         "resource_name_test",
	}
	resourceRepository := postgres.NewResourceRepository(s.store.DB())
	err = resourceRepository.BulkUpsert(ctx, []*domain.Resource{s.dummyResource})
	s.Require().NoError(err)
}

func (s *AppealRepositoryTestSuite) TearDownSuite() {
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

func (s *AppealRepositoryTestSuite) TestGetByID() {
	s.Run("should return error if record not found", func() {
		someID := uuid.New().String()
		expectedError := appeal.ErrAppealNotFound

		actualResult, actualError := s.repository.GetByID(context.Background(), someID)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return records on success", func() {
		dummyAppeal := &domain.Appeal{
			ResourceID:    s.dummyResource.ID,
			PolicyID:      s.dummyPolicy.ID,
			PolicyVersion: s.dummyPolicy.Version,
			AccountID:     "user@example.com",
			AccountType:   domain.DefaultAppealAccountType,
			Role:          "role_test",
			Permissions:   []string{"permission_test"},
			CreatedBy:     "user@example.com",
		}

		ctx := context.Background()
		err := s.repository.BulkUpsert(ctx, []*domain.Appeal{dummyAppeal})
		s.Require().NoError(err)

		actualRecord, actualError := s.repository.GetByID(ctx, dummyAppeal.ID)

		s.Nil(actualError)
		s.Equal(dummyAppeal.ID, actualRecord.ID)
	})

	s.Run("should run query based on filters", func() {
		timeNowPlusAnHour := time.Now().Add(time.Hour)
		dummyAppeals := []*domain.Appeal{
			{
				ResourceID:    s.dummyResource.ID,
				PolicyID:      s.dummyPolicy.ID,
				PolicyVersion: s.dummyPolicy.Version,
				AccountID:     "user@example.com",
				AccountType:   domain.DefaultAppealAccountType,
				Role:          "role_test",
				Status:        domain.AppealStatusApproved,
				Permissions:   []string{"permission_test"},
				CreatedBy:     "user@example.com",
				Options: &domain.AppealOptions{
					ExpirationDate: &time.Time{},
				},
			},
			{
				ResourceID:    s.dummyResource.ID,
				PolicyID:      s.dummyPolicy.ID,
				PolicyVersion: s.dummyPolicy.Version,
				AccountID:     "user2@example.com",
				AccountType:   domain.DefaultAppealAccountType,
				Status:        domain.AppealStatusCanceled,
				Role:          "role_test",
				Permissions:   []string{"permission_test_2"},
				CreatedBy:     "user2@example.com",
				Options: &domain.AppealOptions{
					ExpirationDate: &timeNowPlusAnHour,
				},
			},
		}
		testCases := []struct {
			filters        *domain.ListAppealsFilter
			expectedArgs   []driver.Value
			expectedResult []*domain.Appeal
		}{
			{
				filters: &domain.ListAppealsFilter{
					Q: "user",
				},
				expectedResult: []*domain.Appeal{dummyAppeals[0], dummyAppeals[1]},
			},
			{
				filters: &domain.ListAppealsFilter{
					AccountTypes: []string{"x-account-type"},
				},
				expectedResult: []*domain.Appeal{dummyAppeals[0], dummyAppeals[1]},
			},
		}

		for _, tc := range testCases {
			_, actualError := s.repository.Find(context.Background(), tc.filters)
			s.Nil(actualError)
		}
	})
}

func (s *AppealRepositoryTestSuite) TestGetAppealsTotalCount() {
	s.Run("should return 0", func() {
		_, actualError := s.repository.GetAppealsTotalCount(context.Background(), &domain.ListAppealsFilter{})
		s.Nil(actualError)
	})
}

func (s *AppealRepositoryTestSuite) TestFind() {
	timeNowPlusAnHour := time.Now().Add(time.Hour)
	dummyAppeals := []*domain.Appeal{
		{
			ResourceID:    s.dummyResource.ID,
			PolicyID:      s.dummyPolicy.ID,
			PolicyVersion: s.dummyPolicy.Version,
			AccountID:     "user@example.com",
			AccountType:   domain.DefaultAppealAccountType,
			Role:          "role_test",
			Status:        domain.AppealStatusApproved,
			Permissions:   []string{"permission_test"},
			CreatedBy:     "user@example.com",
			Options: &domain.AppealOptions{
				ExpirationDate: &time.Time{},
			},
		},
		{
			ResourceID:    s.dummyResource.ID,
			PolicyID:      s.dummyPolicy.ID,
			PolicyVersion: s.dummyPolicy.Version,
			AccountID:     "user2@example.com",
			AccountType:   domain.DefaultAppealAccountType,
			Status:        domain.AppealStatusCanceled,
			Role:          "role_test",
			Permissions:   []string{"permission_test_2"},
			CreatedBy:     "user2@example.com",
			Options: &domain.AppealOptions{
				ExpirationDate: &timeNowPlusAnHour,
			},
		},
	}

	err := s.repository.BulkUpsert(context.Background(), dummyAppeals)
	s.Require().NoError(err)

	s.Run("should return error if filters validation returns an error", func() {
		invalidFilters := &domain.ListAppealsFilter{
			Statuses: []string{},
		}

		actualAppeals, actualError := s.repository.Find(context.Background(), invalidFilters)

		s.Error(actualError)
		s.Nil(actualAppeals)
	})

	s.Run("should return error if got any from db", func() {
		expectedError := errors.New("ERROR: invalid input syntax for type uuid: \"not-an-uuid\" (SQLSTATE 22P02)")

		actualResult, actualError := s.repository.Find(context.Background(), &domain.ListAppealsFilter{
			ResourceID: "not-an-uuid",
		})

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should run query based on filters", func() {
		timeNow := time.Now()
		testCases := []struct {
			filters        *domain.ListAppealsFilter
			expectedArgs   []driver.Value
			expectedResult []*domain.Appeal
		}{
			{
				filters:        &domain.ListAppealsFilter{},
				expectedResult: dummyAppeals,
			},
			{
				filters: &domain.ListAppealsFilter{
					CreatedBy: "user@email.com",
				},
				expectedResult: []*domain.Appeal{dummyAppeals[0]},
			},
			{
				filters: &domain.ListAppealsFilter{
					AccountIDs: []string{"user@email.com"},
				},
				expectedResult: []*domain.Appeal{dummyAppeals[0]},
			},
			{
				filters: &domain.ListAppealsFilter{
					AccountID: "user@email.com",
				},
				expectedResult: []*domain.Appeal{dummyAppeals[0]},
			},
			{
				filters: &domain.ListAppealsFilter{
					Statuses: []string{domain.AppealStatusApproved, domain.AppealStatusPending},
				},
				expectedResult: []*domain.Appeal{dummyAppeals[0]},
			},
			{
				filters: &domain.ListAppealsFilter{
					ResourceID: s.dummyResource.ID,
				},
				expectedResult: dummyAppeals,
			},
			{
				filters: &domain.ListAppealsFilter{
					Role: "test-role",
				},
				expectedResult: dummyAppeals,
			},
			{
				filters: &domain.ListAppealsFilter{
					ExpirationDateLessThan: timeNow,
				},
				expectedResult: []*domain.Appeal{dummyAppeals[0]},
			},
			{
				filters: &domain.ListAppealsFilter{
					ExpirationDateGreaterThan: timeNow,
				},
				expectedResult: []*domain.Appeal{dummyAppeals[0]},
			},
			{
				filters: &domain.ListAppealsFilter{
					ProviderTypes: []string{s.dummyProvider.Type},
				},
				expectedResult: dummyAppeals,
			},
			{
				filters: &domain.ListAppealsFilter{
					ProviderURNs: []string{s.dummyProvider.URN},
				},
				expectedResult: dummyAppeals,
			},
			{
				filters: &domain.ListAppealsFilter{
					ResourceTypes: []string{s.dummyResource.Type},
				},
				expectedResult: dummyAppeals,
			},
			{
				filters: &domain.ListAppealsFilter{
					ResourceURNs: []string{s.dummyResource.URN},
				},
				expectedResult: dummyAppeals,
			},
			{
				filters: &domain.ListAppealsFilter{
					OrderBy: []string{"status"},
				},
				expectedResult: []*domain.Appeal{dummyAppeals[0], dummyAppeals[1]},
			},
			{
				filters: &domain.ListAppealsFilter{
					OrderBy: []string{"updated_at:desc"},
				},
				expectedResult: []*domain.Appeal{dummyAppeals[1], dummyAppeals[0]},
			},
			{
				filters: &domain.ListAppealsFilter{
					Q: "user",
				},
				expectedResult: []*domain.Appeal{dummyAppeals[1], dummyAppeals[0]},
			},
			{
				filters: &domain.ListAppealsFilter{
					AccountTypes: []string{"x-account-type"},
				},
				expectedResult: []*domain.Appeal{dummyAppeals[1], dummyAppeals[0]},
			},
		}

		for _, tc := range testCases {
			_, actualError := s.repository.Find(context.Background(), tc.filters)
			s.Nil(actualError)
		}
	})

	s.Run("Should return an array size and offset of n on success", func() {
		testCases := []struct {
			filters        *domain.ListAppealsFilter
			expectedArgs   []driver.Value
			expectedResult []*domain.Appeal
		}{
			{
				filters: &domain.ListAppealsFilter{
					Size:   1,
					Offset: 0,
				},
				expectedResult: []*domain.Appeal{dummyAppeals[0]},
			},
			{
				filters: &domain.ListAppealsFilter{
					Offset: 1,
				},
				expectedResult: []*domain.Appeal{dummyAppeals[1]},
			},
		}
		for _, tc := range testCases {
			_, actualError := s.repository.Find(context.Background(), tc.filters)
			s.Nil(actualError)
		}
	})
}

func (s *AppealRepositoryTestSuite) TestBulkUpsert() {
	s.Run("should return error if appeals input is invalid", func() {
		invalidAppeals := []*domain.Appeal{
			{
				Details: map[string]interface{}{
					"foo": make(chan int), // invalid value
				},
			},
		}

		actualErr := s.repository.BulkUpsert(context.Background(), invalidAppeals)

		s.EqualError(actualErr, "json: unsupported type: chan int")
	})

	dummyAppeals := []*domain.Appeal{
		{
			ResourceID:    s.dummyResource.ID,
			PolicyID:      s.dummyPolicy.ID,
			PolicyVersion: s.dummyPolicy.Version,
			AccountID:     "user@example.com",
			AccountType:   domain.DefaultAppealAccountType,
			Role:          "role_test",
			Status:        domain.AppealStatusApproved,
			Permissions:   []string{"permission_test"},
			CreatedBy:     "user@example.com",
			Description:   "The answer is 42",
		},
		{
			ResourceID:    s.dummyResource.ID,
			PolicyID:      s.dummyPolicy.ID,
			PolicyVersion: s.dummyPolicy.Version,
			AccountID:     "user2@example.com",
			AccountType:   domain.DefaultAppealAccountType,
			Status:        domain.AppealStatusCanceled,
			Role:          "role_test",
			Permissions:   []string{"permission_test_2"},
			CreatedBy:     "user2@example.com",
		},
	}

	s.Run("should return nil error on success", func() {
		actualError := s.repository.BulkUpsert(context.Background(), dummyAppeals)
		s.Nil(actualError)
	})
}

func (s *AppealRepositoryTestSuite) TestUpdate() {
	s.Run("should return error if appeal input is invalid", func() {
		invalidAppeal := &domain.Appeal{
			ID: uuid.New().String(),
			Details: map[string]interface{}{
				"foo": make(chan int), // invalid value
			},
		}

		actualError := s.repository.Update(context.Background(), invalidAppeal)

		s.EqualError(actualError, "json: unsupported type: chan int")
	})

	s.Run("should return nil on success", func() {
		dummyAppeals := []*domain.Appeal{
			{
				ResourceID:    s.dummyResource.ID,
				PolicyID:      s.dummyPolicy.ID,
				PolicyVersion: s.dummyPolicy.Version,
				AccountID:     "user@example.com",
				AccountType:   domain.DefaultAppealAccountType,
				Role:          "role_test",
				Status:        domain.AppealStatusApproved,
				Permissions:   []string{"permission_test"},
				CreatedBy:     "user@example.com",
			},
			{
				ResourceID:    s.dummyResource.ID,
				PolicyID:      s.dummyPolicy.ID,
				PolicyVersion: s.dummyPolicy.Version,
				AccountID:     "user2@example.com",
				AccountType:   domain.DefaultAppealAccountType,
				Status:        domain.AppealStatusCanceled,
				Role:          "role_test",
				Permissions:   []string{"permission_test_2"},
				CreatedBy:     "user2@example.com",
			},
		}

		ctx := context.Background()
		actualError := s.repository.BulkUpsert(ctx, dummyAppeals)
		s.Nil(actualError)

		err := s.repository.Update(ctx, dummyAppeals[0])
		s.Nil(err)
	})
}

func TestAppealRepository(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, new(AppealRepositoryTestSuite))
}
