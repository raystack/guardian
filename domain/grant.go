package domain

import (
	"errors"
	"time"
)

type GrantStatus string

const (
	GrantStatusActive   GrantStatus = "active"
	GrantStatusInactive GrantStatus = "inactive"
)

type Grant struct {
	ID             string      `json:"id" yaml:"id"`
	Status         GrantStatus `json:"status" yaml:"status"`
	AccountID      string      `json:"account_id" yaml:"account_id"`
	AccountType    string      `json:"account_type" yaml:"account_type"`
	ResourceID     string      `json:"resource_id" yaml:"resource_id"`
	Role           string      `json:"role" yaml:"role"`
	Permissions    []string    `json:"permissions" yaml:"permissions"`
	ExpirationDate *time.Time  `json:"expiration_date" yaml:"expiration_date"`
	AppealID       string      `json:"appeal_id" yaml:"appeal_id"`
	RevokedBy      string      `json:"revoked_by,omitempty" yaml:"revoked_by,omitempty"`
	RevokedAt      *time.Time  `json:"revoked_at,omitempty" yaml:"revoked_at,omitempty"`
	RevokeReason   string      `json:"revoke_reason,omitempty" yaml:"revoke_reason,omitempty"`
	CreatedBy      string      `json:"created_by" yaml:"created_by"`
	CreatedAt      time.Time   `json:"created_at" yaml:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" yaml:"updated_at"`

	Resource *Resource `json:"resource,omitempty" yaml:"resource,omitempty"`
	Appeal   *Appeal   `json:"appeal,omitempty" yaml:"appeal,omitempty"`
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
	OrderBy                   []string
	ExpirationDateLessThan    time.Time
	ExpirationDateGreaterThan time.Time
}

type RevokeGrantsFilter struct {
	AccountIDs    []string `validate:"omitempty,required"`
	ProviderTypes []string `validate:"omitempty,min=1"`
	ProviderURNs  []string `validate:"omitempty,min=1"`
	ResourceTypes []string `validate:"omitempty,min=1"`
	ResourceURNs  []string `validate:"omitempty,min=1"`
}
