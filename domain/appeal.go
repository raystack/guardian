package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/goto/guardian/pkg/evaluator"
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
	PermanentDurationLabel   = "Permanent"
)

var (
	ErrApproverInvalidType = errors.New("invalid approver type, expected an email string or array of email string")
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
	Description   string                 `json:"description" yaml:"description"`

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

	duration, err := a.GetDuration()
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

func (a *Appeal) GetDuration() (time.Duration, error) {
	if a.IsDurationEmpty() {
		return 0 * time.Second, nil
	}

	duration, err := time.ParseDuration(a.Options.Duration)
	if err != nil {
		return 0 * time.Second, err
	}

	return duration, nil
}

func (a *Appeal) IsDurationEmpty() bool {
	return a.Options == nil || a.Options.Duration == "" || a.Options.Duration == "0h"
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

func (a *Appeal) ApplyPolicy(p *Policy) error {
	approvals := []*Approval{}
	for i, step := range p.Steps {
		approval, err := step.ToApproval(a, p, i)
		if err != nil {
			return err
		}
		approvals = append(approvals, approval)
	}

	a.Approvals = approvals
	a.Init(p)
	a.Policy = p

	return nil
}

func (a *Appeal) AdvanceApproval(policy *Policy) error {
	if policy == nil {
		return fmt.Errorf("appeal has no policy")
	}

	stepNameIndex := map[string]int{}
	for i, s := range policy.Steps {
		stepNameIndex[s.Name] = i
	}

	for i, approval := range a.Approvals {
		if approval.Status == ApprovalStatusRejected {
			break
		}
		if approval.Status == ApprovalStatusPending {
			stepConfig := policy.Steps[approval.Index]

			appealMap, err := structToMap(a)
			if err != nil {
				return fmt.Errorf("parsing appeal struct to map: %w", err)
			}

			if stepConfig.When != "" {
				v, err := evaluator.Expression(stepConfig.When).EvaluateWithVars(map[string]interface{}{
					"appeal": appealMap,
				})
				if err != nil {
					return err
				}

				isFalsy := reflect.ValueOf(v).IsZero()
				if isFalsy {
					approval.Status = ApprovalStatusSkipped
					if i < len(a.Approvals)-1 {
						a.Approvals[i+1].Status = ApprovalStatusPending
					}
				}
			}

			if approval.Status != ApprovalStatusSkipped && stepConfig.Strategy == ApprovalStepStrategyAuto {
				v, err := evaluator.Expression(stepConfig.ApproveIf).EvaluateWithVars(map[string]interface{}{
					"appeal": appealMap,
				})
				if err != nil {
					return err
				}

				isFalsy := reflect.ValueOf(v).IsZero()
				if isFalsy {
					if stepConfig.AllowFailed {
						approval.Status = ApprovalStatusSkipped
						if i+1 <= len(a.Approvals)-1 {
							a.Approvals[i+1].Status = ApprovalStatusPending
						}
					} else {
						approval.Status = ApprovalStatusRejected
						approval.Reason = stepConfig.RejectionReason
						a.Status = AppealStatusRejected
					}
				} else {
					approval.Status = ApprovalStatusApproved
					if i+1 <= len(a.Approvals)-1 {
						a.Approvals[i+1].Status = ApprovalStatusPending
					}
				}
			}
		}
		if i == len(a.Approvals)-1 && (approval.Status == ApprovalStatusSkipped || approval.Status == ApprovalStatusApproved) {
			a.Status = AppealStatusApproved
		}
	}

	return nil
}

func structToMap(item interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	if item != nil {
		jsonString, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(jsonString, &result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func uniqueSlice(arr []string) []string {
	keys := map[string]bool{}
	result := []string{}

	for _, v := range arr {
		if _, exist := keys[v]; !exist {
			result = append(result, v)
			keys[v] = true
		}
	}
	return result
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
