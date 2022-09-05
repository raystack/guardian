package postgres_test

import (
	"context"
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
	"github.com/odpf/guardian/core/grant"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/guardian/mocks"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type GrantRepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository *postgres.GrantRepository

	timeNow     time.Time
	columnNames []string
}

func TestGrantRepository(t *testing.T) {
	suite.Run(t, new(GrantRepositoryTestSuite))
}

func (s *GrantRepositoryTestSuite) setup() {
	db, mock, err := mocks.NewStore()
	s.Require().NoError(err)
	sqldb, err := db.DB()
	s.Require().NoError(err)
	s.dbmock = mock
	s.sqldb = sqldb
	s.repository = postgres.NewGrantRepository(db)

	s.timeNow = time.Now()
	s.columnNames = []string{
		"id", "status", "account_id", "account_type", "resource_id", "role", "permissions",
		"expiration_date", "appeal_id", "revoked_by", "revoked_at", "revoke_reason",
		"created_by", "created_at", "updated_at",
	}
}

func (s *GrantRepositoryTestSuite) toRow(a domain.Grant) []driver.Value {
	permissions := fmt.Sprintf("{%s}", strings.Join(a.Permissions, ","))
	return []driver.Value{
		a.ID, a.Status, a.AccountID, a.AccountType, a.ResourceID, a.Role, permissions,
		a.ExpirationDate, a.AppealID, a.RevokedBy, a.RevokedAt, a.RevokeReason,
		a.CreatedBy, a.CreatedAt, a.UpdatedAt,
	}
}

