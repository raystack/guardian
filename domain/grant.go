package domain

import (
	"errors"
	"sort"
	"strings"
	"time"
)

type GrantStatus string

const (
	GrantStatusActive   GrantStatus = "active"
	GrantStatusInactive GrantStatus = "inactive"
)

type GrantSource string

const (
	GrantSourceAppeal GrantSource = "appeal"
	GrantSourceImport GrantSource = "import"
)

type Grant struct {
	ID               string      `json:"id" yaml:"id"`
	Status           GrantStatus `json:"status" yaml:"status"`
	StatusInProvider GrantStatus `json:"status_in_provider" yaml:"status_in_provider"`
	AccountID        string      `json:"account_id" yaml:"account_id"`
	AccountType      string      `json:"account_type" yaml:"account_type"`
	ResourceID       string      `json:"resource_id" yaml:"resource_id"`
	Role             string      `json:"role" yaml:"role"`
	Permissions      []string    `json:"permissions" yaml:"permissions"`
	IsPermanent      bool        `json:"is_permanent" yaml:"is_permanent"`
	ExpirationDate   *time.Time  `json:"expiration_date" yaml:"expiration_date"`
	AppealID         string      `json:"appeal_id" yaml:"appeal_id"`
	Source           GrantSource `json:"source" yaml:"source"`
	RevokedBy        string      `json:"revoked_by,omitempty" yaml:"revoked_by,omitempty"`
	RevokedAt        *time.Time  `json:"revoked_at,omitempty" yaml:"revoked_at,omitempty"`
	RevokeReason     string      `json:"revoke_reason,omitempty" yaml:"revoke_reason,omitempty"`
	CreatedBy        string      `json:"created_by" yaml:"created_by"`
	Owner            string      `json:"owner" yaml:"owner"`
	CreatedAt        time.Time   `json:"created_at" yaml:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at" yaml:"updated_at"`

	Resource *Resource `json:"resource,omitempty" yaml:"resource,omitempty"`
	Appeal   *Appeal   `json:"appeal,omitempty" yaml:"appeal,omitempty"`
}

func (g Grant) PermissionsKey() string {
	permissions := make([]string, len(g.Permissions))
	copy(permissions, g.Permissions)
	sort.Strings(permissions)
	return strings.Join(permissions, ";")
}

func (g Grant) IsEligibleForExtension(extensionDurationRule time.Duration) bool {
	if g.ExpirationDate != nil && !g.ExpirationDate.IsZero() {
		return time.Until(*g.ExpirationDate) <= extensionDurationRule
	}
	return true
}

func (g *Grant) Revoke(actor, reason string) error {
	if g == nil {
		return errors.New("grant is nil")
	}
	if actor == "" {
		return errors.New("actor shouldn't be empty")
	}

	g.Status = GrantStatusInactive
	g.StatusInProvider = GrantStatusInactive
	g.RevokedBy = actor
	g.RevokeReason = reason
	now := time.Now()
	g.RevokedAt = &now
	return nil
}

type ListGrantsFilter struct {
	Statuses                  []string
	AccountIDs                []string
	AccountTypes              []string
	ResourceIDs               []string
	Roles                     []string
	Permissions               []string
	ProviderTypes             []string
	ProviderURNs              []string
	ResourceTypes             []string
	ResourceURNs              []string
	CreatedBy                 string
	Owner                     string
	OrderBy                   []string
	ExpirationDateLessThan    time.Time
	ExpirationDateGreaterThan time.Time
	IsPermanent               *bool
}

type RevokeGrantsFilter struct {
	AccountIDs    []string `validate:"omitempty,required"`
	ProviderTypes []string `validate:"omitempty,min=1"`
	ProviderURNs  []string `validate:"omitempty,min=1"`
	ResourceTypes []string `validate:"omitempty,min=1"`
	ResourceURNs  []string `validate:"omitempty,min=1"`
}

type AccessEntry struct {
	AccountID   string
	AccountType string
	Permission  string
}

func (ae AccessEntry) ToGrant(resource Resource) Grant {
	return Grant{
		ResourceID:       resource.ID,
		Status:           GrantStatusActive,
		StatusInProvider: GrantStatusActive,
		AccountID:        ae.AccountID,
		AccountType:      ae.AccountType,
		CreatedBy:        SystemActorName,
		Permissions:      []string{ae.Permission},
		Source:           GrantSourceImport,
		IsPermanent:      true,
	}
}

// MapResourceAccess is list of UserAccess grouped by resource urn
type MapResourceAccess map[string][]AccessEntry
