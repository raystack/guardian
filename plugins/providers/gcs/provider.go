package gcs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/iam"
	"cloud.google.com/go/storage"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Provider struct {
	typeName string
	Clients  map[string]GcsClient
	crypto   domain.Crypto
}

func NewProvider(typeName string, crypto domain.Crypto) *Provider {
	return &Provider{
		typeName: typeName,
		Clients:  map[string]GcsClient{},
		crypto:   crypto,
	}
}

func (p *Provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc, p.crypto)

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

	if err := creds.Decrypt(p.crypto); err != nil {
		return nil, err
	}

	ctx := context.TODO()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(creds.ServiceAccountKey)))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	resourceTypes := []string{}
	for _, rc := range pc.Resources {
		resourceTypes = append(resourceTypes, rc.Type)
	}

	var resources []*domain.Resource

	projectID := strings.Replace(creds.ResourceName, "projects/", "", 1)

	it := client.Buckets(ctx, projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		if utils.ContainsString(resourceTypes, ResourceTypeBucket) {
			b := &Bucket{Name: battrs.Name}
			bucketResource := b.toDomain()
			bucketResource.ProviderType = pc.Type
			bucketResource.ProviderURN = pc.URN
			resources = append(resources, bucketResource)
		}

		if utils.ContainsString(resourceTypes, ResourceTypeObject) {
			objIt := client.Bucket(battrs.Name).Objects(ctx, nil)
			for {
				oattrs, err := objIt.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					return nil, err
				}

				o := &Object{Name: oattrs.Name}
				objectResource := o.toDomain()
				objectResource.ProviderType = pc.Type
				objectResource.ProviderURN = pc.URN
				resources = append(resources, objectResource)
			}
		}
	}

	return resources, nil
}

func (p *Provider) GetType() string {
	return p.typeName
}

func (p *Provider) GrantAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	if err := validateProviderConfigAndAppealParams(pc, a); err != nil {
		return fmt.Errorf("invalid provider/appeal config: %w", err)
	}

	permissions, err := getPermissions(pc.Resources, a)
	if err != nil {
		return fmt.Errorf("error in getting permissions: %w", err)
	}

	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return fmt.Errorf("error in decoding credentials%w", err)
	}

	if err := creds.Decrypt(p.crypto); err != nil {
		return fmt.Errorf("error in decrypting credentials%w", err)
	}

	ctx := context.TODO()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(creds.ServiceAccountKey)))
	if err != nil {
		return fmt.Errorf("error in getting new client: %w", err)
	}
	defer client.Close()

	user := a.AccountID

	if a.Resource.Type == ResourceTypeBucket {
		bucketName := a.Resource.URN
		b := new(Bucket)
		bucket := client.Bucket(bucketName)
		if err := b.fromDomain(a.Resource); err != nil {
			return fmt.Errorf("from Domain func error: %w", err)
		}
		for _, p := range permissions {
			resolvedRole, err := resolveRole(string(p))
			if err != nil {
				if errors.Is(err, ErrPermissionAlreadyExists) {
					return nil
				}
				return fmt.Errorf("error in resolving permissions: %w", err)
			}

			policy, err := bucket.IAM().Policy(ctx)
			if err != nil {
				return fmt.Errorf("Bucket(%q).IAM().Policy: %v", bucketName, err)
			}

			identity := fmt.Sprintf("user:%s", user)           //TODO, the identity should have "group:" or "user:"..  user, serviceAccount also valid
			var role iam.RoleName = iam.RoleName(resolvedRole) //TODO : discuss the roles and edit this    "roles/storage.objectViewer"

			policy.Add(identity, role)
			if err := bucket.IAM().SetPolicy(ctx, policy); err != nil {
				return fmt.Errorf("Bucket(%q).IAM().SetPolicy: %v", bucketName, err)
			}
		}
	}
	return nil
}

func (p *Provider) RevokeAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	if err := validateProviderConfigAndAppealParams(pc, a); err != nil {
		return fmt.Errorf("invalid provider/appeal config: %w", err)
	}

	permissions, err := getPermissions(pc.Resources, a)
	if err != nil {
		return fmt.Errorf("error in getting permissions: %w", err)
	}

	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return fmt.Errorf("error in decoding credentials%w", err)
	}

	if err := creds.Decrypt(p.crypto); err != nil {
		return fmt.Errorf("error in decrypting credentials%w", err)
	}

	ctx := context.TODO()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(creds.ServiceAccountKey)))
	if err != nil {
		return fmt.Errorf("error in getting new client: %w", err)
	}
	defer client.Close()

	user := a.AccountID

	if a.Resource.Type == ResourceTypeBucket {
		bucketName := a.Resource.URN
		b := new(Bucket)
		bucket := client.Bucket(bucketName)
		if err := b.fromDomain(a.Resource); err != nil {
			return fmt.Errorf("from Domain func error: %w", err)
		}
		for _, p := range permissions {
			resolvedRole, err := resolveRole(string(p))
			if err != nil {
				if errors.Is(err, ErrPermissionAlreadyExists) {
					return nil
				}
				return fmt.Errorf("error in resolving permissions: %w", err)
			}

			policy, err := bucket.IAM().Policy(ctx)
			if err != nil {
				return fmt.Errorf("Bucket(%q).IAM().Policy: %v", bucketName, err)
			}

			identity := fmt.Sprintf("user:%s", user)           //TODO, the identity should have "group:" or "user:"..  user, serviceAccount also valid
			var role iam.RoleName = iam.RoleName(resolvedRole) //TODO : discuss the roles and edit this    "roles/storage.objectViewer"

			policy.Remove(identity, role)
			if err := bucket.IAM().SetPolicy(ctx, policy); err != nil {
				return fmt.Errorf("Bucket(%q).IAM().SetPolicy: %v", bucketName, err)
			}
		}
	}
	return nil
}

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	return provider.GetRoles(pc, resourceType)
}

func (p *Provider) GetAccountTypes() []string {
	return []string{AccountTypeUser, AccountTypeServiceAccount, AccountTypeGroup, AccountTypeDomain}
}

func validateProviderConfigAndAppealParams(pc *domain.ProviderConfig, a *domain.Appeal) error {
	if pc == nil {
		return ErrNilProviderConfig
	}
	if a == nil {
		return ErrNilAppeal
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

func getPermissions(resourceConfigs []*domain.ResourceConfig, a *domain.Appeal) ([]Permission, error) {
	var resourceConfig *domain.ResourceConfig
	for _, rc := range resourceConfigs {
		if rc.Type == a.Resource.Type {
			resourceConfig = rc
		}
	}
	if resourceConfig == nil {
		return nil, ErrInvalidResourceType
	}

	var role *domain.Role
	for _, r := range resourceConfig.Roles {
		if r.ID == a.Role {
			role = r
		}
	}
	if role == nil {
		return nil, ErrInvalidRole
	}

	var permissions []Permission
	for _, p := range role.Permissions {
		var permission Permission
		if err := mapstructure.Decode(p, &permission); err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func resolveRole(role string) (string, error) {
	switch role {
	case BucketRoleReader:
		return "roles/storage.legacyBucketReader", nil
	case BucketRoleWriter:
		return "roles/storage.legacyBucketWriter", nil
	case BucketRoleOwner:
		return "roles/storage.legacyBucketOwner", nil
	case BucketRoleObjectAdmin:
		return "roles/storage.objectAdmin", nil
	case BucketRoleAdmin:
		return "roles/storage.admin", nil
	default:
		return "", ErrInvalidRole
	}
}
