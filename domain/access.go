package domain

import (
	"errors"
	"time"
)

type AccessStatus string

const (
	AccessStatusActive   AccessStatus = "active"
	AccessStatusInactive AccessStatus = "inactive"
)

type Access struct {
	ID             string
	Status         AccessStatus
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
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Resource *Resource `json:"resource" yaml:"resource"`
	Appeal   *Appeal   `json:"appeal" yaml:"appeal"`
}

func (a Access) IsEligibleForExtension(extensionDurationRule time.Duration) bool {
	if a.ExpirationDate != nil && !a.ExpirationDate.IsZero() {
		return time.Until(*a.ExpirationDate) <= extensionDurationRule
	}
	return true
}

func (a *Access) Revoke(actor, reason string) error {
	if a == nil {
		return errors.New("access is nil")
	}
	if actor == "" {
		return errors.New("actor shouldn't be empty")
	}

	a.Status = AccessStatusInactive
	a.RevokedBy = actor
	a.RevokeReason = reason
	now := time.Now()
	a.RevokedAt = &now
	return nil
}

type ListAccessesFilter struct {
	Statuses     []string
	AccountIDs   []string
	AccountTypes []string
	ResourceIDs  []string
	Permissions  []string
}
