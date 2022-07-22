package postgres

import (
	"errors"
	"fmt"

	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres/model"
	"github.com/odpf/guardian/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	AppealStatusDefaultSort = []string{
		domain.AppealStatusPending,
		domain.AppealStatusActive,
		domain.AppealStatusRejected,
		domain.AppealStatusTerminated,
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
func (r *AppealRepository) GetByID(id string) (*domain.Appeal, error) {
	m := new(model.Appeal)
	if err := r.db.
		Preload("Approvals", func(db *gorm.DB) *gorm.DB {
			return db.Order("Approvals.index ASC")
		}).
		Preload("Approvals.Approvers").
		Preload("Resource").
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

func (r *AppealRepository) Find(filters *domain.ListAppealsFilter) ([]*domain.Appeal, error) {
	if err := utils.ValidateStruct(filters); err != nil {
		return nil, err
	}

	db := r.db
	if filters.CreatedBy != "" {
		db = db.Where(`"created_by" = ?`, filters.CreatedBy)
	}
	if filters.AccountID != nil {
		db = db.Where(`"account_id" IN ?`, filters.AccountID)
	}
	if filters.Statuses != nil {
		db = db.Where(`"status" IN ?`, filters.Statuses)
	}
	if filters.ResourceID != "" {
		db = db.Where(`"resource_id" = ?`, filters.ResourceID)
	}
	if filters.Role != "" {
		db = db.Where(`"role" = ?`, filters.Role)
	}
	if !filters.ExpirationDateLessThan.IsZero() {
		db = db.Where(`"options" -> 'expiration_date' < ?`, filters.ExpirationDateLessThan)
	}
	if !filters.ExpirationDateGreaterThan.IsZero() {
		db = db.Where(`"options" -> 'expiration_date' > ?`, filters.ExpirationDateGreaterThan)
	}
	if filters.OrderBy != nil {
		db = addOrderByClause(db, filters.OrderBy, addOrderByClauseOptions{
			statusColumnName: `"status"`,
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
	if err := db.Find(&models).Error; err != nil {
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
func (r *AppealRepository) BulkUpsert(appeals []*domain.Appeal) error {
	models := []*model.Appeal{}
	for _, a := range appeals {
		m := new(model.Appeal)
		if err := m.FromDomain(a); err != nil {
			return err
		}
		models = append(models, m)
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
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
func (r *AppealRepository) Update(a *domain.Appeal) error {
	m := new(model.Appeal)
	if err := m.FromDomain(a); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
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
