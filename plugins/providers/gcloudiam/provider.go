package gcloudiam

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
)

type Provider struct {
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

	return c.EncryptCredentials()
}

func (p *Provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return nil, err
	}

	roles, err := p.GetRoles(pc, ResourceTypeProject)
	if err != nil {
		return nil, err
	}

	rolesMap := make(map[string]*domain.Role)
	for _, role := range roles {
		rolesMap[role.ID] = role
	}

	resources := []*domain.Resource{}
	for _, rc := range pc.Resources {
		for _, ro := range rc.Roles {
			if val, ok := rolesMap[ro.ID]; ok {
				resource := &domain.Resource{
					ProviderType: pc.Type,
					ProviderURN:  pc.URN,
					Type:         rc.Type,
					URN:          fmt.Sprintf("%s:%s ", creds.ResourceName, val.ID),
					Name:         val.Name,
				}
				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

func (p *Provider) GrantAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
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
		if err := client.GrantAccess(a.AccountType, a.AccountID, a.Role); err != nil {
			if errors.Is(err, ErrPermissionAlreadyExists) {
				return nil
			}
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

	client, err := p.getIamClient(pc)
	if err != nil {
		return err
	}

	if a.Resource.Type == ResourceTypeProject || a.Resource.Type == ResourceTypeOrganization {
		if err := client.RevokeAccess(a.AccountType, a.AccountID, a.Role); err != nil {
			if errors.Is(err, ErrPermissionNotFound) {
				return nil
			}
			return err
		}
		return nil
	}

	return ErrInvalidResourceType
}

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	if resourceType != ResourceTypeProject && resourceType != ResourceTypeOrganization {
		return nil, ErrInvalidResourceType
	}

	client, err := p.getIamClient(pc)
	if err != nil {
		return nil, err
	}

	iamRoles, err := client.GetRoles()
	if err != nil {
		return nil, err
	}

	var roles []*domain.Role
	for _, r := range iamRoles {
		roles = append(roles, &domain.Role{
			ID:          r.Name,
			Name:        r.Title,
			Description: r.Description,
		})
	}

	return roles, nil
}

func (p *Provider) GetAccountTypes() []string {
	return []string{
		AccountTypeUser,
		AccountTypeServiceAccount,
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
