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
	"github.com/odpf/guardian/core/access"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/guardian/mocks"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type AccessRepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository *postgres.AccessRepository

	timeNow     time.Time
	columnNames []string
}

func TestAccessRepository(t *testing.T) {
	suite.Run(t, new(AccessRepositoryTestSuite))
}

func (s *AccessRepositoryTestSuite) setup() {
	db, mock, err := mocks.NewStore()
	s.Require().NoError(err)
	sqldb, err := db.DB()
	s.Require().NoError(err)
	s.dbmock = mock
	s.sqldb = sqldb
	s.repository = postgres.NewAccessRepository(db)

	s.timeNow = time.Now()
	s.columnNames = []string{
		"id", "status", "account_id", "account_type", "resource_id", "role", "permissions",
		"expiration_date", "appeal_id", "revoked_by", "revoked_at", "revoke_reason",
		"created_by", "created_at", "updated_at",
	}
}

func (s *AccessRepositoryTestSuite) toRow(a domain.Access) []driver.Value {
	permissions := fmt.Sprintf("{%s}", strings.Join(a.Permissions, ","))
	return []driver.Value{
		a.ID, a.Status, a.AccountID, a.AccountType, a.ResourceID, a.Role, permissions,
		a.ExpirationDate, a.AppealID, a.RevokedBy, a.RevokedAt, a.RevokeReason,
		a.CreatedBy, a.CreatedAt, a.UpdatedAt,
	}
}

