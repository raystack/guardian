package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/goto/guardian/core/activity"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/internal/store/postgres"
	"github.com/goto/guardian/pkg/log"
	"github.com/stretchr/testify/suite"
)

type ActivityRepositoryTestSuite struct {
	suite.Suite

	store              *postgres.Store
	repository         *postgres.ActivityRepository
	providerRepository *postgres.ProviderRepository
	resourceRepository *postgres.ResourceRepository

	dummyProvider *domain.Provider
	dummyResource *domain.Resource
}

func TestActivityRepository(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, new(ActivityRepositoryTestSuite))
}

func (s *ActivityRepositoryTestSuite) SetupSuite() {
	logger := log.NewCtxLogger("info", []string{"test"})
	store, pool, resource, err := newTestStore(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.store = store
	s.repository = postgres.NewActivityRepository(store.DB())
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
		ID:   uuid.NewString(),
		Type: "test-provider",
		URN:  "test-provider-urn",
	}
	err = s.providerRepository.Create(context.Background(), s.dummyProvider)
	s.Require().NoError(err)
	s.dummyResource = &domain.Resource{
		ID:           uuid.NewString(),
		ProviderType: s.dummyProvider.Type,
		ProviderURN:  s.dummyProvider.URN,
	}
	err = s.resourceRepository.BulkUpsert(context.Background(), []*domain.Resource{s.dummyResource})
	s.Require().NoError(err)
}

func (s *ActivityRepositoryTestSuite) AfterTest(_, _ string) {
	if err := s.store.DB().Exec("DELETE FROM activities").Error; err != nil {
		s.T().Fatal(err)
	}
}

