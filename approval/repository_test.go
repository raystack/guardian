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

	approvalColumnNames []string
	approverColumnNames []string
}

func (s *RepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = approval.NewRepository(db)

	s.approvalColumnNames = []string{
		"id",
		"name",
		"appeal_id",
		"status",
		"policy_id",
		"policy_version",
		"created_at",
		"updated_at",
	}
	s.approverColumnNames = []string{
		"id",
		"email",
		"appeal_id",
		"approval_id",
		"created_at",
		"updated_at",
	}
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func (s *RepositoryTestSuite) TestGetPendingApprovals() {
	s.Run("return error if got any when getting approver list by email", func() {
		expectedError := errors.New("db error")
		email := "user@email.com"
		s.dbmock.ExpectQuery(".*").
			WithArgs(email).
			WillReturnError(expectedError)

		actualRecords, actualError := s.repository.GetPendingApprovals(email)

		s.Nil(actualRecords)
		s.EqualError(actualError, expectedError.Error())
	})

	approverEmail := "user@email.com"
	expectedApproversQuery := regexp.QuoteMeta(`SELECT * FROM "approvers" WHERE email = $1 AND "approvers"."deleted_at" IS NULL`)

	s.Run("return error if got any when getting user's pending approvals", func() {
		expectedApprovers := []*domain.Approver{
			{
				ID:    1,
				Email: approverEmail,
			},
		}
		expectedUserApproverRows := sqlmock.NewRows(s.approverColumnNames)
		for _, a := range expectedApprovers {
			expectedUserApproverRows.AddRow(
				a.ID,
				a.Email,
				a.AppealID,
				a.ApprovalID,
				a.CreatedAt,
				a.UpdatedAt,
			)
		}
		s.dbmock.ExpectQuery(expectedApproversQuery).
			WithArgs(approverEmail).
			WillReturnRows(expectedUserApproverRows)

		expectedError := errors.New("db error")
		s.dbmock.ExpectQuery(".*").
			WillReturnError(expectedError)

		actualRecords, actualError := s.repository.GetPendingApprovals(approverEmail)

		s.Nil(actualRecords)
		s.EqualError(actualError, expectedError.Error())
	})

	expectedApprovalsQuery := regexp.QuoteMeta(`SELECT * FROM "approvals" WHERE ("appeal_id","index") IN (SELECT appeal_id, min("index") FROM "approvals" WHERE status = $1 AND "approvals"."deleted_at" IS NULL GROUP BY "appeal_id") AND "approvals"."id" IN ($2,$3) AND "approvals"."deleted_at" IS NULL`)

	s.Run("return records on success", func() {
		expectedApprovers := []*domain.Approver{
			{
				ID:         1,
				Email:      approverEmail,
				ApprovalID: 11,
			},
			{
				ID:         2,
				Email:      approverEmail,
				ApprovalID: 12,
			},
		}
		expectedUserApproverRows := sqlmock.NewRows(s.approverColumnNames)
		for _, a := range expectedApprovers {
			expectedUserApproverRows.AddRow(
				a.ID,
				a.Email,
				a.AppealID,
				a.ApprovalID,
				a.CreatedAt,
				a.UpdatedAt,
			)
		}
		s.dbmock.ExpectQuery(expectedApproversQuery).
			WithArgs(approverEmail).
			WillReturnRows(expectedUserApproverRows)

		expectedApprovals := []*domain.Approval{
			{
				ID:     11,
				Status: domain.ApprovalStatusPending,
			},
			{
				ID:     12,
				Status: domain.ApprovalStatusPending,
			},
		}
		expectedUserApprovalRows := sqlmock.NewRows(s.approvalColumnNames)
		for _, a := range expectedApprovals {
			expectedUserApprovalRows.AddRow(
				a.ID,
				a.Name,
				a.AppealID,
				a.Status,
				a.PolicyID,
				a.PolicyVersion,
				a.CreatedAt,
				a.UpdatedAt,
			)
		}

		expectedArgs := []driver.Value{domain.ApprovalStatusPending, 11, 12}
		s.dbmock.ExpectQuery(expectedApprovalsQuery).
			WithArgs(expectedArgs...).
			WillReturnRows(expectedUserApprovalRows)

		actualRecords, actualError := s.repository.GetPendingApprovals(approverEmail)

		s.Equal(expectedApprovals, actualRecords)
		s.Nil(actualError)
	})
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
