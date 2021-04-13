package provider

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

// GetByID record by ID
func (r *Repository) GetByID(id uint) (*domain.Provider, error) {
	if id == 0 {
		return nil, ErrEmptyIDParam
	}

	m := &model.Provider{
		ID: id,
	}
	if err := r.db.Take(m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	p, err := m.ToDomain()
	if err != nil {
		return nil, err
	}

	return p, nil
}

// GetOne returns provider by type and urn
func (r *Repository) GetOne(pType, urn string) (*domain.Provider, error) {
	if pType == "" {
		return nil, ErrEmptyProviderType
	}
	if urn == "" {
		return nil, ErrEmptyProviderURN
	}

	m := &model.Provider{
		Type: pType,
		URN:  urn,
	}
	if err := r.db.Take(m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	p, err := m.ToDomain()
	if err != nil {
		return nil, err
	}

	return p, err
}

// Update record by ID
func (r *Repository) Update(p *domain.Provider) error {
	if p.ID == 0 {
		return ErrEmptyIDParam
	}

	m := new(model.Provider)
	if err := m.FromDomain(p); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(m).Updates(*m).Error; err != nil {
			return err
		}

		newRecord, err := m.ToDomain()
		if err != nil {
			return err
		}

		*p = *newRecord

		return nil
	})
}

// Delete record by ID
func (r *Repository) Delete(id uint) error {
	return nil
}