func (s *AccessRepositoryTestSuite) TestList() {
	s.Run("should return list of access on success", func() {
		s.setup()

		expectedAccesses := []domain.Access{
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
		expectedQuery := regexp.QuoteMeta(`SELECT "accesses"."id","accesses"."status","accesses"."account_id","accesses"."account_type","accesses"."resource_id","accesses"."role","accesses"."permissions","accesses"."expiration_date","accesses"."appeal_id","accesses"."revoked_by","accesses"."revoked_at","accesses"."revoke_reason","accesses"."created_by","accesses"."created_at","accesses"."updated_at","accesses"."deleted_at","Resource"."id" AS "Resource__id","Resource"."provider_type" AS "Resource__provider_type","Resource"."provider_urn" AS "Resource__provider_urn","Resource"."type" AS "Resource__type","Resource"."urn" AS "Resource__urn","Resource"."name" AS "Resource__name","Resource"."details" AS "Resource__details","Resource"."labels" AS "Resource__labels","Resource"."created_at" AS "Resource__created_at","Resource"."updated_at" AS "Resource__updated_at","Resource"."deleted_at" AS "Resource__deleted_at","Resource"."is_deleted" AS "Resource__is_deleted","Appeal"."id" AS "Appeal__id","Appeal"."resource_id" AS "Appeal__resource_id","Appeal"."policy_id" AS "Appeal__policy_id","Appeal"."policy_version" AS "Appeal__policy_version","Appeal"."status" AS "Appeal__status","Appeal"."account_id" AS "Appeal__account_id","Appeal"."account_type" AS "Appeal__account_type","Appeal"."created_by" AS "Appeal__created_by","Appeal"."creator" AS "Appeal__creator","Appeal"."role" AS "Appeal__role","Appeal"."permissions" AS "Appeal__permissions","Appeal"."options" AS "Appeal__options","Appeal"."labels" AS "Appeal__labels","Appeal"."details" AS "Appeal__details","Appeal"."revoked_by" AS "Appeal__revoked_by","Appeal"."revoked_at" AS "Appeal__revoked_at","Appeal"."revoke_reason" AS "Appeal__revoke_reason","Appeal"."created_at" AS "Appeal__created_at","Appeal"."updated_at" AS "Appeal__updated_at","Appeal"."deleted_at" AS "Appeal__deleted_at" FROM "accesses" LEFT JOIN "resources" "Resource" ON "accesses"."resource_id" = "Resource"."id" LEFT JOIN "appeals" "Appeal" ON "accesses"."appeal_id" = "Appeal"."id" WHERE "accesses"."account_id" IN ($1) AND "accesses"."account_type" IN ($2) AND "accesses"."resource_id" IN ($3) AND "accesses"."status" IN ($4) AND "accesses"."role" IN ($5) AND "accesses"."permissions" @> $6 AND "accesses"."created_by" = $7 AND "Resource"."provider_type" IN ($8) AND "Resource"."provider_urn" IN ($9) AND "Resource"."type" IN ($10) AND "Resource"."urn" IN ($11) AND "accesses"."deleted_at" IS NULL ORDER BY ARRAY_POSITION(ARRAY[$12,$13,$14,$15,$16], "accesses"."status")`)
		expectedRows := sqlmock.NewRows(s.columnNames).AddRow(s.toRow(expectedAccesses[0])...)
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

		accesses, err := s.repository.List(context.Background(), domain.ListAccessesFilter{
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
		s.Equal(expectedAccesses, accesses)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("sould return error if db returns an error", func() {
		s.setup()

		expectedError := errors.New("db error")
		s.dbmock.ExpectQuery(".*").WillReturnError(expectedError)

		accesses, err := s.repository.List(context.Background(), domain.ListAccessesFilter{})

		s.ErrorIs(err, expectedError)
		s.Nil(accesses)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}

func (s *AccessRepositoryTestSuite) TestGetByID() {
	s.Run("should return access details on success", func() {
		s.setup()

		expectedID := uuid.New().String()
		expectedAccess := &domain.Access{
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
		expectedQuery := regexp.QuoteMeta(`SELECT "accesses"."id","accesses"."status","accesses"."account_id","accesses"."account_type","accesses"."resource_id","accesses"."role","accesses"."permissions","accesses"."expiration_date","accesses"."appeal_id","accesses"."revoked_by","accesses"."revoked_at","accesses"."revoke_reason","accesses"."created_by","accesses"."created_at","accesses"."updated_at","accesses"."deleted_at","Resource"."id" AS "Resource__id","Resource"."provider_type" AS "Resource__provider_type","Resource"."provider_urn" AS "Resource__provider_urn","Resource"."type" AS "Resource__type","Resource"."urn" AS "Resource__urn","Resource"."name" AS "Resource__name","Resource"."details" AS "Resource__details","Resource"."labels" AS "Resource__labels","Resource"."created_at" AS "Resource__created_at","Resource"."updated_at" AS "Resource__updated_at","Resource"."deleted_at" AS "Resource__deleted_at","Resource"."is_deleted" AS "Resource__is_deleted","Appeal"."id" AS "Appeal__id","Appeal"."resource_id" AS "Appeal__resource_id","Appeal"."policy_id" AS "Appeal__policy_id","Appeal"."policy_version" AS "Appeal__policy_version","Appeal"."status" AS "Appeal__status","Appeal"."account_id" AS "Appeal__account_id","Appeal"."account_type" AS "Appeal__account_type","Appeal"."created_by" AS "Appeal__created_by","Appeal"."creator" AS "Appeal__creator","Appeal"."role" AS "Appeal__role","Appeal"."permissions" AS "Appeal__permissions","Appeal"."options" AS "Appeal__options","Appeal"."labels" AS "Appeal__labels","Appeal"."details" AS "Appeal__details","Appeal"."revoked_by" AS "Appeal__revoked_by","Appeal"."revoked_at" AS "Appeal__revoked_at","Appeal"."revoke_reason" AS "Appeal__revoke_reason","Appeal"."created_at" AS "Appeal__created_at","Appeal"."updated_at" AS "Appeal__updated_at","Appeal"."deleted_at" AS "Appeal__deleted_at" FROM "accesses" LEFT JOIN "resources" "Resource" ON "accesses"."resource_id" = "Resource"."id" LEFT JOIN "appeals" "Appeal" ON "accesses"."appeal_id" = "Appeal"."id" WHERE "accesses"."id" = $1 AND "accesses"."deleted_at" IS NULL ORDER BY "accesses"."id" LIMIT 1`)
		expectedRows := sqlmock.NewRows(s.columnNames).AddRow(s.toRow(*expectedAccess)...)
		s.dbmock.ExpectQuery(expectedQuery).
			WithArgs(expectedID).
			WillReturnRows(expectedRows)

		access, err := s.repository.GetByID(context.Background(), expectedID)

		s.NoError(err)
		s.Equal(expectedAccess, access)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return not found error if record not found", func() {
		s.setup()

		expectedError := gorm.ErrRecordNotFound
		s.dbmock.ExpectQuery(".*").WillReturnError(expectedError)

		actualAccess, err := s.repository.GetByID(context.Background(), "")

		s.ErrorIs(err, access.ErrAccessNotFound)
		s.Nil(actualAccess)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})

	s.Run("should return error if db returns an error", func() {
		s.setup()

		expectedError := errors.New("db error")
		s.dbmock.ExpectQuery(".*").WillReturnError(expectedError)

		access, err := s.repository.GetByID(context.Background(), "")

		s.ErrorIs(err, expectedError)
		s.Nil(access)
		s.NoError(s.dbmock.ExpectationsWereMet())
	})
}
