package postgres_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/odpf/guardian/core/resource"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ResourceRepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository *postgres.ResourceRepository

	columnNames []string
}

func (s *ResourceRepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = postgres.NewResourceRepository(db)

	s.columnNames = []string{
		"id",
		"provider_type",
		"provider_urn",
		"type",
		"urn",
		"details",
		"labels",
		"created_at",
		"updated_at",
	}
}

func (s *ResourceRepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func (s *ResourceRepositoryTestSuite) TestFind() {
	s.Run("should pass conditions based on filters", func() {
		resourceID1 := uuid.New().String()
		resourceID2 := uuid.New().String()
		testCases := []struct {
			filters       map[string]interface{}
			expectedQuery string
			expectedArgs  []driver.Value
		}{
			{
				filters:       map[string]interface{}{},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "is_deleted" = $1 AND "resources"."deleted_at" IS NULL`),
				expectedArgs:  []driver.Value{false},
			},
			{
				filters: map[string]interface{}{
					"ids": []string{resourceID1, resourceID2},
				},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "resources"."id" IN ($1,$2) AND "is_deleted" = $3 AND "resources"."deleted_at" IS NULL`),
				expectedArgs:  []driver.Value{resourceID1, resourceID2, false},
			},
			{
				filters: map[string]interface{}{
					"type": "test-type",
				},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "is_deleted" = $1 AND "type" = $2 AND "resources"."deleted_at" IS NULL`),
				expectedArgs:  []driver.Value{false, "test-type"},
			},
			{
				filters: map[string]interface{}{
					"name": "test-name",
				},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "is_deleted" = $1 AND "name" = $2 AND "resources"."deleted_at" IS NULL`),
				expectedArgs:  []driver.Value{false, "test-name"},
			},
			{
				filters: map[string]interface{}{
					"provider_type": "test-provider-type",
				},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "is_deleted" = $1 AND "provider_type" = $2 AND "resources"."deleted_at" IS NULL`),
				expectedArgs:  []driver.Value{false, "test-provider-type"},
			},
			{
				filters: map[string]interface{}{
					"provider_urn": "test-provider-urn",
				},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "is_deleted" = $1 AND "provider_urn" = $2 AND "resources"."deleted_at" IS NULL`),
				expectedArgs:  []driver.Value{false, "test-provider-urn"},
			},
			{
				filters: map[string]interface{}{
					"urn": "test-urn",
				},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "is_deleted" = $1 AND "urn" = $2 AND "resources"."deleted_at" IS NULL`),
				expectedArgs:  []driver.Value{false, "test-urn"},
			},
			{
				filters: map[string]interface{}{
					"details": map[string]string{
						"foo": "bar",
					},
				},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "is_deleted" = $1 AND "details" #>> $2 = $3 AND "resources"."deleted_at" IS NULL`),
				expectedArgs:  []driver.Value{false, "{foo}", "bar"},
			},
		}

		for _, tc := range testCases {
			s.dbmock.ExpectQuery(tc.expectedQuery).WithArgs(tc.expectedArgs...).WillReturnRows(sqlmock.NewRows(s.columnNames))

			_, actualError := s.repository.Find(tc.filters)

			s.Nil(actualError)
			s.NoError(s.dbmock.ExpectationsWereMet())
		}
	})

	s.Run("should return error if filters has invalid value", func() {
		invalidFilters := map[string]interface{}{
			"name": make(chan int), // invalid value
		}
		actualRecords, actualError := s.repository.Find(invalidFilters)

		s.Error(actualError)
		s.Nil(actualRecords)
	})

	s.Run("should return error if filters validation returns an error", func() {
		invalidFilters := map[string]interface{}{
			"ids": []string{},
		}
		actualRecords, actualError := s.repository.Find(invalidFilters)

		s.Error(actualError)
		s.Nil(actualRecords)
	})

	expectedQuery := regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "is_deleted" = $1 AND "resources"."deleted_at" IS NULL`)
	expectedArgs := []driver.Value{false}
	s.Run("should return error if db returns error", func() {
		expectedError := errors.New("unexpected error")

		s.dbmock.ExpectQuery(expectedQuery).WithArgs(expectedArgs...).
			WillReturnError(expectedError)

		actualRecords, actualError := s.repository.Find(map[string]interface{}{})

		s.EqualError(actualError, expectedError.Error())
		s.Nil(actualRecords)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return list of records on success", func() {
		timeNow := time.Now()
		resourceID := uuid.New().String()
		expectedRecords := []*domain.Resource{
			{
				ID:           resourceID,
				ProviderType: "provider_type_test",
				ProviderURN:  "provider_urn_test",
				Type:         "type_test",
				URN:          "urn_test",
				CreatedAt:    timeNow,
				UpdatedAt:    timeNow,
			},
		}
		expectedRows := sqlmock.NewRows(s.columnNames).
			AddRow(
				resourceID,
				"provider_type_test",
				"provider_urn_test",
				"type_test",
				"urn_test",
				"null",
				"null",
				timeNow,
				timeNow,
			)

		s.dbmock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)

		actualRecords, actualError := s.repository.Find(map[string]interface{}{})

		s.Equal(expectedRecords, actualRecords)
		s.Nil(actualError)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func (s *ResourceRepositoryTestSuite) TestGetOne() {
	s.Run("should return error if id is empty", func() {
		expectedError := resource.ErrEmptyIDParam

		actualResult, actualError := s.repository.GetOne("")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if record not found", func() {
		expectedDBError := gorm.ErrRecordNotFound
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedDBError)
		expectedError := resource.ErrRecordNotFound

		actualResult, actualError := s.repository.GetOne("1")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return error if got error from db", func() {
		expectedError := errors.New("unexpected error")
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedError)

		actualResult, actualError := s.repository.GetOne("1")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	expectedQuery := regexp.QuoteMeta(`SELECT * FROM "resources" WHERE id = $1 AND "resources"."deleted_at" IS NULL LIMIT 1`)
	s.Run("should return record and nil error on success", func() {
		expectedID := uuid.New().String()
		timeNow := time.Now()
		expectedRows := sqlmock.NewRows(s.columnNames).
			AddRow(
				expectedID,
				"provider_type_test",
				"provider_urn_test",
				"type_test",
				"urn_test",
				"null",
				"null",
				timeNow,
				timeNow,
			)
		s.dbmock.ExpectQuery(expectedQuery).
			WithArgs(expectedID).
			WillReturnRows(expectedRows)

		_, actualError := s.repository.GetOne(expectedID)

		s.Nil(actualError)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func (s *ResourceRepositoryTestSuite) TestBulkUpsert() {
	s.Run("should return records with with existing or new IDs", func() {
		resources := []*domain.Resource{
			{
				ProviderType: "provider_test",
				ProviderURN:  "provider_urn_test",
				Type:         "resource_type",
				URN:          "resource_type.resource_name",
				Name:         "resource_name",
			},
			{
				ProviderType: "provider_test",
				ProviderURN:  "provider_urn_test",
				Type:         "resource_type",
				URN:          "resource_type.resource_name_2",
				Name:         "resource_name_2",
			},
		}

		expectedQuery := regexp.QuoteMeta(`INSERT INTO "resources" ("provider_type","provider_urn","type","urn","name","details","labels","created_at","updated_at","deleted_at","is_deleted") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11),($12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22) ON CONFLICT ("provider_type","provider_urn","type","urn") DO UPDATE SET "name"="excluded"."name","details"="excluded"."details","updated_at"="excluded"."updated_at","is_deleted"="excluded"."is_deleted" RETURNING "id"`)
		expectedArgs := []driver.Value{}
		for _, r := range resources {
			expectedArgs = append(expectedArgs,
				r.ProviderType,
				r.ProviderURN,
				r.Type,
				r.URN,
				r.Name,
				"null",
				"null",
				utils.AnyTime{},
				utils.AnyTime{},
				gorm.DeletedAt{},
				false,
			)
		}
		expectedIDs := []string{
			uuid.New().String(),
			uuid.New().String(),
		}
		expectedRows := sqlmock.NewRows([]string{"id"})
		for _, id := range expectedIDs {
			expectedRows.AddRow(id)
		}
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectQuery(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnRows(expectedRows)
		s.dbmock.ExpectCommit()

		err := s.repository.BulkUpsert(resources)

		s.Nil(err)
		for i, r := range resources {
			s.Equal(expectedIDs[i], r.ID)
		}
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return nil error if resources input is empty", func() {
		var resources []*domain.Resource

		err := s.repository.BulkUpsert(resources)

		s.Nil(err)
	})

	s.Run("should return error if resources is invalid", func() {
		invalidResources := []*domain.Resource{
			{
				Details: map[string]interface{}{
					"foo": make(chan int), // invalid value
				},
			},
		}

		actualError := s.repository.BulkUpsert(invalidResources)

		s.EqualError(actualError, "json: unsupported type: chan int")
	})
}

func (s *ResourceRepositoryTestSuite) TestUpdate() {
	s.Run("should return error if id is empty", func() {
		expectedError := resource.ErrEmptyIDParam

		actualError := s.repository.Update(&domain.Resource{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if resource is invalid", func() {
		invalidResource := &domain.Resource{
			ID: uuid.New().String(),
			Details: map[string]interface{}{
				"foo": make(chan int), // invalid value
			},
		}
		actualError := s.repository.Update(invalidResource)

		s.EqualError(actualError, "json: unsupported type: chan int")
	})

	s.Run("should return error if got error from transaction", func() {
		expectedError := errors.New("db error")
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(".*").
			WillReturnError(expectedError)
		s.dbmock.ExpectRollback()

		actualError := s.repository.Update(&domain.Resource{ID: uuid.New().String()})

		s.EqualError(actualError, expectedError.Error())
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	expectedQuery := regexp.QuoteMeta(`UPDATE "resources" SET "id"=$1,"details"=$2,"labels"=$3,"updated_at"=$4 WHERE id = $5`)
	s.Run("should return error if got error from transaction", func() {
		expectedID := uuid.New().String()
		resource := &domain.Resource{
			ID: expectedID,
		}

		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(expectedQuery).
			WillReturnResult(sqlmock.NewResult(1, 1))
		s.dbmock.ExpectCommit()

		err := s.repository.Update(resource)

		actualID := resource.ID

		s.Nil(err)
		s.Equal(expectedID, actualID)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func (s *ResourceRepositoryTestSuite) TestDelete() {
	s.Run("should return error if ID param is empty", func() {
		err := s.repository.Delete("")

		s.Error(err)
		s.ErrorIs(err, resource.ErrEmptyIDParam)
	})

	s.Run("should return error if db.Delete returns error", func() {
		expectedError := errors.New("test error")
		s.dbmock.ExpectExec(".*").WillReturnError(expectedError)

		err := s.repository.Delete("abc")

		s.Error(err)
		s.ErrorIs(err, expectedError)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return error if resource not found", func() {
		s.dbmock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))

		err := s.repository.Delete("abc")

		s.Error(err)
		s.ErrorIs(err, resource.ErrRecordNotFound)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return nil on success", func() {
		expectedID := "abcd"
		s.dbmock.ExpectExec(regexp.QuoteMeta(`UPDATE "resources" SET "deleted_at"=$1 WHERE id = $2 AND "resources"."deleted_at" IS NULL`)).
			WithArgs(utils.AnyTime{}, expectedID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repository.Delete(expectedID)

		s.Nil(err)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func (s *ResourceRepositoryTestSuite) TestBatchDelete() {
	s.Run("should return error if ID param is empty", func() {
		err := s.repository.BatchDelete(nil)

		s.Error(err)
		s.ErrorIs(err, resource.ErrEmptyIDParam)
	})

	s.Run("should return error if db.Delete returns error", func() {
		expectedError := errors.New("test error")
		s.dbmock.ExpectExec(".*").WillReturnError(expectedError)

		err := s.repository.BatchDelete([]string{"abc"})

		s.Error(err)
		s.ErrorIs(err, expectedError)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return error if resource(s) not found", func() {
		s.dbmock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))

		err := s.repository.BatchDelete([]string{"abc"})

		s.Error(err)
		s.ErrorIs(err, resource.ErrRecordNotFound)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return nil on success", func() {
		expectedIDs := []string{"abcd", "efgh"}
		s.dbmock.ExpectExec(regexp.QuoteMeta(`UPDATE "resources" SET "deleted_at"=$1 WHERE "resources"."id" IN ($2,$3) AND "resources"."deleted_at" IS NULL`)).
			WithArgs(utils.AnyTime{}, expectedIDs[0], expectedIDs[1]).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repository.BatchDelete(expectedIDs)

		s.Nil(err)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func TestResourceRepository(t *testing.T) {
	suite.Run(t, new(ResourceRepositoryTestSuite))
}
