package bigquery

import (
	"github.com/odpf/guardian/domain"
)

// Provider for bigquery
type Provider struct {
	typeName string
	crypto   domain.Crypto
}

// NewProvider returns bigquery provider
func NewProvider(typeName string, crypto domain.Crypto) *Provider {
	return &Provider{
		typeName: typeName,
		crypto:   crypto,
	}
}

// GetType returns the provider type
func (p *Provider) GetType() string {
	return p.typeName
}

// CreateConfig validates provider config
func (p *Provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc, p.crypto)

	if err := c.ParseAndValidate(); err != nil {
		return err
	}

	return c.EncryptCredentials()
}
