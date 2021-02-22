package domain

import (
	"net/http"
	"time"
)

// RoleConfig is the configuration to define a role and mapping the permissions in the provider
type RoleConfig struct {
	ID          string                    `json:"id" yaml:"id" validate:"required"`
	Name        string                    `json:"name" yaml:"name" validate:"required"`
	Description string                    `json:"description,omitempty" yaml:"description"`
	Permissions []*map[string]interface{} `json:"permissions" yaml:"permissions" validate:"required"`
}

// PolicyConfig is the configuration that defines which policy is being used in the provider
type PolicyConfig struct {
	ID      string `json:"id" yaml:"id" validate:"required"`
	Version int    `json:"version" yaml:"version" validate:"required"`
}

// ResourceConfig is the configuration for a resource type within a provider
type ResourceConfig struct {
	Type   string        `json:"type" yaml:"type" validate:"required"`
	Policy *PolicyConfig `json:"policy" yaml:"policy"`
	Roles  []*RoleConfig `json:"roles" yaml:"roles" validate:"required"`
}

// ProviderConfig is the configuration for a data provider
type ProviderConfig struct {
	Type      string            `json:"type" yaml:"type" validate:"required"`
	URN       string            `json:"urn" yaml:"urn" validate:"required"`
	Labels    map[string]string `json:"labels" yaml:"labels"`
	Auth      interface{}       `json:"auth" yaml:"auth" validate:"required"`
	Appeal    map[string]string `json:"appeal" yaml:"appeal"`
	Resources []*ResourceConfig `json:"resources" yaml:"resources" validate:"required"`
}

// Provider domain structure
type Provider struct {
	ID        uint            `json:"id"`
	Type      string          `json:"type"`
	URN       string          `json:"urn"`
	Config    *ProviderConfig `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// ProviderRepository interface
type ProviderRepository interface {
	Create(*Provider) error
	Update(*Provider) error
	Find(filters map[string]interface{}) ([]*Provider, error)
	GetOne(uint) (*Provider, error)
	Delete(uint) error
}

// ProviderService interface
type ProviderService interface {
	Create(*Provider) error
}

// ProviderHandler interface
type ProviderHandler interface {
	Create(http.ResponseWriter, *http.Request)
}
