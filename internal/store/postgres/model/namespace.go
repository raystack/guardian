package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/guardian/domain"
	"gorm.io/datatypes"
)

type Namespace struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string
	State     string
	Metadata  datatypes.JSON
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Namespace) TableName() string {
	return "namespaces"
}

func (m *Namespace) FromDomain(a *domain.Namespace) error {
	if a.ID != "" {
		id, err := uuid.Parse(a.ID)
		if err != nil {
			return fmt.Errorf("failed to parse id: %w", err)
		}
		m.ID = id
	}

	m.Name = a.Name
	m.State = a.State
	if a.Metadata != nil {
		metadata, err := json.Marshal(a.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal provider namespace metadata: %w", err)
		}
		m.Metadata = metadata
	}
	m.CreatedAt = a.CreatedAt
	m.UpdatedAt = a.UpdatedAt
	return nil
}

func (m *Namespace) ToDomain(a *domain.Namespace) error {
	if a == nil {
		return fmt.Errorf("namespace target can't be nil")
	}
	a.ID = m.ID.String()
	a.Name = m.Name
	a.State = m.State
	a.UpdatedAt = m.UpdatedAt
	a.CreatedAt = m.CreatedAt

	if m.Metadata != nil {
		if err := json.Unmarshal(m.Metadata, &a.Metadata); err != nil {
			return fmt.Errorf("failed to unmarshal provider activity metadata: %w", err)
		}
	}
	return nil
}
