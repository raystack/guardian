package appeal

import (
	"errors"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/model"
	"gorm.io/gorm"
)

// Repository talks to the store to read or insert data
type Repository struct {
	db *gorm.DB
}

// NewRepository returns repository struct
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db}
}

// GetByID returns appeal record by id along with the approvals and the approvers
func (r *Repository) GetByID(id uint) (*domain.Appeal, error) {
	m := new(model.Appeal)
	if err := r.db.Preload("Approvals.Approvers").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	a, err := m.ToDomain()
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (r *Repository) Find(filters map[string]interface{}) ([]*domain.Appeal, error) {
	whereConditions := map[string]interface{}{}
	if filters["user"] != nil && filters["user"] != "" {
		whereConditions["user"] = filters["user"]
	}
	var models []*model.Appeal
	if err := r.db.Debug().Find(&models, whereConditions).Error; err != nil {
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

// Create new record to database
func (r *Repository) BulkInsert(appeals []*domain.Appeal) error {
	models := []*model.Appeal{}
	for _, a := range appeals {
		m := new(model.Appeal)
		if err := m.FromDomain(a); err != nil {
			return err
		}
		models = append(models, m)
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(models).Error; err != nil {
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
