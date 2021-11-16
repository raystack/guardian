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

type IAMManager interface {
	ParseConfig(*IAMConfig) (SensitiveConfig, error)
	GetClient(SensitiveConfig) (IAMClient, error)
}

// IAMClient interface
type IAMClient interface {
	GetUser(id string) (interface{}, error)
}
