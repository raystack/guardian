package appeal_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/odpf/guardian/appeal"
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
	repository *appeal.Repository

	columnNames         []string
	approvalColumnNames []string
	approverColumnNames []string
}

func (s *RepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = appeal.NewRepository(db)

	s.columnNames = []string{
		"id",
		"resource_id",
		"policy_id",
		"policy_version",
		"status",
		"user",
		"role",
		"options",
		"labels",
		"created_at",
		"updated_at",
	}
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

func (s *RepositoryTestSuite) TestGetByID() {
	expectedQuery := regexp.QuoteMeta(`SELECT * FROM "appeals" WHERE "appeals"."id" = $1 AND "appeals"."deleted_at" IS NULL ORDER BY "appeals"."id" LIMIT 1`)

	s.Run("should return error if got any from db", func() {
		expectedID := uint(1)
		expectedError := errors.New("db error")
		s.dbmock.
			ExpectQuery(expectedQuery).
			WithArgs(expectedID).
			WillReturnError(expectedError)

		actualResult, actualError := s.repository.GetByID(expectedID)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return nil result and nil error if record not found", func() {
		expectedID := uint(1)
		expectedError := gorm.ErrRecordNotFound
		s.dbmock.
			ExpectQuery(expectedQuery).
			WithArgs(expectedID).
			WillReturnError(expectedError)

		actualResult, actualError := s.repository.GetByID(expectedID)

		s.Nil(actualResult)
		s.Nil(actualError)
	})

	s.Run("should return records on success", func() {
		timeNow := time.Now()
		testCases := []struct {
			expectedID     uint
			expectedRecord *domain.Appeal
		}{
			{
				expectedID: 1,
				expectedRecord: &domain.Appeal{
					ID:            1,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
					Approvals: []*domain.Approval{
						{
							ID:            11,
							Name:          "approval_1",
							AppealID:      1,
							Status:        "pending",
							PolicyID:      "policy_1",
							PolicyVersion: 1,
							CreatedAt:     timeNow,
							UpdatedAt:     timeNow,
						},
						{
							ID:            12,
							Name:          "approval_2",
							AppealID:      1,
							Status:        "pending",
							PolicyID:      "policy_1",
							PolicyVersion: 1,
							CreatedAt:     timeNow,
							UpdatedAt:     timeNow,
						},
					},
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				},
			},
		}

		for _, tc := range testCases {
			expectedRows := sqlmock.NewRows(s.columnNames).
				AddRow(
					tc.expectedRecord.ID,
					tc.expectedRecord.ResourceID,
					tc.expectedRecord.PolicyID,
					tc.expectedRecord.PolicyVersion,
					tc.expectedRecord.Status,
					tc.expectedRecord.User,
					tc.expectedRecord.Role,
					"null",
					"null",
					timeNow,
					timeNow,
				)
			s.dbmock.
				ExpectQuery(expectedQuery).
				WithArgs(tc.expectedID).
				WillReturnRows(expectedRows)

			expectedApprovalsPreloadQuery := regexp.QuoteMeta(`SELECT * FROM "approvals" WHERE "approvals"."appeal_id" = $1 AND "approvals"."deleted_at" IS NULL`)
			expectedApprovalRows := sqlmock.NewRows(s.approvalColumnNames)
			for _, a := range tc.expectedRecord.Approvals {
				expectedApprovalRows.AddRow(
					a.ID,
					a.Name,
					a.AppealID,
					a.Status,
					a.PolicyID,
					a.PolicyVersion,
					timeNow,
					timeNow,
				)
			}
			s.dbmock.
				ExpectQuery(expectedApprovalsPreloadQuery).
				WithArgs(tc.expectedID).
				WillReturnRows(expectedApprovalRows)

			expectedApproversPreloadQuery := regexp.QuoteMeta(`SELECT * FROM "approvers" WHERE "approvers"."approval_id" IN ($1,$2) AND "approvers"."deleted_at" IS NULL`)
			expectedApproverRows := sqlmock.NewRows(s.approverColumnNames)
			for _, a := range tc.expectedRecord.Approvals {
				for _, approver := range a.Approvers {
					expectedApproverRows.AddRow(
						1,
						tc.expectedID,
						a.ID,
						approver,
						timeNow,
						timeNow,
					)
				}
			}
			s.dbmock.
				ExpectQuery(expectedApproversPreloadQuery).
				WithArgs(11, 12).
				WillReturnRows(expectedApproverRows)

			actualRecord, actualError := s.repository.GetByID(tc.expectedID)

			s.Nil(actualError)
			s.Equal(tc.expectedRecord, actualRecord)
		}

	})
}

func (s *RepositoryTestSuite) TestFind() {
	s.Run("should return error if got any from db", func() {
		expectedError := errors.New("db error")
		s.dbmock.
			ExpectQuery(".*").
			WillReturnError(expectedError)

		actualResult, actualError := s.repository.Find(map[string]interface{}{})

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should run query based on filters", func() {
		timeNow := time.Now()
		testCases := []struct {
			filters       map[string]interface{}
			expectedQuery string
			expectedArgs  []driver.Value
		}{
			{
				filters:       map[string]interface{}{},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "appeals" WHERE "appeals"."deleted_at" IS NULL`),
			},
			{
				filters: map[string]interface{}{
					"user": "user@email.com",
				},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "appeals" WHERE "user" = $1 AND "appeals"."deleted_at" IS NULL`),
				expectedArgs:  []driver.Value{"user@email.com"},
			},
			{
				filters: map[string]interface{}{
					"statuses": []string{domain.AppealStatusActive, domain.AppealStatusTerminated},
				},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "appeals" WHERE "status" IN ($1,$2) AND "appeals"."deleted_at" IS NULL`),
				expectedArgs:  []driver.Value{domain.AppealStatusActive, domain.AppealStatusTerminated},
			},
			{
				filters: map[string]interface{}{
					"expiration_date_lt": timeNow,
				},
				expectedQuery: regexp.QuoteMeta(`SELECT * FROM "appeals" WHERE "options" -> 'expiration_date' < $1 AND "appeals"."deleted_at" IS NULL`),
				expectedArgs:  []driver.Value{timeNow},
			},
		}

		for _, tc := range testCases {
			s.dbmock.
				ExpectQuery(tc.expectedQuery).
				WithArgs(tc.expectedArgs...).
				WillReturnRows(sqlmock.NewRows(s.columnNames))

			_, actualError := s.repository.Find(tc.filters)

			s.Nil(actualError)
		}
	})

	s.Run("should return records on success", func() {
		expectedQuery := regexp.QuoteMeta(`SELECT * FROM "appeals" WHERE "appeals"."deleted_at" IS NULL`)
		expectedFilters := map[string]interface{}{}
		expectedRecords := []*domain.Appeal{
			{
				ID:            1,
				ResourceID:    1,
				PolicyID:      "policy_1",
				PolicyVersion: 1,
				Status:        domain.AppealStatusPending,
				User:          "user@email.com",
				Role:          "role_name",
			},
			{
				ID:            2,
				ResourceID:    2,
				PolicyID:      "policy_1",
				PolicyVersion: 1,
				Status:        domain.AppealStatusPending,
				User:          "user@email.com",
				Role:          "role_name",
			},
		}
		expecterRows := sqlmock.NewRows(s.columnNames)
		for _, r := range expectedRecords {
			expecterRows.AddRow(
				r.ID,
				r.ResourceID,
				r.PolicyID,
				r.PolicyVersion,
				r.Status,
				r.User,
				r.Role,
				"null",
				"null",
				r.CreatedAt,
				r.UpdatedAt,
			)
		}
		s.dbmock.
			ExpectQuery(expectedQuery).
			WillReturnRows(expecterRows)

		actualRecords, actualError := s.repository.Find(expectedFilters)

		s.Equal(expectedRecords, actualRecords)
		s.Nil(actualError)
	})
}

func (s *RepositoryTestSuite) TestBulkInsert() {
	expectedQuery := regexp.QuoteMeta(`INSERT INTO "appeals" ("resource_id","policy_id","policy_version","status","user","role","options","labels","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11),($12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22) RETURNING "id"`)

	appeals := []*domain.Appeal{
		{
			User:       "test@email.com",
			Role:       "role_name",
			ResourceID: 1,
		},
		{
			User:       "test2@email.com",
			Role:       "role_name",
			ResourceID: 3,
		},
	}

	expectedError := errors.New("transaction error")
	expectedArgs := []driver.Value{}
	for _, a := range appeals {
		expectedArgs = append(expectedArgs,
			a.ResourceID,
			a.PolicyID,
			a.PolicyVersion,
			a.Status,
			a.User,
			a.Role,
			"null",
			"null",
			utils.AnyTime{},
			utils.AnyTime{},
			gorm.DeletedAt{},
		)
	}

	s.Run("should return error if got any from transaction", func() {
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectQuery(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnError(expectedError)
		s.dbmock.ExpectRollback()

		actualError := s.repository.BulkInsert(appeals)

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

		actualError := s.repository.BulkInsert(appeals)

		s.Nil(actualError)
		for i, a := range appeals {
			s.Equal(expectedIDs[i], a.ID)
		}
	})
}

func (s *RepositoryTestSuite) TestUpdate() {
	s.Run("should return error if got error from transaction", func() {
		expectedError := errors.New("db error")
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(".*").
			WillReturnError(expectedError)
		s.dbmock.ExpectRollback()

		actualError := s.repository.Update(&domain.Appeal{ID: 1})

		s.EqualError(actualError, expectedError.Error())
	})

	expectedUpdateApprovalsQuery := regexp.QuoteMeta(`INSERT INTO "approvals" ("name","index","appeal_id","status","actor","policy_id","policy_version","created_at","updated_at","deleted_at","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11),($12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22) ON CONFLICT ("id") DO UPDATE SET "name"="excluded"."name","index"="excluded"."index","appeal_id"="excluded"."appeal_id","status"="excluded"."status","actor"="excluded"."actor","policy_id"="excluded"."policy_id","policy_version"="excluded"."policy_version","created_at"="excluded"."created_at","updated_at"="excluded"."updated_at","deleted_at"="excluded"."deleted_at" RETURNING "id"`)
	expectedUpdateAppealQuery := regexp.QuoteMeta(`UPDATE "appeals" SET "resource_id"=$1,"policy_id"=$2,"policy_version"=$3,"status"=$4,"user"=$5,"role"=$6,"options"=$7,"labels"=$8,"created_at"=$9,"updated_at"=$10,"deleted_at"=$11 WHERE "id" = $12`)
	s.Run("should return nil on success", func() {
		expectedID := uint(1)
		appeal := &domain.Appeal{
			ID: expectedID,
			Approvals: []*domain.Approval{
				{
					ID:       11,
					AppealID: expectedID,
				},
				{
					ID:       12,
					AppealID: expectedID,
				},
			},
		}

		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(expectedUpdateAppealQuery).
			WillReturnResult(sqlmock.NewResult(int64(expectedID), 1))
		var expectedApprovalArgs []driver.Value
		expectedApprovalRows := sqlmock.NewRows([]string{"id"})
		for _, approval := range appeal.Approvals {
			expectedApprovalArgs = append(expectedApprovalArgs,
				approval.Name,
				approval.Index,
				approval.AppealID,
				approval.Status,
				approval.Actor,
				approval.PolicyID,
				approval.PolicyVersion,
				utils.AnyTime{},
				utils.AnyTime{},
				gorm.DeletedAt{},
				approval.ID,
			)
			expectedApprovalRows.AddRow(
				approval.ID,
			)
		}
		s.dbmock.ExpectQuery(expectedUpdateApprovalsQuery).
			WithArgs(expectedApprovalArgs...).
			WillReturnRows(expectedApprovalRows)
		s.dbmock.ExpectCommit()

		err := s.repository.Update(appeal)

		s.Nil(err)
	})
}

func TestRepository(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
