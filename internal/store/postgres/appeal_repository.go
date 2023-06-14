package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/guardian/core/appeal"
	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/internal/store/postgres/model"
	"github.com/raystack/guardian/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	AppealStatusDefaultSort = []string{
		domain.AppealStatusPending,
		domain.AppealStatusApproved,
		domain.AppealStatusRejected,
		domain.AppealStatusCanceled,
	}
)

// AppealRepository talks to the store to read or insert data
type AppealRepository struct {
	db *gorm.DB
}

// NewAppealRepository returns repository struct
func NewAppealRepository(db *gorm.DB) *AppealRepository {
	return &AppealRepository{db}
}

// GetByID returns appeal record by id along with the approvals and the approvers
func (r *AppealRepository) GetByID(ctx context.Context, id string) (*domain.Appeal, error) {
	m := new(model.Appeal)
	if err := r.db.
		WithContext(ctx).
		Preload("Approvals", func(db *gorm.DB) *gorm.DB {
			return db.Order("Approvals.index ASC")
		}).
		Preload("Approvals.Approvers").
		Preload("Resource").
		Preload("Grant").
		First(&m, "id = ?", id).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appeal.ErrAppealNotFound
		}
		return nil, err
	}

	a, err := m.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("parsing appeal: %w", err)
	}

	return a, nil
}

func (r *AppealRepository) Find(ctx context.Context, filters *domain.ListAppealsFilter) ([]*domain.Appeal, error) {
	if err := utils.ValidateStruct(filters); err != nil {
		return nil, err
	}

	db := r.db.WithContext(ctx)
	if filters.CreatedBy != "" {
		db = db.Where(`"appeals"."created_by" = ?`, filters.CreatedBy)
	}
	accounts := make([]string, 0)
	if filters.AccountID != "" {
		accounts = append(accounts, filters.AccountID)
	}
	if filters.AccountIDs != nil {
		accounts = append(accounts, filters.AccountIDs...)
	}
	if len(accounts) > 0 {
		db = db.Where(`"appeals"."account_id" IN ?`, accounts)
	}
	if filters.Statuses != nil {
		db = db.Where(`"appeals"."status" IN ?`, filters.Statuses)
	}
	if filters.ResourceID != "" {
		db = db.Where(`"appeals"."resource_id" = ?`, filters.ResourceID)
	}
	if filters.Role != "" {
		db = db.Where(`"appeals"."role" = ?`, filters.Role)
	}
	if !filters.ExpirationDateLessThan.IsZero() {
		db = db.Where(`"options" -> 'expiration_date' < ?`, filters.ExpirationDateLessThan)
	}
	if !filters.ExpirationDateGreaterThan.IsZero() {
		db = db.Where(`"options" -> 'expiration_date' > ?`, filters.ExpirationDateGreaterThan)
	}
	if filters.OrderBy != nil {
		db = addOrderByClause(db, filters.OrderBy, addOrderByClauseOptions{
			statusColumnName: `"appeals"."status"`,
			statusesOrder:    AppealStatusDefaultSort,
		})
	}

	db = db.Joins("Resource")
	if filters.ProviderTypes != nil {
		db = db.Where(`"Resource"."provider_type" IN ?`, filters.ProviderTypes)
	}
	if filters.ProviderURNs != nil {
		db = db.Where(`"Resource"."provider_urn" IN ?`, filters.ProviderURNs)
	}
	if filters.ResourceTypes != nil {
		db = db.Where(`"Resource"."type" IN ?`, filters.ResourceTypes)
	}
	if filters.ResourceURNs != nil {
		db = db.Where(`"Resource"."urn" IN ?`, filters.ResourceURNs)
	}

	var models []*model.Appeal
	if err := db.Joins("Grant").Find(&models).Error; err != nil {
		return nil, err
	}

	records := []*domain.Appeal{}
	for _, m := range models {
		a, err := m.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("parsing appeal: %w", err)
		}

		records = append(records, a)
	}

	return records, nil
}

// BulkUpsert new record to database
func (r *AppealRepository) BulkUpsert(ctx context.Context, appeals []*domain.Appeal) error {
	models := []*model.Appeal{}
	for _, a := range appeals {
		m := new(model.Appeal)
		if err := m.FromDomain(a); err != nil {
			return err
		}
		models = append(models, m)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.OnConflict{UpdateAll: true}).
			Create(models).
			Error; err != nil {
			return err
		}

		for i, m := range models {
			newAppeal, err := m.ToDomain()
			if err != nil {
				return fmt.Errorf("parsing appeal: %w", err)
			}

			*appeals[i] = *newAppeal
		}

		return nil
	})
}

// Update an approval step
func (r *AppealRepository) Update(ctx context.Context, a *domain.Appeal) error {
	m := new(model.Appeal)
	if err := m.FromDomain(a); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Approvals.Approvers").Session(&gorm.Session{FullSaveAssociations: true}).Save(&m).Error; err != nil {
			return err
		}

		newRecord, err := m.ToDomain()
		if err != nil {
			return fmt.Errorf("parsing appeal: %w", err)
		}

		*a = *newRecord

		return nil
	})
}
