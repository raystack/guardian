package grafana

import (
	"context"

	"github.com/mitchellh/mapstructure"
	pv "github.com/raystack/guardian/core/provider"
	"github.com/raystack/guardian/domain"
)

type provider struct {
	pv.UnimplementedClient
	pv.PermissionManager

	typeName string
	Clients  map[string]GrafanaClient
	crypto   domain.Crypto
}

func NewProvider(typeName string, crypto domain.Crypto) *provider {
	return &provider{
		typeName: typeName,
		Clients:  map[string]GrafanaClient{},
		crypto:   crypto,
	}
}

func (p *provider) GetType() string {
	return p.typeName
}

// GetDefaultRoles returns a list of roles supported by the provider
func (p *provider) GetDefaultRoles(ctx context.Context, name string, resourceType string) ([]string, error) {
	return []string{}, nil
}

func (p *provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc, p.crypto)

	if err := c.ParseAndValidate(); err != nil {
		return err
	}

	return c.EncryptCredentials()
}

func (p *provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return nil, err
	}

	client, err := p.getClient(pc.URN, creds)
	if err != nil {
		return nil, err
	}

	resources := []*domain.Resource{}

	folders, err := client.GetFolders()
	if err != nil {
		return nil, err
	}
	for _, f := range folders {
		dashboards, err := client.GetDashboards(f.ID)
		if err != nil {
			return nil, err
		}
		for _, d := range dashboards {
			db := d.ToDomain()
			db.ProviderType = pc.Type
			db.ProviderURN = pc.URN
			resources = append(resources, db)
		}
	}
	return resources, nil
}

func (p *provider) GrantAccess(pc *domain.ProviderConfig, a domain.Grant) error {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}
	client, err := p.getClient(pc.URN, creds)
	if err != nil {
		return err
	}

	permissions := getPermissions(a)
	if a.Resource.Type == ResourceTypeDashboard {
		d := new(Dashboard)
		if err := d.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := client.GrantDashboardAccess(d, a.AccountID, string(p)); err != nil {
				return err
			}
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (p *provider) RevokeAccess(pc *domain.ProviderConfig, a domain.Grant) error {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}
	client, err := p.getClient(pc.URN, creds)
	if err != nil {
		return err
	}

	permissions := getPermissions(a)
	if a.Resource.Type == ResourceTypeDashboard {
		d := new(Dashboard)
		if err := d.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := client.RevokeDashboardAccess(d, a.AccountID, string(p)); err != nil {
				return err
			}
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (p *provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	return pv.GetRoles(pc, resourceType)
}

func (p *provider) GetAccountTypes() []string {
	return []string{
		AccountTypeUser,
	}
}

func (p *provider) getClient(providerURN string, credentials Credentials) (GrafanaClient, error) {
	if p.Clients[providerURN] != nil {
		return p.Clients[providerURN], nil
	}

	if err := credentials.Decrypt(p.crypto); err != nil {
		return nil, err
	}

	org := providerURN
	client, err := NewClient(&ClientConfig{
		Host:     credentials.Host,
		Username: credentials.Username,
		Password: credentials.Password,
		Org:      org,
	})
	if err != nil {
		return nil, err
	}

	p.Clients[providerURN] = client
	return client, nil
}

func getPermissions(a domain.Grant) []Permission {
	var permissions []Permission
	for _, p := range a.Permissions {
		permissions = append(permissions, Permission(p))
	}
	return permissions
}
