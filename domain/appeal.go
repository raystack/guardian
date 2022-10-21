package domain

import (
	"fmt"
	"time"
)

const (
	AppealActionNameApprove = "approve"
	AppealActionNameReject  = "reject"

	AppealStatusPending  = "pending"
	AppealStatusCanceled = "canceled"
	AppealStatusApproved = "approved"
	AppealStatusRejected = "rejected"

	SystemActorName = "system"

	DefaultAppealAccountType = "user"
)

// AppealOptions
type AppealOptions struct {
	ExpirationDate *time.Time `json:"expiration_date,omitempty" yaml:"expiration_date,omitempty"`
	Duration       string     `json:"duration" yaml:"duration"`
}

// Appeal struct
type Appeal struct {
	ID            string                 `json:"id" yaml:"id"`
	ResourceID    string                 `json:"resource_id" yaml:"resource_id"`
	PolicyID      string                 `json:"policy_id" yaml:"policy_id"`
	PolicyVersion uint                   `json:"policy_version" yaml:"policy_version"`
	Status        string                 `json:"status" yaml:"status"`
	AccountID     string                 `json:"account_id" yaml:"account_id"`
	AccountType   string                 `json:"account_type" yaml:"account_type" default:"user"`
	CreatedBy     string                 `json:"created_by" yaml:"created_by"`
	Creator       interface{}            `json:"creator" yaml:"creator"`
	Role          string                 `json:"role" yaml:"role"`
	Permissions   []string               `json:"permissions" yaml:"permissions"`
	Options       *AppealOptions         `json:"options" yaml:"options"`
	Details       map[string]interface{} `json:"details" yaml:"details"`
	Labels        map[string]string      `json:"labels" yaml:"labels"`

	Policy    *Policy     `json:"-" yaml:"-"`
	Resource  *Resource   `json:"resource,omitempty" yaml:"resource,omitempty"`
	Approvals []*Approval `json:"approvals,omitempty" yaml:"approvals,omitempty"`
	Grant     *Grant      `json:"grant,omitempty" yaml:"grant,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

func (a *Appeal) GetNextPendingApproval() *Approval {
	for _, approval := range a.Approvals {
		if approval.Status == ApprovalStatusPending && approval.IsManualApproval() {
			return approval
		}
	}
	return nil
}

func (a *Appeal) Init(policy *Policy) {
	a.Status = AppealStatusPending
	a.PolicyID = policy.ID
	a.PolicyVersion = policy.Version
}

func (a *Appeal) Cancel() {
	a.Status = AppealStatusCanceled
}

func (a *Appeal) Approve() error {
	a.Status = AppealStatusApproved

	if a.Options == nil || a.Options.Duration == "" {
		return nil
	}

	duration, err := time.ParseDuration(a.Options.Duration)
	if err != nil {
		return err
	}

	// for permanent access duration is equal to zero
	if duration == 0*time.Second {
		return nil
	}

	expirationDate := time.Now().Add(duration)
	a.Options.ExpirationDate = &expirationDate
	return nil
}

func (a *Appeal) Reject() {
	a.Status = AppealStatusRejected
}

func (a *Appeal) SetDefaults() {
	if a.AccountType == "" {
		a.AccountType = DefaultAppealAccountType
	}
}

func (a *Appeal) GetApproval(id string) *Approval {
	for _, approval := range a.Approvals {
		if approval.ID == id || approval.Name == id {
			return approval
		}
	}
	return nil
}

func (a Appeal) ToGrant() (*Grant, error) {
	grant := &Grant{
		Status:      GrantStatusActive,
		AccountID:   a.AccountID,
		AccountType: a.AccountType,
		ResourceID:  a.ResourceID,
		Role:        a.Role,
		Permissions: a.Permissions,
		AppealID:    a.ID,
		CreatedBy:   a.CreatedBy,
	}

	if a.Options != nil && a.Options.Duration != "" {
		duration, err := time.ParseDuration(a.Options.Duration)
		if err != nil {
			return nil, fmt.Errorf("parsing duration %q: %w", a.Options.Duration, err)
		}
		if duration == 0 {
			grant.IsPermanent = true
		} else {
			expDate := time.Now().Add(duration)
			grant.ExpirationDate = &expDate
		}
	} else {
		grant.IsPermanent = true
	}

	return grant, nil
}

type ApprovalActionType string

const (
	ApprovalActionApprove ApprovalActionType = "approve"
	ApprovalActionReject  ApprovalActionType = "reject"
)

type ApprovalAction struct {
	AppealID     string `validate:"required" json:"appeal_id"`
	ApprovalName string `validate:"required" json:"approval_name"`
	Actor        string `validate:"email" json:"actor"`
	Action       string `validate:"required,oneof=approve reject" json:"action"`
	Reason       string `json:"reason"`
}

type ListAppealsFilter struct {
	CreatedBy                 string    `mapstructure:"created_by" validate:"omitempty,required"`
	AccountID                 string    `mapstructure:"account_id" validate:"omitempty,required"`
	AccountIDs                []string  `mapstructure:"account_ids" validate:"omitempty,required"`
	ResourceID                string    `mapstructure:"resource_id" validate:"omitempty,required"`
	Role                      string    `mapstructure:"role" validate:"omitempty,required"`
	Statuses                  []string  `mapstructure:"statuses" validate:"omitempty,min=1"`
	ExpirationDateLessThan    time.Time `mapstructure:"expiration_date_lt" validate:"omitempty,required"`
	ExpirationDateGreaterThan time.Time `mapstructure:"expiration_date_gt" validate:"omitempty,required"`
	ProviderTypes             []string  `mapstructure:"provider_types" validate:"omitempty,min=1"`
	ProviderURNs              []string  `mapstructure:"provider_urns" validate:"omitempty,min=1"`
	ResourceTypes             []string  `mapstructure:"resource_types" validate:"omitempty,min=1"`
	ResourceURNs              []string  `mapstructure:"resource_urns" validate:"omitempty,min=1"`
	OrderBy                   []string  `mapstructure:"order_by" validate:"omitempty,min=1"`
}
