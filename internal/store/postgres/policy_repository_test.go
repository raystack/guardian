package postgres_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/odpf/guardian/core/policy"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type PolicyRepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository *postgres.PolicyRepository

	rows []string
}

func (s *PolicyRepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = postgres.NewPolicyRepository(db)

	s.rows = []string{
		"id",
		"version",
		"description",
		"steps",
		"appeal",
		"labels",
		"requirements",
		"iam",
		"created_at",
		"updated_at",
	}
}

func (s *PolicyRepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func (s *PolicyRepositoryTestSuite) TestCreate() {
	s.Run("should return error if payload is invalid", func() {
		policy := &domain.Policy{
			IAM: &domain.IAMConfig{
				Config: make(chan int),
			},
		}
		actualError := s.repository.Create(policy)

		s.EqualError(actualError, "serializing policy: json: unsupported type: chan int")
	})

	expectedQuery := regexp.QuoteMeta(`INSERT INTO "policies" ("id","version","description","steps","labels","requirements","iam","created_at","updated_at","deleted_at","appeal") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "appeal"`)

	s.Run("should return error if got error from db transaction", func() {
		p := &domain.Policy{}

		expectedArgs := []driver.Value{
			p.ID,
			p.Version,
			p.Description,
			"null",
			"null",
			"null",
			"null",
			utils.AnyTime{},
			utils.AnyTime{},
			gorm.DeletedAt{},
			"{\"duration_options\":null}",
		}
		expectedError := errors.New("unexpected error")

		s.dbmock.ExpectBegin()
		s.dbmock.
			ExpectQuery(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnError(expectedError)
		s.dbmock.ExpectRollback()
		s.dbmock.MatchExpectationsInOrder(false)

		actualError := s.repository.Create(p)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return nil error on success", func() {
		p := &domain.Policy{}

		expectedArgs := []driver.Value{
			p.ID,
			p.Version,
			p.Description,
			"null",
			"{\"duration_options\":null}",
			"null",
			"null",
			"null",
			utils.AnyTime{},
			utils.AnyTime{},
			gorm.DeletedAt{},
		}
		expectedResult := sqlmock.NewResult(1, 1)

		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnResult(expectedResult)
		s.dbmock.ExpectCommit()

		err := s.repository.Create(p)

		s.Nil(err)
	})
}

func (s *PolicyRepositoryTestSuite) TestFind() {
	expectedQuery := regexp.QuoteMeta(`SELECT * FROM "policies" WHERE "policies"."deleted_at" IS NULL`)

	s.Run("should return error if db returns error", func() {
		expectedError := errors.New("unexpected error")

		s.dbmock.ExpectQuery(expectedQuery).
			WillReturnError(expectedError)

		actualPolicies, actualError := s.repository.Find()

		s.EqualError(actualError, expectedError.Error())
		s.Nil(actualPolicies)
	})

	s.Run("should return list of policies on success", func() {
		now := time.Now()
		expectedPolicies := []*domain.Policy{
			{
				ID:          "",
				Version:     1,
				Description: "",
				Appeal:      domain.PolicyAppealConfig{DurationOptions: nil},
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}
		expectedRows := sqlmock.NewRows(s.rows).
			AddRow(
				"",
				1,
				"",
				"null",
				"{\"duration_options\":null}",
				"null",
				"null",
				"null",
				now,
				now,
			)

		s.dbmock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)

		actualPolicies, actualError := s.repository.Find()

		s.Equal(expectedPolicies, actualPolicies)
		s.Nil(actualError)
	})
}

func (s *PolicyRepositoryTestSuite) TestGetOne() {
	s.Run("should return error if record not found", func() {
		expectedDBError := gorm.ErrRecordNotFound
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedDBError)
		expectedError := policy.ErrPolicyNotFound

		actualResult, actualError := s.repository.GetOne("", 0)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from db", func() {
		expectedError := errors.New("unexpected error")
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedError)

		actualResult, actualError := s.repository.GetOne("", 0)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should pass args based on the version param", func() {
		testCases := []struct {
			name            string
			expectedID      string
			expectedVersion uint
			expectedQuery   string
			expectedArgs    []driver.Value
		}{
			{
				name:            "should not apply version condition if version param given is 0",
				expectedID:      "test-id",
				expectedVersion: 0,
				expectedQuery:   regexp.QuoteMeta(`SELECT * FROM "policies" WHERE id = $1 AND "policies"."deleted_at" IS NULL ORDER BY version desc,"policies"."id" LIMIT 1`),
				expectedArgs:    []driver.Value{"test-id"},
			},
			{
				name:            "should apply version condition if version param is exists",
				expectedID:      "test-id",
				expectedVersion: 1,
				expectedQuery:   regexp.QuoteMeta(`SELECT * FROM "policies" WHERE (id = $1 AND version = $2) AND "policies"."deleted_at" IS NULL ORDER BY version desc,"policies"."id" LIMIT 1`),
				expectedArgs:    []driver.Value{"test-id", 1},
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				now := time.Now()
				expectedRowValues := []driver.Value{
					tc.expectedID,
					tc.expectedVersion,
					"",
					"null",
					"{\"duration_options\":null}",
					"null",
					"null",
					"null",
					now,
					now,
				}
				s.dbmock.ExpectQuery(tc.expectedQuery).
					WithArgs(tc.expectedArgs...).
					WillReturnRows(sqlmock.NewRows(s.rows).AddRow(expectedRowValues...))

				_, actualError := s.repository.GetOne(tc.expectedID, tc.expectedVersion)

				s.Nil(actualError)
				s.dbmock.ExpectationsWereMet()
			})
		}
	})
}

func TestPolicyRepository(t *testing.T) {
	suite.Run(t, new(PolicyRepositoryTestSuite))
}
