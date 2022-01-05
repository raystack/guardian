package provider_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type RepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository *provider.Repository

	rows []string
}

func (s *RepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = provider.NewRepository(db)

	s.rows = []string{
		"id",
		"type",
		"urn",
		"config",
		"created_at",
		"updated_at",
	}
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func (s *RepositoryTestSuite) TestCreate() {
	expectedQuery := regexp.QuoteMeta(`INSERT INTO "providers" ("type","urn","config","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)

	s.Run("should update model's ID with the returned ID", func() {
		config := &domain.ProviderConfig{}
		provider := &domain.Provider{
			Config: config,
		}

		expectedID := uint(1)
		expectedRows := sqlmock.NewRows([]string{"id"}).
			AddRow(expectedID)
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)
		s.dbmock.ExpectCommit()

		err := s.repository.Create(provider)

		actualID := provider.ID

		s.Nil(err)
		s.Equal(expectedID, actualID)
	})
}

func (s *RepositoryTestSuite) TestFind() {
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
		expectedRecords := []*domain.Provider{
			{
				ID:        1,
				Type:      "type_test",
				URN:       "urn_test",
				Config:    &domain.ProviderConfig{},
				CreatedAt: now,
				UpdatedAt: now,
			},
		}
		expectedRows := sqlmock.NewRows(s.rows).
			AddRow(
				1,
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

func (s *RepositoryTestSuite) TestGetByID() {
	s.Run("should return error if id is empty", func() {
		expectedError := provider.ErrEmptyIDParam

		actualResult, actualError := s.repository.GetByID(0)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if record not found", func() {
		expectedDBError := gorm.ErrRecordNotFound
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedDBError)
		expectedError := provider.ErrRecordNotFound

		actualResult, actualError := s.repository.GetByID(1)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from db", func() {
		expectedError := errors.New("unexpected error")
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedError)

		actualResult, actualError := s.repository.GetByID(1)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	expectedQuery := regexp.QuoteMeta(`SELECT * FROM "providers" WHERE id = $1 AND "providers"."deleted_at" IS NULL ORDER BY "providers"."id" LIMIT 1`)
	s.Run("should return record and nil error on success", func() {
		expectedID := uint(10)
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

func (s *RepositoryTestSuite) TestUpdate() {
	s.Run("should return error if id is empty", func() {
		expectedError := provider.ErrEmptyIDParam

		actualError := s.repository.Update(&domain.Provider{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from transaction", func() {
		expectedError := errors.New("db error")
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(".*").
			WillReturnError(expectedError)
		s.dbmock.ExpectRollback()

		actualError := s.repository.Update(&domain.Provider{ID: 1, Type: "test-type", URN: "test-urn"})

		s.EqualError(actualError, expectedError.Error())
	})

	expectedQuery := regexp.QuoteMeta(`UPDATE "providers" SET "id"=$1,"type"=$2,"urn"=$3,"config"=$4,"updated_at"=$5 WHERE "type" = $6 AND "urn" = $7`)
	s.Run("should return error if got error from transaction", func() {
		config := &domain.ProviderConfig{}
		expectedID := uint(1)
		provider := &domain.Provider{
			ID:     expectedID,
			Type:   "test-type",
			URN:    "test-urn",
			Config: config,
		}

		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(expectedQuery).
			WillReturnResult(sqlmock.NewResult(int64(expectedID), 1))
		s.dbmock.ExpectCommit()

		err := s.repository.Update(provider)

		actualID := provider.ID

		s.Nil(err)
		s.Equal(expectedID, actualID)
	})
}

func TestRepository(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
