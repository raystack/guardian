package domain

import (
	"time"
)

const (
	AppealActionNameApprove = "approve"
	AppealActionNameReject  = "reject"

	AppealStatusPending    = "pending"
	AppealStatusCanceled   = "canceled"
	AppealStatusActive     = "active"
	AppealStatusRejected   = "rejected"
	AppealStatusTerminated = "terminated"

	SystemActorName = "system"
)

// AppealOptions
type AppealOptions struct {
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

// Appeal struct
type Appeal struct {
	ID            uint                   `json:"id"`
	ResourceID    uint                   `json:"resource_id"`
	PolicyID      string                 `json:"policy_id"`
	PolicyVersion uint                   `json:"policy_version"`
	Status        string                 `json:"status"`
	User          string                 `json:"user"`
	Role          string                 `json:"role"`
	Options       *AppealOptions         `json:"options"`
	Details       map[string]interface{} `json:"details"`
	Labels        map[string]string      `json:"labels"`

	RevokedBy    string    `json:"revoked_by"`
	RevokedAt    time.Time `json:"revoked_at"`
	RevokeReason string    `json:"revoke_reason"`

	Policy    *Policy     `json:"-"`
	Resource  *Resource   `json:"resource,omitempty"`
	Approvals []*Approval `json:"approvals,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *Appeal) GetNextPendingApproval() *Approval {
	for _, approval := range a.Approvals {
		if approval.Status == ApprovalStatusPending && approval.IsManualApproval() {
			return approval
		}
	}
	return nil
}

func (a *Appeal) Terminate() {
	a.Status = AppealStatusTerminated
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
	Find(map[string]interface{}) ([]*Appeal, error) // TODO: create ListAppealsFilter as the filter param type
	GetByID(uint) (*Appeal, error)
	Update(*Appeal) error
}

// AppealService interface
type AppealService interface {
	Create([]*Appeal) error
	Find(map[string]interface{}) ([]*Appeal, error)
	GetByID(uint) (*Appeal, error)
	MakeAction(ApprovalAction) (*Appeal, error)
	Cancel(uint) (*Appeal, error)
	Revoke(id uint, actor, reason string) (*Appeal, error)
}
