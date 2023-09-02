package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/raystack/guardian/core/provider"
	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/internal/store/postgres"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/suite"
)

type ProviderRepositoryTestSuite struct {
	suite.Suite
	store              *postgres.Store
	pool               *dockertest.Pool
	resource           *dockertest.Resource
	repository         *postgres.ProviderRepository
	resourceRepository *postgres.ResourceRepository
	providerRepository *postgres.ProviderRepository
}

func (s *ProviderRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewLogrus(log.LogrusWithLevel("debug"))
	s.store, s.pool, s.resource, err = newTestStore(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.repository = postgres.NewProviderRepository(s.store)
	s.resourceRepository = postgres.NewResourceRepository(s.store)
	s.providerRepository = postgres.NewProviderRepository(s.store)
}

func (s *ProviderRepositoryTestSuite) TearDownSuite() {
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

func (s *ProviderRepositoryTestSuite) TestCreate() {
	s.Run("should update model's ID with the returned ID", func() {
		config := &domain.ProviderConfig{
			Parameters: []*domain.ProviderParameter{
				{
					Key:         "username",
					Label:       "Username",
					Required:    true,
					Description: "Please enter your username",
				},
			},
		}
		p := &domain.Provider{
			Config: config,
		}

		err := s.repository.Create(context.Background(), p)
		s.Nil(err)
		s.NotEmpty(p.ID)
	})

	s.Run("should return error if provider is invalid", func() {
		invalidProvider := &domain.Provider{
			Config: &domain.ProviderConfig{
				Credentials: make(chan int), // invalid credentials
			},
		}

		actualError := s.repository.Create(context.Background(), invalidProvider)

		s.EqualError(actualError, "json: unsupported type: chan int")
	})

	s.Run("should return error if db returns an error", func() {
		err := setup(s.store)
		s.NoError(err)

		ctx := context.Background()
		p := &domain.Provider{}
		err1 := s.repository.Create(ctx, p)
		s.Nil(err1)
		s.NotEmpty(p.ID)

		err2 := s.repository.Create(ctx, p)
		s.NotNil(err2)
		s.EqualError(err2, "ERROR: duplicate key value violates unique constraint \"providers_pkey\" (SQLSTATE 23505)")
	})
}

func (s *ProviderRepositoryTestSuite) TestFind() {
	err1 := setup(s.store)
	s.Nil(err1)

	s.Run("should return list of records on success", func() {
		ctx := context.Background()
		expectedRecords := []*domain.Provider{
			{
				Type:   "type_test",
				URN:    "urn_test",
				Config: &domain.ProviderConfig{},
			},
		}
		for _, p := range expectedRecords {
			err := s.repository.Create(ctx, p)
			if err != nil {
				s.Nil(err)
			}
		}

		actualRecords, actualError := s.repository.Find(ctx)

		s.Nil(actualError)
		s.NotEmpty(actualRecords)
		s.Equal(len(expectedRecords), len(actualRecords))
	})
}

func (s *ProviderRepositoryTestSuite) TestGetByID() {
	s.Run("should return error if id is empty", func() {
		expectedError := provider.ErrEmptyIDParam

		actualResult, actualError := s.repository.GetByID(context.Background(), "")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if record not found", func() {
		expectedError := provider.ErrRecordNotFound

		sampleUUID := uuid.New().String()
		actualResult, actualError := s.repository.GetByID(context.Background(), sampleUUID)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return record and nil error on success", func() {
		err := setup(s.store)
		s.Nil(err)

		p := &domain.Provider{
			Config: &domain.ProviderConfig{
				Parameters: []*domain.ProviderParameter{
					{
						Key:         "username",
						Label:       "Username",
						Required:    true,
						Description: "Please enter your username",
					},
				},
			},
		}

		err = s.repository.Create(context.Background(), p)
		s.Nil(err)
		s.NotEmpty(p.ID)

		actual, actualError := s.repository.GetByID(context.Background(), p.ID)

		s.Nil(actualError)
		if diff := cmp.Diff(p, actual, cmpopts.EquateApproxTime(time.Microsecond)); diff != "" {
			s.T().Errorf("result not match, diff: %v", diff)
		}
	})
}

func (s *ProviderRepositoryTestSuite) TestGetOne() {
	s.Run("should return provider details on success", func() {
		expectedType := "test-provider-type"
		expectedURN := "test-provider-urn"
		expectedProvider := &domain.Provider{
			Type: expectedType,
			URN:  expectedURN,
			Config: &domain.ProviderConfig{
				Type:                expectedType,
				URN:                 expectedURN,
				AllowedAccountTypes: []string{"test-account-type"},
				Credentials: map[string]interface{}{
					"foo": "bar",
				},
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess:         true,
					AllowActiveAccessExtensionIn: "24h",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: "test-resource-type",
						Policy: &domain.PolicyConfig{
							ID:      "test-policy-id",
							Version: 1,
						},
						Roles: []*domain.Role{
							{
								ID:          "test-id",
								Name:        "test-name",
								Description: "test-description",
								Permissions: []interface{}{"test-permission"},
							},
						},
					},
				},
			},
		}

		ctx := context.Background()
		err := s.repository.Create(ctx, expectedProvider)
		s.Nil(err)
		s.NotEmpty(expectedProvider.ID)

		actualProvider, actualError := s.repository.GetOne(ctx, expectedType, expectedURN)

		s.NoError(actualError)
		s.Equal(expectedProvider.Config, actualProvider.Config)
	})

	s.Run("should return error if provider type is empty", func() {
		actualProvider, actualError := s.repository.GetOne(context.Background(), "", "test-urn")

		s.ErrorIs(actualError, provider.ErrEmptyProviderType)
		s.Nil(actualProvider)
	})

	s.Run("should return error if provider urn is empty", func() {
		actualProvider, actualError := s.repository.GetOne(context.Background(), "test-type", "")

		s.ErrorIs(actualError, provider.ErrEmptyProviderURN)
		s.Nil(actualProvider)
	})

	s.Run("should return not found error if record not found", func() {
		actualProvider, actualError := s.repository.GetOne(context.Background(), "test-type", "test-urn")

		s.ErrorIs(actualError, provider.ErrRecordNotFound)
		s.Nil(actualProvider)
	})
}

