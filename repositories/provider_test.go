package repositories_test

import (
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/odpf/guardian/models"
)

func (s *RepositoryTestSuite) TestProviderRepository() {
	s.Run("create", func() {
		expectedQuery := regexp.QuoteMeta(`INSERT INTO "providers" ("config","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4) RETURNING "id"`)

		s.Run("should update model's ID with the returned ID", func() {
			config := "config string"
			provider := &models.Provider{
				Config: config,
			}

			expectedID := uint(1)
			expectedRows := sqlmock.NewRows([]string{"id"}).
				AddRow(expectedID)
			s.dbmock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)

			err := s.repositories.Provider.Create(provider)

			actualID := provider.ID

			s.Nil(err)
			s.Equal(expectedID, actualID)
		})
	})
}
