package provider_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/provider"
	"github.com/stretchr/testify/suite"
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

func TestRepository(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
