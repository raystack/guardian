package policy_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/policy"
	"github.com/odpf/guardian/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type RepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository *policy.Repository
}

func (s *RepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = policy.NewRepository(db)
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func (s *RepositoryTestSuite) TestCreate() {
	expectedQuery := regexp.QuoteMeta(`INSERT INTO "policies" ("id","version","description","steps","labels","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`)

	s.Run("should return error if got error from db transaction", func() {
		p := &domain.Policy{}

		expectedArgs := []driver.Value{
			p.ID,
			p.Version,
			p.Description,
			"null",
			"null",
			utils.AnyTime{},
			utils.AnyTime{},
			gorm.DeletedAt{},
		}
		expectedError := errors.New("unexpected error")

		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnError(expectedError)
		s.dbmock.ExpectRollback()

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

func TestRepository(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
