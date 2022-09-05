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
	ID             string
	Status         GrantStatus
	AccountID      string
	AccountType    string
	ResourceID     string
	Role           string
	Permissions    []string
	ExpirationDate *time.Time
	AppealID       string
	RevokedBy      string
	RevokedAt      *time.Time
	RevokeReason   string
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Resource *Resource `json:"resource" yaml:"resource"`
	Appeal   *Appeal   `json:"appeal" yaml:"appeal"`
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
