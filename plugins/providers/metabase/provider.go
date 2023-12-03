package metabase

import (
	"context"
	"strings"

	"github.com/mitchellh/mapstructure"
	pv "github.com/raystack/guardian/core/provider"
	"github.com/raystack/guardian/domain"
	"github.com/raystack/salt/log"
)

type provider struct {
	pv.UnimplementedClient
	pv.PermissionManager

	typeName string
	Clients  map[string]MetabaseClient
	crypto   domain.Crypto
	logger   log.Logger
}

func NewProvider(typeName string, crypto domain.Crypto, logger log.Logger) *provider {
	return &provider{
		typeName: typeName,
		Clients:  map[string]MetabaseClient{},
		crypto:   crypto,
		logger:   logger,
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

	var resourceTypes = make(map[string]bool, 0)
	for _, rc := range pc.Resources {
		resourceTypes[rc.Type] = true
	}

	resources := []*domain.Resource{}

	var databases []*Database
	var collections []*Collection
	if _, ok := resourceTypes[ResourceTypeDatabase]; ok {
		databases, err = client.GetDatabases()
		if err != nil {
			return nil, err
		}
		resources = p.addDatabases(pc, databases, resources)
	}

	if _, ok := resourceTypes[ResourceTypeTable]; ok {
		if databases == nil {
			databases, err = client.GetDatabases()
		}
		if err != nil {
			return nil, err
		}
		resources = p.addTables(pc, databases, resources)
	}

	if _, ok := resourceTypes[ResourceTypeCollection]; ok {
		collections, err = client.GetCollections()
		if err != nil {
			return nil, err
		}
		resources = p.addCollection(pc, collections, resources)
	}

	groups, databaseResourceGroups, collectionResourceGroups, err := client.GetGroups()
	if err != nil {
		return nil, err
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
		databaseResourceMap := make(map[string]*domain.Resource, 0)
		collectionResourceMap := make(map[string]*domain.Resource, 0)

		if databases == nil {
			databases, err = client.GetDatabases()
			if err != nil {
				return nil, err
			}
		}
		for _, database := range databases {
			resource := database.ToDomain()
			databaseResourceMap[resource.URN] = resource
		}

		if collections == nil {
			collections, err = client.GetCollections()
			if err != nil {
				return nil, err
			}
		}
		for _, collection := range collections {
			resource := collection.ToDomain()
			collectionResourceMap[resource.URN] = resource
		}

		for _, g := range groups {
			if strings.HasPrefix(g.Name, GuardianGroupPrefix) {
				continue
			}

			for _, groupResource := range g.DatabaseResources {
				resourceId := groupResource.Urn
				if resource, ok := databaseResourceMap[resourceId]; ok {
					groupResource.Name = resource.Name
					groupResource.Type = resource.Type
				}
			}

			for _, groupResource := range g.CollectionResources {
				resourceId := groupResource.Urn
				if resource, ok := collectionResourceMap[resourceId]; ok {
					groupResource.Name = resource.Name
					groupResource.Type = resource.Type
				}
			}

			group := g.ToDomain()
			group.ProviderType = pc.Type
			group.ProviderURN = pc.URN
			resources = append(resources, group)
		}
	}

	return resources, nil
}

func (p *provider) addCollection(pc *domain.ProviderConfig, collections []*Collection, resources []*domain.Resource) []*domain.Resource {
	for _, c := range collections {
		db := c.ToDomain()
		db.ProviderType = pc.Type
		db.ProviderURN = pc.URN
		resources = append(resources, db)
	}
	return resources
}

func (p *provider) addDatabases(pc *domain.ProviderConfig, databases []*Database, resources []*domain.Resource) []*domain.Resource {
	for _, d := range databases {
		db := d.ToDomain()
		db.ProviderType = pc.Type
		db.ProviderURN = pc.URN
		resources = append(resources, db)
	}
	return resources
}

func (p *provider) addTables(pc *domain.ProviderConfig, databases []*Database, resources []*domain.Resource) []*domain.Resource {
	for _, d := range databases {
		db := d.ToDomain()
		db.ProviderType = pc.Type
		db.ProviderURN = pc.URN

		for _, t := range d.Tables {
			t.Database = db
			table := t.ToDomain()
			table.ProviderType = pc.Type
			table.ProviderURN = pc.URN
			resources = append(resources, table)
		}
	}
	return resources
}

func (p *provider) GrantAccess(pc *domain.ProviderConfig, a domain.Grant) error {
	// TODO: validate provider config and appeal

	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}
	client, err := p.getClient(pc.URN, creds)
	if err != nil {
		return err
	}

	groups, _, _, err := client.GetGroups()
	if err != nil {
		return err
	}

	groupMap := make(map[string]*Group, 0)
	for _, group := range groups {
		groupMap[group.Name] = group
	}

	permissions := getPermissions(a)
	if a.Resource.Type == ResourceTypeDatabase {
		d := new(Database)
		if err := d.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := client.GrantDatabaseAccess(d, a.AccountID, string(p), groupMap); err != nil {
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
	} else if a.Resource.Type == ResourceTypeGroup {
		g := new(Group)
		if err := g.FromDomain(a.Resource); err != nil {
			return err
		}

		if err := client.GrantGroupAccess(g.ID, a.AccountID); err != nil {
			return err
		}
		return nil
	} else if a.Resource.Type == ResourceTypeTable {
		t := new(Table)
		if err := t.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := client.GrantTableAccess(t, a.AccountID, string(p), groupMap); err != nil {
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
	} else if a.Resource.Type == ResourceTypeGroup {
		g := new(Group)
		if err := g.FromDomain(a.Resource); err != nil {
			return err
		}

		if err := client.RevokeGroupAccess(g.ID, a.AccountID); err != nil {
			return err
		}

		return nil
	} else if a.Resource.Type == ResourceTypeTable {
		t := new(Table)
		if err := t.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := client.RevokeTableAccess(t, a.AccountID, string(p)); err != nil {
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
	}, p.logger)
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