func (s *ProviderRepositoryTestSuite) TestGetTypes() {
	s.Run("should return error if results empty", func() {
		expectedError := errors.New("no provider types found")

		actualResult, actualError := s.repository.GetTypes(context.Background())

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return providerTypes and nil error on success", func() {
		expectedProviderTypes := map[string][]string{
			"metabase": {"group", "collection", "database"},
			"bigquery": {"dataset", "table"},
		}

		ctx := context.Background()
		err := s.providerRepository.Create(ctx, &domain.Provider{
			Type: "bigquery",
			URN:  "my-bigquery",
		})
		s.Require().NoError(err)
		err = s.providerRepository.Create(ctx, &domain.Provider{
			Type: "metabase",
			URN:  "my-metabase",
		})
		s.Require().NoError(err)

		err = s.resourceRepository.BulkUpsert(context.Background(), []*domain.Resource{
			{ProviderType: "bigquery", ProviderURN: "my-bigquery", Type: "dataset"},
			{ProviderType: "bigquery", ProviderURN: "my-bigquery", Type: "table"},
			{ProviderType: "metabase", ProviderURN: "my-metabase", Type: "group"},
			{ProviderType: "metabase", ProviderURN: "my-metabase", Type: "collection"},
			{ProviderType: "metabase", ProviderURN: "my-metabase", Type: "database", URN: "db1"},
			{ProviderType: "metabase", ProviderURN: "my-metabase", Type: "database", URN: "db2"},
		})
		s.Require().NoError(err)

		actualResult, actualError := s.repository.GetTypes(ctx)

		for _, pt := range actualResult {
			s.ElementsMatch(expectedProviderTypes[pt.Name], pt.ResourceTypes)
		}
		s.Nil(actualError)
	})
}

func (s *ProviderRepositoryTestSuite) TestUpdate() {
	s.Run("should return error if id is empty", func() {
		expectedError := provider.ErrEmptyIDParam

		actualError := s.repository.Update(context.Background(), &domain.Provider{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if provider is invalid", func() {
		invalidProvider := &domain.Provider{
			ID: uuid.New().String(),
			Config: &domain.ProviderConfig{
				Credentials: make(chan int), // invalid credentials
			},
		}

		actualError := s.repository.Update(context.Background(), invalidProvider)

		s.EqualError(actualError, "json: unsupported type: chan int")
	})

	s.Run("should return nil error on successful update", func() {
		expectedID := uuid.New().String()
		p := &domain.Provider{
			ID:     expectedID,
			Type:   "test-type",
			URN:    "test-urn",
			Config: &domain.ProviderConfig{},
		}

		err := s.repository.Update(context.Background(), p)
		actualID := p.ID

		s.Nil(err)
		s.Equal(expectedID, actualID)
	})
}

func (s *ProviderRepositoryTestSuite) TestDelete() {
	err1 := setup(s.store)
	s.Nil(err1)

	s.Run("should return error if ID param is empty", func() {
		err := s.repository.Delete(context.Background(), "")

		s.Error(err)
		s.ErrorIs(err, provider.ErrEmptyIDParam)
	})

	s.Run("should return error if resource not found", func() {
		id := uuid.New().String()
		err := s.repository.Delete(context.Background(), id)

		s.Error(err)
		s.ErrorIs(err, provider.ErrRecordNotFound)
	})

	s.Run("should return nil on success", func() {
		p := &domain.Provider{
			Config: &domain.ProviderConfig{},
		}

		ctx := context.Background()
		err := s.repository.Create(ctx, p)
		s.Nil(err)
		s.NotEmpty(p.ID)

		err = s.repository.Delete(ctx, p.ID)
		s.Nil(err)
	})
}

func TestProviderRepository(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, new(ProviderRepositoryTestSuite))
}
