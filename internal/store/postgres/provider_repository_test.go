package postgres_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ProviderRepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository *postgres.ProviderRepository

	rows []string
}

func (s *ProviderRepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = postgres.NewProviderRepository(db)

	s.rows = []string{
		"id",
		"type",
		"urn",
		"config",
		"created_at",
		"updated_at",
	}
}

func (s *ProviderRepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func (s *ProviderRepositoryTestSuite) TestCreate() {
	expectedQuery := regexp.QuoteMeta(`INSERT INTO "providers" ("type","urn","config","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)

	s.Run("should update model's ID with the returned ID", func() {
		config := &domain.ProviderConfig{}
		provider := &domain.Provider{
			Config: config,
		}

		expectedID := uuid.New().String()
		expectedRows := sqlmock.NewRows([]string{"id"}).
			AddRow(expectedID)
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)
		s.dbmock.ExpectCommit()

		err := s.repository.Create(provider)

		actualID := provider.ID

		s.Nil(err)
		s.Equal(expectedID, actualID)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return error if provider is invalid", func() {
		invalidProvider := &domain.Provider{
			Config: &domain.ProviderConfig{
				Credentials: make(chan int), // invalid credentials
			},
		}

		actualError := s.repository.Create(invalidProvider)

		s.EqualError(actualError, "json: unsupported type: chan int")
	})

	s.Run("should return error if db returns an error", func() {
		expectedError := errors.New("unexpected error")
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedError)
		s.dbmock.ExpectRollback()

		actualError := s.repository.Create(&domain.Provider{})

		s.ErrorIs(actualError, expectedError)
	})
}

func (s *ProviderRepositoryTestSuite) TestFind() {
	expectedQuery := regexp.QuoteMeta(`SELECT * FROM "providers" WHERE "providers"."deleted_at" IS NULL`)

	s.Run("should return error if db returns error", func() {
		expectedError := errors.New("unexpected error")

		s.dbmock.ExpectQuery(expectedQuery).
			WillReturnError(expectedError)

		actualRecords, actualError := s.repository.Find()

		s.EqualError(actualError, expectedError.Error())
		s.Nil(actualRecords)
	})

	s.Run("should return list of records on success", func() {
		now := time.Now()
		providerID := uuid.New().String()
		expectedRecords := []*domain.Provider{
			{
				ID:        providerID,
				Type:      "type_test",
				URN:       "urn_test",
				Config:    &domain.ProviderConfig{},
				CreatedAt: now,
				UpdatedAt: now,
			},
		}
		expectedRows := sqlmock.NewRows(s.rows).
			AddRow(
				providerID,
				"type_test",
				"urn_test",
				"null",
				now,
				now,
			)

		s.dbmock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)

		actualRecords, actualError := s.repository.Find()

		s.Equal(expectedRecords, actualRecords)
		s.Nil(actualError)
	})
}

func (s *ProviderRepositoryTestSuite) TestGetByID() {
	s.Run("should return error if id is empty", func() {
		expectedError := provider.ErrEmptyIDParam

		actualResult, actualError := s.repository.GetByID("")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if record not found", func() {
		expectedDBError := gorm.ErrRecordNotFound
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedDBError)
		expectedError := provider.ErrRecordNotFound

		actualResult, actualError := s.repository.GetByID("1")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from db", func() {
		expectedError := errors.New("unexpected error")
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedError)

		actualResult, actualError := s.repository.GetByID("1")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	expectedQuery := regexp.QuoteMeta(`SELECT * FROM "providers" WHERE id = $1 AND "providers"."deleted_at" IS NULL ORDER BY "providers"."id" LIMIT 1`)
	s.Run("should return record and nil error on success", func() {
		expectedID := uuid.New().String()
		timeNow := time.Now()
		expectedRows := sqlmock.NewRows(s.rows).
			AddRow(
				expectedID,
				"type_test",
				"urn_test",
				"null",
				timeNow,
				timeNow,
			)
		s.dbmock.ExpectQuery(expectedQuery).
			WillReturnRows(expectedRows)

		_, actualError := s.repository.GetByID(expectedID)

		s.Nil(actualError)
		s.dbmock.ExpectationsWereMet()
	})
}

func (s *ProviderRepositoryTestSuite) TestGetOne() {
	s.Run("should return provider details on success", func() {
		timeNow := time.Now()

		expectedType := "test-provider-type"
		expectedURN := "test-provider-urn"
		expectedProvider := &domain.Provider{
			ID:   uuid.New().String(),
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
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		expectedConfig, err := json.Marshal(expectedProvider.Config)
		s.Require().NoError(err)

		expectedQuery := regexp.QuoteMeta(`SELECT * FROM "providers" WHERE type = $1 AND urn = $2 AND "providers"."deleted_at" IS NULL LIMIT 1`)
		expectedRows := sqlmock.NewRows(s.rows).AddRow(expectedProvider.ID, expectedProvider.Type, expectedProvider.URN, string(expectedConfig), expectedProvider.CreatedAt, expectedProvider.UpdatedAt)
		s.dbmock.ExpectQuery(expectedQuery).
			WithArgs(expectedType, expectedURN).
			WillReturnRows(expectedRows)

		actualProvider, actualError := s.repository.GetOne(expectedType, expectedURN)

		s.NoError(actualError)
		s.Equal(expectedProvider, actualProvider)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return error if provider type is empty", func() {
		actualProvider, actualError := s.repository.GetOne("", "test-urn")

		s.ErrorIs(actualError, provider.ErrEmptyProviderType)
		s.Nil(actualProvider)
	})

	s.Run("should return error if provider urn is empty", func() {
		actualProvider, actualError := s.repository.GetOne("test-type", "")

		s.ErrorIs(actualError, provider.ErrEmptyProviderURN)
		s.Nil(actualProvider)
	})

	s.Run("should return not found error if record not found", func() {
		s.dbmock.ExpectQuery(".*").WillReturnError(gorm.ErrRecordNotFound)
		actualProvider, actualError := s.repository.GetOne("test-type", "test-urn")

		s.ErrorIs(actualError, provider.ErrRecordNotFound)
		s.Nil(actualProvider)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return error if db returns any error", func() {
		expectedError := errors.New("unexpected error")
		s.dbmock.ExpectQuery(".*").WillReturnError(expectedError)
		actualProvider, actualError := s.repository.GetOne("test-type", "test-urn")

		s.ErrorIs(actualError, expectedError)
		s.Nil(actualProvider)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func (s *ProviderRepositoryTestSuite) TestGetTypes() {
	s.Run("should return error if results empty", func() {
		expectedError := errors.New("no provider types found")

		s.dbmock.ExpectQuery("select distinct provider_type, type as resource_type from resources").WillReturnRows(sqlmock.NewRows([]string{"provider_type", "resource_type"}))

		actualResult, actualError := s.repository.GetTypes()

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return providerTypes and nil error on success", func() {
		expectedResult := []domain.ProviderType{
			{Name: "bigquery", ResourceTypes: []string{"dataset", "table"}},
			{Name: "metabase", ResourceTypes: []string{"group", "collection", "database"}},
		}
		expectedRows := sqlmock.NewRows([]string{"provider_type", "resource_type"}).
			AddRow("bigquery", "dataset").
			AddRow("bigquery", "table").
			AddRow("metabase", "group").
			AddRow("metabase", "collection").
			AddRow("metabase", "database")

		s.dbmock.ExpectQuery("select distinct provider_type, type as resource_type from resources").WillReturnRows(expectedRows)

		actualResult, actualError := s.repository.GetTypes()

		s.ElementsMatch(expectedResult, actualResult)
		s.Nil(actualError)
	})
}

func (s *ProviderRepositoryTestSuite) TestUpdate() {
	s.Run("should return error if id is empty", func() {
		expectedError := provider.ErrEmptyIDParam

		actualError := s.repository.Update(&domain.Provider{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if provider is invalid", func() {
		invalidProvider := &domain.Provider{
			ID: uuid.New().String(),
			Config: &domain.ProviderConfig{
				Credentials: make(chan int), // invalid credentials
			},
		}

		actualError := s.repository.Update(invalidProvider)

		s.EqualError(actualError, "json: unsupported type: chan int")
	})

	s.Run("should return error if got error from transaction", func() {
		expectedError := errors.New("db error")
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(".*").
			WillReturnError(expectedError)
		s.dbmock.ExpectRollback()

		actualError := s.repository.Update(&domain.Provider{ID: uuid.New().String(), Type: "test-type", URN: "test-urn"})

		s.EqualError(actualError, expectedError.Error())
	})

	expectedQuery := regexp.QuoteMeta(`UPDATE "providers" SET "id"=$1,"type"=$2,"urn"=$3,"config"=$4,"updated_at"=$5 WHERE "id" = $6`)
	s.Run("should return error if got error from transaction", func() {
		config := &domain.ProviderConfig{}
		expectedID := uuid.New().String()
		provider := &domain.Provider{
			ID:     expectedID,
			Type:   "test-type",
			URN:    "test-urn",
			Config: config,
		}

		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(expectedQuery).
			WillReturnResult(sqlmock.NewResult(1, 1))
		s.dbmock.ExpectCommit()

		err := s.repository.Update(provider)

		actualID := provider.ID

		s.Nil(err)
		s.Equal(expectedID, actualID)
	})
}

func (s *ProviderRepositoryTestSuite) TestDelete() {
	s.Run("should return error if ID param is empty", func() {
		err := s.repository.Delete("")

		s.Error(err)
		s.ErrorIs(err, provider.ErrEmptyIDParam)
	})

	s.Run("should return error if db.Delete returns error", func() {
		expectedError := errors.New("test error")
		s.dbmock.ExpectExec(".*").WillReturnError(expectedError)

		err := s.repository.Delete("abc")

		s.Error(err)
		s.ErrorIs(err, expectedError)
	})

	s.Run("should return error if resource not found", func() {
		s.dbmock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))

		err := s.repository.Delete("abc")

		s.Error(err)
		s.ErrorIs(err, provider.ErrRecordNotFound)
	})

	s.Run("should return nil on success", func() {
		expectedID := "abcd"
		s.dbmock.ExpectExec(regexp.QuoteMeta(`UPDATE "providers" SET "deleted_at"=$1 WHERE id = $2 AND "providers"."deleted_at" IS NULL`)).
			WithArgs(utils.AnyTime{}, expectedID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repository.Delete(expectedID)

		s.Nil(err)
	})
}

func TestProviderRepository(t *testing.T) {
	suite.Run(t, new(ProviderRepositoryTestSuite))
}
