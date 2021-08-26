package gcloudiam

import (
	"context"
	"fmt"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

type iamClient struct {
	projectID                   string
	orgID                       string
	cloudResourceManagerService *cloudresourcemanager.Service
	iamService                  *iam.Service
}

func newIamClient(credentialsJSON []byte, projectID, orgID string) (*iamClient, error) {
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
		projectID:                   projectID,
		orgID:                       orgID,
		cloudResourceManagerService: cloudResourceManagerService,
		iamService:                  iamService,
	}, nil
}

func (c *iamClient) GetRoles(ctx context.Context, orgID string) ([]*Role, error) {
	var roles []*Role

	req := c.iamService.Roles.List()
	if err := req.Pages(ctx, func(page *iam.ListRolesResponse) error {
		for _, role := range page.Roles {
			roles = append(roles, c.fromIamRole(role))
		}
		return nil
	}); err != nil {
		return nil, err
	}

	parentProject := fmt.Sprintf("projects/%s", c.projectID)
	projectRolesReq := c.iamService.Projects.Roles.List(parentProject)
	if err := projectRolesReq.Pages(ctx, func(page *iam.ListRolesResponse) error {
		for _, role := range page.Roles {
			roles = append(roles, c.fromIamRole(role))
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if c.orgID != "" {
		parentOrg := fmt.Sprintf("organizations/%s", c.orgID)
		orgRolesReq := c.iamService.Organizations.Roles.List(parentOrg)
		if err := orgRolesReq.Pages(ctx, func(page *iam.ListRolesResponse) error {
			for _, role := range page.Roles {
				roles = append(roles, c.fromIamRole(role))
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return roles, nil
}

func (c *iamClient) GrantAccess(r *Role, user string) error {
	policy, err := c.cloudResourceManagerService.Projects.GetIamPolicy(c.projectID, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		return err
	}

	member := fmt.Sprintf("user:%s", user)
	roleExists := false
	for _, b := range policy.Bindings {
		if b.Role == r.Name {
			roleExists = true
			if containsString(b.Members, member) {
				return ErrPermissionAlreadyExists
			}
			b.Members = append(b.Members, member)
		}
	}
	if !roleExists {
		policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
			Role:    r.Name,
			Members: []string{member},
		})
	}

	setIamPolicyRequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}
	_, err = c.cloudResourceManagerService.Projects.SetIamPolicy(c.projectID, setIamPolicyRequest).Do()
	return err
}

func (c *iamClient) RevokeAccess(r *Role, user string) error {
	policy, err := c.cloudResourceManagerService.Projects.GetIamPolicy(c.projectID, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		return err
	}

	member := fmt.Sprintf("user:%s", user)
	for _, b := range policy.Bindings {

		if b.Role == r.Name {
			var removeIndex int
			for i, m := range b.Members {
				if m == member {
					removeIndex = i
				}
			}
			if removeIndex == 0 {
				return ErrPermissionNotFound
			}
			b.Members = append(b.Members[:removeIndex], b.Members[removeIndex+1:]...)
		}
	}

	setIamPolicyRequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}
	_, err = c.cloudResourceManagerService.Projects.SetIamPolicy(c.projectID, setIamPolicyRequest).Do()
	return err
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
