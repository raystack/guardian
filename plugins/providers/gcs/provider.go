package gcs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/iam"
	"github.com/goto/guardian/core/provider"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/utils"
	"github.com/mitchellh/mapstructure"
)

//go:generate mockery --name=GCSClient --exported --with-expecter
type GCSClient interface {
	GetBuckets(context.Context) ([]*Bucket, error)
	GrantBucketAccess(ctx context.Context, b Bucket, identity string, roleName iam.RoleName) error
	RevokeBucketAccess(ctx context.Context, b Bucket, identity string, roleName iam.RoleName) error
	ListAccess(context.Context, []*domain.Resource) (domain.MapResourceAccess, error)
}

//go:generate mockery --name=Crypto --exported --with-expecter
type Crypto interface {
	domain.Crypto
}

type Provider struct {
	provider.UnimplementedClient
	provider.PermissionManager

	typeName string
	Clients  map[string]GCSClient
	crypto   Crypto
}

func NewProvider(typeName string, crypto Crypto) *Provider {
	return &Provider{
		typeName: typeName,
		Clients:  map[string]GCSClient{},
		crypto:   crypto,
	}
}

func (p *Provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc, p.crypto)

	if err := c.parseAndValidate(); err != nil {
		return err
	}

	credentials, ok := c.ProviderConfig.Credentials.(*Credentials)
	if !ok {
		return ErrInvalidCredentialsType
	}

	if err := credentials.Encrypt(c.crypto); err != nil {
		return fmt.Errorf("encrypting creds: %w", err)
	}

	c.ProviderConfig.Credentials = credentials
	return nil
}

func (p *Provider) GetResources(ctx context.Context, pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	client, err := p.getGCSClient(ctx, *pc)
	if err != nil {
		return nil, err
	}

	resourceTypes := []string{}
	for _, rc := range pc.Resources {
		resourceTypes = append(resourceTypes, rc.Type)
	}

	var resources []*domain.Resource
	buckets, err := client.GetBuckets(ctx)
	if err != nil {
		return nil, err
	}
	for _, b := range buckets {
		bucketResource := b.toDomain()
		bucketResource.ProviderType = pc.Type
		bucketResource.ProviderURN = pc.URN
		if utils.ContainsString(resourceTypes, ResourceTypeBucket) {
			resources = append(resources, bucketResource)
		}
	}

	return resources, nil
}

func (p *Provider) GetType() string {
	return p.typeName
}

func (p *Provider) GrantAccess(ctx context.Context, pc *domain.ProviderConfig, a domain.Grant) error {
	if err := validateProviderConfigAndAppealParams(pc, a); err != nil {
		return fmt.Errorf("invalid provider/appeal config: %w", err)
	}

	permissions := getPermissions(a)

	client, err := p.getGCSClient(ctx, *pc)
	if err != nil {
		return fmt.Errorf("error in getting new client: %w", err)
	}
	// identity is AccountType : AccountID, eg: "serviceAccount:test@email.com"
	identity := fmt.Sprintf("%s:%s", a.AccountType, a.AccountID)
	if a.Resource.Type == ResourceTypeBucket {
		b := new(Bucket)
		if err := b.fromDomain(a.Resource); err != nil {
			return fmt.Errorf("from Domain func error: %w", err)
		}
		for _, p := range permissions {
			role := iam.RoleName(string(p))
			if err := client.GrantBucketAccess(ctx, *b, identity, role); err != nil {
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

func (p *Provider) RevokeAccess(ctx context.Context, pc *domain.ProviderConfig, a domain.Grant) error {
	if err := validateProviderConfigAndAppealParams(pc, a); err != nil {
		return fmt.Errorf("invalid provider/appeal config: %w", err)
	}

	permissions := getPermissions(a)

	client, err := p.getGCSClient(ctx, *pc)
	if err != nil {
		return fmt.Errorf("error in getting new client: %w", err)
	}

	// identity is AccountType : AccountID, eg: "serviceAccount:test@email.com"
	identity := fmt.Sprintf("%s:%s", a.AccountType, a.AccountID)
	if a.Resource.Type == ResourceTypeBucket {
		b := new(Bucket)
		if err := b.fromDomain(a.Resource); err != nil {
			return fmt.Errorf("from Domain func error: %w", err)
		}
		for _, p := range permissions {
			var role iam.RoleName = iam.RoleName(string(p))
			if err := client.RevokeBucketAccess(ctx, *b, identity, role); err != nil {
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

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	return provider.GetRoles(pc, resourceType)
}

func (p *Provider) GetAccountTypes() []string {
	return []string{AccountTypeUser, AccountTypeServiceAccount, AccountTypeGroup, AccountTypeDomain}
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

func (p *Provider) ListAccess(ctx context.Context, pc domain.ProviderConfig, resources []*domain.Resource) (domain.MapResourceAccess, error) {
	client, err := p.getGCSClient(ctx, pc)
	if err != nil {
		return nil, err
	}

	return client.ListAccess(ctx, resources)
}

func (p *Provider) getGCSClient(ctx context.Context, pc domain.ProviderConfig) (GCSClient, error) {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return nil, fmt.Errorf("decoding credentials: %w", err)
	}

	if err := creds.Decrypt(p.crypto); err != nil {
		return nil, fmt.Errorf("decrypting credentials: %w", err)
	}

	projectID := strings.Replace(creds.ResourceName, "projects/", "", 1)
	if p.Clients[projectID] != nil {
		return p.Clients[projectID], nil
	}

	client, err := newGCSClient(ctx, projectID, []byte(creds.ServiceAccountKey))
	if err != nil {
		return nil, err
	}

	p.Clients[projectID] = client
	return client, nil
}
