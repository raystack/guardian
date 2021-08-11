package model

import (
	"encoding/json"
	"time"

	"github.com/odpf/guardian/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Resource is the database model for resource
type Resource struct {
	ID           uint   `gorm:"autoIncrement;uniqueIndex"`
	ProviderType string `gorm:"primaryKey"`
	ProviderURN  string `gorm:"primaryKey"`
	Type         string `gorm:"primaryKey"`
	URN          string `gorm:"primaryKey"`
	Name         string
	Details      datatypes.JSON
	Labels       datatypes.JSON

	Provider Provider `gorm:"ForeignKey:ProviderType,ProviderURN;References:Type,URN"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
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

	m.ID = r.ID
	m.ProviderType = r.ProviderType
	m.ProviderURN = r.ProviderURN
	m.Type = r.Type
	m.URN = r.URN
	m.Name = r.Name
	m.Details = datatypes.JSON(details)
	m.Labels = datatypes.JSON(labels)
	m.CreatedAt = r.CreatedAt
	m.UpdatedAt = r.UpdatedAt

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
		ID:           m.ID,
		ProviderType: m.ProviderType,
		ProviderURN:  m.ProviderURN,
		Type:         m.Type,
		URN:          m.URN,
		Name:         m.Name,
		Details:      details,
		Labels:       labels,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}
