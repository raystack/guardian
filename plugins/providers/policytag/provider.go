package policytag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
)

type PolicyTagClient interface {
	GetPolicies(ctx context.Context) ([]*Policy, error)
	GrantPolicyAccess(ctx context.Context, tag *Policy, user, role string) error
	RevokePolicyAccess(ctx context.Context, tag *Policy, user, role string) error
	ListAccess(ctx context.Context, resources []*domain.Resource) (domain.MapResourceAccess, error)
}

type encryptor interface {
	domain.Crypto
}

// Provider for policy tag
type Provider struct {
	Clients   map[string]PolicyTagClient
	typeName  string
	encryptor encryptor
}

// NewProvider returns policy tag provider
func NewProvider(typeName string, c encryptor) *Provider {
	return &Provider{
		typeName:  typeName,
		Clients:   map[string]PolicyTagClient{},
		encryptor: c,
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

func (p *Provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return nil, err
	}

	resourceTypes := pc.GetResourceTypes()
	var resources []*domain.Resource
	if containsString(resourceTypes, ResourceTypeTag) {
		client, err := p.getPolicyTagClient(creds)
		if err != nil {
			return nil, err
		}
		ctx := context.Background()
		policies, err := client.GetPolicies(ctx)
		if err != nil {
			return nil, err
		}
		for _, policy := range policies {
			resource := policy.ToDomain()
			resource.ProviderType = pc.Type
			resource.ProviderURN = pc.URN
			resources = append(resources, resource)
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
	client, err := p.getPolicyTagClient(creds)
	if err != nil {
		return err
	}

	permissions := getPermissions(a)

	ctx := context.Background()
	if a.Resource.Type == ResourceTypeTag {
		policy := new(Policy)
		policy.FromDomain(a.Resource)

		for _, p := range permissions {
			if err := client.GrantPolicyAccess(ctx, policy, fmt.Sprintf("%s:%s", AccountTypeUser, a.AccountID), string(p)); err != nil {
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
	client, err := p.getPolicyTagClient(creds)
	if err != nil {
		return err
	}

	permissions := getPermissions(a)
	ctx := context.Background()

	if a.Resource.Type == ResourceTypeTag {
		policy := new(Policy)
		policy.FromDomain(a.Resource)
		for _, p := range permissions {
			if err := client.RevokePolicyAccess(ctx, policy, fmt.Sprintf("%s:%s", AccountTypeUser, a.AccountID), string(p)); err != nil {
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

func (p *Provider) GetPermissions(pc *domain.ProviderConfig, resourceType, role string) ([]interface{}, error) {
	permissions := make([]interface{}, 0)
	if resourceType != ResourceTypeTag {
		return nil, ErrInvalidResourceType
	}

	if role != FineGrainReaderPermissionRole {
		return nil, ErrInvalidRole
	}
	permissions = append(permissions, FineGrainReaderPermission)
	return permissions, nil
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
	client, err := p.getPolicyTagClient(creds)
	if err != nil {
		return nil, fmt.Errorf("initializing bigquery client: %w", err)
	}

	return client.ListAccess(ctx, resources)
}

func (p *Provider) getPolicyTagClient(credentials Credentials) (PolicyTagClient, error) {
	var projectID, taxonomyLocation string
	parseVariableCounts, err := fmt.Sscanf(strings.ReplaceAll(credentials.ResourceName, "/", " "),
		"projects %s locations %s", &projectID, &taxonomyLocation)
	if err != nil || parseVariableCounts != 2 {
		return nil, ErrInvalidResourceFormatType
	}
	if p.Clients[projectID] != nil {
		return p.Clients[projectID], nil
	}

	err = credentials.Decrypt(p.encryptor)
	if err != nil {
		return nil, ErrUnableToDecryptCredentials
	}
	client, err := newPolicyTagClient(projectID, taxonomyLocation, []byte(credentials.ServiceAccountKey))
	if err != nil {
		return nil, err
	}

	p.Clients[projectID] = client
	return client, nil
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
