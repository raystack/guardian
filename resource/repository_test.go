package resource_test

import (
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/resource"
	"github.com/odpf/guardian/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type RepositoryTestSuite struct {
	suite.Suite
	sqldb      *sql.DB
	dbmock     sqlmock.Sqlmock
	repository *resource.Repository
}

func (s *RepositoryTestSuite) SetupTest() {
	db, mock, _ := mocks.NewStore()
	s.sqldb, _ = db.DB()
	s.dbmock = mock
	s.repository = resource.NewRepository(db)
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.sqldb.Close()
}

func (s *RepositoryTestSuite) TestBulkUpsert() {
	s.Run("should return records with with existing or new IDs", func() {
		resources := []*domain.Resource{
			{
				ProviderType: "provider_test",
				ProviderURN:  "provider_urn_test",
				Type:         "resource_type",
				URN:          "resource_type.resource_name",
				Name:         "resource_name",
			},
			{
				ProviderType: "provider_test",
				ProviderURN:  "provider_urn_test",
				Type:         "resource_type",
				URN:          "resource_type.resource_name_2",
				Name:         "resource_name_2",
			},
		}

		expectedQuery := regexp.QuoteMeta(`INSERT INTO "resources" ("provider_type","provider_urn","type","urn","name","details","labels","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10),($11,$12,$13,$14,$15,$16,$17,$18,$19,$20) ON CONFLICT ("provider_type","provider_urn","type","urn") DO UPDATE SET "name"="excluded"."name","details"="excluded"."details","labels"="excluded"."labels","updated_at"="excluded"."updated_at" RETURNING "id"`)
		expectedArgs := []driver.Value{}
		for _, r := range resources {
			expectedArgs = append(expectedArgs,
				r.ProviderType,
				r.ProviderURN,
				r.Type,
				r.URN,
				r.Name,
				"null",
				"null",
				utils.AnyTime{},
				utils.AnyTime{},
				gorm.DeletedAt{},
			)
		}
		expectedIDs := []uint{1, 2}
		expectedRows := sqlmock.NewRows([]string{"id"})
		for _, id := range expectedIDs {
			expectedRows.AddRow(id)
		}
		s.dbmock.ExpectBegin()
		s.dbmock.ExpectQuery(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnRows(expectedRows)
		s.dbmock.ExpectCommit()

		err := s.repository.BulkUpsert(resources)

		s.Nil(err)
		for i, r := range resources {
			s.Equal(expectedIDs[i], r.ID)
		}
	})
}

func TestRepository(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
