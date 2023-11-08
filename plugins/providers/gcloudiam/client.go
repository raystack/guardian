package gcloudiam

import (
	"context"
	"fmt"
	"strings"

	"github.com/goto/guardian/domain"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

const (
	ResourceNameOrganizationPrefix = "organizations/"
	ResourceNameProjectPrefix      = "projects/"
)

type iamClient struct {
	resourceName                string
	cloudResourceManagerService *cloudresourcemanager.Service
	iamService                  *iam.Service
}

func newIamClient(credentialsJSON []byte, resourceName string) (*iamClient, error) {
	ctx := context.Background()
	cloudResourceManagerService, err := cloudresourcemanager.NewService(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, err
	}

	iamService, err := iam.NewService(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, err
	}

	return &iamClient{
		resourceName:                resourceName,
		cloudResourceManagerService: cloudResourceManagerService,
		iamService:                  iamService,
	}, nil
}

func (c *iamClient) ListServiceAccounts(ctx context.Context) ([]*iam.ServiceAccount, error) {
	res, err := c.iamService.Projects.ServiceAccounts.List(c.resourceName).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return res.Accounts, nil
}

func (c *iamClient) GetGrantableRoles(ctx context.Context, resourceType string) ([]*iam.Role, error) {
	var fullResourceName string
	switch resourceType {
	case ResourceTypeOrganization:
		orgID := strings.Replace(c.resourceName, ResourceNameOrganizationPrefix, "", 1)
		fullResourceName = fmt.Sprintf("//cloudresourcemanager.googleapis.com/organizations/%s", orgID)

	case ResourceTypeProject:
		projectID := strings.Replace(c.resourceName, ResourceNameProjectPrefix, "", 1)
		fullResourceName = fmt.Sprintf("//cloudresourcemanager.googleapis.com/projects/%s", projectID)

	case ResourceTypeServiceAccount:
		projectID := strings.Replace(c.resourceName, ResourceNameProjectPrefix, "", 1)
		res, err := c.iamService.Projects.ServiceAccounts.List(c.resourceName).PageSize(1).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("getting a sample of service account: %w", err)
		}
		if res.Accounts == nil || len(res.Accounts) == 0 {
			return nil, fmt.Errorf("no service accounts found in project %s", projectID)
		}
		fullResourceName = fmt.Sprintf("//iam.googleapis.com/%s", res.Accounts[0].Name)

	default:
		return nil, fmt.Errorf("unknown resource type %s", resourceType)
	}

	roles := []*iam.Role{}
	nextPageToken := ""
	for {
		req := &iam.QueryGrantableRolesRequest{
			FullResourceName: fullResourceName,
			PageToken:        nextPageToken,
		}
		res, err := c.iamService.Roles.QueryGrantableRoles(req).Context(ctx).Do()
		if err != nil {
			return nil, err
		}
		roles = append(roles, res.Roles...)
		if nextPageToken = res.NextPageToken; nextPageToken == "" {
			break
		}
	}

	return roles, nil
}

func (c *iamClient) GrantAccess(ctx context.Context, accountType, accountID, role string) error {
	policy, err := c.getIamPolicy(ctx)
	if err != nil {
		return err
	}

	member := fmt.Sprintf("%s:%s", accountType, accountID)
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
		policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
			Role:    role,
			Members: []string{member},
		})
	}

	_, err = c.setIamPolicy(ctx, policy)
	return err
}

func (c *iamClient) RevokeAccess(ctx context.Context, accountType, accountID, role string) error {
	policy, err := c.getIamPolicy(ctx)
	if err != nil {
		return err
	}

	member := fmt.Sprintf("%s:%s", accountType, accountID)
	for _, b := range policy.Bindings {
		if b.Role == role {
			removeIndex := -1
			for i, m := range b.Members {
				if m == member {
					removeIndex = i
				}
			}
			if removeIndex == -1 {
				return ErrPermissionNotFound
			}
			b.Members = append(b.Members[:removeIndex], b.Members[removeIndex+1:]...)
		}
	}

	c.setIamPolicy(ctx, policy)
	return err
}

