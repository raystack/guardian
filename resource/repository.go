package resource

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/model"
	"github.com/odpf/guardian/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type findFilters struct {
	IDs          []uint            `mapstructure:"ids" validate:"omitempty,min=1"`
	IsDeleted    bool              `mapstructure:"is_deleted" validate:"omitempty"`
	ProviderType string            `mapstructure:"provider_type" validate:"omitempty"`
	ProviderURN  string            `mapstructure:"provider_urn" validate:"omitempty"`
	Name         string            `mapstructure:"name" validate:"omitempty"`
	ResourceURN  string            `mapstructure:"urn" validate:"omitempty"`
	ResourceType string            `mapstructure:"type" validate:"omitempty"`
	Details      map[string]string `mapstructure:"details"`
}

// Repository talks to the store/database to read/insert data
type Repository struct {
	db *gorm.DB
}

// NewRepository returns *Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db}
}

// Find records based on filters
func (r *Repository) Find(filters map[string]interface{}) ([]*domain.Resource, error) {
	var conditions findFilters
	if err := mapstructure.Decode(filters, &conditions); err != nil {
		return nil, err
	}
	if err := utils.ValidateStruct(conditions); err != nil {
		return nil, err
	}

	db := r.db
	if conditions.IDs != nil {
		db = db.Where(conditions.IDs)
	}
	if !conditions.IsDeleted {
		db = db.Where(`"is_deleted" = ?`, conditions.IsDeleted)
	}
	if conditions.ResourceType != "" {
		db = db.Where(`"type" = ?`, conditions.ResourceType)
	}
	if conditions.Name != "" {
		db = db.Where(`"name" = ?`, conditions.Name)
	}
	if conditions.ProviderType != "" {
		db = db.Where(`"provider_type" = ?`, conditions.ProviderType)
	}
	if conditions.ProviderURN != "" {
		db = db.Where(`"provider_urn" = ?`, conditions.ProviderURN)
	}
	if conditions.ResourceURN != "" {
		db = db.Where(`"urn" = ?`, conditions.ResourceURN)
	}
	for path, v := range conditions.Details {
		pathArr := "{" + strings.Join(strings.Split(path, "."), ",") + "}"
		db = db.Where(`"details" #>> ? = ?`, pathArr, v)
	}

	var models []*model.Resource
	if err := db.Find(&models).Error; err != nil {
		return nil, err
	}

	records := []*domain.Resource{}
	for _, m := range models {
		r, err := m.ToDomain()
		if err != nil {
			return nil, err
		}

		records = append(records, r)
	}

	return records, nil
}

// GetOne record by ID
func (r *Repository) GetOne(id uint) (*domain.Resource, error) {
	if id == 0 {
		return nil, ErrEmptyIDParam
	}

	var m model.Resource
	if err := r.db.Where("id = ?", id).Take(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	res, err := m.ToDomain()
	if err != nil {
		return nil, err
	}

	return res, nil
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
			DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at", "is_deleted"}),
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

// Update record by ID
func (r *Repository) Update(resource *domain.Resource) error {
	if resource.ID == 0 {
		return ErrEmptyIDParam
	}

	m := new(model.Resource)
	if err := m.FromDomain(resource); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(m).Where("id = ?", m.ID).Updates(*m).Error; err != nil {
			return err
		}

		newRecord, err := m.ToDomain()
		if err != nil {
			return err
		}

		*resource = *newRecord

		return nil
	})
}
