package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/raystack/guardian/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Policy is the database model for policy
type Policy struct {
	ID           string `gorm:"primaryKey"`
	Title        string
	NamespaceID  uuid.UUID `gorm:"type:uuid"`
	Version      uint      `gorm:"primaryKey"`
	Description  string
	Steps        datatypes.JSON
	AppealConfig datatypes.JSON
	Labels       datatypes.JSON
	Requirements datatypes.JSON
	IAM          datatypes.JSON
	CreatedBy    string
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
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

	requirements, err := json.Marshal(p.Requirements)
	if err != nil {
		return err
	}

	appeal, err := json.Marshal(p.AppealConfig)
	if err != nil {
		return err
	}

	iam, err := json.Marshal(p.IAM)
	if err != nil {
		return err
	}

	m.ID = p.ID
	m.Version = p.Version
	m.Title = p.Title
	m.Description = p.Description
	m.Steps = datatypes.JSON(steps)
	m.AppealConfig = datatypes.JSON(appeal)
	m.Labels = datatypes.JSON(labels)
	m.Requirements = datatypes.JSON(requirements)
	m.IAM = datatypes.JSON(iam)
	m.CreatedAt = p.CreatedAt
	m.UpdatedAt = p.UpdatedAt
	m.CreatedBy = p.CreatedBy

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

	var requirements []*domain.Requirement
	if m.Requirements != nil {
		if err := json.Unmarshal(m.Requirements, &requirements); err != nil {
			return nil, err
		}
	}

	var appealConfig *domain.PolicyAppealConfig
	if m.AppealConfig != nil {
		if err := json.Unmarshal(m.AppealConfig, &appealConfig); err != nil {
			return nil, err
		}
	}

	var iam *domain.IAMConfig
	if m.IAM != nil {
		if err := json.Unmarshal(m.IAM, &iam); err != nil {
			return nil, err
		}
	}

	return &domain.Policy{
		ID:           m.ID,
		Title:        m.Title,
		Version:      m.Version,
		Description:  m.Description,
		Steps:        steps,
		AppealConfig: appealConfig,
		Labels:       labels,
		Requirements: requirements,
		IAM:          iam,
		CreatedBy:    m.CreatedBy,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}
