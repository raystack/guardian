package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/odpf/guardian/core/activity"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ActivityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db}
}

func (r *ActivityRepository) Find(ctx context.Context, filter domain.ListProviderActivitiesFilter) ([]*domain.Activity, error) {
	var activities []*model.Activity
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

	var results []*domain.Activity
	for _, activity := range activities {
		pa, err := activity.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model %q to domain: %w", activity.ID, err)
		}
		results = append(results, pa)
	}
	return results, nil
}

func (r *ActivityRepository) GetOne(ctx context.Context, id string) (*domain.Activity, error) {
	var m model.Activity
	if err := r.db.
		WithContext(ctx).
		Joins("Provider").
		Joins("Resource").
		Where(`"provider_activities"."id" = ?`, id).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, activity.ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain()
}

func (r *ActivityRepository) BulkUpsert(ctx context.Context, activities []*domain.Activity) error {
	models := make([]*model.Activity, len(activities))
	for i, a := range activities {
		models[i] = &model.Activity{}
		if err := models[i].FromDomain(a); err != nil {
			return fmt.Errorf("failed to convert domain to model: %w", err)
		}
	}

	if len(models) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "provider_id"},
				{Name: "provider_activity_id"},
			},
			UpdateAll: true,
		}).Create(models).Error; err != nil {
			return fmt.Errorf("failed to upsert provider activities: %w", err)
		}

		for i, m := range models {
			newActivity, err := m.ToDomain()
			if err != nil {
				return fmt.Errorf("failed to convert model %q to domain: %w", m.ID, err)
			}
			*activities[i] = *newActivity
		}

		return nil
	})
}
