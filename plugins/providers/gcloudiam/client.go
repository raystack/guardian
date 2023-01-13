package gcloudiam

import (
	"context"
	"fmt"
	"strings"

	"github.com/odpf/guardian/domain"
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

func (c *iamClient) GetRoles() ([]*Role, error) {
	var roles []*Role

	ctx := context.TODO()
	req := c.iamService.Roles.List()
	if err := req.Pages(ctx, func(page *iam.ListRolesResponse) error {
		for _, role := range page.Roles {
			roles = append(roles, c.fromIamRole(role))
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if strings.HasPrefix(c.resourceName, ResourceNameProjectPrefix) {
		projectRolesReq := c.iamService.Projects.Roles.List(c.resourceName)
		if err := projectRolesReq.Pages(ctx, func(page *iam.ListRolesResponse) error {
			for _, role := range page.Roles {
				roles = append(roles, c.fromIamRole(role))
			}
			return nil
		}); err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(c.resourceName, ResourceNameOrganizationPrefix) {
		orgRolesReq := c.iamService.Organizations.Roles.List(c.resourceName)
		if err := orgRolesReq.Pages(ctx, func(page *iam.ListRolesResponse) error {
			for _, role := range page.Roles {
				roles = append(roles, c.fromIamRole(role))
			}
			return nil
		}); err != nil {
			return nil, err
		}
	} else {
		return nil, ErrInvalidResourceName
	}

	return roles, nil
}

func (c *iamClient) GrantAccess(accountType, accountID, role string) error {
	ctx := context.TODO()
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

func (c *iamClient) RevokeAccess(accountType, accountID, role string) error {
	ctx := context.TODO()
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

func (c *iamClient) fromIamRole(r *iam.Role) *Role {
	return &Role{
		Name:        r.Name,
		Title:       r.Title,
		Description: r.Description,
	}
}

func containsString(arr []string, v string) bool {
	for _, item := range arr {
		if item == v {
			return true
		}
	}
	return false
}
