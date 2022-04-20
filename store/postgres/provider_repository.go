package postgres

import (
	"errors"

	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/store/postgres/model"
	"gorm.io/gorm"
)

// ProviderRepository talks to the store to read or insert data
type ProviderRepository struct {
	db *gorm.DB
}

// NewProviderRepository returns repository struct
func NewProviderRepository(db *gorm.DB) *ProviderRepository {
	return &ProviderRepository{db}
}

// Create new record to database
func (r *ProviderRepository) Create(p *domain.Provider) error {
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
func (r *ProviderRepository) Find() ([]*domain.Provider, error) {
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
func (r *ProviderRepository) GetByID(id string) (*domain.Provider, error) {
	if id == "" {
		return nil, provider.ErrEmptyIDParam
	}

	var m model.Provider
	if err := r.db.First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, provider.ErrRecordNotFound
		}
		return nil, err
	}

	p, err := m.ToDomain()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *ProviderRepository) GetTypes() ([]domain.ProviderType, error) {
	var results []struct {
		ProviderType string
		ResourceType string
	}

	r.db.Raw("select distinct provider_type, type as resource_type from resources").Scan(&results)

	if len(results) == 0 {
		return nil, errors.New("no provider types found")
	}

	providerTypesMap := make(map[string][]string)
	for _, res := range results {
		if val, ok := providerTypesMap[res.ProviderType]; ok {
			providerTypesMap[res.ProviderType] = append(val, res.ResourceType)
		} else {
			providerTypesMap[res.ProviderType] = []string{res.ResourceType}
		}
	}

	var providerTypes []domain.ProviderType
	for providerType, resourceTypes := range providerTypesMap {
		providerTypes = append(providerTypes, domain.ProviderType{
			Name:          providerType,
			ResourceTypes: resourceTypes,
		})
	}

	return providerTypes, nil
}

// GetOne returns provider by type and urn
func (r *ProviderRepository) GetOne(pType, urn string) (*domain.Provider, error) {
	if pType == "" {
		return nil, provider.ErrEmptyProviderType
	}
	if urn == "" {
		return nil, provider.ErrEmptyProviderURN
	}

	m := &model.Provider{}
	db := r.db.Where("type = ?", pType).Where("urn = ?", urn)
	if err := db.Take(m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, provider.ErrRecordNotFound
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
func (r *ProviderRepository) Update(p *domain.Provider) error {
	if p.ID == "" {
		return provider.ErrEmptyIDParam
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
func (r *ProviderRepository) Delete(id string) error {
	return nil
}
