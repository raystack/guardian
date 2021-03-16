package resource_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/resource"
	"github.com/odpf/guardian/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type RepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository *resource.Repository

	columnNames []string
}

func (s *RepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = resource.NewRepository(db)

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

func (s *RepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func (s *RepositoryTestSuite) TestFind() {
	expectedQuery := regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "resources"."deleted_at" IS NULL`)

	s.Run("should return error if db returns error", func() {
		expectedError := errors.New("unexpected error")

		s.dbmock.ExpectQuery(expectedQuery).
			WillReturnError(expectedError)

		actualRecords, actualError := s.repository.Find()

		s.EqualError(actualError, expectedError.Error())
		s.Nil(actualRecords)
	})

	s.Run("should return list of records on success", func() {
		timeNow := time.Now()
		expectedRecords := []*domain.Resource{
			{
				ID:           1,
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
				1,
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

		actualRecords, actualError := s.repository.Find()

		s.Equal(expectedRecords, actualRecords)
		s.Nil(actualError)
	})
}

func (s *RepositoryTestSuite) TestGetOne() {
	s.Run("should return error if id is empty", func() {
		expectedError := resource.ErrEmptyIDParam

		actualResult, actualError := s.repository.GetOne(0)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return nil record and nil error if record not found", func() {
		expectedError := gorm.ErrRecordNotFound
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedError)

		actualResult, actualError := s.repository.GetOne(1)

		s.Nil(actualResult)
		s.Nil(actualError)
	})

	s.Run("should return error if got error from db", func() {
		expectedError := errors.New("unexpected error")
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedError)

		actualResult, actualError := s.repository.GetOne(1)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	expectedQuery := regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "resources"."deleted_at" IS NULL LIMIT 1`)
	s.Run("should return record and nil error on success", func() {
		expectedID := uint(10)
		timeNow := time.Now()
		expectedRows := sqlmock.NewRows(s.columnNames).
			AddRow(
				1,
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
			WillReturnRows(expectedRows)

		_, actualError := s.repository.GetOne(expectedID)

		s.Nil(actualError)
		s.dbmock.ExpectationsWereMet()
	})
}

func (s *RepositoryTestSuite) TestBulkUpsert() {
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

		expectedQuery := regexp.QuoteMeta(`INSERT INTO "resources" ("provider_type","provider_urn","type","urn","name","details","labels","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10),($11,$12,$13,$14,$15,$16,$17,$18,$19,$20) ON CONFLICT ("provider_type","provider_urn","type","urn") DO UPDATE SET "name"="excluded"."name","details"="excluded"."details","labels"="excluded"."labels","updated_at"="excluded"."updated_at" RETURNING "id"`)
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
			)
		}
		expectedIDs := []uint{1, 2}
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
	})
}

func (s *RepositoryTestSuite) TestUpdate() {
	s.Run("should return error if id is empty", func() {
		expectedError := resource.ErrEmptyIDParam

		actualError := s.repository.Update(&domain.Resource{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from transaction", func() {
		expectedError := errors.New("db error")
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(".*").
			WillReturnError(expectedError)
		s.dbmock.ExpectRollback()

		actualError := s.repository.Update(&domain.Resource{ID: 1})

		s.EqualError(actualError, expectedError.Error())
	})

	expectedQuery := regexp.QuoteMeta(`UPDATE "resources" SET "id"=$1,"details"=$2,"labels"=$3,"updated_at"=$4 WHERE id = $5`)
	s.Run("should return error if got error from transaction", func() {
		expectedID := uint(1)
		resource := &domain.Resource{
			ID: expectedID,
		}

		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(expectedQuery).
			WillReturnResult(sqlmock.NewResult(int64(expectedID), 1))
		s.dbmock.ExpectCommit()

		err := s.repository.Update(resource)

		actualID := resource.ID

		s.Nil(err)
		s.Equal(expectedID, actualID)
	})
}

func TestRepository(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
