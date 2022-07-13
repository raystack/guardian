package domain

type IAMProviderType string

const (
	IAMProviderTypeShield IAMProviderType = "shield"
	IAMProviderTypeHTTP   IAMProviderType = "http"
)

type IAMConfig struct {
	Provider      IAMProviderType   `json:"provider" yaml:"provider" validate:"required,oneof=http shield"`
	Config        interface{}       `json:"config" yaml:"config" validate:"required"`
	Schema        map[string]string `json:"schema" yaml:"schema"`
	AccountStatus string            `json:"account_status" yaml:"account_status"`
}

type IAMManager interface {
	ParseConfig(*IAMConfig) (SensitiveConfig, error)
	GetClient(SensitiveConfig) (IAMClient, error)
}

// IAMClient interface
type IAMClient interface {
	GetUser(id string) (interface{}, error)
}
