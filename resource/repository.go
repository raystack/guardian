package resource

import (
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository talks to the store/database to read/insert data
type Repository struct {
	db *gorm.DB
}

// NewRepository returns *Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db}
}

// BulkUpsert inserts records if the records are not exist, or updates the records if they are already exist
func (r *Repository) BulkUpsert(resources []*domain.Resource) error {
	var models []*model.Resource
	for _, r := range resources {
		m := new(model.Resource)
		if err := m.FromDomain(r); err != nil {
			return err
		}

		models = append(models, m)
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		upsertClause := clause.OnConflict{
			Columns: []clause.Column{
				{Name: "provider_type"},
				{Name: "provider_urn"},
				{Name: "type"},
				{Name: "urn"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"name", "details", "labels", "updated_at"}),
		}
		if err := r.db.Clauses(upsertClause).Create(models).Error; err != nil {
			return err
		}

		for i, m := range models {
			r, err := m.ToDomain()
			if err != nil {
				return err
			}
			*resources[i] = *r
		}

		return nil
	})
}
