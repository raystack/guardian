package provider

import (
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

// Create new record to database
func (r *Repository) Create(p *domain.Provider) error {
	m := new(model.Provider)
	if err := m.FromDomain(p); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if result := tx.Create(m); result.Error != nil {
			return result.Error
		}

		newProvider, err := m.ToDomain()
		if err != nil {
			return err
		}

		*p = *newProvider

		return nil
	})
}

// Update record by ID
func (r *Repository) Update(p *domain.Provider) error {
	return nil
}

// Find records based on filters
func (r *Repository) Find() ([]*domain.Provider, error) {
	providers := []*domain.Provider{}

	var models []*model.Provider
	if err := r.db.Find(&models).Error; err != nil {
		return nil, err
	}
	for _, m := range models {
		p, err := m.ToDomain()
		if err != nil {
			return nil, err
		}

		providers = append(providers, p)
	}

	return providers, nil
}

// GetOne record by ID
func (r *Repository) GetOne(id uint) (*domain.Provider, error) {
	return nil, nil
}

// Delete record by ID
func (r *Repository) Delete(id uint) error {
	return nil
}
