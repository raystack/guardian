package domain

type IAMProviderType string

const (
	IAMProviderTypeShield IAMProviderType = "shield"
	IAMProviderTypeHTTP   IAMProviderType = "http"
)

type IAMConfig struct {
	Provider IAMProviderType `json:"provider" yaml:"provider" validate:"required,oneof=http shield"`
	Config   interface{}     `json:"config" yaml:"config" validate:"required"`
}

// IAMClient interface
type IAMClient interface {
	GetUser(id string) (interface{}, error)
}
