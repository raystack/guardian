package model

import (
	"time"

	"github.com/odpf/guardian/domain"
	"gorm.io/gorm"
)

// Approval database model
type Approval struct {
	ID            uint   `gorm:"primaryKey"`
	Name          string `gorm:"index"`
	AppealID      uint
	Status        string
	PolicyID      string
	PolicyVersion uint

	Approvers []Approver

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// FromDomain transforms *domain.Approval values into the model
func (m *Approval) FromDomain(a *domain.Approval) error {
	var approvers []Approver
	if a.Approvers != nil {
		for _, approver := range a.Approvers {
			m := new(Approver)
			if err := m.FromDomain(&domain.Approver{Email: approver}); err != nil {
				return err
			}
			approvers = append(approvers, *m)
		}
	}

	m.ID = a.ID
	m.Name = a.Name
	m.AppealID = a.AppealID
	m.Status = a.Status
	m.PolicyID = a.PolicyID
	m.PolicyVersion = a.PolicyVersion
	m.Approvers = approvers
	m.CreatedAt = a.CreatedAt
	m.UpdatedAt = a.UpdatedAt

	return nil
}

// ToDomain transforms model into *domain.Approval
func (m *Approval) ToDomain() (*domain.Approval, error) {
	var approvers []string
	if m.Approvers != nil {
		for _, a := range m.Approvers {
			approver, err := a.ToDomain()
			if err != nil {
				return nil, err
			}
			approvers = append(approvers, approver.Email)
		}
	}

	return &domain.Approval{
		ID:            m.ID,
		Name:          m.Name,
		AppealID:      m.AppealID,
		Status:        m.Status,
		PolicyID:      m.PolicyID,
		PolicyVersion: m.PolicyVersion,
		Approvers:     approvers,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}
