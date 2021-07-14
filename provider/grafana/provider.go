package grafana

import (
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
)

type provider struct {
	typeName string
	clients  map[string]*client
	crypto   domain.Crypto
}

func NewProvider(typeName string, crypto domain.Crypto) *provider {
	return &provider{
		typeName: typeName,
		clients:  map[string]*client{},
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

	folders, err := client.getFolders()
	if err != nil {
		return nil, err
	}
	for _, f := range folders {

		dashboards, err := client.getDashboards(f.ID)
		if err != nil {
			return nil, err
		}
		for _, d := range dashboards {
			db := d.toDomain()
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
		if err := d.fromDomain(a.Resource); err != nil {
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
		if err := d.fromDomain(a.Resource); err != nil {
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

func (p *provider) getClient(providerURN string, credentials Credentials) (*client, error) {
	if p.clients[providerURN] != nil {
		return p.clients[providerURN], nil
	}

	credentials.Decrypt(p.crypto)
	client, err := NewClient(&ClientConfig{
		Host:   credentials.Host,
		ApiKey: credentials.ApiKey,
	})
	if err != nil {
		return nil, err
	}

	p.clients[providerURN] = client
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

	var roleConfig *domain.RoleConfig
	for _, rc := range resourceConfig.Roles {
		if rc.ID == a.Role {
			roleConfig = rc
		}
	}
	if roleConfig == nil {
		return nil, ErrInvalidRole
	}

	var permissions []PermissionConfig
	for _, p := range roleConfig.Permissions {
		var permission PermissionConfig
		if err := mapstructure.Decode(p, &permission); err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}
