package domain

import "time"

// Resource struct
type Resource struct {
	ID           uint                   `json:"id"`
	ProviderType string                 `json:"provider_type"`
	ProviderURN  string                 `json:"provider_urn"`
	Type         string                 `json:"type"`
	URN          string                 `json:"urn"`
	Name         string                 `json:"name"`
	Details      map[string]interface{} `json:"details"`
	Labels       map[string]string      `json:"labels"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ResourceRepository interface
type ResourceRepository interface {
	Find(filters map[string]interface{}) ([]*Resource, error)
	GetOne(uint) (*Resource, error)
	BulkUpsert([]*Resource) error
	Update(*Resource) error
}

// ResourceService interface
type ResourceService interface {
	Find(filters map[string]interface{}) ([]*Resource, error)
	GetOne(uint) (*Resource, error)
	BulkUpsert([]*Resource) error
	Update(*Resource) error
}
