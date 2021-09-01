package bigquery

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/provider"
)

// Provider for bigquery
type Provider struct {
	typeName   string
	bqClients  map[string]*bigQueryClient
	iamClients map[string]*iamClient
	crypto     domain.Crypto
}

// NewProvider returns bigquery provider
func NewProvider(typeName string, crypto domain.Crypto) *Provider {
	return &Provider{
		typeName:   typeName,
		bqClients:  map[string]*bigQueryClient{},
		iamClients: map[string]*iamClient{},
		crypto:     crypto,
	}
}

// GetType returns the provider type
func (p *Provider) GetType() string {
	return p.typeName
}

// CreateConfig validates provider config
func (p *Provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc, p.crypto)

	if err := c.ParseAndValidate(); err != nil {
		return err
	}

	return c.EncryptCredentials()
}

// GetResources returns BigQuery dataset and table resources
func (p *Provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	client, err := p.getBigQueryClient(pc.URN, Credentials(pc.Credentials.(string)))
	if err != nil {
		return nil, err
	}

	resources := []*domain.Resource{}
	ctx := context.Background()
	datasets, err := client.GetDatasets(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range datasets {
		dataset := d.toDomain()
		dataset.ProviderType = pc.Type
		dataset.ProviderURN = pc.URN
		resources = append(resources, dataset)

		tables, err := client.GetTables(ctx, dataset.Name)
		if err != nil {
			return nil, err
		}
		for _, t := range tables {
			table := t.toDomain()
			table.ProviderType = pc.Type
			table.ProviderURN = pc.URN
			resources = append(resources, table)
		}
	}

	return resources, nil
}

func (p *Provider) GrantAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	if err := validateProviderConfigAndAppealParams(pc, a); err != nil {
		return err
	}

	permissions, err := getPermissions(pc.Resources, a)
	if err != nil {
		return err
	}

	bqClient, err := p.getBigQueryClient(pc.URN, Credentials(pc.Credentials.(string)))
	if err != nil {
		return err
	}

	iamClient, err := p.getIamClient(pc.URN, Credentials(pc.Credentials.(string)))
	if err != nil {
		return err
	}

	ctx := context.TODO()
	if a.Resource.Type == ResourceTypeDataset {
		d := new(Dataset)
		if err := d.fromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Target == "" {
				if err := bqClient.GrantDatasetAccess(ctx, d, a.User, p.Name); err != nil {
					return err
				}
			} else {
				if err := iamClient.GrantAccess(ctx, p.Target, a.User, p.Name); err != nil {
					return err
				}
			}
		}

		return nil
	} else if a.Resource.Type == ResourceTypeTable {
		t := new(Table)
		if err := t.fromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Target == "" {
				if err := bqClient.GrantTableAccess(ctx, t, a.User, p.Name); err != nil {
					return err
				}
			} else {
				if err := iamClient.GrantAccess(ctx, p.Target, a.User, p.Name); err != nil {
					return err
				}
			}
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (p *Provider) RevokeAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	if err := validateProviderConfigAndAppealParams(pc, a); err != nil {
		return err
	}

	permissions, err := getPermissions(pc.Resources, a)
	if err != nil {
		return err
	}

	bqClient, err := p.getBigQueryClient(pc.URN, Credentials(pc.Credentials.(string)))
	if err != nil {
		return err
	}

	iamClient, err := p.getIamClient(pc.URN, Credentials(pc.Credentials.(string)))
	if err != nil {
		return err
	}

	ctx := context.TODO()
	if a.Resource.Type == ResourceTypeDataset {
		d := new(Dataset)
		if err := d.fromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Target == "" {
				if err := bqClient.RevokeDatasetAccess(ctx, d, a.User, p.Name); err != nil {
					return err
				}
			} else {
				if err := iamClient.RevokeAccess(ctx, p.Target, a.User, p.Name); err != nil {
					return err
				}
			}
		}

		return nil
	} else if a.Resource.Type == ResourceTypeTable {
		t := new(Table)
		if err := t.fromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Target == "" {
				if err := bqClient.RevokeTableAccess(ctx, t, a.User, p.Name); err != nil {
					return err
				}
			} else {
				if err := iamClient.RevokeAccess(ctx, p.Target, a.User, p.Name); err != nil {
					return err
				}
			}
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	return provider.GetRoles(pc, resourceType)
}

func (p *Provider) getBigQueryClient(projectID string, credentials Credentials) (*bigQueryClient, error) {
	if p.bqClients[projectID] != nil {
		return p.bqClients[projectID], nil
	}

	credentials.Decrypt(p.crypto)
	client, err := newBigQueryClient(projectID, []byte(credentials))
	if err != nil {
		return nil, err
	}

	p.bqClients[projectID] = client
	return client, nil
}

func (p *Provider) getIamClient(projectID string, credentials Credentials) (*iamClient, error) {
	if p.iamClients[projectID] != nil {
		return p.iamClients[projectID], nil
	}

	credentials.Decrypt(p.crypto)
	client, err := newCloudResourceManagerClient([]byte(credentials))
	if err != nil {
		return nil, err
	}

	p.iamClients[projectID] = client
	return client, nil
}

func validateProviderConfigAndAppealParams(pc *domain.ProviderConfig, a *domain.Appeal) error {
	if pc == nil {
		return ErrNilProviderConfig
	}
	if a == nil {
		return ErrNilAppeal
	}
	if a.Resource == nil {
		return ErrNilResource
	}
	if a.Resource.ProviderType != pc.Type {
		return ErrProviderTypeMismatch
	}
	if a.Resource.ProviderURN != pc.URN {
		return ErrProviderURNMismatch
	}
	return nil
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
