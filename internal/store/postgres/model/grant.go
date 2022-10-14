package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/odpf/guardian/domain"
)

var GrantColumns = ColumnNames{
	"id",
	"status",
	"status_in_provider",
	"account_id",
	"account_type",
	"resource_id",
	"role",
	"permissions",
	"is_permanent",
	"expiration_date",
	"appeal_id",
	"source",
	"revoked_by",
	"revoked_at",
	"revoke_reason",
	"owner",
	"created_at",
	"updated_at",
	"deleted_at",
}

type Grant struct {
	ID               uuid.UUID      `db:"id"`
	Status           string         `db:"status"`
	StatusInProvider string         `db:"status_in_provider"`
	AccountID        string         `db:"account_id"`
	AccountType      string         `db:"account_type"`
	ResourceID       string         `db:"resource_id"`
	Role             string         `db:"role"`
	Permissions      pq.StringArray `db:"permissions"`
	IsPermanent      bool           `db:"is_permanent"`
	ExpirationDate   sql.NullTime   `db:"expiration_date"`
	AppealID         *string        `db:"appeal_id"`
	Source           string         `db:"source"`
	RevokedBy        string         `db:"revoked_by"`
	RevokedAt        sql.NullTime   `db:"revoked_at"`
	RevokeReason     string         `db:"revoke_reason"`
	Owner            string         `db:"owner"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
	DeletedAt        sql.NullTime   `db:"deleted_at"`

	Resource *Resource `db:"resource"`
	Appeal   *Appeal   `db:"appeal"`
}

func (m *Grant) Values() []interface{} {
	return []interface{}{
		m.ID,
		m.Status,
		m.StatusInProvider,
		m.AccountID,
		m.AccountType,
		m.ResourceID,
		m.Role,
		m.Permissions,
		m.IsPermanent,
		m.ExpirationDate,
		m.AppealID,
		m.Source,
		m.RevokedBy,
		m.RevokedAt,
		m.RevokeReason,
		m.Owner,
		m.CreatedAt,
		m.UpdatedAt,
		m.DeletedAt,
	}
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
		m.ExpirationDate = sql.NullTime{Time: *g.ExpirationDate, Valid: true}
	}

	if g.RevokedAt != nil {
		m.RevokedAt = sql.NullTime{Time: *g.RevokedAt, Valid: true}
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
	m.Owner = g.CreatedBy
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

	if m.ExpirationDate.Valid {
		d := m.ExpirationDate.Time
		grant.ExpirationDate = &d
	}
	if m.RevokedAt.Valid {
		grant.RevokedAt = &m.RevokedAt.Time
	}

	return grant, nil
}
