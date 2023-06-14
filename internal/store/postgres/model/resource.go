package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/guardian/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Resource is the database model for resource
type Resource struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ParentID     *string   `gorm:"type:uuid"`
	ProviderType string    `gorm:"uniqueIndex:resource_index"`
	ProviderURN  string    `gorm:"uniqueIndex:resource_index"`
	Type         string    `gorm:"uniqueIndex:resource_index"`
	URN          string    `gorm:"uniqueIndex:resource_index"`
	Name         string
	Details      datatypes.JSON
	Labels       datatypes.JSON

	Children []Resource `gorm:"ForeignKey:ParentID;References:ID"`
	Provider Provider   `gorm:"ForeignKey:ProviderType,ProviderURN;References:Type,URN"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	IsDeleted bool
}

// TableName overrides the table name
func (Resource) TableName() string {
	return "resources"
}

func (r *Resource) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.AddClause(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "provider_type"},
			{Name: "provider_urn"},
			{Name: "type"},
			{Name: "urn"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"name", "details", "updated_at", "is_deleted", "parent_id"}),
	})

	return nil
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

	if r.ID != "" {
		uuid, err := uuid.Parse(r.ID)
		if err != nil {
			return fmt.Errorf("parsing uuid: %w", err)
		}
		m.ID = uuid
	}

	m.ParentID = r.ParentID
	m.ProviderType = r.ProviderType
	m.ProviderURN = r.ProviderURN
	m.Type = r.Type
	m.URN = r.URN
	m.Name = r.Name
	m.Details = datatypes.JSON(details)
	m.Labels = datatypes.JSON(labels)
	m.CreatedAt = r.CreatedAt
	m.UpdatedAt = r.UpdatedAt
	m.IsDeleted = r.IsDeleted

	if r.Children != nil && len(r.Children) > 0 {
		m.Children = make([]Resource, len(r.Children))
		for i, child := range r.Children {
			if err := m.Children[i].FromDomain(child); err != nil {
				return fmt.Errorf("parsing child: %w", err)
			}
		}
	}

	return nil
}

// ToDomain transforms model into *domain.Provider
func (m *Resource) ToDomain() (*domain.Resource, error) {
	r := &domain.Resource{
		ID:           m.ID.String(),
		ParentID:     m.ParentID,
		ProviderType: m.ProviderType,
		ProviderURN:  m.ProviderURN,
		Type:         m.Type,
		URN:          m.URN,
		Name:         m.Name,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		IsDeleted:    m.IsDeleted,
	}

	if m.Details != nil {
		if err := json.Unmarshal(m.Details, &r.Details); err != nil {
			return nil, err
		}
	}

	if m.Labels != nil {
		if err := json.Unmarshal(m.Labels, &r.Labels); err != nil {
			return nil, err
		}
	}

	if m.Children != nil && len(m.Children) > 0 {
		r.Children = make([]*domain.Resource, len(m.Children))
		for i, child := range m.Children {
			child, err := child.ToDomain()
			if err != nil {
				return nil, fmt.Errorf("parsing child: %w", err)
			}
			r.Children[i] = child
		}
	}

	return r, nil
}
