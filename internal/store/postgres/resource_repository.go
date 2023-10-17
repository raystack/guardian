package postgres

import (
	"context"
	"strings"

	"github.com/raystack/guardian/core/resource"
	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/internal/store/postgres/model"
	"github.com/raystack/guardian/utils"
	"gorm.io/gorm"
)

// ResourceRepository talks to the store/database to read/insert data
type ResourceRepository struct {
	store *Store
}

// NewResourceRepository returns *Repository
func NewResourceRepository(db *Store) *ResourceRepository {
	return &ResourceRepository{db}
}

// Find records based on filters
func (r *ResourceRepository) Find(ctx context.Context, filter domain.ListResourcesFilter) ([]*domain.Resource, error) {
	if err := utils.ValidateStruct(filter); err != nil {
		return nil, err
	}

	var models []*model.Resource
	err := r.store.Tx(ctx, func(tx *gorm.DB) error {
		if filter.IDs != nil {
			tx = tx.Where(filter.IDs)
		}
		if !filter.IsDeleted {
			tx = tx.Where(`"is_deleted" = ?`, filter.IsDeleted)
		}
		if filter.ResourceType != "" {
			tx = tx.Where(`"type" = ?`, filter.ResourceType)
		}
		if filter.Name != "" {
			tx = tx.Where(`"name" = ?`, filter.Name)
		}
		if filter.ProviderType != "" {
			tx = tx.Where(`"provider_type" = ?`, filter.ProviderType)
		}
		if filter.ProviderURN != "" {
			tx = tx.Where(`"provider_urn" = ?`, filter.ProviderURN)
		}
		if filter.ResourceURN != "" {
			tx = tx.Where(`"urn" = ?`, filter.ResourceURN)
		}
		if filter.ResourceURNs != nil {
			tx = tx.Where(`"urn" IN ?`, filter.ResourceURNs)
		}
		if filter.ResourceTypes != nil {
			tx = tx.Where(`"type" IN ?`, filter.ResourceTypes)
		}
		for path, v := range filter.Details {
			pathArr := "{" + strings.Join(strings.Split(path, "."), ",") + "}"
			tx = tx.Where(`"details" #>> ? = ?`, pathArr, v)
		}
		return tx.Find(&models).Error
	})
	if err != nil {
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
	if err := r.store.Tx(ctx, func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).Take(&m).Error
	}); err != nil {
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
		m.NamespaceID = namespaceFromContext(ctx)
		models = append(models, m)
	}

	if len(models) > 0 {
		return r.store.Tx(ctx, func(tx *gorm.DB) error {
			// upsert clause is moved to model.Resource.BeforeCreate() (gorm's hook) to apply the same for associations (model.Resource.Children)
			if err := tx.Session(&gorm.Session{CreateBatchSize: 1000}).Create(models).Error; err != nil {
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
	m.NamespaceID = namespaceFromContext(ctx)

	return r.store.Tx(ctx, func(tx *gorm.DB) error {
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
	var result *gorm.DB
	_ = r.store.Tx(ctx, func(tx *gorm.DB) error {
		result = tx.Where("id = ?", id).Delete(&model.Resource{})
		return nil
	})
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
	var result *gorm.DB
	_ = r.store.Tx(ctx, func(tx *gorm.DB) error {
		result = tx.Delete(&model.Resource{}, ids)
		return nil
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return resource.ErrRecordNotFound
	}

	return nil
}
