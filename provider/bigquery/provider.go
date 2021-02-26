package bigquery

import "github.com/odpf/guardian/domain"

// Provider for bigquery
type Provider struct {
	typeName string
}

// NewProvider returns bigquery provider
func NewProvider(typeName string) *Provider {
	return &Provider{
		typeName: typeName,
	}
}

// GetType returns the provider type
func (p *Provider) GetType() string {
	return p.typeName
}

// ValidateConfig validates provider config
func (p *Provider) ValidateConfig(pc *domain.ProviderConfig) error {
	return NewConfig(pc).Validate()
}
