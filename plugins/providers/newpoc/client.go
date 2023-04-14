package newpoc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/goto/guardian/domain"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
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
	config                      *Config
	iamService                  *iam.Service
	cloudResourceManagerService *cloudresourcemanager.Service
}

func NewClient(cfg *Config, opts ...option.ClientOption) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	// validator := validator.New() // TODO: use option to override validator
	// if err := cfg.Validate(context.TODO(), validator); err != nil {
	// 	return nil, err
	// }

	ctx := context.Background()
	options := []option.ClientOption{
		option.WithCredentialsJSON([]byte(cfg.credentials.ServiceAccountKey)),
	}
	options = append(options, opts...)
	iamService, err := iam.NewService(ctx, options...)
	if err != nil {
		return nil, err
	}
	cloudResourceManagerService, err := cloudresourcemanager.NewService(ctx, options...)
	if err != nil {
		return nil, err
	}

	c := &Client{
		config:                      cfg,
		iamService:                  iamService,
		cloudResourceManagerService: cloudResourceManagerService,
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

func (c *Client) ListResources(ctx context.Context) ([]domain.Resourceable, error) {
	return []domain.Resourceable{
		&resource{
			Type: c.config.resourceType,
			URN:  c.config.credentials.ResourceName,
		},
	}, nil
}

func (c *Client) GrantAccess(ctx context.Context, g domain.Grant) error {
	for _, permission := range g.Permissions {
		policy, err := c.getIamPolicy(ctx)
		if err != nil {
			return err
		}

		member := fmt.Sprintf("%s:%s", g.AccountType, g.AccountID)
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

		if _, err = c.setIamPolicy(ctx, policy); err != nil {
			return err
		}
	}
	return nil

}

func (c *Client) RevokeAccess(ctx context.Context, g domain.Grant) error {
	if g.Resource.Type != ResourceTypeProject && g.Resource.Type != ResourceTypeOrganization {
		return ErrInvalidResourceType
	}

	for _, permission := range g.Permissions {
		policy, err := c.getIamPolicy(ctx)
		if err != nil {
			return err
		}

		member := fmt.Sprintf("%s:%s", g.AccountType, g.AccountID)
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

		if _, err := c.setIamPolicy(ctx, policy); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) ListAccess(ctx context.Context, resources []*domain.Resource) (domain.MapResourceAccess, error) {
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

func (c *Client) getIamPolicy(ctx context.Context) (*cloudresourcemanager.Policy, error) {
	switch c.config.resourceType {
	case ResourceTypeProject:
		return c.cloudResourceManagerService.Projects.
			GetIamPolicy(c.config.resourceID, &cloudresourcemanager.GetIamPolicyRequest{}).
			Context(ctx).Do()
	case ResourceTypeOrganization:
		return c.cloudResourceManagerService.Organizations.
			GetIamPolicy(c.config.resourceID, &cloudresourcemanager.GetIamPolicyRequest{}).
			Context(ctx).Do()
	default:
		return nil, fmt.Errorf("%w: %q", ErrInvalidResourceType, c.config.resourceType)
	}
}

func (c *Client) setIamPolicy(ctx context.Context, policy *cloudresourcemanager.Policy) (*cloudresourcemanager.Policy, error) {
	setIamPolicyRequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}
	switch c.config.resourceType {
	case ResourceTypeProject:
		return c.cloudResourceManagerService.Projects.
			SetIamPolicy(c.config.resourceID, setIamPolicyRequest).
			Context(ctx).Do()
	case ResourceTypeOrganization:
		return c.cloudResourceManagerService.Organizations.
			SetIamPolicy(c.config.resourceID, setIamPolicyRequest).
			Context(ctx).Do()
	default:
		return nil, fmt.Errorf("%w: %q", ErrInvalidResourceType, c.config.resourceType)
	}
}

func (c *Client) listGrantableRoles(ctx context.Context) ([]string, error) {
	req := &iam.QueryGrantableRolesRequest{
		FullResourceName: fmt.Sprintf("//cloudresourcemanager.googleapis.com/%s", c.config.credentials.ResourceName),
	}
	res, err := c.iamService.Roles.QueryGrantableRoles(req).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	roles := make([]string, len(res.Roles))
	for i, r := range res.Roles {
		roles[i] = r.Name
	}
	return roles, nil
}
