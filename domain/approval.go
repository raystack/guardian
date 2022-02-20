package domain

import (
	"errors"
	"time"
)

const (
	ApprovalStatusPending  = "pending"
	ApprovalStatusBlocked  = "blocked"
	ApprovalStatusSkipped  = "skipped"
	ApprovalStatusApproved = "approved"
	ApprovalStatusRejected = "rejected"
)

type Approval struct {
	ID            string  `json:"id" yaml:"id"`
	Name          string  `json:"name" yaml:"name"`
	Index         int     `json:"-" yaml:"-"`
	AppealID      string  `json:"appeal_id" yaml:"appeal_id"`
	Status        string  `json:"status" yaml:"status"`
	Actor         *string `json:"actor" yaml:"actor"`
	Reason        string  `json:"reason,omitempty" yaml:"reason,omitempty"`
	PolicyID      string  `json:"policy_id" yaml:"policy_id"`
	PolicyVersion uint    `json:"policy_version" yaml:"policy_version"`

	Approvers []string `json:"approvers,omitempty" yaml:"approvers,omitempty"`
	Appeal    *Appeal  `json:"appeal,omitempty" yaml:"appeal,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

func (a *Approval) Init(policy *Policy, index int, approvers []string) error {
	if index > len(policy.Steps)-1 {
		return errors.New("approval step index out of range")
	}
	approvalStep := policy.Steps[index]

	a.Status = ApprovalStatusPending
	if index > 0 {
		a.Status = ApprovalStatusBlocked
	}
	a.Index = index
	a.Name = approvalStep.Name
	a.PolicyID = policy.ID
	a.PolicyVersion = policy.Version
	a.Approvers = approvers
	return nil
}

func (a *Approval) Approve() {
	a.Status = ApprovalStatusApproved
}

func (a *Approval) Reject() {
	a.Status = ApprovalStatusRejected
}

func (a *Approval) Skip() {
	a.Status = ApprovalStatusSkipped
}

func (a *Approval) IsManualApproval() bool {
	return len(a.Approvers) > 0
}

type ListApprovalsFilter struct {
	AccountID string   `mapstructure:"account_id" validate:"omitempty,required"`
	CreatedBy string   `mapstructure:"created_by" validate:"omitempty,required"`
	Statuses  []string `mapstructure:"statuses" validate:"omitempty,min=1"`
	OrderBy   []string `mapstructure:"order_by" validate:"omitempty,min=1"`
}

type ApprovalService interface {
	BulkInsert([]*Approval) error
	ListApprovals(*ListApprovalsFilter) ([]*Approval, error)
	AdvanceApproval(appeal *Appeal) error
}
