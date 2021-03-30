package model

import (
	"time"

	"github.com/odpf/guardian/domain"
	"gorm.io/gorm"
)

// Approval database model
type Approval struct {
	ID            uint `gorm:"primaryKey"`
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
	m.ID = a.ID
	m.AppealID = a.AppealID
	m.Status = a.Status
	m.PolicyID = a.PolicyID
	m.PolicyVersion = a.PolicyVersion
	m.CreatedAt = a.CreatedAt
	m.UpdatedAt = a.UpdatedAt

	return nil
}

// ToDomain transforms model into *domain.Approval
func (m *Approval) ToDomain() (*domain.Approval, error) {
	return &domain.Approval{
		ID:            m.ID,
		AppealID:      m.AppealID,
		Status:        m.Status,
		PolicyID:      m.PolicyID,
		PolicyVersion: m.PolicyVersion,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}