func (s *GrantRepositoryTestSuite) TestList() {
	s.Run("should return list of grant on success", func() {
		s.setup()

		expectedGrants := []domain.Grant{
			{
				ID:             uuid.New().String(),
				Status:         "test-status",
				AccountID:      "test-account-id",
				AccountType:    "test-account-type",
				ResourceID:     uuid.New().String(),
				Role:           "test-role",
				Permissions:    []string{"test-permission"},
				ExpirationDate: &s.timeNow,
				AppealID:       uuid.New().String(),
				RevokedBy:      "test-revoked-by",
				RevokedAt:      &s.timeNow,
				RevokeReason:   "test-revoke-reason",
				CreatedAt:      s.timeNow,
				UpdatedAt:      s.timeNow,
			},
		}
		expectedQuery := regexp.QuoteMeta(`SELECT "grants"."id","grants"."status","grants"."account_id","grants"."account_type","grants"."resource_id","grants"."role","grants"."permissions","grants"."expiration_date","grants"."appeal_id","grants"."revoked_by","grants"."revoked_at","grants"."revoke_reason","grants"."created_by","grants"."created_at","grants"."updated_at","grants"."deleted_at","Resource"."id" AS "Resource__id","Resource"."provider_type" AS "Resource__provider_type","Resource"."provider_urn" AS "Resource__provider_urn","Resource"."type" AS "Resource__type","Resource"."urn" AS "Resource__urn","Resource"."name" AS "Resource__name","Resource"."details" AS "Resource__details","Resource"."labels" AS "Resource__labels","Resource"."created_at" AS "Resource__created_at","Resource"."updated_at" AS "Resource__updated_at","Resource"."deleted_at" AS "Resource__deleted_at","Resource"."is_deleted" AS "Resource__is_deleted","Appeal"."id" AS "Appeal__id","Appeal"."resource_id" AS "Appeal__resource_id","Appeal"."policy_id" AS "Appeal__policy_id","Appeal"."policy_version" AS "Appeal__policy_version","Appeal"."status" AS "Appeal__status","Appeal"."account_id" AS "Appeal__account_id","Appeal"."account_type" AS "Appeal__account_type","Appeal"."created_by" AS "Appeal__created_by","Appeal"."creator" AS "Appeal__creator","Appeal"."role" AS "Appeal__role","Appeal"."permissions" AS "Appeal__permissions","Appeal"."options" AS "Appeal__options","Appeal"."labels" AS "Appeal__labels","Appeal"."details" AS "Appeal__details","Appeal"."revoked_by" AS "Appeal__revoked_by","Appeal"."revoked_at" AS "Appeal__revoked_at","Appeal"."revoke_reason" AS "Appeal__revoke_reason","Appeal"."created_at" AS "Appeal__created_at","Appeal"."updated_at" AS "Appeal__updated_at","Appeal"."deleted_at" AS "Appeal__deleted_at" FROM "grants" LEFT JOIN "resources" "Resource" ON "grants"."resource_id" = "Resource"."id" LEFT JOIN "appeals" "Appeal" ON "grants"."appeal_id" = "Appeal"."id" WHERE "grants"."account_id" IN ($1) AND "grants"."account_type" IN ($2) AND "grants"."resource_id" IN ($3) AND "grants"."status" IN ($4) AND "grants"."role" IN ($5) AND "grants"."permissions" @> $6 AND "grants"."created_by" = $7 AND "Resource"."provider_type" IN ($8) AND "Resource"."provider_urn" IN ($9) AND "Resource"."type" IN ($10) AND "Resource"."urn" IN ($11) AND "grants"."deleted_at" IS NULL ORDER BY ARRAY_POSITION(ARRAY[$12,$13,$14,$15,$16], "grants"."status")`)
		expectedRows := sqlmock.NewRows(s.columnNames).AddRow(s.toRow(expectedGrants[0])...)
		s.dbmock.ExpectQuery(expectedQuery).
			WithArgs(
				"test-account-id",
				"test-account-type",
				"test-resource-id",
				"test-status",
				"test-role",
				pq.StringArray([]string{"test-permission"}),
				"test-created-by",
				"test-provider-type",
				"test-provider-urn",
				"test-resource-type",
				"test-resource-urn",
				postgres.AppealStatusDefaultSort[0],
				postgres.AppealStatusDefaultSort[1],
				postgres.AppealStatusDefaultSort[2],
				postgres.AppealStatusDefaultSort[3],
				postgres.AppealStatusDefaultSort[4],
			).
			WillReturnRows(expectedRows)

		grants, err := s.repository.List(context.Background(), domain.ListGrantsFilter{
			Statuses:      []string{"test-status"},
			AccountIDs:    []string{"test-account-id"},
			AccountTypes:  []string{"test-account-type"},
			ResourceIDs:   []string{"test-resource-id"},
			Roles:         []string{"test-role"},
			Permissions:   []string{"test-permission"},
			ProviderTypes: []string{"test-provider-type"},
			ProviderURNs:  []string{"test-provider-urn"},
			ResourceTypes: []string{"test-resource-type"},
			ResourceURNs:  []string{"test-resource-urn"},
			CreatedBy:     "test-created-by",
			OrderBy:       []string{"status"},
		})

		s.NoError(err)
		s.Equal(expectedGrants, grants)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("sould return error if db returns an error", func() {
		s.setup()

		expectedError := errors.New("db error")
		s.dbmock.ExpectQuery(".*").WillReturnError(expectedError)

		grants, err := s.repository.List(context.Background(), domain.ListGrantsFilter{})

		s.ErrorIs(err, expectedError)
		s.Nil(grants)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func (s *GrantRepositoryTestSuite) TestGetByID() {
	s.Run("should return grant details on success", func() {
		s.setup()

		expectedID := uuid.New().String()
		expectedGrant := &domain.Grant{
			ID:             expectedID,
			Status:         "test-status",
			AccountID:      "test-account-id",
			AccountType:    "test-account-type",
			ResourceID:     uuid.New().String(),
			Role:           "test-role",
			Permissions:    []string{"test-permission"},
			ExpirationDate: &s.timeNow,
			AppealID:       uuid.New().String(),
			RevokedBy:      "test-revoked-by",
			RevokedAt:      &s.timeNow,
			RevokeReason:   "test-revoke-reason",
			CreatedAt:      s.timeNow,
			UpdatedAt:      s.timeNow,
		}
		expectedQuery := regexp.QuoteMeta(`SELECT "grants"."id","grants"."status","grants"."account_id","grants"."account_type","grants"."resource_id","grants"."role","grants"."permissions","grants"."expiration_date","grants"."appeal_id","grants"."revoked_by","grants"."revoked_at","grants"."revoke_reason","grants"."created_by","grants"."created_at","grants"."updated_at","grants"."deleted_at","Resource"."id" AS "Resource__id","Resource"."provider_type" AS "Resource__provider_type","Resource"."provider_urn" AS "Resource__provider_urn","Resource"."type" AS "Resource__type","Resource"."urn" AS "Resource__urn","Resource"."name" AS "Resource__name","Resource"."details" AS "Resource__details","Resource"."labels" AS "Resource__labels","Resource"."created_at" AS "Resource__created_at","Resource"."updated_at" AS "Resource__updated_at","Resource"."deleted_at" AS "Resource__deleted_at","Resource"."is_deleted" AS "Resource__is_deleted","Appeal"."id" AS "Appeal__id","Appeal"."resource_id" AS "Appeal__resource_id","Appeal"."policy_id" AS "Appeal__policy_id","Appeal"."policy_version" AS "Appeal__policy_version","Appeal"."status" AS "Appeal__status","Appeal"."account_id" AS "Appeal__account_id","Appeal"."account_type" AS "Appeal__account_type","Appeal"."created_by" AS "Appeal__created_by","Appeal"."creator" AS "Appeal__creator","Appeal"."role" AS "Appeal__role","Appeal"."permissions" AS "Appeal__permissions","Appeal"."options" AS "Appeal__options","Appeal"."labels" AS "Appeal__labels","Appeal"."details" AS "Appeal__details","Appeal"."revoked_by" AS "Appeal__revoked_by","Appeal"."revoked_at" AS "Appeal__revoked_at","Appeal"."revoke_reason" AS "Appeal__revoke_reason","Appeal"."created_at" AS "Appeal__created_at","Appeal"."updated_at" AS "Appeal__updated_at","Appeal"."deleted_at" AS "Appeal__deleted_at" FROM "grants" LEFT JOIN "resources" "Resource" ON "grants"."resource_id" = "Resource"."id" LEFT JOIN "appeals" "Appeal" ON "grants"."appeal_id" = "Appeal"."id" WHERE "grants"."id" = $1 AND "grants"."deleted_at" IS NULL ORDER BY "grants"."id" LIMIT 1`)
		expectedRows := sqlmock.NewRows(s.columnNames).AddRow(s.toRow(*expectedGrant)...)
		s.dbmock.ExpectQuery(expectedQuery).
			WithArgs(expectedID).
			WillReturnRows(expectedRows)

		grant, err := s.repository.GetByID(context.Background(), expectedID)

		s.NoError(err)
		s.Equal(expectedGrant, grant)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return not found error if record not found", func() {
		s.setup()

		expectedError := gorm.ErrRecordNotFound
		s.dbmock.ExpectQuery(".*").WillReturnError(expectedError)

		actualGrant, err := s.repository.GetByID(context.Background(), "")

		s.ErrorIs(err, grant.ErrGrantNotFound)
		s.Nil(actualGrant)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return error if db returns an error", func() {
		s.setup()

		expectedError := errors.New("db error")
		s.dbmock.ExpectQuery(".*").WillReturnError(expectedError)

		grant, err := s.repository.GetByID(context.Background(), "")

		s.ErrorIs(err, expectedError)
		s.Nil(grant)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}
