package policy

import (
	"fmt"

	"github.com/odpf/guardian/domain"
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

// Create new record to database
func (r *Repository) Create(p *domain.Policy) error {
	m := new(Model)
	if err := m.fromDomain(p); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if result := tx.Create(m); result.Error != nil {
			return result.Error
		}

		newPolicy, err := m.toDomain()
		if err != nil {
			return err
		}

		*p = *newPolicy

		return nil
	})
}

// Find records based on filters
func (r *Repository) Find() ([]*domain.Policy, error) {
	policies := []*domain.Policy{}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var models []*Model
		if err := tx.Find(&models).Error; err != nil {
			return err
		}
		for _, m := range models {
			p, err := m.toDomain()
			fmt.Printf("===== %#v --- %#v\n", p, err)
			if err != nil {
				return err
			}

			policies = append(policies, p)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return policies, nil
}

// GetOne returns a policy record based on the id and version params.
// If version is 0, the latest version will be returned
func (r *Repository) GetOne(id string, version int) (*domain.Policy, error) {
	policy := &domain.Policy{}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		m := &Model{}
		condition := "id = ?"
		args := []interface{}{id}
		if version != 0 {
			condition = "id = ? AND version = ?"
			args = append(args, version)
		}

		conds := append([]interface{}{condition}, args...)
		if err := tx.Last(m, conds...).Error; err != nil {
			return err
		}

		p, err := m.toDomain()
		if err != nil {
			return err
		}

		policy = p
		return nil
	})
	if err != nil {
		return nil, err
	}

	return policy, nil
}
