package domain

import "time"

type Approval struct {
	ID            uint   `json:"id"`
	AppealID      uint   `json:"appeal_id"`
	Status        string `json:"status"`
	PolicyID      string `json:"policy_id"`
	PolicyVersion uint   `json:"policy_version"`

	Approvers []*Approver `json:"approvers"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ApprovalRepository interface {
	BulkInsert([]*Approval) error
}

type ApprovalService interface {
	BulkInsert([]*Approval) error
}
