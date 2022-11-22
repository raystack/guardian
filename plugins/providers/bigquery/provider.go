package bigquery

import (
	"context"
	"errors"
	"fmt"
	"strings"

	bq "cloud.google.com/go/bigquery"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
)

var (
	// BigQueryAuditMetadataMethods are listed from this documentations:
	// https://cloud.google.com/bigquery/docs/reference/auditlogs
	BigQueryAuditMetadataMethods = []string{
		"google.cloud.bigquery.v2.TableService.InsertTable",
		"google.cloud.bigquery.v2.TableService.UpdateTable",
		"google.cloud.bigquery.v2.TableService.PatchTable",
		"google.cloud.bigquery.v2.TableService.DeleteTable",
		"google.cloud.bigquery.v2.DatasetService.InsertDataset",
		"google.cloud.bigquery.v2.DatasetService.UpdateDataset",
		"google.cloud.bigquery.v2.DatasetService.PatchDataset",
		"google.cloud.bigquery.v2.DatasetService.DeleteDataset",
		"google.cloud.bigquery.v2.TableDataService.List",
		"google.cloud.bigquery.v2.JobService.InsertJob",
		"google.cloud.bigquery.v2.JobService.Query",
		"google.cloud.bigquery.v2.JobService.GetQueryResults",
	}
)

//go:generate mockery --name=BigQueryClient --exported --with-expecter
type BigQueryClient interface {
	GetDatasets(context.Context) ([]*Dataset, error)
	GetTables(ctx context.Context, datasetID string) ([]*Table, error)
	GrantDatasetAccess(ctx context.Context, d *Dataset, user, role string) error
	RevokeDatasetAccess(ctx context.Context, d *Dataset, user, role string) error
	GrantTableAccess(ctx context.Context, t *Table, accountType, accountID, role string) error
	RevokeTableAccess(ctx context.Context, t *Table, accountType, accountID, role string) error
	ResolveDatasetRole(role string) (bq.AccessRole, error)
	ListAccess(ctx context.Context, resources []*domain.Resource) (domain.MapResourceAccess, error)
}

type cloudLoggingClientI interface {
	ListLogEntries(context.Context, domain.ImportActivitiesFilter) ([]*Activity, error)
}

//go:generate mockery --name=encryptor --exported --with-expecter
type encryptor interface {
	domain.Crypto
}

// Provider for bigquery
type Provider struct {
	provider.PermissionManager

	typeName            string
	Clients             map[string]BigQueryClient
	cloudLoggingClients map[string]cloudLoggingClientI
	encryptor           encryptor
}

// NewProvider returns bigquery provider
func NewProvider(typeName string, c encryptor) *Provider {
	return &Provider{
		typeName:  typeName,
		Clients:   map[string]BigQueryClient{},
		encryptor: c,
	}
}

// GetType returns the provider type
func (p *Provider) GetType() string {
	return p.typeName
}

// CreateConfig validates provider config
func (p *Provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc, p.encryptor)

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

	resourceTypes := pc.GetResourceTypes()

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
				dataset.Children = append(dataset.Children, table)
			}
		}
	}

	return resources, nil
}

func (p *Provider) GrantAccess(pc *domain.ProviderConfig, a domain.Grant) error {
	if err := validateProviderConfigAndAppealParams(pc, a); err != nil {
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

	permissions := getPermissions(a)
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

func (p *Provider) RevokeAccess(pc *domain.ProviderConfig, a domain.Grant) error {
	if err := validateProviderConfigAndAppealParams(pc, a); err != nil {
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

	permissions := getPermissions(a)
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

func (p *Provider) ListAccess(ctx context.Context, pc domain.ProviderConfig, resources []*domain.Resource) (domain.MapResourceAccess, error) {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}
	bqClient, err := p.getBigQueryClient(creds)
	if err != nil {
		return nil, fmt.Errorf("initializing bigquery client: %w", err)
	}

	return bqClient.ListAccess(ctx, resources)
}

func (p *Provider) GetActivities(ctx context.Context, pd domain.Provider, filter domain.ImportActivitiesFilter) ([]*domain.Activity, error) {
	creds, err := ParseCredentials(pd.Config.Credentials, p.encryptor)
	if err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	client, err := NewCloudLoggingClient(ctx, creds.ProjectID(), []byte(creds.ServiceAccountKey))
	if err != nil {
		return nil, fmt.Errorf("initializing cloud logging client: %w", err)
	}
	defer client.Close()

	entries, err := client.ListLogEntries(ctx, ImportActivitiesFilter{
		ImportActivitiesFilter: filter,
		Types:                  BigQueryAuditMetadataMethods,
		Authorizations:         []string{},
	})
	if err != nil {
		return nil, fmt.Errorf("listing log entries: %w", err)
	}

	activities := make([]*domain.Activity, 0, len(entries))
	for _, e := range entries {
		pa, err := e.ToActivity(pd)
		if err != nil {
			return nil, fmt.Errorf("converting log entry to provider activity: %w", err)
		}

		activities = append(activities, pa)
	}

	return activities, nil
}

func (p *Provider) getBigQueryClient(credentials Credentials) (BigQueryClient, error) {
	projectID := strings.Replace(credentials.ResourceName, "projects/", "", 1)
	if p.Clients[projectID] != nil {
		return p.Clients[projectID], nil
	}

	credentials.Decrypt(p.encryptor)
	client, err := newBigQueryClient(projectID, []byte(credentials.ServiceAccountKey))
	if err != nil {
		return nil, err
	}

	p.Clients[projectID] = client
	return client, nil
}

func validateProviderConfigAndAppealParams(pc *domain.ProviderConfig, a domain.Grant) error {
	if pc == nil {
		return ErrNilProviderConfig
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

func getPermissions(a domain.Grant) []Permission {
	var permissions []Permission
	for _, p := range a.Permissions {
		permissions = append(permissions, Permission(p))
	}
	return permissions
}
