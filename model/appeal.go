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
	Email         string
	Labels        datatypes.JSON

	Resource  Resource `gorm:"ForeignKey:ResourceID;References:ID"`
	Policy    Policy   `gorm:"ForeignKey:PolicyID,PolicyVersion;References:ID,Version"`
	Approvals []Approval

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

	var approvals []Approval
	if a.Approvals != nil {
		for _, approval := range a.Approvals {
			m := new(Approval)
			if err := m.FromDomain(approval); err != nil {
				return err
			}
			approvals = append(approvals, *m)
		}
	}

	m.ID = a.ID
	m.ResourceID = a.ResourceID
	m.PolicyID = a.PolicyID
	m.PolicyVersion = a.PolicyVersion
	m.Status = a.Status
	m.Email = a.Email
	m.Labels = datatypes.JSON(labels)
	m.Approvals = approvals
	m.CreatedAt = a.CreatedAt
	m.UpdatedAt = a.UpdatedAt

	return nil
}

// ToDomain transforms model into *domain.Appeal
func (m *Appeal) ToDomain() (*domain.Appeal, error) {
	var labels map[string]interface{}
	if err := json.Unmarshal(m.Labels, &labels); err != nil {
		return nil, err
	}

	var approvals []*domain.Approval
	if m.Approvals != nil {
		for _, a := range m.Approvals {
			approval, err := a.ToDomain()
			if err != nil {
				return nil, err
			}
			approvals = append(approvals, approval)
		}
	}

	return &domain.Appeal{
		ID:            m.ID,
		ResourceID:    m.ResourceID,
		PolicyID:      m.PolicyID,
		PolicyVersion: m.PolicyVersion,
		Status:        m.Status,
		Email:         m.Email,
		Labels:        labels,
		Approvals:     approvals,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}
