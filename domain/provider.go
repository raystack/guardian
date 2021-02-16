package domain

import "time"

// Provider domain structure
type Provider struct {
	ID        uint      `json:"id"`
	Config    string    `json:"config"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	Create(Provider) error
}
