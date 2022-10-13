package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/guardian/domain"
)

var ResourceColumns = []string{
	`resources.id as "resource.id"`,
	`resources.provider_type as "resource.provider_type"`,
	`resources.provider_urn as "resource.provider_urn"`,
	`resources.type as "resource.type"`,
	`resources.urn as "resource.urn"`,
	`resources.name as "resource.name"`,
	`resources.details as "resource.details"`,
	`resources.labels as "resource.labels"`,
	`resources.created_at as "resource.created_at"`,
	`resources.updated_at as "resource.updated_at"`,
	`resources.is_deleted as "resource.is_deleted"`,
}

// Resource is the database model for resource
type Resource struct {
	ID           uuid.UUID       `db:"id"`
	ProviderType string          `db:"provider_type"`
	ProviderURN  string          `db:"provider_urn"`
	Type         string          `db:"type"`
	URN          string          `db:"urn"`
	Name         string          `db:"name"`
	Details      json.RawMessage `db:"details"`
	Labels       json.RawMessage `db:"labels"`

	Provider Provider `db:"provider"`

	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
	IsDeleted bool         `db:"is_deleted"`
}

// TableName overrides the table name
func (Resource) TableName() string {
	return "resources"
}

// FromDomain uses *domain.Resource values as the model values
func (m *Resource) FromDomain(r *domain.Resource) error {
	details, err := json.Marshal(r.Details)
	if err != nil {
		return err
	}

	labels, err := json.Marshal(r.Labels)
	if err != nil {
		return err
	}

	var id uuid.UUID
	if r.ID != "" {
		uuid, err := uuid.Parse(r.ID)
		if err != nil {
			return fmt.Errorf("parsing uuid: %w", err)
		}
		id = uuid
	}
	m.ID = id
	m.ProviderType = r.ProviderType
	m.ProviderURN = r.ProviderURN
	m.Type = r.Type
	m.URN = r.URN
	m.Name = r.Name
	m.Details = details
	m.Labels = labels
	m.CreatedAt = r.CreatedAt
	m.UpdatedAt = r.UpdatedAt
	m.IsDeleted = r.IsDeleted
	return nil
}

// ToDomain transforms model into *domain.Provider
func (m *Resource) ToDomain() (*domain.Resource, error) {
	var details map[string]interface{}
	if err := json.Unmarshal(m.Details, &details); err != nil {
		return nil, err
	}

	var labels map[string]string
	if err := json.Unmarshal(m.Labels, &labels); err != nil {
		return nil, err
	}

	return &domain.Resource{
		ID:           m.ID.String(),
		ProviderType: m.ProviderType,
		ProviderURN:  m.ProviderURN,
		Type:         m.Type,
		URN:          m.URN,
		Name:         m.Name,
		Details:      details,
		Labels:       labels,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		IsDeleted:    m.IsDeleted,
	}, nil
}
