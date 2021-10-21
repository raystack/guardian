package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mcuadros/go-lookup"
)

const (
	ApproversKeyResource      = "$resource"
	ApproversKeyUserApprovers = "$user_approvers"
)

var (
	ErrInvalidConditionField = errors.New("unable to parse condition's field")
)

// MatchCondition is for determining the requirement of the condition
type MatchCondition struct {
	Eq interface{} `json:"eq" yaml:"eq"`
}

// Condition gets evaluated to determine the approval step resolution whether it is success or failed
type Condition struct {
	Field string          `json:"field" yaml:"field" validate:"required"`
	Match *MatchCondition `json:"match" yaml:"match" validate:"required"`
}

func (c *Condition) IsMatch(a *Appeal) (bool, error) {
	if strings.HasPrefix(c.Field, ApproversKeyResource) {
		jsonString, err := json.Marshal(a.Resource)
		if err != nil {
			return false, err
		}
		var resourceMap map[string]interface{}
		if err := json.Unmarshal(jsonString, &resourceMap); err != nil {
			return false, err
		}

		path := strings.TrimPrefix(c.Field, fmt.Sprintf("%s.", ApproversKeyResource))
		value, err := lookup.LookupString(resourceMap, path)
		if err != nil {
			return false, err
		}

		expectedValue := c.Match.Eq
		return value.Interface() == expectedValue, nil
	}

	return false, fmt.Errorf("evaluating field: %v: %v", c.Field, ErrInvalidConditionField)
}

// Step is an individual process within an approval flow
type Step struct {
	Name        string       `json:"name" yaml:"name" validate:"required"`
	Description string       `json:"description" yaml:"description"`
	Conditions  []*Condition `json:"conditions" yaml:"conditions" validate:"required_without=Approvers,omitempty,min=1,dive"`
	AllowFailed bool         `json:"allow_failed" yaml:"allow_failed"`

	Dependencies []string `json:"dependencies" yaml:"dependencies"`
	Approvers    string   `json:"approvers" yaml:"approvers" validate:"required_without=Conditions"`
}

type RequirementTrigger struct {
	ProviderType string       `json:"provider_type" yaml:"provider_type" validate:"required_without_all"`
	ProviderURN  string       `json:"provider_urn" yaml:"provider_urn" validate:"required_without_all"`
	ResourceType string       `json:"resource_type" yaml:"resource_type" validate:"required_without_all"`
	ResourceURN  string       `json:"resource_urn" yaml:"resource_urn" validate:"required_without_all"`
	Role         string       `json:"role" yaml:"role" validate:"required_without_all"`
	Conditions   []*Condition `json:"conditions" yaml:"conditions" validate:"required_without_all"`
}

func (r *RequirementTrigger) IsMatch(a *Appeal) (bool, error) {
	if r.ProviderType != "" {
		if match, err := regexp.MatchString(r.ProviderType, a.Resource.ProviderType); err != nil {
			return match, err
		} else if !match {
			return match, nil
		}
	}
	if r.ProviderURN != "" {
		if match, err := regexp.MatchString(r.ProviderURN, a.Resource.ProviderURN); err != nil {
			return match, err
		} else if !match {
			return match, nil
		}
	}
	if r.ResourceType != "" {
		if match, err := regexp.MatchString(r.ResourceType, a.Resource.Type); err != nil {
			return match, err
		} else if !match {
			return match, nil
		}
	}
	if r.ResourceURN != "" {
		if match, err := regexp.MatchString(r.ResourceURN, a.Resource.URN); err != nil {
			return match, err
		} else if !match {
			return match, nil
		}
	}
	if r.Role != "" {
		if match, err := regexp.MatchString(r.Role, a.Role); err != nil {
			return match, err
		} else if !match {
			return match, nil
		}
	}
	if r.Conditions != nil {
		for i, c := range r.Conditions {
			if match, err := c.IsMatch(a); err != nil {
				return match, fmt.Errorf("evaluating conditions[%v]: %v", i, err)
			} else if !match {
				return match, nil
			}
		}
	}

	return true, nil
}

type ResourceIdentifier struct {
	ProviderType string `json:"provider_type" yaml:"provider_type" validate:"required_with=ProviderURN Type URN"`
	ProviderURN  string `json:"provider_urn" yaml:"provider_urn" validate:"required_with=ProviderType Type URN"`
	Type         string `json:"type" yaml:"type" validate:"required_with=ProviderType ProviderURN URN"`
	URN          string `json:"urn" yaml:"urn" validate:"required_with=ProviderType ProviderURN Type"`
	ID           uint   `json:"id" yaml:"id" validate:"required_without_all"`
}

type AdditionalAppeal struct {
	Resource *ResourceIdentifier `json:"resource" yaml:"resource"  validate:"required"`
	Role     string              `json:"role" yaml:"role" validate:"required"`
	Options  *AppealOptions      `json:"options" yaml:"options"`
	Policy   *PolicyConfig       `json:"policy" yaml:"policy"`
}

type Requirement struct {
	On      *RequirementTrigger `json:"on" yaml:"on" validate:"required"`
	Appeals []*AdditionalAppeal `json:"appeals" yaml:"appeals" validate:"required,min=1,dive"`
}

// Policy is the approval policy configuration
type Policy struct {
	ID           string            `json:"id" yaml:"id" validate:"required"`
	Version      uint              `json:"version" yaml:"version" validate:"required"`
	Description  string            `json:"description" yaml:"description"`
	Steps        []*Step           `json:"steps" yaml:"steps" validate:"required,min=1,dive"`
	Requirements []*Requirement    `json:"requirements" yaml:"requirements" validate:"omitempty,min=1,dive"`
	Labels       map[string]string `json:"labels" yaml:"labels"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// PolicyRepository interface
type PolicyRepository interface {
	Create(*Policy) error
	Find() ([]*Policy, error)
	GetOne(id string, version uint) (*Policy, error)
}

// PolicyService interface
type PolicyService interface {
	Create(*Policy) error
	Find() ([]*Policy, error)
	GetOne(id string, version uint) (*Policy, error)
	Update(*Policy) error
}
