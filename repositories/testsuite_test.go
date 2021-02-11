package repositories_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mockdb "github.com/odpf/guardian/database/mocks"
	"github.com/odpf/guardian/repositories"
	"github.com/stretchr/testify/suite"
)

type RepositoryTestSuite struct {
	suite.Suite
	sqldb        *sql.DB
	dbmock       sqlmock.Sqlmock
	repositories *repositories.Repositories
}

func (s *RepositoryTestSuite) SetupTest() {
	db, mock, _ := mockdb.New()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repositories = repositories.New(db)
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
