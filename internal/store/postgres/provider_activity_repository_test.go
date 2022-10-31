package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/odpf/guardian/core/provideractivity"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/suite"
)

type ProviderActivityRepositoryTestSuite struct {
	suite.Suite

	store              *postgres.Store
	repository         *postgres.ProviderActivityRepository
	providerRepository *postgres.ProviderRepository
	resourceRepository *postgres.ResourceRepository

	dummyProvider *domain.Provider
	dummyResource *domain.Resource
}

func TestProviderActivity(t *testing.T) {
	suite.Run(t, new(ProviderActivityRepositoryTestSuite))
}

func (s *ProviderActivityRepositoryTestSuite) SetupSuite() {
	logger := log.NewLogrus(log.LogrusWithLevel("debug"))
	store, pool, resource, err := newTestStore(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.store = store
	s.repository = postgres.NewProviderActivityRepository(store.DB())
	s.providerRepository = postgres.NewProviderRepository(store.DB())
	s.resourceRepository = postgres.NewResourceRepository(store.DB())

	s.T().Cleanup(func() {
		db, err := s.store.DB().DB()
		if err != nil {
			s.T().Fatal(err)
		}
		if err := db.Close(); err != nil {
			s.T().Fatal(err)
		}
		if err := purgeTestDocker(pool, resource); err != nil {
			s.T().Fatal(err)
		}
	})

	s.dummyProvider = &domain.Provider{
		ID: uuid.NewString(),
	}
	err = s.providerRepository.Create(context.Background(), s.dummyProvider)
	s.Require().NoError(err)
	s.dummyResource = &domain.Resource{
		ID: uuid.NewString(),
	}
	err = s.resourceRepository.BulkUpsert(context.Background(), []*domain.Resource{s.dummyResource})
	s.Require().NoError(err)
}

func (s *ProviderActivityRepositoryTestSuite) TestFind() {
	activity := &domain.ProviderActivity{
		ProviderID:     s.dummyProvider.ID,
		ResourceID:     s.dummyResource.ID,
		AccountID:      "user@example.com",
		Timestamp:      time.Now(),
		Type:           "test-type",
		Authorizations: []string{"test-authorization"},
		Metadata:       map[string]interface{}{"foo": "bar"},
	}
	activity2 := &domain.ProviderActivity{
		ProviderID:     s.dummyProvider.ID,
		ResourceID:     s.dummyResource.ID,
		AccountID:      "user2@example.com",
		Timestamp:      time.Now(),
		Type:           "test-type",
		Authorizations: []string{"test-authorization"},
		Metadata:       map[string]interface{}{"foo": "bar"},
	}
	err := s.repository.BulkInsert(context.Background(), []*domain.ProviderActivity{activity, activity2})
	s.Require().NoError(err)

	oneHourAgo := time.Now().Add(-time.Hour)
	now := time.Now()
	testCases := []struct {
		name               string
		filter             domain.ListProviderActivitiesFilter
		expectedActivities []*domain.ProviderActivity
	}{
		{
			"filter by provider ids",
			domain.ListProviderActivitiesFilter{
				ProviderIDs: []string{s.dummyProvider.ID},
			},
			[]*domain.ProviderActivity{activity, activity2},
		},
		{
			"filter by resoruce ids",
			domain.ListProviderActivitiesFilter{
				ResourceIDs: []string{s.dummyResource.ID},
			},
			[]*domain.ProviderActivity{activity, activity2},
		},
		{
			"filter by account ids",
			domain.ListProviderActivitiesFilter{
				AccountIDs: []string{"user@example.com"},
			},
			[]*domain.ProviderActivity{activity},
		},
		{
			"filter by types",
			domain.ListProviderActivitiesFilter{
				Types: []string{"test-type"},
			},
			[]*domain.ProviderActivity{activity, activity2},
		},
		{
			"filter by timestamp",
			domain.ListProviderActivitiesFilter{
				TimestampGte: &oneHourAgo,
				TimestampLte: &now,
			},
			[]*domain.ProviderActivity{activity, activity2},
		},
		{
			"filter by timestamp 2",
			domain.ListProviderActivitiesFilter{
				TimestampGte: &now,
			},
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			actualActivities, err := s.repository.Find(context.Background(), tc.filter)

			s.NoError(err)
			if diff := cmp.Diff(tc.expectedActivities, actualActivities, cmpopts.EquateApproxTime(time.Millisecond)); diff != "" {
				s.FailNow("unexpected activities", diff)
			}
		})
	}
}

func (s *ProviderActivityRepositoryTestSuite) TestGetOne() {
	activity := &domain.ProviderActivity{
		ProviderID:     s.dummyProvider.ID,
		ResourceID:     s.dummyResource.ID,
		AccountID:      "user@example.com",
		Timestamp:      time.Now(),
		Type:           "test-type",
		Authorizations: []string{"test-authorization"},
		Metadata:       map[string]interface{}{"foo": "bar"},
	}
	err := s.repository.BulkInsert(context.Background(), []*domain.ProviderActivity{activity})
	s.Require().NoError(err)

	s.Run("should return activity details", func() {
		expectedAcitivity := &domain.ProviderActivity{}
		*expectedAcitivity = *activity
		expectedAcitivity.Provider = s.dummyProvider
		expectedAcitivity.Resource = s.dummyResource

		actualActivity, err := s.repository.GetOne(context.Background(), activity.ID)

		s.NoError(err)
		if diff := cmp.Diff(expectedAcitivity, actualActivity, cmpopts.EquateApproxTime(time.Millisecond)); diff != "" {
			s.FailNow("unexpected activity", diff)
		}
	})

	s.Run("should return error if activity not found", func() {
		_, err := s.repository.GetOne(context.Background(), uuid.NewString())

		s.ErrorIs(err, provideractivity.ErrNotFound)
	})
}

func (s *ProviderActivityRepositoryTestSuite) TestBulkInsert() {
	s.Run("should return error if an error occured when converting domain.ProviderActivity", func() {
		invalidActivity := &domain.ProviderActivity{
			ProviderID: "invalid-uuid",
			ResourceID: "invalid-uuid",
		}

		err := s.repository.BulkInsert(context.Background(), []*domain.ProviderActivity{invalidActivity})

		s.Error(err)
	})

	activity := &domain.ProviderActivity{
		ProviderID:     s.dummyProvider.ID,
		ResourceID:     s.dummyResource.ID,
		AccountID:      "user@example.com",
		Timestamp:      time.Now(),
		Type:           "test-type",
		Authorizations: []string{"test-authorization"},
		Metadata:       map[string]interface{}{"foo": "bar"},
	}
	activity2 := &domain.ProviderActivity{
		ProviderID:     s.dummyProvider.ID,
		ResourceID:     s.dummyResource.ID,
		AccountID:      "user2@ecample.com",
		Timestamp:      time.Now(),
		Type:           "test-type",
		Authorizations: []string{"test-authorization"},
		Metadata:       map[string]interface{}{"foo": "bar"},
	}

	s.Run("should insert activities and update the IDs", func() {
		err := s.repository.BulkInsert(context.Background(), []*domain.ProviderActivity{activity, activity2})

		s.NoError(err)
		s.NotEmpty(activity.ID)
		s.NotEmpty(activity2.ID)
	})

	s.Run("should return error if provider relation not found", func() {
		activity := &domain.ProviderActivity{
			ProviderID: uuid.NewString(),
		}

		err := s.repository.BulkInsert(context.Background(), []*domain.ProviderActivity{activity})

		s.Error(err)
	})
}
