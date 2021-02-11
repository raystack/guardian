package repositories

import (
	"github.com/odpf/guardian/models"
	"gorm.io/gorm"
)

// Provider repository interface
type Provider interface {
	Create(*models.Provider) error
	Update(*models.Provider) error
	Find(filters map[string]interface{}) ([]*models.Provider, error)
	GetOne(uint) (*models.Provider, error)
	Delete(uint) error
}

// ProviderRepository repository
type ProviderRepository struct {
	db *gorm.DB
}

// NewProvider returns provider repository struct
func NewProvider(db *gorm.DB) *ProviderRepository {
	return &ProviderRepository{db}
}

// Create new record to database
func (r *ProviderRepository) Create(p *models.Provider) error {
	if result := r.db.Create(p); result.Error != nil {
		return result.Error
	}

	return nil
}

// Update record by ID
func (r *ProviderRepository) Update(p *models.Provider) error {
	return nil
}

// Find records based on filters
func (r *ProviderRepository) Find(filters map[string]interface{}) ([]*models.Provider, error) {
	return nil, nil
}

// GetOne record by ID
func (r *ProviderRepository) GetOne(id uint) (*models.Provider, error) {
	return nil, nil
}

// Delete record by ID
func (r *ProviderRepository) Delete(id uint) error {
	return nil
}
