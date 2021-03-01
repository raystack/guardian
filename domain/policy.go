package domain

import (
	"net/http"
	"time"
)

// MatchCondition is for determining the requirement of the condition
type MatchCondition struct {
	Eq interface{} `json:"eq" yaml:"eq"`
}

// Condition gets evaluated to determine the approval step resolution whether it is success or failed
type Condition struct {
	Field string          `json:"field" yaml:"field"`
	Match *MatchCondition `json:"match" yaml:"match" validate:"required"`
}

// Step is an individual process within an approval flow
type Step struct {
	Name        string       `json:"name" yaml:"name"`
	Description string       `json:"description" yaml:"description"`
	Conditions  []*Condition `json:"conditions" yaml:"conditions" validate:"required_without=Approvers,required"`
	AllowFailed bool         `json:"allow_failed" yaml:"allow_failed"`

	Dependencies []string `json:"dependencies" yaml:"dependencies"`
	Approvers    string   `json:"approvers" yaml:"approvers" validate:"required_without=Conditions"`
}

// Policy is the approval policy configuration
type Policy struct {
	ID          string                 `json:"id" yaml:"id" validate:"required"`
	Version     uint                   `json:"version" yaml:"version" validate:"required"`
	Description string                 `json:"description" yaml:"description"`
	Steps       []*Step                `json:"steps" yaml:"steps" validate:"required"`
	Labels      map[string]interface{} `json:"labels" yaml:"labels"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PolicyRepository interface
type PolicyRepository interface {
	Create(*Policy) error
	Find() ([]*Policy, error)
	GetOne(id string, version int) (*Policy, error)
}

// PolicyService interface
type PolicyService interface {
	Create(*Policy) error
	Find() ([]*Policy, error)
	GetOne(id string, version int) (*Policy, error)
}

// PolicyHandler interface
type PolicyHandler interface {
	Create(http.ResponseWriter, *http.Request)
}
