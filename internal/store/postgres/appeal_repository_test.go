package postgres_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type AppealRepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository *postgres.AppealRepository

	columnNames         []string
	approvalColumnNames []string
	approverColumnNames []string
	resourceColumnNames []string
}

func (s *AppealRepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = postgres.NewAppealRepository(db)

	s.columnNames = []string{
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
		"permissions",
		"options",
		"labels",
		"details",
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
	s.resourceColumnNames = []string{
		"id",
		"provider_type",
		"provider_urn",
		"type",
		"urn",
		"details",
		"labels",
		"created_at",
		"updated_at",
	}
}

func (s *AppealRepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func (s *AppealRepositoryTestSuite) TestGetByID() {
	expectedQuery := regexp.QuoteMeta(`SELECT * FROM "appeals" WHERE id = $1 AND "appeals"."deleted_at" IS NULL ORDER BY "appeals"."id" LIMIT 1`)

	s.Run("should return error if got any from db", func() {
		expectedID := uuid.New().String()
		expectedError := errors.New("db error")
		s.dbmock.
			ExpectQuery(expectedQuery).
			WithArgs(expectedID).
			WillReturnError(expectedError)

		actualResult, actualError := s.repository.GetByID(expectedID)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return error if record not found", func() {
		expectedID := uuid.New().String()
		expectedDBError := gorm.ErrRecordNotFound
		s.dbmock.
			ExpectQuery(expectedQuery).
			WithArgs(expectedID).
			WillReturnError(expectedDBError)
		expectedError := appeal.ErrAppealNotFound

		actualResult, actualError := s.repository.GetByID(expectedID)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return error if returned record is invalid", func() {
		expectedID := uuid.New().String()
		expectedRows := sqlmock.NewRows(s.columnNames).
			AddRow(
				"",
				"",
				"",
				0,
				"",
				"",
				"",
				"",
				"null",
				"",
				"{}",
				`{"duration":999}`, // invalid options
				"null",
				"null",
				nil,
				nil,
			)
		s.dbmock.
			ExpectQuery(expectedQuery).
			WithArgs(expectedID).
			WillReturnRows(expectedRows)

		actualResult, actualError := s.repository.GetByID(expectedID)

		s.Nil(actualResult)
		s.EqualError(actualError, "parsing appeal: json: cannot unmarshal number into Go struct field AppealOptions.duration of type string")
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return records on success", func() {
		timeNow := time.Now()
		expectedID := uuid.New().String()
		approvalID1 := uuid.New().String()
		approvalID2 := uuid.New().String()
		testCases := []struct {
			expectedID     string
			expectedRecord *domain.Appeal
		}{
			{
				expectedID: expectedID,
				expectedRecord: &domain.Appeal{
					ID:            expectedID,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
					Approvals: []*domain.Approval{
						{
							ID:            approvalID1,
							Name:          "approval_1",
							AppealID:      expectedID,
							Status:        "pending",
							PolicyID:      "policy_1",
							PolicyVersion: 1,
							CreatedAt:     timeNow,
							UpdatedAt:     timeNow,
						},
						{
							ID:            approvalID2,
							Name:          "approval_2",
							AppealID:      expectedID,
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
					tc.expectedRecord.AccountID,
					tc.expectedRecord.AccountType,
					tc.expectedRecord.CreatedBy,
					"null",
					tc.expectedRecord.Role,
					pq.StringArray(tc.expectedRecord.Permissions),
					"null",
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
				WithArgs(approvalID1, approvalID2).
				WillReturnRows(expectedApproverRows)

			actualRecord, actualError := s.repository.GetByID(tc.expectedID)

			s.Nil(actualError)
			s.Equal(tc.expectedRecord, actualRecord)
			s.NoError(s.dbmock.ExpectationsWereMet())
		}
	})
}

func (s *AppealRepositoryTestSuite) TestFind() {
	s.Run("should return error if filters validation returns an error", func() {
		invalidFilters := &domain.ListAppealsFilter{
			Statuses: []string{},
		}

		actualAppeals, actualError := s.repository.Find(invalidFilters)

		s.Error(actualError)
		s.Nil(actualAppeals)
	})

	s.Run("should return error if got any from db", func() {
		expectedError := errors.New("db error")
		s.dbmock.
			ExpectQuery(".*").
			WillReturnError(expectedError)

		actualResult, actualError := s.repository.Find(&domain.ListAppealsFilter{})

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should run query based on filters", func() {
		selectAppealsJoinsWithResourceSql := `SELECT "appeals"."id","appeals"."resource_id","appeals"."policy_id","appeals"."policy_version","appeals"."status","appeals"."account_id","appeals"."account_type","appeals"."created_by","appeals"."creator","appeals"."role","appeals"."permissions","appeals"."options","appeals"."labels","appeals"."details","appeals"."revoked_by","appeals"."revoked_at","appeals"."revoke_reason","appeals"."created_at","appeals"."updated_at","appeals"."deleted_at","Resource"."id" AS "Resource__id","Resource"."provider_type" AS "Resource__provider_type","Resource"."provider_urn" AS "Resource__provider_urn","Resource"."type" AS "Resource__type","Resource"."urn" AS "Resource__urn","Resource"."name" AS "Resource__name","Resource"."details" AS "Resource__details","Resource"."labels" AS "Resource__labels","Resource"."created_at" AS "Resource__created_at","Resource"."updated_at" AS "Resource__updated_at","Resource"."deleted_at" AS "Resource__deleted_at","Resource"."is_deleted" AS "Resource__is_deleted" FROM "appeals" LEFT JOIN "resources" "Resource" ON "appeals"."resource_id" = "Resource"."id" WHERE`
		timeNow := time.Now()
		testCases := []struct {
			filters             *domain.ListAppealsFilter
			expectedClauseQuery string
			expectedArgs        []driver.Value
		}{
			{
				filters:             &domain.ListAppealsFilter{},
				expectedClauseQuery: `"appeals"."deleted_at" IS NULL`,
			},
			{
				filters: &domain.ListAppealsFilter{
					CreatedBy: "user@email.com",
				},
				expectedClauseQuery: `"created_by" = $1 AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{"user@email.com"},
			},
			{
				filters: &domain.ListAppealsFilter{
					AccountIDs: []string{"user@email.com"},
				},
				expectedClauseQuery: `"account_id" IN ($1) AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{"user@email.com"},
			},
			{
				filters: &domain.ListAppealsFilter{
					AccountID: "user@email.com",
				},
				expectedClauseQuery: `"account_id" IN ($1) AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{"user@email.com"},
			},
			{
				filters: &domain.ListAppealsFilter{
					Statuses: []string{domain.AppealStatusActive, domain.AppealStatusTerminated},
				},
				expectedClauseQuery: `"status" IN ($1,$2) AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{domain.AppealStatusActive, domain.AppealStatusTerminated},
			},
			{
				filters: &domain.ListAppealsFilter{
					ResourceID: "1",
				},
				expectedClauseQuery: `"resource_id" = $1 AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{"1"},
			},
			{
				filters: &domain.ListAppealsFilter{
					Role: "test-role",
				},
				expectedClauseQuery: `"role" = $1 AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{"test-role"},
			},
			{
				filters: &domain.ListAppealsFilter{
					ExpirationDateLessThan: timeNow,
				},
				expectedClauseQuery: `"options" -> 'expiration_date' < $1 AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{timeNow},
			},
			{
				filters: &domain.ListAppealsFilter{
					ExpirationDateGreaterThan: timeNow,
				},
				expectedClauseQuery: `"options" -> 'expiration_date' > $1 AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{timeNow},
			},
			{
				filters: &domain.ListAppealsFilter{
					ProviderTypes: []string{"test-provider-type"},
				},
				expectedClauseQuery: `"Resource"."provider_type" IN ($1) AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{"test-provider-type"},
			},
			{
				filters: &domain.ListAppealsFilter{
					ProviderURNs: []string{"test-provider-urn"},
				},
				expectedClauseQuery: `"Resource"."provider_urn" IN ($1) AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{"test-provider-urn"},
			},
			{
				filters: &domain.ListAppealsFilter{
					ResourceTypes: []string{"test-resource-type"},
				},
				expectedClauseQuery: `"Resource"."type" IN ($1) AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{"test-resource-type"},
			},
			{
				filters: &domain.ListAppealsFilter{
					ResourceURNs: []string{"test-resource-urn"},
				},
				expectedClauseQuery: `"Resource"."urn" IN ($1) AND "appeals"."deleted_at" IS NULL`,
				expectedArgs:        []driver.Value{"test-resource-urn"},
			},
			{
				filters: &domain.ListAppealsFilter{
					OrderBy: []string{"status"},
				},
				expectedClauseQuery: `"appeals"."deleted_at" IS NULL ORDER BY ARRAY_POSITION(ARRAY[$1,$2,$3,$4,$5], "status")`,
				expectedArgs: []driver.Value{
					postgres.AppealStatusDefaultSort[0],
					postgres.AppealStatusDefaultSort[1],
					postgres.AppealStatusDefaultSort[2],
					postgres.AppealStatusDefaultSort[3],
					postgres.AppealStatusDefaultSort[4],
				},
			},
			{
				filters: &domain.ListAppealsFilter{
					OrderBy: []string{"updated_at:desc"},
				},
				expectedClauseQuery: `"appeals"."deleted_at" IS NULL ORDER BY "updated_at" desc`,
			},
		}

		for _, tc := range testCases {
			expectedQuery := regexp.QuoteMeta(strings.Join([]string{selectAppealsJoinsWithResourceSql, tc.expectedClauseQuery}, " "))
			s.dbmock.
				ExpectQuery(expectedQuery).
				WithArgs(tc.expectedArgs...).
				WillReturnRows(sqlmock.NewRows(s.columnNames))

			_, actualError := s.repository.Find(tc.filters)

			s.Nil(actualError)
			s.NoError(s.dbmock.ExpectationsWereMet())
		}
	})

	s.Run("should return error if returned appeal is invalid", func() {
		aggregatedColumns := s.columnNames
		for _, c := range s.resourceColumnNames {
			aggregatedColumns = append(aggregatedColumns, fmt.Sprintf("Resource__%s", c))
		}
		expectedRows := sqlmock.NewRows(aggregatedColumns).AddRow(
			// appeal
			uuid.New().String(),
			"test-resource-id",
			"test-policy-id",
			1,
			"test-status",
			"test-account-id",
			"test-account-type",
			"test-created-by",
			"null",
			"test-role",
			"{}",
			`{"duration":999}`, // invalid options
			"null",
			"null",
			time.Now(),
			time.Now(),

			// resource
			uuid.New().String(),
			"test-provider-type",
			"test-provider-urn",
			"test-resource-type",
			"test-resource-urn",
			"null",
			"null",
			time.Now(),
			time.Now(),
		)
		s.dbmock.ExpectQuery(".*").WillReturnRows(expectedRows)

		actualRecords, actualError := s.repository.Find(&domain.ListAppealsFilter{})

		s.Nil(actualRecords)
		s.EqualError(actualError, "parsing appeal: json: cannot unmarshal number into Go struct field AppealOptions.duration of type string")
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return records on success", func() {
		expectedQuery := regexp.QuoteMeta(`SELECT "appeals"."id","appeals"."resource_id","appeals"."policy_id","appeals"."policy_version","appeals"."status","appeals"."account_id","appeals"."account_type","appeals"."created_by","appeals"."creator","appeals"."role","appeals"."permissions","appeals"."options","appeals"."labels","appeals"."details","appeals"."revoked_by","appeals"."revoked_at","appeals"."revoke_reason","appeals"."created_at","appeals"."updated_at","appeals"."deleted_at","Resource"."id" AS "Resource__id","Resource"."provider_type" AS "Resource__provider_type","Resource"."provider_urn" AS "Resource__provider_urn","Resource"."type" AS "Resource__type","Resource"."urn" AS "Resource__urn","Resource"."name" AS "Resource__name","Resource"."details" AS "Resource__details","Resource"."labels" AS "Resource__labels","Resource"."created_at" AS "Resource__created_at","Resource"."updated_at" AS "Resource__updated_at","Resource"."deleted_at" AS "Resource__deleted_at","Resource"."is_deleted" AS "Resource__is_deleted" FROM "appeals" LEFT JOIN "resources" "Resource" ON "appeals"."resource_id" = "Resource"."id" WHERE "appeals"."deleted_at" IS NULL`)
		resourceID1 := uuid.New().String()
		resourceID2 := uuid.New().String()
		expectedRecords := []*domain.Appeal{
			{
				ID:         uuid.New().String(),
				ResourceID: resourceID1,
				Resource: &domain.Resource{
					ID:           resourceID1,
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
					URN:          "resource_urn",
				},
				PolicyID:      "policy_1",
				PolicyVersion: 1,
				Status:        domain.AppealStatusPending,
				AccountID:     "user@email.com",
				Role:          "role_name",
			},
			{
				ID:         uuid.New().String(),
				ResourceID: resourceID2,
				Resource: &domain.Resource{
					ID:           resourceID2,
					ProviderType: "provider_type",
					ProviderURN:  "provider_urn",
					Type:         "resource_type",
					URN:          "resource_urn",
				},
				PolicyID:      "policy_1",
				PolicyVersion: 1,
				Status:        domain.AppealStatusPending,
				AccountID:     "user@email.com",
				Role:          "role_name",
			},
		}
		aggregatedColumns := s.columnNames
		for _, c := range s.resourceColumnNames {
			aggregatedColumns = append(aggregatedColumns, fmt.Sprintf("Resource__%s", c))
		}
		expectedRows := sqlmock.NewRows(aggregatedColumns)
		for _, r := range expectedRecords {
			expectedRows.AddRow(
				// appeal
				r.ID,
				r.ResourceID,
				r.PolicyID,
				r.PolicyVersion,
				r.Status,
				r.AccountID,
				r.AccountType,
				r.CreatedBy,
				"null",
				r.Role,
				pq.StringArray(r.Permissions),
				"null",
				"null",
				"null",
				r.CreatedAt,
				r.UpdatedAt,

				// resource
				r.Resource.ID,
				r.Resource.ProviderType,
				r.Resource.ProviderURN,
				r.Resource.Type,
				r.Resource.URN,
				"null",
				"null",
				r.Resource.CreatedAt,
				r.Resource.UpdatedAt,
			)
		}
		s.dbmock.
			ExpectQuery(expectedQuery).
			WillReturnRows(expectedRows)

		actualRecords, actualError := s.repository.Find(&domain.ListAppealsFilter{})

		s.Nil(actualError)
		s.Equal(expectedRecords, actualRecords)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func (s *AppealRepositoryTestSuite) TestBulkUpsert() {
	s.Run("should return error if appeals input is invalid", func() {
		invalidAppeals := []*domain.Appeal{
			{
				Details: map[string]interface{}{
					"foo": make(chan int), // invalid value
				},
			},
		}

		actualErr := s.repository.BulkUpsert(invalidAppeals)

		s.EqualError(actualErr, "json: unsupported type: chan int")
	})

	expectedQuery := regexp.QuoteMeta(`INSERT INTO "appeals" ("resource_id","policy_id","policy_version","status","account_id","account_type","created_by","creator","role","permissions","options","labels","details","revoked_by","revoked_at","revoke_reason","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19),($20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38) ON CONFLICT ("id") DO UPDATE SET "resource_id"="excluded"."resource_id","policy_id"="excluded"."policy_id","policy_version"="excluded"."policy_version","status"="excluded"."status","account_id"="excluded"."account_id","account_type"="excluded"."account_type","created_by"="excluded"."created_by","creator"="excluded"."creator","role"="excluded"."role","permissions"="excluded"."permissions","options"="excluded"."options","labels"="excluded"."labels","details"="excluded"."details","revoked_by"="excluded"."revoked_by","revoked_at"="excluded"."revoked_at","revoke_reason"="excluded"."revoke_reason","updated_at"="excluded"."updated_at","deleted_at"="excluded"."deleted_at" RETURNING "id"`)

	appeals := []*domain.Appeal{
		{
			AccountID:   "test@email.com",
			Role:        "role_name",
			Permissions: []string{"test-permission"},
			ResourceID:  uuid.New().String(),
		},
		{
			AccountID:   "test2@email.com",
			Role:        "role_name",
			Permissions: []string{"test-permission"},
			ResourceID:  uuid.New().String(),
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
			a.AccountID,
			a.AccountType,
			a.CreatedBy,
			"null",
			a.Role,
			pq.StringArray(a.Permissions),
			"null",
			"null",
			"null",
			a.RevokedBy,
			utils.AnyTime{},
			a.RevokeReason,
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

		actualError := s.repository.BulkUpsert(appeals)

		s.EqualError(actualError, expectedError.Error())
		s.NoError(s.dbmock.ExpectationsWereMet())
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

		actualError := s.repository.BulkUpsert(appeals)

		s.Nil(actualError)
		for i, a := range appeals {
			s.Equal(expectedIDs[i], a.ID)
		}
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func (s *AppealRepositoryTestSuite) TestUpdate() {
	s.Run("should return error if appeal input is invalid", func() {
		invalidAppeal := &domain.Appeal{
			ID: uuid.New().String(),
			Details: map[string]interface{}{
				"foo": make(chan int), // invalid value
			},
		}

		actualError := s.repository.Update(invalidAppeal)

		s.EqualError(actualError, "json: unsupported type: chan int")
	})

	s.Run("should return error if got error from transaction", func() {
		expectedError := errors.New("db error")
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(".*").
			WillReturnError(expectedError)
		s.dbmock.ExpectRollback()

		actualError := s.repository.Update(&domain.Appeal{ID: uuid.New().String()})

		s.EqualError(actualError, expectedError.Error())
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	expectedUpdateApprovalsQuery := regexp.QuoteMeta(`INSERT INTO "approvals" ("name","index","appeal_id","status","actor","reason","policy_id","policy_version","created_at","updated_at","deleted_at","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12),($13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24) ON CONFLICT ("id") DO UPDATE SET "name"="excluded"."name","index"="excluded"."index","appeal_id"="excluded"."appeal_id","status"="excluded"."status","actor"="excluded"."actor","reason"="excluded"."reason","policy_id"="excluded"."policy_id","policy_version"="excluded"."policy_version","created_at"="excluded"."created_at","updated_at"="excluded"."updated_at","deleted_at"="excluded"."deleted_at" RETURNING "id"`)
	expectedUpdateAppealQuery := regexp.QuoteMeta(`UPDATE "appeals" SET "resource_id"=$1,"policy_id"=$2,"policy_version"=$3,"status"=$4,"account_id"=$5,"account_type"=$6,"created_by"=$7,"creator"=$8,"role"=$9,"permissions"=$10,"options"=$11,"labels"=$12,"details"=$13,"revoked_by"=$14,"revoked_at"=$15,"revoke_reason"=$16,"created_at"=$17,"updated_at"=$18,"deleted_at"=$19 WHERE "id" = $20`)
	s.Run("should return nil on success", func() {
		expectedID := uuid.New().String()
		appeal := &domain.Appeal{
			ID: expectedID,
			Approvals: []*domain.Approval{
				{
					ID:       uuid.New().String(),
					AppealID: expectedID,
				},
				{
					ID:       uuid.New().String(),
					AppealID: expectedID,
				},
			},
		}

		s.dbmock.ExpectBegin()
		s.dbmock.ExpectExec(expectedUpdateAppealQuery).
			WillReturnResult(sqlmock.NewResult(1, 1))
		var expectedApprovalArgs []driver.Value
		expectedApprovalRows := sqlmock.NewRows([]string{"id"})
		for _, approval := range appeal.Approvals {
			expectedApprovalArgs = append(expectedApprovalArgs,
				approval.Name,
				approval.Index,
				approval.AppealID,
				approval.Status,
				approval.Actor,
				approval.Reason,
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
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func TestAppealRepository(t *testing.T) {
	suite.Run(t, new(AppealRepositoryTestSuite))
}
