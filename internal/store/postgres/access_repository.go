package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/odpf/guardian/core/access"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres/model"
	"gorm.io/gorm"
)

type AccessRepository struct {
	db *gorm.DB
}

func NewAccessRepository(db *gorm.DB) *AccessRepository {
	return &AccessRepository{db}
}

func (r *AccessRepository) List(ctx context.Context, filter domain.ListAccessesFilter) ([]domain.Access, error) {
	db := r.db.WithContext(ctx)
	if filter.AccountIDs != nil {
		db = db.Where(`"accesses"."account_id" IN ?`, filter.AccountIDs)
	}
	if filter.AccountTypes != nil {
		db = db.Where(`"accesses"."account_type" IN ?`, filter.AccountTypes)
	}
	if filter.ResourceIDs != nil {
		db = db.Where(`"accesses"."resource_id" IN ?`, filter.ResourceIDs)
	}
	if filter.Statuses != nil {
		db = db.Where(`"accesses"."status" IN ?`, filter.Statuses)
	}

	var models []model.Access
	if err := db.Joins("Resource").Joins("Appeal").Find(&models).Error; err != nil {
		return nil, err
	}

	var accesses []domain.Access
	for _, m := range models {
		a, err := m.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("parsing access %q: %w", a.ID, err)
		}
		accesses = append(accesses, *a)
	}

	return accesses, nil
}

func (r *AccessRepository) GetByID(ctx context.Context, id string) (*domain.Access, error) {
	m := new(model.Access)
	if err := r.db.WithContext(ctx).Joins("Resource").Joins("Appeal").First(&m, `"accesses"."id" = ?`, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, access.ErrAccessNotFound
		}
		return nil, err
	}

	a, err := m.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("parsing access %q: %w", a.ID, err)
	}
	return a, nil
}

func (r *AccessRepository) Update(ctx context.Context, a *domain.Access) error {
	if a == nil || a.ID == "" {
		return access.ErrEmptyIDParam
	}

	m := new(model.Access)
	if err := m.FromDomain(*a); err != nil {
		return fmt.Errorf("parsing access payload: %w", err)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(m).Updates(*m).Error; err != nil {
			return err
		}

		newAccess, err := m.ToDomain()
		if err != nil {
			return fmt.Errorf("parsing access: %w", err)
		}
		*a = *newAccess
		return nil
	})
}
