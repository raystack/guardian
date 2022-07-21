package gcs

import (
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
)

type Provider struct {
	typeName string
	Clients  map[string]GcsClient
	crypto   domain.Crypto
}

func NewProvider(typeName string, crypto domain.Crypto) *Provider {
	return &Provider{
		typeName: typeName,
		Clients:  map[string]GcsClient{},
		crypto:   crypto,
	}
}

func (p *Provider) CreateConfig(pc *domain.ProviderConfig) error {
	return nil
}

func (p *Provider) GetType() string {
	return p.typeName
}

func (p *Provider) GrantAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	return nil
}

func (p *Provider) RevokeAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	return nil
}

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	return provider.GetRoles(pc, resourceType)
}

func (p *Provider) GetAccountTypes() []string {
	return []string{"user", "serviceAccount", "group", "domain"}
}
