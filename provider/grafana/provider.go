package grafana

import (
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	pv "github.com/odpf/guardian/provider"
)

type provider struct {
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

func (p *provider) GrantAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	permissions, err := getPermissions(pc.Resources, a)
	if err != nil {
		return err
	}

	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}
	client, err := p.getClient(pc.URN, creds)
	if err != nil {
		return err
	}

	if a.Resource.Type == ResourceTypeDashboard {
		d := new(Dashboard)
		if err := d.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := client.GrantDashboardAccess(d, a.User, p.Name); err != nil {
				return err
			}
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (p *provider) RevokeAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	permissions, err := getPermissions(pc.Resources, a)
	if err != nil {
		return err
	}

	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}
	client, err := p.getClient(pc.URN, creds)
	if err != nil {
		return err
	}

	if a.Resource.Type == ResourceTypeDashboard {
		d := new(Dashboard)
		if err := d.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := client.RevokeDashboardAccess(d, a.User, p.Name); err != nil {
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

func getPermissions(resourceConfigs []*domain.ResourceConfig, a *domain.Appeal) ([]PermissionConfig, error) {
	var resourceConfig *domain.ResourceConfig
	for _, rc := range resourceConfigs {
		if rc.Type == a.Resource.Type {
			resourceConfig = rc
		}
	}
	if resourceConfig == nil {
		return nil, ErrInvalidResourceType
	}

	var role *domain.Role
	for _, r := range resourceConfig.Roles {
		if r.ID == a.Role {
			role = r
		}
	}
	if role == nil {
		return nil, ErrInvalidRole
	}

	var permissions []PermissionConfig
	for _, p := range role.Permissions {
		var permission PermissionConfig
		if err := mapstructure.Decode(p, &permission); err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}
