package bigquery

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	bq "cloud.google.com/go/bigquery"
	"github.com/goto/guardian/core/provider"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/pkg/slices"
	"github.com/goto/salt/log"
	"github.com/mitchellh/mapstructure"
	"github.com/patrickmn/go-cache"
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
	GetRolePermissions(context.Context, string) ([]string, error)
}

//go:generate mockery --name=cloudLoggingClientI --exported --with-expecter
type cloudLoggingClientI interface {
	Close() error
	ListLogEntries(context.Context, ImportActivitiesFilter) ([]*Activity, error)
}

//go:generate mockery --name=encryptor --exported --with-expecter
type encryptor interface {
	domain.Crypto
}

// Provider for bigquery
type Provider struct {
	provider.PermissionManager

	typeName   string
	Clients    map[string]BigQueryClient
	LogClients map[string]cloudLoggingClientI
	encryptor  encryptor
	logger     log.Logger

	mu         sync.Mutex
	rolesCache *cache.Cache
}

// NewProvider returns bigquery provider
func NewProvider(typeName string, c encryptor, logger log.Logger) *Provider {
	return &Provider{
		typeName:   typeName,
		Clients:    map[string]BigQueryClient{},
		LogClients: map[string]cloudLoggingClientI{},
		encryptor:  c,
		logger:     logger,

		mu:         sync.Mutex{},
		rolesCache: cache.New(5*time.Minute, 10*time.Minute),
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
	logClient, err := p.getCloudLoggingClient(ctx, *pd.Config)
	if err != nil {
		return nil, fmt.Errorf("initializing cloud logging client: %w", err)
	}

	entries, err := logClient.ListLogEntries(ctx, ImportActivitiesFilter{
		ImportActivitiesFilter: filter,
		Types:                  BigQueryAuditMetadataMethods,
	})
	if err != nil {
		return nil, fmt.Errorf("listing log entries: %w", err)
	}

	activities := make([]*domain.Activity, 0, len(entries))
	if len(entries) == 0 {
		return activities, nil
	}

	gcloudRolesMap, err := p.getGcloudRoles(ctx, pd)
	if err != nil {
		return nil, fmt.Errorf("getting gcloud permissions roles: %w", err)
	}

	for _, e := range entries {
		a, err := e.ToDomainActivity(pd)
		if err != nil {
			return nil, fmt.Errorf("converting log entry to provider activity: %w", err)
		}

		for _, gcloudPermission := range a.Authorizations {
			if gcloudRoles, exists := gcloudRolesMap[a.Resource.Type][gcloudPermission]; exists {
				a.RelatedPermissions = append(a.RelatedPermissions, gcloudRoles...)
			}
		}
		a.RelatedPermissions = slices.UniqueStringSlice(a.RelatedPermissions)
		sort.Strings(a.RelatedPermissions)

		activities = append(activities, a)
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

func (p *Provider) getCloudLoggingClient(ctx context.Context, pd domain.ProviderConfig) (cloudLoggingClientI, error) {
	decryptedCreds, err := ParseCredentials(pd.Credentials, p.encryptor)
	if err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	projectID := strings.Replace(decryptedCreds.ResourceName, "projects/", "", 1)
	if p.LogClients[projectID] != nil {
		return p.LogClients[projectID], nil
	}

	client, err := NewCloudLoggingClient(ctx, projectID, []byte(decryptedCreds.ServiceAccountKey))
	if err != nil {
		return nil, err
	}

	p.LogClients[projectID] = client
	return client, nil
}

// getGcloudRoles returns map[resourceType][gcloudPermission]gcloudRoles
func (p *Provider) getGcloudRoles(ctx context.Context, pd domain.Provider) (map[string]map[string][]string, error) {
	result := map[string]map[string][]string{}

	gcloudRolesMap := map[string][]string{}
	for _, rc := range pd.Config.Resources {
		for _, r := range rc.Roles {
			for _, p := range r.Permissions {
				gcloudRole := p.(string)
				gcloudRolesMap[rc.Type] = append(gcloudRolesMap[rc.Type], gcloudRole)
			}
		}
	}

	var wg sync.WaitGroup
	chDone := make(chan bool)
	chErr := make(chan error)
	for resourceType, gcloudRoles := range gcloudRolesMap {
		for _, r := range gcloudRoles {
			wg.Add(1)
			go func(rt, gcloudRole string) {
				gcloudPermissions, err := p.getGcloudPermissions(ctx, pd, gcloudRole)
				if err != nil {
					chErr <- fmt.Errorf("getting gcloud permissions roles: %w", err)
					return
				}

				p.mu.Lock()
				for _, p := range gcloudPermissions {
					if _, ok := result[rt]; !ok {
						result[rt] = map[string][]string{}
					}
					if _, ok := result[rt][p]; !ok {
						result[rt][p] = []string{}
					}
					result[rt][p] = append(result[rt][p], gcloudRole)
				}
				p.mu.Unlock()

				wg.Done()
			}(resourceType, r)
		}
	}

	go func() {
		wg.Wait()
		close(chDone)
	}()
	select {
	case <-chDone:
		return result, nil
	case err := <-chErr:
		close(chErr)
		return nil, err
	}
}

// getGcloudPermissions returns list of gcloud permissions for given gcloud role
func (p *Provider) getGcloudPermissions(ctx context.Context, pd domain.Provider, gcloudRole string) ([]string, error) {
	roleID := translateDatasetRoleToBigQueryRole(gcloudRole)
	if permissions, exists := p.rolesCache.Get(roleID); exists {
		p.logger.Debug("getting permissions from cache", "role", roleID)
		return permissions.([]string), nil
	}

	p.logger.Debug("getting permissions from gcloud", "role", roleID)
	creds, err := ParseCredentials(pd.Config.Credentials, p.encryptor)
	if err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	client, err := p.getBigQueryClient(*creds)
	if err != nil {
		return nil, err
	}

	permissions, err := client.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	p.rolesCache.Set(roleID, permissions, 1*time.Hour)
	return permissions, nil
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

func translateDatasetRoleToBigQueryRole(role string) string {
	switch role {
	case DatasetRoleOwner:
		return "roles/bigquery.admin"
	case DatasetRoleWriter:
		return "roles/bigquery.dataEditor"
	case DatasetRoleReader:
		return "roles/bigquery.dataViewer"
	default:
		return role
	}
}
