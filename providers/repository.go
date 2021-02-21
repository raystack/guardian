package providers

import (
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
func (r *Repository) Create(p *domain.Provider) error {
	m := new(Model)
	m.fromDomain(p)
	if result := r.db.Create(m); result.Error != nil {
		return result.Error
	}

	p.ID = m.ID

	return nil
}

// Update record by ID
func (r *Repository) Update(p *domain.Provider) error {
	return nil
}

// Find records based on filters
func (r *Repository) Find(filters map[string]interface{}) ([]*domain.Provider, error) {
	return nil, nil
}

// GetOne record by ID
func (r *Repository) GetOne(id uint) (*domain.Provider, error) {
	return nil, nil
}

// Delete record by ID
func (r *Repository) Delete(id uint) error {
	return nil
}
