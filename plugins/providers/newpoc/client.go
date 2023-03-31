package newpoc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/goto/guardian/domain"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/iam/v1"
)

const (
	AccountTypeUser           = "user"
	AccountTypeServiceAccount = "serviceAccount"
	AccountTypeGroup          = "group"

	ResourceNameOrganizationPrefix = "organizations/"
	ResourceNameProjectPrefix      = "projects/"
)

// Client implements BasicProviderClient
type Client struct {
	providerConfig              *domain.ProviderConfig
	cloudResourceManagerService *cloudresourcemanager.Service
	iamService                  *iam.Service
}

type ClientDependencies struct {
	ProviderConfig              *domain.ProviderConfig
	CloudResourceManagerService *cloudresourcemanager.Service
	IamService                  *iam.Service
}

func NewClient(deps *ClientDependencies) (*Client, error) {
	if deps == nil {
		return nil, errors.New("dependencies can't be nil")
	}

	if deps.ProviderConfig == nil {
		return nil, errors.New("provider config can't be nil")
	}

	if deps.CloudResourceManagerService == nil {
		return nil, errors.New("cloud resource manager service can't be nil")
	}

	if deps.IamService == nil {
		return nil, errors.New("iam service can't be nil")
	}

	c := &Client{
		providerConfig:              deps.ProviderConfig,
		cloudResourceManagerService: deps.CloudResourceManagerService,
		iamService:                  deps.IamService,
	}

	return c, nil
}

func (c *Client) GetAllowedAccountTypes(ctx context.Context) []string {
	return []string{
		AccountTypeUser,
		AccountTypeServiceAccount,
		AccountTypeGroup,
	}
}

func (c *Client) ListResources(ctx context.Context) ([]IResource, error) {
	var creds credentials
	if err := mapstructure.Decode(c.providerConfig.Credentials, &creds); err != nil {
		return nil, err
	}

	var t string
	if strings.HasPrefix(creds.ResourceName, "project") {
		t = ResourceTypeProject
	} else if strings.HasPrefix(creds.ResourceName, "organization") {
		t = ResourceTypeOrganization
	}

	return []IResource{
		&domain.Resource{
			ProviderType: c.providerConfig.Type,
			ProviderURN:  c.providerConfig.URN,
			Type:         t,
			URN:          creds.ResourceName,
			Name:         fmt.Sprintf("%s - GCP IAM", creds.ResourceName),
		},
	}, nil
}

func (c *Client) GrantAccess(ctx context.Context, r IResource, accountID string, permissions []string) error {
	var creds credentials
	if err := mapstructure.Decode(c.providerConfig.Credentials, &creds); err != nil {
		return err
	}

	if r.GetType() == ResourceTypeProject || r.GetType() == ResourceTypeOrganization {
		for _, permission := range permissions {
			policy, err := c.getIamPolicy(ctx, creds.ResourceName)
			if err != nil {
				return err
			}

			member := accountID
			roleExists := false
			for _, b := range policy.Bindings {
				if b.Role == permission {
					roleExists = true
					if containsString(b.Members, member) {
						// Permission already exists
						continue
					}
					b.Members = append(b.Members, member)
				}
			}
			if !roleExists {
				policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
					Role:    permission,
					Members: []string{member},
				})
			}

			_, err = c.setIamPolicy(ctx, creds.ResourceName, policy)
			if err != nil {
				return err
			}
		}
	}
	return ErrInvalidResourceType
}

func (c *Client) RevokeAccess(ctx context.Context, r IResource, accountID string, permissions []string) error {
	var creds credentials
	if err := mapstructure.Decode(c.providerConfig.Credentials, &creds); err != nil {
		return err
	}

	if r.GetType() == ResourceTypeProject || r.GetType() == ResourceTypeOrganization {
		for _, permission := range permissions {
			policy, err := c.getIamPolicy(ctx, creds.ResourceName)
			if err != nil {
				return err
			}

			member := accountID

			for _, b := range policy.Bindings {
				if b.Role == permission {
					removeIndex := -1
					for i, m := range b.Members {
						if m == member {
							removeIndex = i
						}
					}
					if removeIndex == -1 {
						// permission doesn't exist
						continue
					}
					b.Members = append(b.Members[:removeIndex], b.Members[removeIndex+1:]...)
				}
			}

			c.setIamPolicy(ctx, creds.ResourceName, policy)
			return err
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (c *Client) getIamPolicy(ctx context.Context, resourceName string) (*cloudresourcemanager.Policy, error) {
	if strings.HasPrefix(resourceName, ResourceNameProjectPrefix) {
		projectID := strings.Replace(resourceName, ResourceNameProjectPrefix, "", 1)
		return c.cloudResourceManagerService.Projects.
			GetIamPolicy(projectID, &cloudresourcemanager.GetIamPolicyRequest{}).
			Context(ctx).Do()
	} else if strings.HasPrefix(resourceName, ResourceNameOrganizationPrefix) {
		orgID := strings.Replace(resourceName, ResourceNameOrganizationPrefix, "", 1)
		return c.cloudResourceManagerService.Organizations.
			GetIamPolicy(orgID, &cloudresourcemanager.GetIamPolicyRequest{}).
			Context(ctx).Do()
	}
	return nil, ErrInvalidResourceName
}

func (c *Client) setIamPolicy(ctx context.Context, resourceName string, policy *cloudresourcemanager.Policy) (*cloudresourcemanager.Policy, error) {
	setIamPolicyRequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}
	if strings.HasPrefix(resourceName, ResourceNameProjectPrefix) {
		projectID := strings.Replace(resourceName, ResourceNameProjectPrefix, "", 1)
		return c.cloudResourceManagerService.Projects.
			SetIamPolicy(projectID, setIamPolicyRequest).
			Context(ctx).Do()
	} else if strings.HasPrefix(resourceName, ResourceNameOrganizationPrefix) {
		orgID := strings.Replace(resourceName, ResourceNameOrganizationPrefix, "", 1)
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
