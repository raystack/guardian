package bigquery

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
)

type iamClient struct {
	cloudResourceManagerService *cloudresourcemanager.Service
}

func newCloudResourceManagerClient(credentialsJSON []byte) (*iamClient, error) {
	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, credentialsJSON, cloudresourcemanager.CloudPlatformScope)
	if err != nil {
		return nil, err
	}
	client := oauth2.NewClient(ctx, creds.TokenSource)

	cloudResourceManagerService, err := cloudresourcemanager.New(client)
	if err != nil {
		return nil, err
	}

	return &iamClient{
		cloudResourceManagerService: cloudResourceManagerService,
	}, nil
}

func newDefaultCloudResourceManagerClient() (*iamClient, error) {
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, cloudresourcemanager.CloudPlatformScope)
	if err != nil {
		return nil, err
	}

	cloudResourceManagerService, err := cloudresourcemanager.New(client)
	if err != nil {
		return nil, err
	}

	return &iamClient{
		cloudResourceManagerService: cloudResourceManagerService,
	}, nil
}

func (c *iamClient) GrantAccess(ctx context.Context, projectID, user, role string) error {
	policy, err := c.cloudResourceManagerService.Projects.GetIamPolicy(projectID, &cloudresourcemanager.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return err
	}

	member := fmt.Sprintf("user:%s", user)
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

	setIamPolicyRequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}
	_, err = c.cloudResourceManagerService.Projects.SetIamPolicy(projectID, setIamPolicyRequest).Context(ctx).Do()
	return err
}

func (c *iamClient) RevokeAccess(ctx context.Context, projectID, user, role string) error {
	policy, err := c.cloudResourceManagerService.Projects.GetIamPolicy(projectID, &cloudresourcemanager.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return err
	}

	member := fmt.Sprintf("user:%s", user)
	var accessRemoved bool
	for _, b := range policy.Bindings {
		if b.Role == role {
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
			accessRemoved = true
		}
	}
	if accessRemoved {
		return ErrPermissionNotFound
	}

	setIamPolicyRequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}
	_, err = c.cloudResourceManagerService.Projects.SetIamPolicy(projectID, setIamPolicyRequest).Context(ctx).Do()
	return err
}