func (c *iamClient) GrantServiceAccountAccess(ctx context.Context, sa, accountType, accountID, role string) error {
	policy, err := c.iamService.Projects.ServiceAccounts.
		GetIamPolicy(sa).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy of service account %q: %w", sa, err)
	}

	member := fmt.Sprintf("%s:%s", accountType, accountID)
	roleExists := false
	for _, b := range policy.Bindings {
		if b.Role == role {
			if containsString(b.Members, member) {
				return ErrPermissionAlreadyExists
			}
			b.Members = append(b.Members, member)
		}
	}
	if !roleExists {
		policy.Bindings = append(policy.Bindings, &iam.Binding{
			Role:    role,
			Members: []string{member},
		})
	}

	if _, err := c.iamService.Projects.ServiceAccounts.
		SetIamPolicy(sa, &iam.SetIamPolicyRequest{Policy: policy}).
		Context(ctx).Do(); err != nil {
		return fmt.Errorf("setting IAM policy of service account %q: %w", sa, err)
	}

	return nil
}

func (c *iamClient) RevokeServiceAccountAccess(ctx context.Context, sa, accountType, accountID, role string) error {
	policy, err := c.iamService.Projects.ServiceAccounts.
		GetIamPolicy(sa).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy of service account %q: %w", sa, err)
	}

	member := fmt.Sprintf("%s:%s", accountType, accountID)
	for _, b := range policy.Bindings {
		if b.Role == role {
			removeIndex := -1
			for i, m := range b.Members {
				if m == member {
					removeIndex = i
				}
			}
			if removeIndex == -1 {
				return ErrPermissionNotFound
			}
			b.Members = append(b.Members[:removeIndex], b.Members[removeIndex+1:]...)
		}
	}

	if _, err := c.iamService.Projects.ServiceAccounts.
		SetIamPolicy(sa, &iam.SetIamPolicyRequest{Policy: policy}).
		Context(ctx).Do(); err != nil {
		return fmt.Errorf("setting IAM policy of service account %q: %w", sa, err)
	}

	return nil
}

func (c *iamClient) ListAccess(ctx context.Context, resources []*domain.Resource) (domain.MapResourceAccess, error) {
	policy, err := c.getIamPolicy(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting IAM policy: %w", err)
	}

	access := make(domain.MapResourceAccess)
	for _, resource := range resources {
		for _, binding := range policy.Bindings {
			for _, member := range binding.Members {
				account := strings.Split(member, ":")
				ae := domain.AccessEntry{
					AccountType: account[0],
					AccountID:   account[1],
					Permission:  binding.Role,
				}
				access[resource.URN] = append(access[resource.URN], ae)
			}
		}
	}

	return access, nil
}

func (c *iamClient) getIamPolicy(ctx context.Context) (*cloudresourcemanager.Policy, error) {
	if strings.HasPrefix(c.resourceName, ResourceNameProjectPrefix) {
		projectID := strings.Replace(c.resourceName, ResourceNameProjectPrefix, "", 1)
		return c.cloudResourceManagerService.Projects.
			GetIamPolicy(projectID, &cloudresourcemanager.GetIamPolicyRequest{}).
			Context(ctx).Do()
	} else if strings.HasPrefix(c.resourceName, ResourceNameOrganizationPrefix) {
		orgID := strings.Replace(c.resourceName, ResourceNameOrganizationPrefix, "", 1)
		return c.cloudResourceManagerService.Organizations.
			GetIamPolicy(orgID, &cloudresourcemanager.GetIamPolicyRequest{}).
			Context(ctx).Do()
	}
	return nil, ErrInvalidResourceName
}

func (c *iamClient) setIamPolicy(ctx context.Context, policy *cloudresourcemanager.Policy) (*cloudresourcemanager.Policy, error) {
	setIamPolicyRequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}
	if strings.HasPrefix(c.resourceName, ResourceNameProjectPrefix) {
		projectID := strings.Replace(c.resourceName, ResourceNameProjectPrefix, "", 1)
		return c.cloudResourceManagerService.Projects.
			SetIamPolicy(projectID, setIamPolicyRequest).
			Context(ctx).Do()
	} else if strings.HasPrefix(c.resourceName, ResourceNameOrganizationPrefix) {
		orgID := strings.Replace(c.resourceName, ResourceNameOrganizationPrefix, "", 1)
		return c.cloudResourceManagerService.Organizations.
			SetIamPolicy(orgID, setIamPolicyRequest).
			Context(ctx).Do()
	}
	return nil, ErrInvalidResourceName
}

func containsString(arr []string, v string) bool {
	for _, item := range arr {
		if item == v {
			return true
		}
	}
	return false
}
