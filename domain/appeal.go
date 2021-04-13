package domain

import "time"

const (
	AppealActionNameApprove = "approve"
	AppealActionNameReject  = "reject"

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

	Resource  *Resource   `json:"resource,omitempty"`
	Approvals []*Approval `json:"approvals,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ApprovalAction struct {
	AppealID     uint   `validate:"required"`
	ApprovalName string `validate:"required"`
	Actor        string `validate:"email"`
	Action       string `validate:"required,oneof=approve reject"`
}

// AppealRepository interface
type AppealRepository interface {
	BulkInsert([]*Appeal) error
	Find(map[string]interface{}) ([]*Appeal, error)
	GetByID(uint) (*Appeal, error)
	Update(*Appeal) error
}

// AppealService interface
type AppealService interface {
	Create([]*Appeal) error
	Find(map[string]interface{}) ([]*Appeal, error)
	GetByID(uint) (*Appeal, error)
	GetPendingApprovals(user string) ([]*Approval, error)
	MakeAction(ApprovalAction) (*Appeal, error)
}
