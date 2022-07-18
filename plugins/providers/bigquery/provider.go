package bigquery

import (
	"context"
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
)

// Provider for bigquery
type Provider struct {
	typeName string
	Clients  map[string]BigQueryClient
	crypto   domain.Crypto
}

// NewProvider returns bigquery provider
func NewProvider(typeName string, crypto domain.Crypto) *Provider {
	return &Provider{
		typeName: typeName,
		Clients:  map[string]BigQueryClient{},
		crypto:   crypto,
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
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return nil, err
	}

	client, err := p.getBigQueryClient(creds)
	if err != nil {
		return nil, err
	}

	var resourceTypes []string
	for _, rc := range pc.Resources {
		resourceTypes = append(resourceTypes, rc.Type)
	}

	resources := []*domain.Resource{}
	ctx := context.Background()
	datasets, err := client.GetDatasets(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range datasets {
		dataset := d.ToDomain()
		dataset.ProviderType = pc.Type
		dataset.ProviderURN = pc.URN

		if containsString(resourceTypes, ResourceTypeDataset) {
			resources = append(resources, dataset)
		}

		if containsString(resourceTypes, ResourceTypeTable) {
			tables, err := client.GetTables(ctx, dataset.Name)
			if err != nil {
				return nil, err
			}
			for _, t := range tables {
				table := t.ToDomain()
				table.ProviderType = pc.Type
				table.ProviderURN = pc.URN
				resources = append(resources, table)
			}
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

	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}
	bqClient, err := p.getBigQueryClient(creds)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	if a.Resource.Type == ResourceTypeDataset {
		d := new(Dataset)
		if err := d.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := bqClient.GrantDatasetAccess(ctx, d, a.AccountID, string(p)); err != nil {
				if errors.Is(err, ErrPermissionAlreadyExists) {
					return nil
				}
				return err
			}
		}

		return nil
	} else if a.Resource.Type == ResourceTypeTable {
		t := new(Table)
		if err := t.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := bqClient.GrantTableAccess(ctx, t, a.AccountType, a.AccountID, string(p)); err != nil {
				if errors.Is(err, ErrPermissionAlreadyExists) {
					return nil
				}
				return err
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

	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}
	bqClient, err := p.getBigQueryClient(creds)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	if a.Resource.Type == ResourceTypeDataset {
		d := new(Dataset)
		if err := d.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := bqClient.RevokeDatasetAccess(ctx, d, a.AccountID, string(p)); err != nil {
				if errors.Is(err, ErrPermissionNotFound) {
					return nil
				}
				return err
			}
		}

		return nil
	} else if a.Resource.Type == ResourceTypeTable {
		t := new(Table)
		if err := t.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if err := bqClient.RevokeTableAccess(ctx, t, a.AccountType, a.AccountID, string(p)); err != nil {
				if errors.Is(err, ErrPermissionNotFound) {
					return nil
				}
				return err
			}
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	return provider.GetRoles(pc, resourceType)
}

func (p *Provider) GetAccountTypes() []string {
	return []string{
		AccountTypeUser,
		AccountTypeServiceAccount,
	}
}

func (p *Provider) getBigQueryClient(credentials Credentials) (BigQueryClient, error) {
	projectID := strings.Replace(credentials.ResourceName, "projects/", "", 1)
	if p.Clients[projectID] != nil {
		return p.Clients[projectID], nil
	}

	credentials.Decrypt(p.crypto)
	client, err := newBigQueryClient(projectID, []byte(credentials.ServiceAccountKey))
	if err != nil {
		return nil, err
	}

	p.Clients[projectID] = client
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
