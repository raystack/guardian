package newpoc

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/goto/guardian/domain"
	"github.com/mitchellh/mapstructure"
)

const (
	ProviderType = "gcloud_iam"
)

var (
	ErrShouldHaveOneResource  = errors.New("gcloud_iam should have one resource")
	ErrInvalidCredentials     = errors.New("invalid credentials type")
	ErrRolesShouldNotBeEmpty  = errors.New("gcloud_iam provider should not have empty roles")
	ErrProviderShouldNotBeNil = errors.New("provider should not be nil")

	resourceTypeValidation = fmt.Sprintf("oneof=%s %s", ResourceTypeProject, ResourceTypeOrganization)
)

type credentials struct {
	ServiceAccountKey string `mapstructure:"service_account_key" json:"service_account_key" validate:"required"`
	ResourceName      string `mapstructure:"resource_name" json:"resource_name" validate:"startswith=projects/|startswith=organizations/"`
}

func (c *credentials) Decode(v interface{}) error {
	if decodedCreds, ok := v.(*credentials); ok {
		*c = *decodedCreds
		return nil
	}
	return mapstructure.Decode(v, c)
}

func (c *credentials) Validate(validator *validator.Validate) error {
	if err := validator.Struct(c); err != nil {
		return err
	}
	return nil
}

type Config struct {
	pc           *domain.ProviderConfig
	credentials  *credentials
	resourceType string
	resourceID   string
}

func NewConfig(pc *domain.ProviderConfig) (*Config, error) {
	if pc.Type != ProviderType {
		return nil, fmt.Errorf("%w: expected provider type: %q", ErrInvalidProviderType, ProviderType)
	}

	creds := new(credentials)
	if err := creds.Decode(pc.Credentials); err != nil {
		return nil, fmt.Errorf("decoding credentials: %w", err)
	}
	pc.Credentials = creds

	resourceType, resourceID, err := getResourceIdentifier(creds.ResourceName)
	if err != nil {
		return nil, err
	}

	return &Config{
		pc:           pc,
		credentials:  creds,
		resourceType: resourceType,
		resourceID:   resourceID,
	}, nil
}

func (c *Config) GetProviderConfig() *domain.ProviderConfig {
	return c.pc
}

func (c *Config) Validate(ctx context.Context, validator *validator.Validate) error {
	if c.pc == nil {
		return ErrProviderShouldNotBeNil
	}

	// validate credentials
	if err := c.credentials.Validate(validator); err != nil {
		return fmt.Errorf("validating credentials: %w", err)
	}

	// validate resource config
	if len(c.pc.Resources) != 1 {
		return ErrShouldHaveOneResource
	}
	rc := c.pc.Resources[0]
	if err := validator.Var(rc.Type, resourceTypeValidation); err != nil {
		return fmt.Errorf("validating resource type %q: %w", rc.Type, err)
	}
	if len(rc.Roles) == 0 {
		return ErrRolesShouldNotBeEmpty
	}

	// validate permissions (gcloud roles)
	tmpClient, err := NewClient(c) // TODO: client should be overrideable with an existing client instance through option param
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}
	grantableRoles, err := tmpClient.listGrantableRoles(ctx)
	if err != nil {
		return fmt.Errorf("listing grantable roles: %w", err)
	}
	grantableRolesMap := make(map[string]bool)
	for _, r := range grantableRoles {
		grantableRolesMap[r] = true
	}
	for _, role := range rc.Roles {
		for _, permission := range role.Permissions {
			permissionString, ok := permission.(string)
			if !ok {
				return fmt.Errorf("invalid permission type for %q: %T", permission, permission)
			}
			if !grantableRolesMap[permissionString] {
				return fmt.Errorf("permission %q is not grantable to %q", permissionString, c.credentials.ResourceName)
			}
		}
	}

	return nil
}

// Encrypt encrypts the service account key in ProviderConfig.Credentials
func (c *Config) Encrypt(encryptor domain.Encryptor) error {
	credentialsString, ok := c.pc.Credentials.(map[string]interface{})["service_account_key"].(string)
	if !ok {
		return fmt.Errorf("invalid credentials type: %T", c.pc.Credentials)
	}

	encryptedSA, err := encryptor.Encrypt(credentialsString)
	if err != nil {
		return err
	}

	c.pc.Credentials.(map[string]interface{})["service_account_key"] = encryptedSA
	return nil
}

func (c *Config) Decrypt(decryptor domain.Decryptor) error {
	credentialsString, ok := c.pc.Credentials.(map[string]interface{})["service_account_key"].(string)
	if !ok {
		return fmt.Errorf("invalid credentials type: %T", c.pc.Credentials)
	}

	decryptedSA, err := decryptor.Decrypt(credentialsString)
	if err != nil {
		return err
	}

	c.pc.Credentials.(map[string]interface{})["service_account_key"] = decryptedSA
	return nil
}
