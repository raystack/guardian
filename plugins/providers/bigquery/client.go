package bigquery

import (
	"context"
	"errors"
	"fmt"
	"strings"

	bq "cloud.google.com/go/bigquery"
	"github.com/raystack/guardian/domain"
	bqApi "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type bigQueryClient struct {
	projectID  string
	client     *bq.Client
	iamService *iam.Service
	apiClient  *bqApi.Service
}

func NewBigQueryClient(projectID string, opts ...option.ClientOption) (*bigQueryClient, error) {
	ctx := context.Background()
	client, err := bq.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, err
	}

	apiClient, err := bqApi.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	iamService, err := iam.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &bigQueryClient{
		projectID:  projectID,
		client:     client,
		iamService: iamService,
		apiClient:  apiClient,
	}, nil
}

// GetDatasets returns all datasets within a project
func (c *bigQueryClient) GetDatasets(ctx context.Context) ([]*Dataset, error) {
	var results []*Dataset

	req := c.apiClient.Datasets.List(c.projectID)
	if err := req.Pages(ctx, func(page *bqApi.DatasetList) error {
		for _, dataset := range page.Datasets {
			d := &Dataset{
				ProjectID: dataset.DatasetReference.ProjectId,
				DatasetID: dataset.DatasetReference.DatasetId,
				Labels:    dataset.Labels,
			}
			results = append(results, d)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return results, nil
}

// GetTables returns all tables within a dataset
func (c *bigQueryClient) GetTables(ctx context.Context, datasetID string) ([]*Table, error) {
	var results []*Table

	req := c.apiClient.Tables.List(c.projectID, datasetID)
	if err := req.Pages(ctx, func(page *bqApi.TableList) error {
		for _, table := range page.Tables {
			t := &Table{
				ProjectID: table.TableReference.ProjectId,
				DatasetID: table.TableReference.DatasetId,
				TableID:   table.TableReference.TableId,
				Labels:    table.Labels,
			}
			results = append(results, t)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return results, nil
}

func (c *bigQueryClient) ResolveDatasetRole(role string) (bq.AccessRole, error) {
	switch role {
	case DatasetRoleReader:
		return bq.ReaderRole, nil
	case DatasetRoleWriter:
		return bq.WriterRole, nil
	case DatasetRoleOwner:
		return bq.OwnerRole, nil
	default:
		return "", ErrInvalidRole
	}
}

func (c *bigQueryClient) GrantDatasetAccess(ctx context.Context, d *Dataset, user, role string) error {
	dataset := c.client.Dataset(d.DatasetID)
	metadata, err := dataset.Metadata(ctx)
	if err != nil {
		return err
	}

	bqRole, err := c.ResolveDatasetRole(role)
	if err != nil {
		return err
	}
	for _, a := range metadata.Access {
		if a.Entity == user && a.Role == bqRole {
			return ErrPermissionAlreadyExists
		}
	}
	update := bq.DatasetMetadataToUpdate{
		Access: append(metadata.Access, &bq.AccessEntry{
			Role:       bqRole,
			EntityType: bq.UserEmailEntity,
			Entity:     user,
		}),
	}

	_, err = dataset.Update(ctx, update, metadata.ETag)
	return err
}

func (c *bigQueryClient) RevokeDatasetAccess(ctx context.Context, d *Dataset, user, role string) error {
	dataset := c.client.Dataset(d.DatasetID)
	metadata, err := dataset.Metadata(ctx)
	if err != nil {
		return err
	}
	bqRole, err := c.ResolveDatasetRole(role)
	if err != nil {
		return err
	}

	remainingAccessEntries := []*bq.AccessEntry{}
	for _, a := range metadata.Access {
		if a.Entity == user && a.Role == bqRole {
			continue
		}
		remainingAccessEntries = append(remainingAccessEntries, a)
	}
	if len(remainingAccessEntries) == len(metadata.Access) {
		return ErrPermissionNotFound
	}

	update := bq.DatasetMetadataToUpdate{
		Access: remainingAccessEntries,
	}

	_, err = dataset.Update(ctx, update, metadata.ETag)
	return err
}

func (c *bigQueryClient) GrantTableAccess(ctx context.Context, t *Table, accountType, accountID, role string) error {
	resourceName := fmt.Sprintf("projects/%s/datasets/%s/tables/%s", c.projectID, t.DatasetID, t.TableID)
	member := fmt.Sprintf("%s:%s", accountType, accountID)

	tableService := c.apiClient.Tables
	getIamPolicyRequest := &bqApi.GetIamPolicyRequest{
		Options: &bqApi.GetPolicyOptions{
			RequestedPolicyVersion: 1,
		},
	}
	policy, err := tableService.GetIamPolicy(resourceName, getIamPolicyRequest).Do()
	if err != nil {
		return err
	}
	roleExists := false
	for _, b := range policy.Bindings {
		if b.Role == role {
			roleExists = true
			if containsString(b.Members, member) {
				return ErrPermissionAlreadyExists
			}
			b.Members = append(b.Members, member)
		}
	}
	if !roleExists {
		policy.Bindings = append(policy.Bindings, &bqApi.Binding{
			Role:    role,
			Members: []string{member},
		})
	}

	setIamPolicyRequest := &bqApi.SetIamPolicyRequest{
		Policy: policy,
	}
	_, err = tableService.SetIamPolicy(resourceName, setIamPolicyRequest).Do()
	return err
}

func (c *bigQueryClient) RevokeTableAccess(ctx context.Context, t *Table, accountType, accountID, role string) error {
	resourceName := fmt.Sprintf("projects/%s/datasets/%s/tables/%s", c.projectID, t.DatasetID, t.TableID)
	member := fmt.Sprintf("%s:%s", accountType, accountID)

	tableService := c.apiClient.Tables
	getIamPolicyRequest := &bqApi.GetIamPolicyRequest{
		Options: &bqApi.GetPolicyOptions{
			RequestedPolicyVersion: 1,
		},
	}
	policy, err := tableService.GetIamPolicy(resourceName, getIamPolicyRequest).Context(ctx).Do()
	if err != nil {
		return err
	}

	isRoleFound := false
	for _, b := range policy.Bindings {
		if b.Role == role {
			isRoleFound = true
			isMemberFound := false
			updatedMembers := []string{}
			for _, m := range b.Members {
				if m == member {
					isMemberFound = true
					continue
				}
				updatedMembers = append(updatedMembers, m)
			}
			if !isMemberFound {
				return ErrPermissionNotFound
			}
			b.Members = updatedMembers
			break
		}
	}
	if !isRoleFound {
		return ErrPermissionNotFound
	}

	setIamPolicyRequest := &bqApi.SetIamPolicyRequest{
		Policy: policy,
	}
	_, err = tableService.SetIamPolicy(resourceName, setIamPolicyRequest).Do()
	return err
}

func (c *bigQueryClient) ListAccess(ctx context.Context, resources []*domain.Resource) (domain.MapResourceAccess, error) {
	access := make(domain.MapResourceAccess)

	for _, r := range resources {
		var accessEntries []domain.AccessEntry

		switch r.Type {
		case ResourceTypeDataset:
			d := new(Dataset)
			d.FromDomain(r)

			md, err := c.client.Dataset(d.DatasetID).Metadata(ctx)
			if err != nil {
				return nil, fmt.Errorf("getting dataset access entries of %q, %w", r.URN, err)
			}

			for _, a := range md.Access {
				ae := datasetAccessEntry(*a)
				accessEntries = append(accessEntries, domain.AccessEntry{
					AccountID:   a.Entity,
					AccountType: ae.getEntityType(),
					Permission:  string(a.Role),
				})
			}
		case ResourceTypeTable:
			t := new(Table)
			t.FromDomain(r)

			resourceName := fmt.Sprintf("projects/%s/datasets/%s/tables/%s", c.projectID, t.DatasetID, t.TableID)
			getIamPolicyRequest := &bqApi.GetIamPolicyRequest{
				Options: &bqApi.GetPolicyOptions{RequestedPolicyVersion: 1},
			}
			policy, err := c.apiClient.Tables.GetIamPolicy(resourceName, getIamPolicyRequest).Context(ctx).Do()
			if err != nil {
				return nil, fmt.Errorf("getting table access entries of %q, %w", r.URN, err)
			}

			for _, b := range policy.Bindings {
				for _, m := range b.Members {
					member := strings.Split(m, ":")
					if len(member) != 2 {
						return nil, errors.New("invalid table access member signature")
					}
					accountType := member[0]
					accountID := member[1]
					accessEntries = append(accessEntries, domain.AccessEntry{
						AccountID:   accountID,
						AccountType: accountType,
						Permission:  b.Role,
					})
				}
			}
		}

		if accessEntries != nil {
			access[r.URN] = accessEntries
		}
	}

	return access, nil
}

func (c *bigQueryClient) GetRolePermissions(ctx context.Context, role string) ([]string, error) {
	var iamRole *iam.Role
	var err error

	switch {
	case strings.HasPrefix(role, "roles/"):
		iamRole, err = c.iamService.Roles.Get(role).Context(ctx).Do()
	case strings.HasPrefix(role, "projects/"):
		iamRole, err = c.iamService.Projects.Roles.Get(role).Context(ctx).Do()
	case strings.HasPrefix(role, "organizations/"):
		iamRole, err = c.iamService.Organizations.Roles.Get(role).Context(ctx).Do()
	default:
		return nil, fmt.Errorf("invalid role signature: %q", role)
	}
	if err != nil {
		return nil, fmt.Errorf("getting role permissions of %q, %w", role, err)
	}

	return iamRole.IncludedPermissions, nil
}

func (c *bigQueryClient) getGrantableRolesForTables() ([]string, error) {
	var resourceName string
	ctx := context.Background()
	datasetIterator := c.client.Datasets(ctx)
	for {
		dataset, err := datasetIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		tableIterator := c.client.Dataset(dataset.DatasetID).Tables(ctx)
		for {
			table, err := tableIterator.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, err
			}

			resourceName = fmt.Sprintf("//bigquery.googleapis.com/projects/%v/datasets/%v/tables/%v", table.ProjectID, table.DatasetID, table.TableID)
			break
		}
		if resourceName != "" {
			break
		}
	}

	if resourceName == "" {
		return nil, ErrEmptyResource
	}

	request := &iam.QueryGrantableRolesRequest{
		FullResourceName: resourceName,
	}
	response, err := c.iamService.Roles.QueryGrantableRoles(request).Do()
	if err != nil {
		return nil, err
	}

	var roles []string
	for _, role := range response.Roles {
		roles = append(roles, role.Name)
	}

	return roles, nil
}

func containsString(arr []string, v string) bool {
	for _, item := range arr {
		if item == v {
			return true
		}
	}
	return false
}
