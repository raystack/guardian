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

type Activity struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ProviderID         uuid.UUID
	ResourceID         uuid.UUID
	ProviderActivityID string
	AccountType        string
	AccountID          string
	Timestamp          time.Time
	Authorizations     pq.StringArray `gorm:"type:text[]"`
	Type               string
	Metadata           datatypes.JSON
	CreatedAt          time.Time `gorm:"autoCreateTime"`

	Provider *Provider `gorm:"ForeignKey:ProviderID;References:ID"`
	Resource *Resource `gorm:"ForeignKey:ResourceID;References:ID"`
}

func (Activity) TableName() string {
	return "activities"
}

func (m *Activity) FromDomain(a *domain.Activity) error {
	if a.ID != "" {
		id, err := uuid.Parse(a.ID)
		if err != nil {
			return fmt.Errorf("failed to parse id: %w", err)
		}
		m.ID = id
	}
	if a.ProviderID != "" {
		id, err := uuid.Parse(a.ProviderID)
		if err != nil {
			return fmt.Errorf("failed to parse provider id: %w", err)
		}
		m.ProviderID = id
	}
	if a.ResourceID != "" {
		id, err := uuid.Parse(a.ResourceID)
		if err != nil {
			return fmt.Errorf("failed to parse resource id: %w", err)
		}
		m.ResourceID = id
	}
	m.ProviderActivityID = a.ProviderActivityID
	m.AccountType = a.AccountType
	m.AccountID = a.AccountID
	m.Timestamp = a.Timestamp
	m.Authorizations = a.Authorizations
	m.Type = a.Type

	if a.Metadata != nil {
		metadata, err := json.Marshal(a.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal provider activity metadata: %w", err)
		}
		m.Metadata = metadata
	}
	m.CreatedAt = a.CreatedAt
	if a.Provider != nil {
		m.Provider = &Provider{}
		if err := m.Provider.FromDomain(a.Provider); err != nil {
			return fmt.Errorf("failed to convert provider: %w", err)
		}
	}
	if a.Resource != nil {
		m.Resource = &Resource{}
		if err := m.Resource.FromDomain(a.Resource); err != nil {
			return fmt.Errorf("failed to convert resource: %w", err)
		}
	}

	return nil
}

func (m *Activity) ToDomain() (*domain.Activity, error) {
	a := &domain.Activity{
		ID:                 m.ID.String(),
		ProviderID:         m.ProviderID.String(),
		ResourceID:         m.ResourceID.String(),
		ProviderActivityID: m.ProviderActivityID,
		AccountType:        m.AccountType,
		AccountID:          m.AccountID,
		Timestamp:          m.Timestamp,
		Authorizations:     m.Authorizations,
		Type:               m.Type,
		CreatedAt:          m.CreatedAt,
	}
	if m.Metadata != nil {
		if err := json.Unmarshal(m.Metadata, &a.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider activity metadata: %w", err)
		}
	}
	if m.Provider != nil {
		p, err := m.Provider.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert provider: %w", err)
		}
		a.Provider = p
	}
	if m.Resource != nil {
		r, err := m.Resource.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert resource: %w", err)
		}
		a.Resource = r
	}
	return a, nil
}
