package metabase

import (
	"github.com/mitchellh/mapstructure"
	pv "github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
)

type provider struct {
	typeName string
	Clients  map[string]MetabaseClient
	crypto   domain.Crypto
}

func NewProvider(typeName string, crypto domain.Crypto) *provider {
	return &provider{
		typeName: typeName,
		Clients:  map[string]MetabaseClient{},
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

	var resourceTypes = make(map[string]bool, 0)
	for _, rc := range pc.Resources {
		resourceTypes[rc.Type] = true
	}

	resources := []*domain.Resource{}

	if _, ok := resourceTypes[ResourceTypeDatabase]; ok && resourceTypes[ResourceTypeDatabase] {
		databases, err := client.GetDatabases()
		if err != nil {
			return nil, err
		}
		for _, d := range databases {
			db := d.ToDomain()
			db.ProviderType = pc.Type
			db.ProviderURN = pc.URN
			resources = append(resources, db)

			if _, ok := resourceTypes[ResourceTypeTable]; ok && resourceTypes[ResourceTypeTable] {
				tables := d.Tables
				for _, t := range tables {
					t.Database = db
					table := t.ToDomain()
					table.ProviderType = pc.Type
					table.ProviderURN = pc.URN
					resources = append(resources, table)
				}
			}
		}
	}

	if _, ok := resourceTypes[ResourceTypeCollection]; ok && resourceTypes[ResourceTypeCollection] {
		collections, err := client.GetCollections()
		if err != nil {
			return nil, err
		}
		for _, c := range collections {
			db := c.ToDomain()
			db.ProviderType = pc.Type
			db.ProviderURN = pc.URN
			resources = append(resources, db)
		}
	}

	groups, databaseResourceGroups, collectionResourceGroups, err := client.GetGroups()
	if err != nil {
		return nil, err
	}

	databaseResourceMap := make(map[string]*domain.Resource, 0)
	collectionResourceMap := make(map[string]*domain.Resource, 0)
	for _, resource := range resources {
		if resource.Type == ResourceTypeDatabase || resource.Type == ResourceTypeTable {
			databaseResourceMap[resource.URN] = resource
		}
		if resource.Type == ResourceTypeCollection {
			collectionResourceMap[resource.URN] = resource
		}
	}

	for _, resource := range resources {
		if resource.Type == ResourceTypeDatabase || resource.Type == ResourceTypeTable {
			if groups, ok := databaseResourceGroups[resource.URN]; ok {
				resource.Details["groups"] = groups
			}
		}
		if resource.Type == ResourceTypeCollection {
			if groups, ok := collectionResourceGroups[resource.URN]; ok {
				resource.Details["groups"] = groups
			}
		}
	}

	if _, ok := resourceTypes[ResourceTypeGroup]; ok && resourceTypes[ResourceTypeGroup] {
		for _, g := range groups {
			for _, resourceMap := range g.DatabaseResources {
				resourceId := resourceMap["urn"].(string)
				if resource, ok := databaseResourceMap[resourceId]; ok {
					resourceMap["name"] = resource.Name
					resourceMap["type"] = resource.Type
				}
			}

			for _, resourceMap := range g.CollectionResources {
				resourceId := resourceMap["urn"].(string)
				if resource, ok := collectionResourceMap[resourceId]; ok {
					resourceMap["name"] = resource.Name
					resourceMap["type"] = resource.Type
				}
			}

			db := g.ToDomain()
			db.ProviderType = pc.Type
			db.ProviderURN = pc.URN
			resources = append(resources, db)
		}
	}

	return resources, nil
}

func (p *provider) GrantAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	// TODO: validate provider config and appeal

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

	if a.Resource.Type == ResourceTypeDatabase {
		d := new(Database)
		if err := d.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := client.GrantDatabaseAccess(d, a.AccountID, string(p)); err != nil {
				return err
			}
		}

		return nil
	} else if a.Resource.Type == ResourceTypeCollection {
		c := new(Collection)
		if err := c.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := client.GrantCollectionAccess(c, a.AccountID, string(p)); err != nil {
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

	if a.Resource.Type == ResourceTypeDatabase {
		d := new(Database)
		if err := d.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := client.RevokeDatabaseAccess(d, a.AccountID, string(p)); err != nil {
				return err
			}
		}

		return nil
	} else if a.Resource.Type == ResourceTypeCollection {
		c := new(Collection)
		if err := c.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := client.RevokeCollectionAccess(c, a.AccountID, string(p)); err != nil {
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

func (p *provider) getClient(providerURN string, credentials Credentials) (MetabaseClient, error) {
	if p.Clients[providerURN] != nil {
		return p.Clients[providerURN], nil
	}

	if err := credentials.Decrypt(p.crypto); err != nil {
		return nil, err
	}
	client, err := NewClient(&ClientConfig{
		Host:     credentials.Host,
		Username: credentials.Username,
		Password: credentials.Password,
	})
	if err != nil {
		return nil, err
	}

	p.Clients[providerURN] = client
	return client, nil
}

func getPermissions(resourceConfigs []*domain.ResourceConfig, a *domain.Appeal) ([]Permission, error) {
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

	var permissions []Permission
	for _, p := range role.Permissions {
		var permission Permission
		if err := mapstructure.Decode(p, &permission); err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}
