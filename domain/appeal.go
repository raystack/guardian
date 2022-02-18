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
	Options       *AppealOptions         `json:"options" yaml:"options"`
	Details       map[string]interface{} `json:"details" yaml:"details"`
	Labels        map[string]string      `json:"labels" yaml:"labels"`

	RevokedBy    string    `json:"revoked_by,omitempty" yaml:"revoked_by,omitempty"`
	RevokedAt    time.Time `json:"revoked_at,omitempty" yaml:"revoked_at,omitempty"`
	RevokeReason string    `json:"revoke_reason,omitempty" yaml:"revoke_reason,omitempty"`

	Policy    *Policy     `json:"-" yaml:"-"`
	Resource  *Resource   `json:"resource,omitempty" yaml:"resource,omitempty"`
	Approvals []*Approval `json:"approvals,omitempty" yaml:"approvals,omitempty"`

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

func (a *Appeal) Terminate() {
	a.Status = AppealStatusTerminated
}

func (a *Appeal) Activate() error {
	a.Status = AppealStatusActive

	if a.Options != nil && a.Options.Duration != "" {
		duration, err := time.ParseDuration(a.Options.Duration)
		if err != nil {
			return err
		}

		expirationDate := time.Now().Add(duration)
		a.Options.ExpirationDate = &expirationDate
	}

	return nil
}

func (a *Appeal) SetDefaults() {
	if a.AccountType == "" {
		a.AccountType = DefaultAppealAccountType
	}
}

type ApprovalAction struct {
	AppealID     string `validate:"required"`
	ApprovalName string `validate:"required"`
	Actor        string `validate:"email"`
	Action       string `validate:"required,oneof=approve reject"`
	Reason       string
}

type ListAppealsFilter struct {
	CreatedBy                 string    `mapstructure:"created_by" validate:"omitempty,required"`
	AccountID                 string    `mapstructure:"account_id" validate:"omitempty,required"`
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

// AppealService interface
type AppealService interface {
	Create([]*Appeal) error
	Find(*ListAppealsFilter) ([]*Appeal, error)
	GetByID(id string) (*Appeal, error)
	MakeAction(ApprovalAction) (*Appeal, error)
	Cancel(id string) (*Appeal, error)
	Revoke(id string, actor, reason string) (*Appeal, error)
}
