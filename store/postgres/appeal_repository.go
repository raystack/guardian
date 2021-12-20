package postgres

import (
	"errors"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/store/model"
	"github.com/odpf/guardian/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type appealFindFilters struct {
	AccountID                 string    `mapstructure:"account_id" validate:"omitempty,required"`
	ResourceID                uint      `mapstructure:"resource_id" validate:"omitempty,required"`
	Role                      string    `mapstructure:"role" validate:"omitempty,required"`
	Statuses                  []string  `mapstructure:"statuses" validate:"omitempty,min=1"`
	ExpirationDateLessThan    time.Time `mapstructure:"expiration_date_lt" validate:"omitempty,required"`
	ExpirationDateGreaterThan time.Time `mapstructure:"expiration_date_gt" validate:"omitempty,required"`
	ProviderTypes             []string  `mapstructure:"provider_types" validate:"omitempty,min=1"`
	ProviderURNs              []string  `mapstructure:"provider_urns" validate:"omitempty,min=1"`
	ResourceTypes             []string  `mapstructure:"resource_types" validate:"omitempty,min=1"`
	ResourceURNs              []string  `mapstructure:"resource_urns" validate:"omitempty,min=1"`
}

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
func (r *AppealRepository) GetByID(id uint) (*domain.Appeal, error) {
	m := new(model.Appeal)
	if err := r.db.
		Preload("Approvals", func(db *gorm.DB) *gorm.DB {
			return db.Order("Approvals.index ASC")
		}).
		Preload("Approvals.Approvers").
		Preload("Resource").
		First(&m, id).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appeal.ErrAppealNotFound
		}
		return nil, err
	}

	a, err := m.ToDomain()
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (r *AppealRepository) Find(filters map[string]interface{}) ([]*domain.Appeal, error) {
	var conditions appealFindFilters
	if err := mapstructure.Decode(filters, &conditions); err != nil {
		return nil, err
	}
	if err := utils.ValidateStruct(conditions); err != nil {
		return nil, err
	}

	db := r.db
	if conditions.AccountID != "" {
		db = db.Where(`"account_id" = ?`, conditions.AccountID)
	}
	if conditions.Statuses != nil {
		db = db.Where(`"status" IN ?`, conditions.Statuses)
	}
	if conditions.ResourceID != 0 {
		db = db.Where(`"resource_id" = ?`, conditions.ResourceID)
	}
	if conditions.Role != "" {
		db = db.Where(`"role" = ?`, conditions.Role)
	}
	if !conditions.ExpirationDateLessThan.IsZero() {
		db = db.Where(`"options" -> 'expiration_date' < ?`, conditions.ExpirationDateLessThan)
	}
	if !conditions.ExpirationDateGreaterThan.IsZero() {
		db = db.Where(`"options" -> 'expiration_date' > ?`, conditions.ExpirationDateGreaterThan)
	}

	db = db.Clauses(clause.OrderBy{
		Expression: clause.Expr{
			SQL:                `ARRAY_POSITION(ARRAY[?], "status"), "updated_at" desc`,
			Vars:               []interface{}{AppealStatusDefaultSort},
			WithoutParentheses: true,
		},
	})

	db = db.Joins("Resource")
	if conditions.ProviderTypes != nil {
		db = db.Where(`"Resource"."provider_type" IN ?`, conditions.ProviderTypes)
	}
	if conditions.ProviderURNs != nil {
		db = db.Where(`"Resource"."provider_urn" IN ?`, conditions.ProviderURNs)
	}
	if conditions.ResourceTypes != nil {
		db = db.Where(`"Resource"."type" IN ?`, conditions.ResourceTypes)
	}
	if conditions.ResourceURNs != nil {
		db = db.Where(`"Resource"."urn" IN ?`, conditions.ResourceURNs)
	}

	var models []*model.Appeal
	if err := db.Find(&models).Error; err != nil {
		return nil, err
	}

	records := []*domain.Appeal{}
	for _, m := range models {
		a, err := m.ToDomain()
		if err != nil {
			return nil, err
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
				return err
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
			return err
		}

		*a = *newRecord

		return nil
	})
}
