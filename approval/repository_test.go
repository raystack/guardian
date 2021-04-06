package approval_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/odpf/guardian/approval"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type RepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository domain.ApprovalRepository
}

func (s *RepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = approval.NewRepository(db)
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func (s *RepositoryTestSuite) TestBulkInsert() {
	expectedQuery := regexp.QuoteMeta(`INSERT INTO "approvals" ("name","index","appeal_id","status","actor","policy_id","policy_version","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10),($11,$12,$13,$14,$15,$16,$17,$18,$19,$20) RETURNING "id"`)

	actor := "user@email.com"
	approvals := []*domain.Approval{
		{
			Name:          "approval_step_1",
			Index:         0,
			AppealID:      1,
			Status:        domain.ApprovalStatusPending,
			Actor:         &actor,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
		},
		{
			Name:          "approval_step_2",
			Index:         1,
			AppealID:      1,
			Status:        domain.ApprovalStatusPending,
			Actor:         &actor,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
		},
	}

	expectedArgs := []driver.Value{}
	for _, a := range approvals {
		expectedArgs = append(expectedArgs,
			a.Name,
			a.Index,
			a.AppealID,
			a.Status,
			a.Actor,
			a.PolicyID,
			a.PolicyVersion,
			utils.AnyTime{},
			utils.AnyTime{},
			gorm.DeletedAt{},
		)
	}

	s.Run("should return error if got any from transaction", func() {
		expectedError := errors.New("transaction error")
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectQuery(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnError(expectedError)
		s.dbmock.ExpectRollback()

		actualError := s.repository.BulkInsert(approvals)

		s.EqualError(actualError, expectedError.Error())
	})

	expectedIDs := []uint{1, 2}
	expectedRows := sqlmock.NewRows([]string{"id"})
	for _, id := range expectedIDs {
		expectedRows.AddRow(id)
	}

	s.Run("should return nil error on success", func() {
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectQuery(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnRows(expectedRows)
		s.dbmock.ExpectCommit()

		actualError := s.repository.BulkInsert(approvals)

		s.Nil(actualError)
		for i, a := range approvals {
			s.Equal(expectedIDs[i], a.ID)
		}
	})
}

func TestRepository(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
