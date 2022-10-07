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
	ID           string         `db:"id"`
	Version      uint           `db:"version"`
	Description  string         `db:"description"`
	Steps        datatypes.JSON `db:"steps"`
	AppealConfig datatypes.JSON `db:"appeal_config"`
	Labels       datatypes.JSON `db:"labels"`
	Requirements datatypes.JSON `db:"requirements"`
	IAM          datatypes.JSON `db:"iam"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	DeletedAt    gorm.DeletedAt `db:"deleted_at"`
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
	m.Description = p.Description
	m.Steps = datatypes.JSON(steps)
	m.AppealConfig = datatypes.JSON(appeal)
	m.Labels = datatypes.JSON(labels)
	m.Requirements = datatypes.JSON(requirements)
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
		Version:      m.Version,
		Description:  m.Description,
		Steps:        steps,
		AppealConfig: appealConfig,
		Labels:       labels,
		Requirements: requirements,
		IAM:          iam,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}