func (s *ActivityRepositoryTestSuite) TestFind() {
	activity := &domain.Activity{
		ProviderID:         s.dummyProvider.ID,
		ResourceID:         s.dummyResource.ID,
		ProviderActivityID: "test-provider-activity-id-1",
		AccountID:          "user@example.com",
		Timestamp:          time.Now(),
		Type:               "test-type",
		Authorizations:     []string{"test-authorization"},
		Metadata:           map[string]interface{}{"foo": "bar"},
	}
	activity2 := &domain.Activity{
		ProviderID:         s.dummyProvider.ID,
		ResourceID:         s.dummyResource.ID,
		ProviderActivityID: "test-provider-activity-id-2",
		AccountID:          "user2@example.com",
		Timestamp:          time.Now(),
		Type:               "test-type",
		Authorizations:     []string{"test-authorization"},
		Metadata:           map[string]interface{}{"foo": "bar"},
	}
	err := s.repository.BulkUpsert(context.Background(), []*domain.Activity{activity, activity2})
	s.Require().NoError(err)

	oneHourAgo := time.Now().Add(-time.Hour)
	now := time.Now()
	testCases := []struct {
		name               string
		filter             domain.ListProviderActivitiesFilter
		expectedActivities []*domain.Activity
	}{
		{
			"filter by provider ids",
			domain.ListProviderActivitiesFilter{
				ProviderIDs: []string{s.dummyProvider.ID},
			},
			[]*domain.Activity{activity, activity2},
		},
		{
			"filter by resoruce ids",
			domain.ListProviderActivitiesFilter{
				ResourceIDs: []string{s.dummyResource.ID},
			},
			[]*domain.Activity{activity, activity2},
		},
		{
			"filter by account ids",
			domain.ListProviderActivitiesFilter{
				AccountIDs: []string{"user@example.com"},
			},
			[]*domain.Activity{activity},
		},
		{
			"filter by types",
			domain.ListProviderActivitiesFilter{
				Types: []string{"test-type"},
			},
			[]*domain.Activity{activity, activity2},
		},
		{
			"filter by timestamp",
			domain.ListProviderActivitiesFilter{
				TimestampGte: &oneHourAgo,
				TimestampLte: &now,
			},
			[]*domain.Activity{activity, activity2},
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

func (s *ActivityRepositoryTestSuite) TestGetOne() {
	a := &domain.Activity{
		ProviderID:     s.dummyProvider.ID,
		ResourceID:     s.dummyResource.ID,
		AccountID:      "user@example.com",
		Timestamp:      time.Now(),
		Type:           "test-type",
		Authorizations: []string{"test-authorization"},
		Metadata:       map[string]interface{}{"foo": "bar"},
	}
	err := s.repository.BulkUpsert(context.Background(), []*domain.Activity{a})
	s.Require().NoError(err)

	s.Run("should return activity details", func() {
		expectedAcitivity := &domain.Activity{}
		*expectedAcitivity = *a
		expectedAcitivity.Provider = s.dummyProvider
		expectedAcitivity.Resource = s.dummyResource

		actualActivity, err := s.repository.GetOne(context.Background(), a.ID)

		s.NoError(err)
		if diff := cmp.Diff(expectedAcitivity, actualActivity, cmpopts.EquateApproxTime(time.Millisecond)); diff != "" {
			s.FailNow("unexpected activity", diff)
		}
	})

	s.Run("should return error if activity not found", func() {
		_, err := s.repository.GetOne(context.Background(), uuid.NewString())

		s.ErrorIs(err, activity.ErrNotFound)
	})
}

func (s *ActivityRepositoryTestSuite) TestBulkUpsert() {
	s.Run("should return error if an error occurred when converting domain.Activity", func() {
		invalidActivity := &domain.Activity{
			ProviderID: "invalid-uuid",
			ResourceID: "invalid-uuid",
		}

		err := s.repository.BulkUpsert(context.Background(), []*domain.Activity{invalidActivity})

		s.Error(err)
	})

	dummyResource := &domain.Resource{
		ProviderType: s.dummyProvider.Type,
		ProviderURN:  s.dummyProvider.URN,
		Type:         "test-resource",
		URN:          "test-urn",
	}
	preExistingActivity := &domain.Activity{
		ProviderID:         s.dummyProvider.ID,
		Resource:           dummyResource,
		ProviderActivityID: "test-provider-activity-id",
		AccountID:          "user@example.com",
		Timestamp:          time.Now(),
		Type:               "test-type",
		Authorizations:     []string{"test-authorization"},
		Metadata:           map[string]interface{}{"foo": "bar"},
	}
	err := s.repository.BulkUpsert(context.Background(), []*domain.Activity{preExistingActivity})
	s.Require().NoError(err)

	activity := &domain.Activity{
		ProviderID:         s.dummyProvider.ID,
		Resource:           dummyResource,
		ProviderActivityID: "test-provider-activity-id",
		AccountID:          "user@example.com",
		Timestamp:          time.Now(),
		Type:               "test-type",
		Authorizations:     []string{"test-authorization"},
		Metadata:           map[string]interface{}{"foo": "bar"},
	}
	activity2 := &domain.Activity{
		ProviderID:         s.dummyProvider.ID,
		Resource:           dummyResource,
		ProviderActivityID: "test-provider-activity-id-2",
		AccountID:          "user2@ecample.com",
		Timestamp:          time.Now(),
		Type:               "test-type",
		Authorizations:     []string{"test-authorization"},
		Metadata:           map[string]interface{}{"foo": "bar"},
	}

	s.Run("should insert activities and update the IDs", func() {
		err := s.repository.BulkUpsert(context.Background(), []*domain.Activity{activity, activity2})

		s.NoError(err)
		s.NotEmpty(activity.ID)
		s.Equal(dummyResource.ID, activity.ResourceID)
		s.NotEmpty(activity2.ID)
		s.Equal(dummyResource.ID, activity2.ResourceID)
	})

	s.Run("should return error if provider relation not found", func() {
		activity := &domain.Activity{
			ProviderID: uuid.NewString(),
		}

		err := s.repository.BulkUpsert(context.Background(), []*domain.Activity{activity})

		s.Error(err)
	})
}
