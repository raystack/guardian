package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/odpf/guardian/core/provideractivity"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres/model"
	"gorm.io/gorm"
)

type ProviderActivityRepository struct {
	db *gorm.DB
}

func NewProviderActivityRepository(db *gorm.DB) *ProviderActivityRepository {
	return &ProviderActivityRepository{db}
}

func (r *ProviderActivityRepository) Find(ctx context.Context, filter domain.ListProviderActivitiesFilter) ([]*domain.ProviderActivity, error) {
	var activities []*model.ProviderActivity
	db := r.db.WithContext(ctx)
	if filter.ProviderIDs != nil {
		db = db.Where(`"provider_id" IN ?`, filter.ProviderIDs)
	}
	if filter.ResourceIDs != nil {
		db = db.Where(`"resource_id" IN ?`, filter.ResourceIDs)
	}
	if filter.AccountIDs != nil {
		db = db.Where(`"account_id" IN ?`, filter.AccountIDs)
	}
	if filter.Types != nil {
		db = db.Where(`"type" IN ?`, filter.Types)
	}
	if filter.TimestampGte != nil {
		db = db.Where(`"timestamp" >= ?`, *filter.TimestampGte)
	}
	if filter.TimestampLte != nil {
		db = db.Where(`"timestamp" <= ?`, *filter.TimestampLte)
	}
	if err := db.Find(&activities).Error; err != nil {
		return nil, err
	}

	var results []*domain.ProviderActivity
	for _, activity := range activities {
		pa, err := activity.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model %q to domain: %w", activity.ID, err)
		}
		results = append(results, pa)
	}
	return results, nil
}

func (r *ProviderActivityRepository) GetOne(ctx context.Context, id string) (*domain.ProviderActivity, error) {
	var activity model.ProviderActivity
	if err := r.db.
		WithContext(ctx).
		Joins("Provider").
		Joins("Resource").
		Where(`"provider_activities"."id" = ?`, id).
		First(&activity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, provideractivity.ErrNotFound
		}
		return nil, err
	}
	return activity.ToDomain()
}

func (r *ProviderActivityRepository) BulkInsert(ctx context.Context, pas []*domain.ProviderActivity) error {
	models := make([]*model.ProviderActivity, len(pas))
	for i, pa := range pas {
		models[i] = &model.ProviderActivity{}
		if err := models[i].FromDomain(pa); err != nil {
			return fmt.Errorf("failed to convert domain to model: %w", err)
		}
	}

	if err := r.db.WithContext(ctx).Create(models).Error; err != nil {
		return fmt.Errorf("failed to insert provider activities: %w", err)
	}

	for i, m := range models {
		newProviderActivity, err := m.ToDomain()
		if err != nil {
			return fmt.Errorf("failed to convert model %q to domain: %w", m.ID, err)
		}
		*pas[i] = *newProviderActivity
	}

	return nil
}
