package domain

import "time"

type AccessStatus string

const (
	AccessStatusActive  AccessStatus = "active"
	AccessStatusRevoked AccessStatus = "revoked"
)

type Access struct {
	ID             string
	Status         AccessStatus
	AccountID      string
	AccountType    string
	ResourceID     string
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

type ListAccessesFilter struct {
	Statuses     []string
	AccountIDs   []string
	AccountTypes []string
	ResourceIDs  []string
}
