package provider

import (
	"encoding/json"
	"time"

	"github.com/odpf/guardian/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Model is the database model for provider
type Model struct {
	ID        uint `gorm:"primaryKey"`
	Type      string
	URN       string
	Config    datatypes.JSON
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the table name
func (Model) TableName() string {
	return "providers"
}

func (m *Model) fromDomain(p *domain.Provider) error {
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

func (m *Model) toDomain() (*domain.Provider, error) {
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
