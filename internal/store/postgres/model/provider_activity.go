package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/odpf/guardian/domain"
	"gorm.io/datatypes"
)

type ProviderActivity struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ProviderID     uuid.UUID
	ResourceID     uuid.UUID
	AccountType    string
	AccountID      string
	Timestamp      time.Time
	Authorizations pq.StringArray `gorm:"type:text[]"`
	Type           string
	Metadata       datatypes.JSON
	CreatedAt      time.Time `gorm:"autoCreateTime"`

	Provider *Provider `gorm:"ForeignKey:ProviderID;References:ID"`
	Resource *Resource `gorm:"ForeignKey:ResourceID;References:ID"`
}

func (ProviderActivity) TableName() string {
	return "provider_activities"
}

func (m *ProviderActivity) FromDomain(pa *domain.ProviderActivity) error {
	if pa.ID != "" {
		id, err := uuid.Parse(pa.ID)
		if err != nil {
			return fmt.Errorf("failed to parse id: %w", err)
		}
		m.ID = id
	}
	if pa.ProviderID != "" {
		id, err := uuid.Parse(pa.ProviderID)
		if err != nil {
			return fmt.Errorf("failed to parse provider id: %w", err)
		}
		m.ProviderID = id
	}
	if pa.ResourceID != "" {
		id, err := uuid.Parse(pa.ResourceID)
		if err != nil {
			return fmt.Errorf("failed to parse resource id: %w", err)
		}
		m.ResourceID = id
	}
	m.AccountType = pa.AccountType
	m.AccountID = pa.AccountID
	m.Timestamp = pa.Timestamp
	m.Authorizations = pa.Authorizations
	m.Type = pa.Type

	if pa.Metadata != nil {
		metadata, err := json.Marshal(pa.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal provider activity metadata: %w", err)
		}
		m.Metadata = metadata
	}
	m.CreatedAt = pa.CreatedAt
	if pa.Provider != nil {
		m.Provider = &Provider{}
		if err := m.Provider.FromDomain(pa.Provider); err != nil {
			return fmt.Errorf("failed to convert provider: %w", err)
		}
	}
	if pa.Resource != nil {
		m.Resource = &Resource{}
		if err := m.Resource.FromDomain(pa.Resource); err != nil {
			return fmt.Errorf("failed to convert resource: %w", err)
		}
	}

	return nil
}

func (m *ProviderActivity) ToDomain() (*domain.ProviderActivity, error) {
	pa := &domain.ProviderActivity{
		ID:             m.ID.String(),
		ProviderID:     m.ProviderID.String(),
		ResourceID:     m.ResourceID.String(),
		AccountType:    m.AccountType,
		AccountID:      m.AccountID,
		Timestamp:      m.Timestamp,
		Authorizations: m.Authorizations,
		Type:           m.Type,
		CreatedAt:      m.CreatedAt,
	}
	if m.Metadata != nil {
		if err := json.Unmarshal(m.Metadata, &pa.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider activity metadata: %w", err)
		}
	}
	if m.Provider != nil {
		p, err := m.Provider.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert provider: %w", err)
		}
		pa.Provider = p
	}
	if m.Resource != nil {
		r, err := m.Resource.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert resource: %w", err)
		}
		pa.Resource = r
	}
	return pa, nil
}
