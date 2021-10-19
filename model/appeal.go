package model

import (
	"encoding/json"
	"time"

	"github.com/odpf/guardian/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Appeal database model
type Appeal struct {
	ID            uint `gorm:"primaryKey"`
	ResourceID    uint
	PolicyID      string
	PolicyVersion uint
	Status        string
	AccountID     string
	Role          string
	Options       datatypes.JSON
	Labels        datatypes.JSON
	Details       datatypes.JSON

	RevokedBy    string
	RevokedAt    time.Time
	RevokeReason string

	Resource  *Resource `gorm:"ForeignKey:ResourceID;References:ID"`
	Policy    Policy    `gorm:"ForeignKey:PolicyID,PolicyVersion;References:ID,Version"`
	Approvals []*Approval

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// FromDomain transforms *domain.Appeal values into the model
func (m *Appeal) FromDomain(a *domain.Appeal) error {
	labels, err := json.Marshal(a.Labels)
	if err != nil {
		return err
	}

	options, err := json.Marshal(a.Options)
	if err != nil {
		return err
	}

	details, err := json.Marshal(a.Details)
	if err != nil {
		return err
	}

	var approvals []*Approval
	if a.Approvals != nil {
		for _, approval := range a.Approvals {
			m := new(Approval)
			if err := m.FromDomain(approval); err != nil {
				return err
			}
			approvals = append(approvals, m)
		}
	}

	if a.Resource != nil {
		r := new(Resource)
		if err := r.FromDomain(a.Resource); err != nil {
			return err
		}
		m.Resource = r
	}

	m.ID = a.ID
	m.ResourceID = a.ResourceID
	m.PolicyID = a.PolicyID
	m.PolicyVersion = a.PolicyVersion
	m.Status = a.Status
	m.AccountID = a.AccountID
	m.Role = a.Role
	m.Options = datatypes.JSON(options)
	m.Labels = datatypes.JSON(labels)
	m.Details = datatypes.JSON(details)
	m.Approvals = approvals
	m.CreatedAt = a.CreatedAt
	m.UpdatedAt = a.UpdatedAt

	return nil
}

// ToDomain transforms model into *domain.Appeal
func (m *Appeal) ToDomain() (*domain.Appeal, error) {
	var labels map[string]string
	if err := json.Unmarshal(m.Labels, &labels); err != nil {
		return nil, err
	}

	var options *domain.AppealOptions
	if m.Options != nil {
		if err := json.Unmarshal(m.Options, &options); err != nil {
			return nil, err
		}
	}

	var details map[string]interface{}
	if m.Details != nil {
		if err := json.Unmarshal(m.Details, &details); err != nil {
			return nil, err
		}
	}

	var approvals []*domain.Approval
	if m.Approvals != nil {
		for _, a := range m.Approvals {
			if a != nil {
				approval, err := a.ToDomain()
				if err != nil {
					return nil, err
				}
				approvals = append(approvals, approval)
			}
		}
	}

	var resource *domain.Resource
	if m.Resource != nil {
		r, err := m.Resource.ToDomain()
		if err != nil {
			return nil, err
		}
		resource = r
	}

	return &domain.Appeal{
		ID:            m.ID,
		ResourceID:    m.ResourceID,
		PolicyID:      m.PolicyID,
		PolicyVersion: m.PolicyVersion,
		Status:        m.Status,
		AccountID:     m.AccountID,
		Role:          m.Role,
		Options:       options,
		Details:       details,
		Labels:        labels,
		Approvals:     approvals,
		Resource:      resource,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}
