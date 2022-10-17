package postgres

import (
	"context"
	"strings"

	"github.com/odpf/guardian/core/resource"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres/model"
	"github.com/odpf/guardian/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ResourceRepository talks to the store/database to read/insert data
type ResourceRepository struct {
	db *gorm.DB
}

// NewResourceRepository returns *Repository
func NewResourceRepository(db *gorm.DB) *ResourceRepository {
	return &ResourceRepository{db}
}

// Find records based on filters
func (r *ResourceRepository) Find(ctx context.Context, filter domain.ListResourcesFilter) ([]*domain.Resource, error) {
	if err := utils.ValidateStruct(filter); err != nil {
		return nil, err
	}

	db := r.db.WithContext(ctx)
	if filter.IDs != nil {
		db = db.Where(filter.IDs)
	}
	if !filter.IsDeleted {
		db = db.Where(`"is_deleted" = ?`, filter.IsDeleted)
	}
	if filter.ResourceType != "" {
		db = db.Where(`"type" = ?`, filter.ResourceType)
	}
	if filter.Name != "" {
		db = db.Where(`"name" = ?`, filter.Name)
	}
	if filter.ProviderType != "" {
		db = db.Where(`"provider_type" = ?`, filter.ProviderType)
	}
	if filter.ProviderURN != "" {
		db = db.Where(`"provider_urn" = ?`, filter.ProviderURN)
	}
	if filter.ResourceURN != "" {
		db = db.Where(`"urn" = ?`, filter.ResourceURN)
	}
	if filter.ResourceURNs != nil {
		db = db.Where(`"urn" IN ?`, filter.ResourceURNs)
	}
	if filter.ResourceTypes != nil {
		db = db.Where(`"type" IN ?`, filter.ResourceTypes)
	}
	for path, v := range filter.Details {
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
func (r *ResourceRepository) GetOne(ctx context.Context, id string) (*domain.Resource, error) {
	if id == "" {
		return nil, resource.ErrEmptyIDParam
	}

	var m model.Resource
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, resource.ErrRecordNotFound
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
func (r *ResourceRepository) BulkUpsert(ctx context.Context, resources []*domain.Resource) error {
	var models []*model.Resource
	for _, r := range resources {
		m := new(model.Resource)
		if err := m.FromDomain(r); err != nil {
			return err
		}

		models = append(models, m)
	}

	if len(models) > 0 {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			upsertClause := clause.OnConflict{
				Columns: []clause.Column{
					{Name: "provider_type"},
					{Name: "provider_urn"},
					{Name: "type"},
					{Name: "urn"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"name", "details", "updated_at", "is_deleted"}),
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

	return nil
}

// Update record by ID
func (r *ResourceRepository) Update(ctx context.Context, res *domain.Resource) error {
	if res.ID == "" {
		return resource.ErrEmptyIDParam
	}

	m := new(model.Resource)
	if err := m.FromDomain(res); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(m).Where("id = ?", m.ID).Updates(*m).Error; err != nil {
			return err
		}

		newRecord, err := m.ToDomain()
		if err != nil {
			return err
		}

		*res = *newRecord

		return nil
	})
}

func (r *ResourceRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return resource.ErrEmptyIDParam
	}

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Resource{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return resource.ErrRecordNotFound
	}

	return nil
}

func (r *ResourceRepository) BatchDelete(ctx context.Context, ids []string) error {
	if ids == nil {
		return resource.ErrEmptyIDParam
	}

	result := r.db.WithContext(ctx).Delete(&model.Resource{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return resource.ErrRecordNotFound
	}

	return nil
}
