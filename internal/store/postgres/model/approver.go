package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/goto/guardian/domain"
	"gorm.io/gorm"
)

// Approver database model
type Approver struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ApprovalID string
	AppealID   string `gorm:"index"`
	Email      string `gorm:"index"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// FromDomain transforms *domain.Approver values into the model
func (m *Approver) FromDomain(a *domain.Approver) error {
	var id uuid.UUID
	if a.ID != "" {
		uuid, err := uuid.Parse(a.ID)
		if err != nil {
			return fmt.Errorf("parsing uuid: %w", err)
		}
		id = uuid
	}
	m.ID = id
	m.ApprovalID = a.ApprovalID
	m.AppealID = a.AppealID
	m.Email = a.Email
	m.CreatedAt = a.CreatedAt
	m.UpdatedAt = a.UpdatedAt

	return nil
}

// ToDomain transforms model into *domain.Approver
func (m *Approver) ToDomain() *domain.Approver {
	return &domain.Approver{
		ID:         m.ID.String(),
		ApprovalID: m.ApprovalID,
		AppealID:   m.AppealID,
		Email:      m.Email,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
