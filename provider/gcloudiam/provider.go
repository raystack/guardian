package gcloudiam

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
)

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

func (p *Provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return nil, err
	}

	client, err := p.getIamClient(pc.URN, creds)
	if err != nil {
		return nil, err
	}

	resources := []*domain.Resource{}

	ctx := context.Background()
	roles, err := client.GetRoles(ctx, creds.OrganizationID)
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		role := r.toDomain()
		role.ProviderType = pc.Type
		role.ProviderURN = pc.URN
		resources = append(resources, role)
	}

	return resources, nil
}

func (p *Provider) GrantAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	// TODO: validate provider config and appeal

	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}

	client, err := p.getIamClient(pc.URN, creds)
	if err != nil {
		return err
	}

	if a.Resource.Type == ResourceTypeRole {
		r := new(Role)
		if err := r.fromDomain(a.Resource); err != nil {
			return err
		}

		if err := client.GrantAccess(r, a.User); err != nil {
			return err
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (p *Provider) RevokeAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}

	client, err := p.getIamClient(pc.URN, creds)
	if err != nil {
		return err
	}

	if a.Resource.Type == ResourceTypeRole {
		r := new(Role)
		if err := r.fromDomain(a.Resource); err != nil {
			return err
		}

		if err := client.RevokeAccess(r, a.User); err != nil {
			return err
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (p *Provider) getIamClient(providerURN string, credentials Credentials) (*iamClient, error) {
	if p.iamClients[providerURN] != nil {
		return p.iamClients[providerURN], nil
	}

	credentials.Decrypt(p.crypto)
	client, err := newIamClient([]byte(credentials.ServiceAccountKey), providerURN, credentials.OrganizationID)
	if err != nil {
		return nil, err
	}

	p.iamClients[providerURN] = client
	return client, nil
}
