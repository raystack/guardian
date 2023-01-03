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
		a := &domain.Activity{}
		if err := activity.ToDomain(a); err != nil {
			return nil, fmt.Errorf("failed to convert model %q to domain: %w", activity.ID, err)
		}
		results = append(results, a)
	}
	return results, nil
}

func (r *ActivityRepository) GetOne(ctx context.Context, id string) (*domain.Activity, error) {
	var m model.Activity
	if err := r.db.
		WithContext(ctx).
		Joins("Provider").
		Joins("Resource").
		Where(`"activities"."id" = ?`, id).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, activity.ErrNotFound
		}
		return nil, err
	}

	a := &domain.Activity{}
	if err := m.ToDomain(a); err != nil {
		return nil, err
	}

	return a, nil
}

func (r *ActivityRepository) BulkUpsert(ctx context.Context, activities []*domain.Activity) error {
	if len(activities) == 0 {
		return nil
	}

	activityModels := make([]*model.Activity, len(activities))
	uniqueResourcesMap := map[string]*model.Resource{}

	for i, a := range activities {
		activityModels[i] = &model.Activity{}
		if err := activityModels[i].FromDomain(a); err != nil {
			return fmt.Errorf("failed to convert domain to model: %w", err)
		}

		// use single resource reference for activities with same resource
		if r := activityModels[i].Resource; r != nil {
			key := getResourceUniqueURN(*r)
			if _, exists := uniqueResourcesMap[key]; !exists {
				uniqueResourcesMap[key] = r
			} else {
				activityModels[i].Resource = uniqueResourcesMap[key]
			}
		}
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// upsert resources separately to avoid resource upsertion duplicate issue
		var resources []*model.Resource
		for _, r := range uniqueResourcesMap {
			resources = append(resources, r)
		}
		if len(resources) > 0 {
			if err := tx.Create(resources).Error; err != nil {
				return fmt.Errorf("failed to upsert resources: %w", err)
			}
		}
		for _, a := range activityModels {
			// set resource id after upsertion
			if a.Resource != nil {
				a.ResourceID = a.Resource.ID
			}
		}

		if err := tx.Omit("Resource").
			Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "provider_id"},
					{Name: "provider_activity_id"},
				},
				UpdateAll: true,
			}).
			Create(activityModels).Error; err != nil {
			return fmt.Errorf("failed to upsert provider activities: %w", err)
		}

		for i, m := range activityModels {
			if err := m.ToDomain(activities[i]); err != nil {
				return fmt.Errorf("failed to convert model %q to domain: %w", m.ID, err)
			}
		}

		return nil
	})
}

func getResourceUniqueURN(r model.Resource) string {
	return fmt.Sprintf(`%s/%s/%s/%s`, r.ProviderType, r.ProviderURN, r.Type, r.URN)
}
