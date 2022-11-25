package gcloudiam

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
)

type Provider struct {
	provider.PermissionManager
	provider.UnimplementedClient

	typeName string
	Clients  map[string]GcloudIamClient
	crypto   domain.Crypto
}

func NewProvider(typeName string, crypto domain.Crypto) *Provider {
	return &Provider{
		typeName: typeName,
		Clients:  map[string]GcloudIamClient{},
		crypto:   crypto,
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

	client, err := p.getIamClient(pc)
	if err != nil {
		return err
	}

	for _, r := range c.ProviderConfig.Resources {
		if err := c.validatePermissions(r, client); err != nil {
			return err
		}
	}

	return c.EncryptCredentials()
}

func (p *Provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return nil, err
	}

	var t string
	if strings.HasPrefix(creds.ResourceName, "project") {
		t = ResourceTypeProject
	} else if strings.HasPrefix(creds.ResourceName, "organization") {
		t = ResourceTypeOrganization
	}

	return []*domain.Resource{
		{
			ProviderType: pc.Type,
			ProviderURN:  pc.URN,
			Type:         t,
			URN:          creds.ResourceName,
			Name:         fmt.Sprintf("%s - GCP IAM", creds.ResourceName),
		},
	}, nil
}

func (p *Provider) GrantAccess(pc *domain.ProviderConfig, a domain.Grant) error {
	// TODO: validate provider config and appeal

	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}

	client, err := p.getIamClient(pc)
	if err != nil {
		return err
	}

	if a.Resource.Type == ResourceTypeProject || a.Resource.Type == ResourceTypeOrganization {
		for _, p := range a.Permissions {
			permission := fmt.Sprint(p)
			if err := client.GrantAccess(a.AccountType, a.AccountID, permission); err != nil {
				if !errors.Is(err, ErrPermissionAlreadyExists) {
					return err
				}
			}
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (p *Provider) RevokeAccess(pc *domain.ProviderConfig, a domain.Grant) error {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}

	client, err := p.getIamClient(pc)
	if err != nil {
		return err
	}

	if a.Resource.Type == ResourceTypeProject || a.Resource.Type == ResourceTypeOrganization {
		for _, p := range a.Permissions {
			permission := fmt.Sprint(p)
			if err := client.RevokeAccess(a.AccountType, a.AccountID, permission); err != nil {
				if !errors.Is(err, ErrPermissionNotFound) {
					return err
				}
			}
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	if resourceType != ResourceTypeProject && resourceType != ResourceTypeOrganization {
		return nil, ErrInvalidResourceType
	}

	return provider.GetRoles(pc, resourceType)
}

func (p *Provider) GetPermissions(_pc *domain.ProviderConfig, _resourceType, role string) ([]interface{}, error) {
	// TODO: validate if role is a valid gcloud iam role
	return p.PermissionManager.GetPermissions(_pc, _resourceType, role)
}

func (p *Provider) GetAccountTypes() []string {
	return []string{
		AccountTypeUser,
		AccountTypeServiceAccount,
		AccountTypeGroup,
	}
}

func (p *Provider) getIamClient(pc *domain.ProviderConfig) (GcloudIamClient, error) {
	var credentials Credentials
	if err := mapstructure.Decode(pc.Credentials, &credentials); err != nil {
		return nil, err
	}
	providerURN := pc.URN

	if p.Clients[providerURN] != nil {
		return p.Clients[providerURN], nil
	}

	credentials.Decrypt(p.crypto)
	client, err := newIamClient([]byte(credentials.ServiceAccountKey), credentials.ResourceName)
	if err != nil {
		return nil, err
	}

	p.Clients[providerURN] = client
	return client, nil
}
