package gcloudiam

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/raystack/guardian/core/provider"
	"github.com/raystack/guardian/domain"
	"golang.org/x/net/context"
	"google.golang.org/api/iam/v1"
)

//go:generate mockery --name=GcloudIamClient --exported --with-expecter
type GcloudIamClient interface {
	GetGrantableRoles(ctx context.Context, resourceType string) ([]*iam.Role, error)
	GrantAccess(accountType, accountID, role string) error
	RevokeAccess(accountType, accountID, role string) error
	ListAccess(ctx context.Context, resources []*domain.Resource) (domain.MapResourceAccess, error)
	ListServiceAccounts(context.Context) ([]*iam.ServiceAccount, error)
	GrantServiceAccountAccess(ctx context.Context, sa, accountType, accountID, roles string) error
	RevokeServiceAccountAccess(ctx context.Context, sa, accountType, accountID, role string) error
}

//go:generate mockery --name=encryptor --exported --with-expecter
type encryptor interface {
	domain.Crypto
}

type Provider struct {
	provider.PermissionManager
	provider.UnimplementedClient

	typeName string
	Clients  map[string]GcloudIamClient
	crypto   encryptor
}

func NewProvider(typeName string, crypto encryptor) *Provider {
	return &Provider{
		typeName: typeName,
		Clients:  map[string]GcloudIamClient{},
		crypto:   crypto,
	}
}

func (p *Provider) GetType() string {
	return p.typeName
}

func (p *Provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc, p.crypto)

	if err := c.ParseAndValidate(); err != nil {
		return err
	}

	client, err := p.getIamClient(pc)
	if err != nil {
		return err
	}

	for _, r := range c.ProviderConfig.Resources {
		if err := c.validatePermissions(r, client); err != nil {
			return err
		}
	}

	return c.EncryptCredentials()
}

func (p *Provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	resources := []*domain.Resource{}

	for _, rc := range pc.Resources {
		switch rc.Type {
		case ResourceTypeProject, ResourceTypeOrganization:
			var creds Credentials
			if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
				return nil, err
			}
			resources = append(resources, &domain.Resource{
				ProviderType: pc.Type,
				ProviderURN:  pc.URN,
				Type:         rc.Type,
				URN:          creds.ResourceName,
				Name:         fmt.Sprintf("%s - GCP IAM", creds.ResourceName),
			})

		case ResourceTypeServiceAccount:
			client, err := p.getIamClient(pc)
			if err != nil {
				return nil, fmt.Errorf("initializing iam client: %w", err)
			}

			serviceAccounts, err := client.ListServiceAccounts(context.TODO())
			if err != nil {
				return nil, fmt.Errorf("listing service accounts: %w", err)
			}

			// TODO: filter

			for _, sa := range serviceAccounts {
				resources = append(resources, &domain.Resource{
					ProviderType: pc.Type,
					ProviderURN:  pc.URN,
					Type:         rc.Type,
					URN:          sa.Name,
					Name:         sa.Email,
				})
			}

		default:
			return nil, ErrInvalidResourceType
		}
	}

	return resources, nil
}

func (p *Provider) GrantAccess(pc *domain.ProviderConfig, g domain.Grant) error {
	// TODO: validate provider config and appeal

	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}

	client, err := p.getIamClient(pc)
	if err != nil {
		return err
	}

	switch g.Resource.Type {
	case ResourceTypeProject, ResourceTypeOrganization:
		for _, p := range g.Permissions {
			if err := client.GrantAccess(g.AccountType, g.AccountID, p); err != nil {
				if !errors.Is(err, ErrPermissionAlreadyExists) {
					return err
				}
			}
		}
		return nil

	case ResourceTypeServiceAccount:
		for _, p := range g.Permissions {
			if err := client.GrantServiceAccountAccess(context.TODO(), g.Resource.URN, g.AccountType, g.AccountID, p); err != nil {
				if !errors.Is(err, ErrPermissionAlreadyExists) {
					return err
				}
			}
		}
		return nil

	default:
		return ErrInvalidResourceType
	}
}

func (p *Provider) RevokeAccess(pc *domain.ProviderConfig, g domain.Grant) error {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}

	client, err := p.getIamClient(pc)
	if err != nil {
		return err
	}

	switch g.Resource.Type {
	case ResourceTypeProject, ResourceTypeOrganization:
		for _, p := range g.Permissions {
			if err := client.RevokeAccess(g.AccountType, g.AccountID, p); err != nil {
				if !errors.Is(err, ErrPermissionNotFound) {
					return err
				}
			}
		}
		return nil

	case ResourceTypeServiceAccount:
		for _, p := range g.Permissions {
			if err := client.RevokeServiceAccountAccess(context.TODO(), g.Resource.URN, g.AccountType, g.AccountID, p); err != nil {
				if !errors.Is(err, ErrPermissionNotFound) {
					return err
				}
			}
		}
		return nil

	default:
		return ErrInvalidResourceType
	}
}

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	return provider.GetRoles(pc, resourceType)
}

func (p *Provider) GetPermissions(_pc *domain.ProviderConfig, _resourceType, role string) ([]interface{}, error) {
	// TODO: validate if role is a valid gcloud iam role
	return p.PermissionManager.GetPermissions(_pc, _resourceType, role)
}

func (p *Provider) GetAccountTypes() []string {
	return []string{
		AccountTypeUser,
		AccountTypeServiceAccount,
		AccountTypeGroup,
	}
}

func (p *Provider) ListAccess(ctx context.Context, pc domain.ProviderConfig, resources []*domain.Resource) (domain.MapResourceAccess, error) {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	client, err := p.getIamClient(&pc)
	if err != nil {
		return nil, fmt.Errorf("initializing iam client: %w", err)
	}

	return client.ListAccess(ctx, resources)
}

func (p *Provider) getIamClient(pc *domain.ProviderConfig) (GcloudIamClient, error) {
	var credentials Credentials
	if err := mapstructure.Decode(pc.Credentials, &credentials); err != nil {
		return nil, err
	}
	providerURN := pc.URN

	if p.Clients[providerURN] != nil {
		return p.Clients[providerURN], nil
	}

	credentials.Decrypt(p.crypto)
	client, err := newIamClient([]byte(credentials.ServiceAccountKey), credentials.ResourceName)
	if err != nil {
		return nil, err
	}

	p.Clients[providerURN] = client
	return client, nil
}
