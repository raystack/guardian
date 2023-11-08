package shield

import (
	"context"

	pv "github.com/goto/guardian/core/provider"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/pkg/log"
	"github.com/mitchellh/mapstructure"
)

type provider struct {
	pv.UnimplementedClient
	pv.PermissionManager

	typeName string
	Clients  map[string]ShieldClient
	logger   log.Logger
}

func (p *provider) GetAccountTypes() []string {
	return []string{
		AccountTypeUser,
	}
}

func NewProvider(typeName string, logger log.Logger) *provider {
	return &provider{
		typeName: typeName,
		Clients:  map[string]ShieldClient{},
		logger:   logger,
	}
}

func (p *provider) GetType() string {
	return p.typeName
}

func (p *provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc)
	return c.ParseAndValidate()
}

func (p *provider) GetResources(ctx context.Context, pc *domain.ProviderConfig) ([]*domain.Resource, error) {
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

	var teams []*Team
	var projects []*Project
	var organizations []*Organization

	if _, ok := resourceTypes[ResourceTypeTeam]; ok {
		teams, err = client.GetTeams(ctx)
		if err != nil {
			return nil, err
		}
		resources = p.addTeams(pc, teams, resources)
	}

	if _, ok := resourceTypes[ResourceTypeProject]; ok {
		projects, err = client.GetProjects(ctx)
		if err != nil {
			return nil, err
		}
		resources = p.addProjects(pc, projects, resources)
	}

	if _, ok := resourceTypes[ResourceTypeOrganization]; ok {
		organizations, err = client.GetOrganizations(ctx)
		if err != nil {
			return nil, err
		}
		resources = p.addOrganizations(pc, organizations, resources)
	}

	return resources, nil
}

func (p *provider) addTeams(pc *domain.ProviderConfig, teams []*Team, resources []*domain.Resource) []*domain.Resource {
	for _, c := range teams {
		t := c.ToDomain()
		t.ProviderType = pc.Type
		t.ProviderURN = pc.URN
		resources = append(resources, t)
	}

	return resources
}

func (p *provider) addProjects(pc *domain.ProviderConfig, projects []*Project, resources []*domain.Resource) []*domain.Resource {
	for _, c := range projects {
		t := c.ToDomain()
		t.ProviderType = pc.Type
		t.ProviderURN = pc.URN
		resources = append(resources, t)
	}
	return resources
}

func (p *provider) addOrganizations(pc *domain.ProviderConfig, organizations []*Organization, resources []*domain.Resource) []*domain.Resource {
	for _, c := range organizations {
		t := c.ToDomain()
		t.ProviderType = pc.Type
		t.ProviderURN = pc.URN
		resources = append(resources, t)
	}
	return resources
}

func (p *provider) getClient(providerURN string, credentials Credentials) (ShieldClient, error) {
	if p.Clients[providerURN] != nil {
		return p.Clients[providerURN], nil
	}

	client, err := NewClient(&ClientConfig{
		Host:       credentials.Host,
		AuthHeader: credentials.AuthHeader,
		AuthEmail:  credentials.AuthEmail,
	}, p.logger)
	if err != nil {
		return nil, err
	}

	p.Clients[providerURN] = client
	return client, nil
}

func (p *provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	return pv.GetRoles(pc, resourceType)
}

func (p *provider) GrantAccess(ctx context.Context, pc *domain.ProviderConfig, a domain.Grant) error {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}
	client, err := p.getClient(pc.URN, creds)
	if err != nil {
		return err
	}

	permissions := a.GetPermissions()

	var user *User
	if user, err = client.GetSelfUser(ctx, a.AccountID); err != nil {
		return nil
	}

	switch a.Resource.Type {
	case ResourceTypeTeam:
		t := new(Team)
		if err := t.FromDomain(a.Resource); err != nil {
			return err
		}
		for _, p := range permissions {
			if err := client.GrantTeamAccess(ctx, t, user.ID, p); err != nil {
				return err
			}
		}
		return nil
	case ResourceTypeProject:
		pj := new(Project)
		if err := pj.FromDomain(a.Resource); err != nil {
			return err
		}
		for _, p := range permissions {
			if err := client.GrantProjectAccess(ctx, pj, user.ID, p); err != nil {
				return err
			}
		}
		return nil
	case ResourceTypeOrganization:
		o := new(Organization)
		if err := o.FromDomain(a.Resource); err != nil {
			return err
		}
		for _, p := range permissions {
			if err := client.GrantOrganizationAccess(ctx, o, user.ID, p); err != nil {
				return err
			}
		}
		return nil
	}

	return ErrInvalidResourceType
}

func (p *provider) RevokeAccess(ctx context.Context, pc *domain.ProviderConfig, a domain.Grant) error {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}
	client, err := p.getClient(pc.URN, creds)
	if err != nil {
		return err
	}

	permissions := a.GetPermissions()

	var user *User
	if user, err = client.GetSelfUser(ctx, a.AccountID); err != nil {
		return nil
	}

	switch a.Resource.Type {
	case ResourceTypeTeam:
		t := new(Team)
		if err := t.FromDomain(a.Resource); err != nil {
			return err
		}
		for _, p := range permissions {
			if err := client.RevokeTeamAccess(ctx, t, user.ID, p); err != nil {
				return err
			}
		}

		return nil
	case ResourceTypeProject:
		pj := new(Project)
		if err := pj.FromDomain(a.Resource); err != nil {
			return err
		}
		for _, p := range permissions {
			if err := client.RevokeProjectAccess(ctx, pj, user.ID, p); err != nil {
				return err
			}
		}

		return nil
	case ResourceTypeOrganization:
		o := new(Organization)
		if err := o.FromDomain(a.Resource); err != nil {
			return err
		}
		for _, p := range permissions {
			if err := client.RevokeOrganizationAccess(ctx, o, user.ID, p); err != nil {
				return err
			}
		}
		return nil
	}

	return ErrInvalidResourceType
}
