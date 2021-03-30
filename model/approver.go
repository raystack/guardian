package model

import (
	"time"

	"github.com/odpf/guardian/domain"
	"gorm.io/gorm"
)

// Approver database model
type Approver struct {
	ID         uint `gorm:"autoIncrement;uniqueIndex"`
	ApprovalID uint
	AppealID   uint   `gorm:"index"`
	Email      string `gorm:"index"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// FromDomain transforms *domain.Approver values into the model
func (m *Approver) FromDomain(a *domain.Approver) error {
	m.ID = a.ID
	m.ApprovalID = a.ApprovalID
	m.AppealID = a.AppealID
	m.Email = a.Email
	m.CreatedAt = a.CreatedAt
	m.UpdatedAt = a.UpdatedAt

	return nil
}

// ToDomain transforms model into *domain.Approver
func (m *Approver) ToDomain() (*domain.Approver, error) {
	return &domain.Approver{
		ID:         m.ID,
		ApprovalID: m.ApprovalID,
		AppealID:   m.AppealID,
		Email:      m.Email,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}, nil
}
