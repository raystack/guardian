package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/guardian/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Provider is the database model for provider
type Provider struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Type      string    `gorm:"uniqueIndex:provider_index"`
	URN       string    `gorm:"uniqueIndex:provider_index"`
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

	var id uuid.UUID
	if p.ID != "" {
		uuid, err := uuid.Parse(p.ID)
		if err != nil {
			return fmt.Errorf("parsing uuid: %w", err)
		}
		id = uuid
	}
	m.ID = id
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
		ID:        m.ID.String(),
		Type:      m.Type,
		URN:       m.URN,
		Config:    config,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}, nil
}
