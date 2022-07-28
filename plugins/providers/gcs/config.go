package gcs

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	"google.golang.org/api/option"
)

type Config struct {
	ProviderConfig *domain.ProviderConfig
	valid          bool

	crypto    domain.Crypto
	validator *validator.Validate
}

func (c *Config) EncryptCredentials() error {
	if err := c.parseAndValidate(); err != nil {
		return err
	}

	credentials, ok := c.ProviderConfig.Credentials.(*Credentials)
	if !ok {
		return ErrInvalidCredentialsType
	}

	if err := credentials.Encrypt(c.crypto); err != nil {
		return err
	}

	c.ProviderConfig.Credentials = credentials
	return nil
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

func (c *Config) ParseAndValidate() error {
	return c.parseAndValidate()
}

func (c *Config) parseAndValidate() error {
	if c.valid {
		return nil
	}

	validationError := []error{}

	credentials, err := c.validateCredentials(c.ProviderConfig.Credentials)
	if err != nil {
		validationError = append(validationError, err)
	} else {
		c.ProviderConfig.Credentials = credentials
	}
	//  Todo- Resource.go
	ctx := context.TODO()
	saKey := credentials.ServiceAccountKey
	fmt.Printf("saKey: %v\n", saKey)
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(saKey)))
	if err != nil {
		return fmt.Errorf("initialising gcs client: %w", err)
	}
	defer client.Close()

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

	c.valid = true
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

// Todo - Resource type and its struct

func (c *Config) validateResourceConfig(resource *domain.ResourceConfig) error {
	resourceTypeValidation := fmt.Sprintf("oneof=%s %s", ResourceTypeBucket, ResourceTypeObject)
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
		nameValidation = "oneof=roles/storage.admin roles/storage.legacyBucketOwner roles/storage.legacyBucketReader roles/storage.legacyBucketWriter roles/storage.legacyObjectOwner roles/storage.legacyObjectReader roles/storage.objectAdmin roles/storage.objectCreator roles/storage.objectViewer"
	} else if resourceType == ResourceTypeObject {
		nameValidation = "oneof=viewer owner" //Todo- check with API
	}

	if err := c.validator.Var(pc, nameValidation); err != nil {
		return nil, err
	}

	return &pc, nil
}
