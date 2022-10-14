package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/guardian/domain"
)

var ApproverColumns = ColumnNames{
	"id",
	"approval_id",
	"appeal_id",
	"email",
	"created_at",
	"updated_at",
}

// Approver database model
type Approver struct {
	ID         uuid.UUID `db:"id"`
	ApprovalID string    `db:"approval_id"`
	AppealID   string    `db:"appeal_id"`
	Email      string    `db:"email"`

	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
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
