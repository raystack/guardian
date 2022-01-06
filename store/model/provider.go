package model

import (
	"encoding/json"
	"time"

	"github.com/odpf/guardian/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Provider is the database model for provider
type Provider struct {
	ID        uint   `gorm:"autoIncrement;uniqueIndex"`
	Type      string `gorm:"primaryKey"`
	URN       string `gorm:"primaryKey"`
	Config    datatypes.JSON
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the table name
func (Provider) TableName() string {
	return "providers"
}

// FromDomain uses *domain.Provider values as the model values
func (m *Provider) FromDomain(p *domain.Provider) error {
	config, err := json.Marshal(p.Config)
	if err != nil {
		return err
	}

	m.ID = p.ID
	m.Type = p.Type
	m.URN = p.URN
	m.Config = datatypes.JSON(config)
	m.CreatedAt = p.CreatedAt
	m.UpdatedAt = p.UpdatedAt

	return nil
}

// ToDomain transforms model into *domain.Provider
func (m *Provider) ToDomain() (*domain.Provider, error) {
	config := &domain.ProviderConfig{}
	if err := json.Unmarshal(m.Config, config); err != nil {
		return nil, err
	}

	return &domain.Provider{
		ID:        m.ID,
		Type:      m.Type,
		URN:       m.URN,
		Config:    config,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}, nil
}
