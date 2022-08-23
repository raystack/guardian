package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/odpf/guardian/domain"
	"gorm.io/gorm"
)

type Access struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Status         string
	AccountID      string
	AccountType    string
	ResourceID     string
	Role           string
	Permissions    pq.StringArray `gorm:"type:text[]"`
	ExpirationDate time.Time
	AppealID       string
	RevokedBy      string
	RevokedAt      time.Time
	RevokeReason   string
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	Resource *Resource `gorm:"ForeignKey:ResourceID;References:ID"`
	Appeal   *Appeal   `gorm:"ForeignKey:AppealID;References:ID"`
}

func (m *Access) FromDomain(a domain.Access) error {
	if a.ID != "" {
		uuid, err := uuid.Parse(a.ID)
		if err != nil {
			return fmt.Errorf("parsing uuid: %w", err)
		}
		m.ID = uuid
	}

	if a.Resource != nil {
		r := new(Resource)
		if err := r.FromDomain(a.Resource); err != nil {
			return fmt.Errorf("parsing resource: %w", err)
		}
		m.Resource = r
	}

	if a.Appeal != nil {
		appeal := new(Appeal)
		if err := appeal.FromDomain(a.Appeal); err != nil {
			return fmt.Errorf("parsing appeal: %w", err)
		}
		m.Appeal = appeal
	}

	if a.ExpirationDate != nil {
		m.ExpirationDate = *a.ExpirationDate
	}

	if a.RevokedAt != nil {
		m.RevokedAt = *a.RevokedAt
	}

	m.Status = string(a.Status)
	m.AccountID = a.AccountID
	m.AccountType = a.AccountType
	m.ResourceID = a.ResourceID
	m.Role = a.Role
	m.Permissions = pq.StringArray(a.Permissions)
	m.AppealID = a.AppealID
	m.RevokedBy = a.RevokedBy
	m.RevokeReason = a.RevokeReason
	m.CreatedAt = a.CreatedAt
	m.UpdatedAt = a.UpdatedAt
	return nil
}

func (m Access) ToDomain() (*domain.Access, error) {
	access := &domain.Access{
		ID:           m.ID.String(),
		Status:       domain.AccessStatus(m.Status),
		AccountID:    m.AccountID,
		AccountType:  m.AccountType,
		ResourceID:   m.ResourceID,
		Role:         m.Role,
		Permissions:  []string(m.Permissions),
		AppealID:     m.AppealID,
		RevokedBy:    m.RevokedBy,
		RevokeReason: m.RevokeReason,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}

	if m.Resource != nil {
		r, err := m.Resource.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("parsing resource: %w", err)
		}
		access.Resource = r
	}

	if m.Appeal != nil {
		a, err := m.Appeal.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("parsing appeal: %w", err)
		}
		access.Appeal = a
	}

	if !m.ExpirationDate.IsZero() {
		access.ExpirationDate = &m.ExpirationDate
	}
	if !m.RevokedAt.IsZero() {
		access.RevokedAt = &m.RevokedAt
	}

	return access, nil
}
