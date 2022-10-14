package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"github.com/odpf/guardian/domain"
)

var AppealColumns = ColumnNames{
	"id",
	"resource_id",
	"policy_id",
	"policy_version",
	"status",
	"account_id",
	"account_type",
	"created_by",
	"creator",
	"role",
	"permissions",
	"options",
	"labels",
	"details",
	"created_at",
	"updated_at",
}

// Appeal database model
type Appeal struct {
	ID            uuid.UUID      `db:"id"`
	ResourceID    string         `db:"resource_id"`
	PolicyID      string         `db:"policy_id"`
	PolicyVersion uint           `db:"policy_version"`
	Status        string         `db:"status"`
	AccountID     string         `db:"account_id"`
	AccountType   string         `db:"account_type"`
	CreatedBy     string         `db:"created_by"`
	Creator       types.JSONText `db:"creator"`
	Role          string         `db:"role"`
	Permissions   pq.StringArray `db:"permissions"`
	Options       types.JSONText `db:"options"`
	Labels        types.JSONText `db:"labels"`
	Details       types.JSONText `db:"details"`

	Resource  *Resource   `db:"resource"`
	Policy    Policy      `db:"-"`
	Approvals []*Approval `db:"approvals"`
	Grant     *Grant      `db:"grant"`

	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func (m *Appeal) Values() []interface{} {
	return []interface{}{
		m.ID,
		m.ResourceID,
		m.PolicyID,
		m.PolicyVersion,
		m.Status,
		m.AccountID,
		m.AccountType,
		m.CreatedBy,
		m.Creator,
		m.Role,
		m.Permissions,
		m.Options,
		m.Labels,
		m.Details,
		m.CreatedAt,
		m.UpdatedAt,
	}
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

	creator, err := json.Marshal(a.Creator)
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

	if a.Grant != nil {
		grant := new(Grant)
		if err := grant.FromDomain(*a.Grant); err != nil {
			return fmt.Errorf("parsing grant: %w", err)
		}
		m.Grant = grant
	}

	var id uuid.UUID
	if a.ID != "" {
		uuid, err := uuid.Parse(a.ID)
		if err != nil {
			return fmt.Errorf("parsing uuid: %w", err)
		}
		id = uuid
	}
	m.ID = id
	m.ResourceID = a.ResourceID
	m.PolicyID = a.PolicyID
	m.PolicyVersion = a.PolicyVersion
	m.Status = a.Status
	m.AccountID = a.AccountID
	m.AccountType = a.AccountType
	m.CreatedBy = a.CreatedBy
	m.Creator = creator
	m.Role = a.Role
	m.Permissions = a.Permissions
	m.Options = options
	m.Labels = labels
	m.Details = details
	m.Approvals = approvals
	m.CreatedAt = a.CreatedAt
	m.UpdatedAt = a.UpdatedAt

	return nil
}

// ToDomain transforms model into *domain.Appeal
func (m *Appeal) ToDomain() (*domain.Appeal, error) {
	var labels map[string]string
	if err := json.Unmarshal(m.Labels, &labels); err != nil {
		return nil, fmt.Errorf("parsing labels: %w", err)
	}

	var options *domain.AppealOptions
	if m.Options != nil {
		if err := json.Unmarshal(m.Options, &options); err != nil {
			return nil, fmt.Errorf("parsing options: %w", err)
		}
	}

	var details map[string]interface{}
	if m.Details != nil {
		if err := json.Unmarshal(m.Details, &details); err != nil {
			return nil, fmt.Errorf("parsing details: %w", err)
		}
	}

	var creator interface{}
	if m.Creator != nil {
		if err := json.Unmarshal(m.Creator, &creator); err != nil {
			return nil, fmt.Errorf("parsing creator: %w", err)
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

	var grant *domain.Grant
	if m.Grant != nil {
		a, err := m.Grant.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("parsing grant: %w", err)
		}
		grant = a
	}

	return &domain.Appeal{
		ID:            m.ID.String(),
		ResourceID:    m.ResourceID,
		PolicyID:      m.PolicyID,
		PolicyVersion: m.PolicyVersion,
		Status:        m.Status,
		AccountID:     m.AccountID,
		AccountType:   m.AccountType,
		CreatedBy:     m.CreatedBy,
		Creator:       creator,
		Role:          m.Role,
		Permissions:   []string(m.Permissions),
		Options:       options,
		Details:       details,
		Labels:        labels,
		Approvals:     approvals,
		Resource:      resource,
		Grant:         grant,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}
