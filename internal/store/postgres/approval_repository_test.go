package postgres_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ApprovalRepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository *postgres.ApprovalRepository

	approvalColumnNames []string
	approverColumnNames []string
	appealColumnNames   []string
	resourceColumnNames []string
}

func (s *ApprovalRepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = postgres.NewApprovalRepository(db)

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
	s.appealColumnNames = []string{
		"id",
		"resource_id",
		"policy_id",
		"policy_version",
		"status",
		"account_id",
		"account_type",
		"created_by",
		"creator",
		"role",
		"options",
		"labels",
		"details",
		"created_at",
		"updated_at",
	}
	s.resourceColumnNames = []string{
		"id",
		"provider_type",
		"provider_urn",
		"type",
		"urn",
		"name",
		"details",
		"labels",
		"created_at",
		"updated_at",
	}
}

func (s *ApprovalRepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func (s *ApprovalRepositoryTestSuite) TestListApprovals() {
	s.Run("should return list of approvals on success", func() {
		timeNow := time.Now()

		appealID := uuid.New().String()
		approvalID := uuid.New().String()
		resourceID := uuid.New().String()

		expectedUser := "user@example.com"
		expectedAccountID := "account@example.com"
		expectedStatuses := []string{"test-status-1", "test-status-2"}
		expectedApprovers := []*domain.Approver{
			{
				ID:         uuid.New().String(),
				ApprovalID: approvalID,
				Email:      expectedUser,
				CreatedAt:  timeNow,
				UpdatedAt:  timeNow,
			},
		}
		expectedApprovals := []*domain.Approval{
			{
				ID:            approvalID,
				Name:          "test-approval-name",
				Index:         0,
				AppealID:      appealID,
				Status:        "test-status",
				PolicyID:      "test-policy-id",
				PolicyVersion: 1,
				Appeal: &domain.Appeal{
					ID:         appealID,
					ResourceID: resourceID,
					Resource: &domain.Resource{
						ID:           resourceID,
						ProviderType: "test-provider-type",
						ProviderURN:  "test-provider-urn",
						Type:         "test-resource-type",
						URN:          "test-resource-urn",
						Name:         "test-name",
						CreatedAt:    timeNow,
						UpdatedAt:    timeNow,
					},
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				},
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			},
		}

		expectedListApproversQuery := regexp.QuoteMeta(`SELECT * FROM "approvers" WHERE email = $1 AND "approvers"."deleted_at" IS NULL`)
		expectedApproverRows := sqlmock.NewRows(s.approverColumnNames)
		for _, a := range expectedApprovers {
			expectedApproverRows.AddRow(a.ID, a.Email, a.AppealID, a.ApprovalID, a.CreatedAt, a.UpdatedAt)
		}
		s.dbmock.ExpectQuery(expectedListApproversQuery).
			WithArgs(expectedUser).
			WillReturnRows(expectedApproverRows)
		expectedListApprovalsQuery := regexp.QuoteMeta(`SELECT "approvals"."id","approvals"."name","approvals"."index","approvals"."appeal_id","approvals"."status","approvals"."actor","approvals"."reason","approvals"."policy_id","approvals"."policy_version","approvals"."created_at","approvals"."updated_at","approvals"."deleted_at","Appeal"."id" AS "Appeal__id","Appeal"."resource_id" AS "Appeal__resource_id","Appeal"."policy_id" AS "Appeal__policy_id","Appeal"."policy_version" AS "Appeal__policy_version","Appeal"."status" AS "Appeal__status","Appeal"."account_id" AS "Appeal__account_id","Appeal"."account_type" AS "Appeal__account_type","Appeal"."created_by" AS "Appeal__created_by","Appeal"."creator" AS "Appeal__creator","Appeal"."role" AS "Appeal__role","Appeal"."options" AS "Appeal__options","Appeal"."labels" AS "Appeal__labels","Appeal"."details" AS "Appeal__details","Appeal"."revoked_by" AS "Appeal__revoked_by","Appeal"."revoked_at" AS "Appeal__revoked_at","Appeal"."revoke_reason" AS "Appeal__revoke_reason","Appeal"."created_at" AS "Appeal__created_at","Appeal"."updated_at" AS "Appeal__updated_at","Appeal"."deleted_at" AS "Appeal__deleted_at" FROM "approvals" LEFT JOIN "appeals" "Appeal" ON "approvals"."appeal_id" = "Appeal"."id" WHERE "approvals"."status" IN ($1,$2) AND "Appeal"."account_id" = $3 AND "Appeal"."status" != $4 AND "approvals"."id" = $5 AND "approvals"."deleted_at" IS NULL ORDER BY ARRAY_POSITION(ARRAY[$6,$7,$8,$9,$10], "approvals"."status"), "updated_at" desc, "created_at"`)
		approvalColumnNames := s.approvalColumnNames
		for _, c := range s.appealColumnNames {
			approvalColumnNames = append(approvalColumnNames, fmt.Sprintf("Appeal__%s", c))
		}
		expectedApprovalRows := sqlmock.NewRows(approvalColumnNames)
		for _, a := range expectedApprovals {
			expectedApprovalRows.AddRow(
				// approval
				a.ID, a.Name, a.AppealID, a.Status, a.PolicyID, a.PolicyVersion, a.CreatedAt, a.UpdatedAt,

				// appeal
				a.Appeal.ID, a.Appeal.ResourceID, a.Appeal.PolicyID, a.Appeal.PolicyVersion, a.Appeal.Status, a.Appeal.AccountID, a.Appeal.AccountType, a.Appeal.CreatedBy, a.Appeal.Creator, a.Appeal.Role, "null", "null", "null", a.CreatedAt, a.UpdatedAt,
			)
		}
		s.dbmock.ExpectQuery(expectedListApprovalsQuery).
			WithArgs(expectedStatuses[0], expectedStatuses[1], expectedAccountID, domain.AppealStatusCanceled, approvalID,
				domain.AppealStatusPending,
				domain.AppealStatusActive,
				domain.AppealStatusRejected,
				domain.AppealStatusTerminated,
				domain.AppealStatusCanceled,
			).
			WillReturnRows(expectedApprovalRows)

		expectedListAppealsQuery := regexp.QuoteMeta(`SELECT * FROM "appeals" WHERE "appeals"."id" = $1 AND "appeals"."deleted_at" IS NULL`)
		expectedAppealRows := sqlmock.NewRows(s.appealColumnNames)
		for _, a := range expectedApprovals {
			expectedAppealRows.AddRow(a.Appeal.ID, a.Appeal.ResourceID, a.Appeal.PolicyID, a.Appeal.PolicyVersion, a.Appeal.Status, a.Appeal.AccountID, a.Appeal.AccountType, a.Appeal.CreatedBy, a.Appeal.Creator, a.Appeal.Role, "null", "null", "null", a.CreatedAt, a.UpdatedAt)
		}
		s.dbmock.ExpectQuery(expectedListAppealsQuery).
			WithArgs(appealID).
			WillReturnRows(expectedAppealRows)

		expectedListResourcesQuery := regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "resources"."id" = $1 AND "resources"."deleted_at" IS NULL`)
		expectedResourceRows := sqlmock.NewRows(s.resourceColumnNames)
		for _, a := range expectedApprovals {
			expectedResourceRows.AddRow(
				a.Appeal.Resource.ID, a.Appeal.Resource.ProviderType, a.Appeal.Resource.ProviderURN, a.Appeal.Resource.Type, a.Appeal.Resource.URN, a.Appeal.Resource.Name, "null", "null", a.Appeal.Resource.CreatedAt, a.Appeal.Resource.UpdatedAt)
		}
		s.dbmock.ExpectQuery(expectedListResourcesQuery).
			WithArgs(resourceID).
			WillReturnRows(expectedResourceRows)

		approvals, err := s.repository.ListApprovals(&domain.ListApprovalsFilter{
			AccountID: expectedAccountID,
			CreatedBy: expectedUser,
			Statuses:  expectedStatuses,
			OrderBy:   []string{"status", "updated_at:desc", "created_at"},
		})

		s.NoError(err)
		s.Equal(expectedApprovals, approvals)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return error if conditions invalid", func() {
		approvals, err := s.repository.ListApprovals(&domain.ListApprovalsFilter{
			AccountID: "",
			CreatedBy: "",
			Statuses:  []string{},
			OrderBy:   []string{},
		})

		s.Error(err)
		s.Nil(approvals)
	})

	s.Run("should return error if db execution returns an error on listing approvers", func() {
		expectedError := errors.New("unexpected error")
		s.dbmock.ExpectQuery(".*").WillReturnError(expectedError)

		approvals, err := s.repository.ListApprovals(&domain.ListApprovalsFilter{})

		s.ErrorIs(err, expectedError)
		s.Nil(approvals)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func (s *ApprovalRepositoryTestSuite) TestBulkInsert() {
	expectedQuery := regexp.QuoteMeta(`INSERT INTO "approvals" ("name","index","appeal_id","status","actor","reason","policy_id","policy_version","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11),($12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22) RETURNING "id"`)

	actor := "user@email.com"
	appealID := uuid.New().String()
	approvals := []*domain.Approval{
		{
			Name:          "approval_step_1",
			Index:         0,
			AppealID:      appealID,
			Status:        domain.ApprovalStatusPending,
			Actor:         &actor,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
		},
		{
			Name:          "approval_step_2",
			Index:         1,
			AppealID:      appealID,
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
			a.Reason,
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

	expectedIDs := []string{
		uuid.New().String(),
		uuid.New().String(),
	}
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

func TestApprovalRepository(t *testing.T) {
	suite.Run(t, new(ApprovalRepositoryTestSuite))
}
