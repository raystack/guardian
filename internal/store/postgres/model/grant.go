package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/raystack/guardian/domain"
	"gorm.io/gorm"
)

type Grant struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Status           string
	StatusInProvider string
	AccountID        string
	AccountType      string
	ResourceID       string
	Role             string
	Permissions      pq.StringArray `gorm:"type:text[]"`
	IsPermanent      bool
	ExpirationDate   time.Time
	AppealID         *string
	Source           string
	RevokedBy        string
	RevokedAt        time.Time
	RevokeReason     string
	Owner            string
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	Resource *Resource `gorm:"ForeignKey:ResourceID;References:ID"`
	Appeal   *Appeal   `gorm:"ForeignKey:AppealID;References:ID"`
}

func (m *Grant) FromDomain(g domain.Grant) error {
	if g.ID != "" {
		uuid, err := uuid.Parse(g.ID)
		if err != nil {
			return fmt.Errorf("parsing uuid: %w", err)
		}
		m.ID = uuid
	}

	if g.Resource != nil {
		r := new(Resource)
		if err := r.FromDomain(g.Resource); err != nil {
			return fmt.Errorf("parsing resource: %w", err)
		}
		m.Resource = r
	}

	if g.AppealID != "" {
		m.AppealID = &g.AppealID
	}

	if g.Appeal != nil {
		appeal := new(Appeal)
		if err := appeal.FromDomain(g.Appeal); err != nil {
			return fmt.Errorf("parsing appeal: %w", err)
		}
		m.Appeal = appeal
	}

	if g.ExpirationDate != nil {
		m.ExpirationDate = *g.ExpirationDate
	}

	if g.RevokedAt != nil {
		m.RevokedAt = *g.RevokedAt
	}

	m.Status = string(g.Status)
	m.StatusInProvider = string(g.StatusInProvider)
	m.AccountID = g.AccountID
	m.AccountType = g.AccountType
	m.ResourceID = g.ResourceID
	m.Role = g.Role
	m.Permissions = pq.StringArray(g.Permissions)
	m.IsPermanent = g.IsPermanent
	m.Source = string(g.Source)
	m.RevokedBy = g.RevokedBy
	m.RevokeReason = g.RevokeReason
	m.Owner = g.Owner
	if m.Owner == "" {
		m.Owner = g.CreatedBy
	}
	m.CreatedAt = g.CreatedAt
	m.UpdatedAt = g.UpdatedAt
	return nil
}

func (m Grant) ToDomain() (*domain.Grant, error) {
	grant := &domain.Grant{
		ID:               m.ID.String(),
		Status:           domain.GrantStatus(m.Status),
		StatusInProvider: domain.GrantStatus(m.StatusInProvider),
		AccountID:        m.AccountID,
		AccountType:      m.AccountType,
		ResourceID:       m.ResourceID,
		Role:             m.Role,
		Permissions:      []string(m.Permissions),
		IsPermanent:      m.IsPermanent,
		Source:           domain.GrantSource(m.Source),
		RevokedBy:        m.RevokedBy,
		RevokeReason:     m.RevokeReason,
		CreatedBy:        m.Owner,
		Owner:            m.Owner,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}

	if m.AppealID != nil {
		grant.AppealID = *m.AppealID
	}

	if m.Resource != nil {
		r, err := m.Resource.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("parsing resource: %w", err)
		}
		grant.Resource = r
	}

	if m.Appeal != nil {
		a, err := m.Appeal.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("parsing appeal: %w", err)
		}
		grant.Appeal = a
	}

	if !m.ExpirationDate.IsZero() {
		grant.ExpirationDate = &m.ExpirationDate
	}
	if !m.RevokedAt.IsZero() {
		grant.RevokedAt = &m.RevokedAt
	}

	return grant, nil
}
