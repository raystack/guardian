package domain

import "time"

const (
	AppealStatusPending    = "pending"
	AppealStatusActive     = "active"
	AppealStatusRejected   = "rejected"
	AppealStatusTerminated = "terminated"
)

// Appeal struct
type Appeal struct {
	ID            uint                   `json:"id"`
	ResourceID    uint                   `json:"resource_id"`
	PolicyID      string                 `json:"policy_id"`
	PolicyVersion uint                   `json:"policy_version"`
	Status        string                 `json:"status"`
	User          string                 `json:"user"`
	Role          string                 `json:"role"`
	Labels        map[string]interface{} `json:"labels"`
	Resource      *Resource              `json:"resource,omitempty"`
	Approvals     []*Approval            `json:"approvals"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// AppealRepository interface
type AppealRepository interface {
	BulkInsert([]*Appeal) error
	GetByID(uint) (*Appeal, error)
}

// AppealService interface
type AppealService interface {
	Create([]*Appeal) error
	GetByID(uint) (*Appeal, error)
}
