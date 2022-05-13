package repositories_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/odpf/guardian/pkg/audit"
	"github.com/odpf/guardian/pkg/audit/repositories"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresRepositoryTestSuite struct {
	suite.Suite

	dbMock     sqlmock.Sqlmock
	dbConn     *sql.DB
	repository *repositories.PostgresRepository
}

func TestPostgresRepository(t *testing.T) {
	suite.Run(t, new(PostgresRepositoryTestSuite))
}

func (s *PostgresRepositoryTestSuite) setupTest() {
	db, dbMock, err := sqlmock.New()
	s.Require().NoError(err)
	s.dbConn = db
	s.dbMock = dbMock
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	s.Require().NoError(err)

	s.repository = repositories.NewPostgresRepository(gormDB)
}

func (s *PostgresRepositoryTestSuite) cleanupTest() {
	s.dbConn.Close()
}

func (s *PostgresRepositoryTestSuite) TestInit() {
	s.Run("should migrate audit log model", func() {
		s.setupTest()
		defer s.cleanupTest()

		s.dbMock.ExpectExec(regexp.QuoteMeta(`CREATE TABLE "audit_logs" ("timestamp" timestamptz,"action" text,"actor" text,"data" JSONB,"metadata" JSONB,"app" JSONB)`)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repository.Init(context.Background())
		s.NoError(err)
		s.dbMock.ExpectationsWereMet()
	})

	s.Run("should return error if migrate returns error", func() {
		s.setupTest()
		defer s.cleanupTest()

		expectedError := errors.New("test error")
		s.dbMock.ExpectExec(".*").WillReturnError(expectedError)

		err := s.repository.Init(context.Background())
		s.ErrorIs(err, expectedError)
		s.dbMock.ExpectationsWereMet()
	})
}

func (s *PostgresRepositoryTestSuite) TestInsert() {
	s.Run("should insert record to db", func() {
		s.setupTest()
		defer s.cleanupTest()

		l := &audit.Log{}

		s.dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "audit_logs" ("timestamp","action","actor","data","metadata","app") VALUES ($1,$2,$3,$4,$5,$6)`)).
			WithArgs(l.Timestamp, l.Action, l.Actor, `null`, `null`, `null`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repository.Insert(context.Background(), l)
		s.NoError(err)
		s.dbMock.ExpectationsWereMet()
	})

	s.Run("should return error if data marshaling returns error", func() {
		s.setupTest()
		defer s.cleanupTest()

		l := &audit.Log{
			Data: make(chan int),
		}

		err := s.repository.Insert(context.Background(), l)
		s.EqualError(err, "marshaling data: json: unsupported type: chan int")
	})

	s.Run("should return error if metadata marshaling returns error", func() {
		s.setupTest()
		defer s.cleanupTest()

		l := &audit.Log{
			Metadata: map[string]interface{}{
				"foo": make(chan int),
			},
		}

		err := s.repository.Insert(context.Background(), l)
		s.EqualError(err, "marshaling metadata: json: unsupported type: chan int")
	})

	s.Run("should return error if db insert returns error", func() {
		s.setupTest()
		defer s.cleanupTest()

		l := &audit.Log{}

		expectedError := errors.New("test error")
		s.dbMock.ExpectExec(".*").WillReturnError(expectedError)

		err := s.repository.Insert(context.Background(), l)
		s.ErrorIs(err, expectedError)
		s.dbMock.ExpectationsWereMet()
	})
}
