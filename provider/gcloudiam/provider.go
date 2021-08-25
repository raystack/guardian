package gcloudiam

import "github.com/odpf/guardian/domain"

type Provider struct {
	typeName   string
	iamClients map[string]*iamClient
	crypto     domain.Crypto
}

func NewProvider(typeName string, crypto domain.Crypto) *Provider {
	return &Provider{
		typeName:   typeName,
		iamClients: map[string]*iamClient{},
		crypto:     crypto,
	}
}

func (p *Provider) GetType() string {
	return p.typeName
}

func (p *Provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc, p.crypto)

	if err := c.ParseAndValidate(); err != nil {
		return err
	}

	return c.EncryptCredentials()
}
