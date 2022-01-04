package model

import (
	"encoding/json"
	"time"

	"github.com/odpf/guardian/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Policy is the database model for policy
type Policy struct {
	ID          string `gorm:"primaryKey"`
	Version     uint   `gorm:"primaryKey"`
	Description string
	Steps       datatypes.JSON
	Labels      datatypes.JSON
	IAM         datatypes.JSON
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the table name
func (Policy) TableName() string {
	return "policies"
}

// FromDomain transforms *domain.Policy values into the model
func (m *Policy) FromDomain(p *domain.Policy) error {
	steps, err := json.Marshal(p.Steps)
	if err != nil {
		return err
	}

	labels, err := json.Marshal(p.Labels)
	if err != nil {
		return err
	}

	iam, err := json.Marshal(p.IAM)
	if err != nil {
		return err
	}

	m.ID = p.ID
	m.Version = p.Version
	m.Description = p.Description
	m.Steps = datatypes.JSON(steps)
	m.Labels = datatypes.JSON(labels)
	m.IAM = datatypes.JSON(iam)
	m.CreatedAt = p.CreatedAt
	m.UpdatedAt = p.UpdatedAt

	return nil
}

// ToDomain transforms model into *domain.Policy
func (m *Policy) ToDomain() (*domain.Policy, error) {
	var steps []*domain.Step
	if err := json.Unmarshal(m.Steps, &steps); err != nil {
		return nil, err
	}

	var labels map[string]string
	if err := json.Unmarshal(m.Labels, &labels); err != nil {
		return nil, err
	}

	var iam *domain.IAMConfig
	if m.IAM != nil {
		if err := json.Unmarshal(m.IAM, &iam); err != nil {
			return nil, err
		}
	}

	return &domain.Policy{
		ID:          m.ID,
		Version:     m.Version,
		Description: m.Description,
		Steps:       steps,
		Labels:      labels,
		IAM:         iam,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}