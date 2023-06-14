package gcs

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/raystack/guardian/domain"
)

const (
	BucketRoleAdmin         = "roles/storage.admin"
	BucketRoleOwner         = "roles/storage.legacyBucketOwner "
	BucketRoleReader        = "roles/storage.legacyBucketReader"
	BucketRoleWriter        = "roles/storage.legacyBucketWriter"
	BucketRoleObjectOwner   = "roles/storage.legacyObjectOwner"
	BucketRoleObjectReader  = "roles/storage.legacyObjectReader"
	BucketRoleObjectAdmin   = "roles/storage.objectAdmin"
	BucketRoleObjectCreator = "roles/storage.objectCreator"
	BucketRoleObjectViewer  = "roles/storage.objectViewer"

	AccountTypeUser           = "user"
	AccountTypeServiceAccount = "serviceAccount"
	AccountTypeGroup          = "group"
	AccountTypeDomain         = "domain"
)

var (
	AllowedAccountTypes = []string{
		AccountTypeUser,
		AccountTypeServiceAccount,
		AccountTypeGroup,
		AccountTypeDomain,
	}
)

type Config struct {
	ProviderConfig *domain.ProviderConfig

	crypto    domain.Crypto
	validator *validator.Validate
}

type Credentials struct {
	ServiceAccountKey string `json:"service_account_key" mapstructure:"service_account_key" validate:"required,base64"`
	ResourceName      string `json:"resource_name" mapstructure:"resource_name" validate:"required"`
}

func (c *Credentials) Decrypt(decryptor domain.Decryptor) error {
	if c == nil {
		return ErrUnableToDecryptNilCredentials
	}

	decryptedServiceAccount, err := decryptor.Decrypt(c.ServiceAccountKey)
	if err != nil {
		return err
	}

	c.ServiceAccountKey = decryptedServiceAccount
	return nil
}

func (c *Credentials) Encrypt(encryptor domain.Encryptor) error {
	if c == nil {
		return ErrUnableToEncryptNilCredentials
	}

	encryptedServiceAccount, err := encryptor.Encrypt(c.ServiceAccountKey)
	if err != nil {
		return err
	}

	c.ServiceAccountKey = encryptedServiceAccount
	return nil
}

type Permission string

func NewConfig(pc *domain.ProviderConfig, crypto domain.Crypto) *Config {
	return &Config{
		ProviderConfig: pc,
		validator:      validator.New(),
		crypto:         crypto,
	}
}

func (c *Config) parseAndValidate() error {
	validationError := []error{}

	credentials, err := c.validateCredentials(c.ProviderConfig.Credentials)
	if err != nil {
		validationError = append(validationError, err)
	} else {
		c.ProviderConfig.Credentials = credentials
	}

	for _, r := range c.ProviderConfig.Resources {
		if err := c.validateResourceConfig(r); err != nil {
			validationError = append(validationError, err)
		}
	}

	if len(validationError) > 0 {
		errorStrings := []string{}
		for _, err := range validationError {
			errorStrings = append(errorStrings, err.Error())
		}
		return errors.New(strings.Join(errorStrings, "\n"))
	}

	return nil
}

func (c *Config) validateCredentials(value interface{}) (*Credentials, error) {
	var credentials Credentials
	if err := mapstructure.Decode(value, &credentials); err != nil {
		return nil, err
	}

	if err := c.validator.Struct(credentials); err != nil {
		return nil, err
	}

	saKeyJson, err := base64.StdEncoding.DecodeString(credentials.ServiceAccountKey)
	if err != nil {
		return nil, err
	}

	credentials.ServiceAccountKey = string(saKeyJson)

	return &credentials, nil
}

func (c *Config) validateResourceConfig(resource *domain.ResourceConfig) error {
	resourceTypeValidation := fmt.Sprintf("oneof=%s", ResourceTypeBucket)
	if err := c.validator.Var(resource.Type, resourceTypeValidation); err != nil {
		return fmt.Errorf("validating resource type: %w", err)
	}

	for _, role := range resource.Roles {
		for i, permission := range role.Permissions {
			if permissionConfig, err := c.validatePermission(resource.Type, permission); err != nil {
				return fmt.Errorf("validating permissions: %w", err)
			} else {
				role.Permissions[i] = permissionConfig
			}
		}
	}
	return nil
}

func (c *Config) validatePermission(resourceType string, value interface{}) (*Permission, error) {
	permissionConfig, ok := value.(string)
	if !ok {
		return nil, ErrInvalidPermissionConfig
	}

	var pc Permission
	if err := mapstructure.Decode(permissionConfig, &pc); err != nil {
		return nil, err
	}

	var nameValidation string
	if resourceType == ResourceTypeBucket {
		nameValidation = fmt.Sprintf("oneof=%s %s %s %s %s %s %s %s %s", BucketRoleAdmin, BucketRoleOwner, BucketRoleReader, BucketRoleWriter, BucketRoleObjectOwner, BucketRoleObjectReader, BucketRoleObjectAdmin, BucketRoleObjectCreator, BucketRoleObjectViewer)
	}
	if err := c.validator.Var(pc, nameValidation); err != nil {
		return nil, err
	}

	return &pc, nil
}
