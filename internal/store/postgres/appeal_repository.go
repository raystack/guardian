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
	store *Store
}

// NewAppealRepository returns repository struct
func NewAppealRepository(store *Store) *AppealRepository {
	return &AppealRepository{
		store: store,
	}
}

// GetByID returns appeal record by id along with the approvals and the approvers
func (r *AppealRepository) GetByID(ctx context.Context, id string) (*domain.Appeal, error) {
	m := new(model.Appeal)
	if err := r.store.Tx(ctx, func(tx *gorm.DB) error {
		return tx.
			Preload("Approvals", func(db *gorm.DB) *gorm.DB {
				return db.Order("Approvals.index ASC")
			}).
			Preload("Approvals.Approvers").
			Preload("Resource").
			Preload("Grant").
			First(&m, "id = ?", id).
			Error
	}); err != nil {
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
	var models []*model.Appeal

	err := r.store.Tx(ctx, func(tx *gorm.DB) error {
		db := tx
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

		return db.Joins("Grant").Find(&models).Error
	})
	if err != nil {
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

func (r *AppealRepository) GetAppealsTotalCount(ctx context.Context, filter *domain.ListAppealsFilter) (int64, error) {
	db := r.store.db.WithContext(ctx)
	db = applyAppealFilter(db, filter)
	var count int64
	err := db.Model(&model.Appeal{}).Count(&count).Error

	return count, err
}

// BulkUpsert new record to database
func (r *AppealRepository) BulkUpsert(ctx context.Context, appeals []*domain.Appeal) error {
	models := []*model.Appeal{}
	for _, a := range appeals {
		m := new(model.Appeal)
		if err := m.FromDomain(a); err != nil {
			return err
		}
		m.NamespaceID = namespaceFromContext(ctx)
		models = append(models, m)
	}

	return r.store.Tx(ctx, func(tx *gorm.DB) error {
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
	m.NamespaceID = namespaceFromContext(ctx)

	return r.store.Tx(ctx, func(tx *gorm.DB) error {
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

func applyAppealFilter(db *gorm.DB, filters *domain.ListAppealsFilter) *gorm.DB {
	db = db.Joins("JOIN resources ON appeals.resource_id = resources.id")
	if filters.Q != "" {
		// NOTE: avoid adding conditions before this grouped where clause.
		// Otherwise, it will be wrapped in parentheses and the query will be invalid.
		db = db.Where(db.
			Where(`"appeals"."account_id" LIKE ?`, fmt.Sprintf("%%%s%%", filters.Q)).
			Or(`"appeals"."role" LIKE ?`, fmt.Sprintf("%%%s%%", filters.Q)).
			Or(`"resources"."urn" LIKE ?`, fmt.Sprintf("%%%s%%", filters.Q)),
		)
	}
	if filters.Statuses != nil {
		db = db.Where(`"appeals"."status" IN ?`, filters.Statuses)
	}
	if filters.AccountTypes != nil {
		db = db.Where(`"appeals"."account_type" IN ?`, filters.AccountTypes)
	}
	if filters.ResourceTypes != nil {
		db = db.Where(`"resources"."type" IN ?`, filters.ResourceTypes)
	}
	if filters.Size > 0 {
		db = db.Limit(filters.Size)
	}
	if filters.Offset > 0 {
		db = db.Offset(filters.Offset)
	}
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

	return db
}
