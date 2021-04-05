package domain

import "time"

const (
	ApprovalStatusPending  = "pending"
	ApprovalStatusSkipped  = "skipped"
	ApprovalStatusApproved = "approved"
	ApprovalStatusRejected = "rejected"
)

type Approval struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	AppealID      uint   `json:"appeal_id"`
	Status        string `json:"status"`
	PolicyID      string `json:"policy_id"`
	PolicyVersion uint   `json:"policy_version"`

	Approvers []string `json:"approvers,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ApprovalRepository interface {
	BulkInsert([]*Approval) error
}

type ApprovalService interface {
	BulkInsert([]*Approval) error
}