package gcs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
)

type Config struct {
	ProviderConfig *domain.ProviderConfig
	valid          bool //ask why required?

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

type Permission string

func NewConfig(pc *domain.ProviderConfig, crypto domain.Crypto) *Config {
	return &Config{
		ProviderConfig: pc,
		validator:      validator.New(),
		crypto:         crypto,
	}
}

func (c *Config) parseAndValidate() error {
	if c.valid {
		return nil
	}

	validationError := []error{}

	if credentials, err := c.validateCredentials(c.ProviderConfig.Credentials); err != nil {
		validationError = append(validationError, err)
	} else {
		c.ProviderConfig.Credentials = credentials
	}
	//  Todo- Resource.go
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

	return &credentials, nil
}

// Todo - Resource type and its struct

func (c *Config) validateResourceConfig(resource *domain.ResourceConfig) error {
	resourceTypeValidation := fmt.Sprintf("oneof=%s%s", ResourceTypeBucket, ResourceTypeObject)
	if err := c.validator.Var(resource.Type, resourceTypeValidation); err != nil {
		return err
	}

	for _, role := range resource.Roles {
		for i, permission := range role.Permissions {
			if permissionConfig, err := c.validatePermission(resource.Type, permission); err != nil {
				return err
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
		nameValidation = "oneof=view edit admin" //Todo- check with API
	} else if resourceType == ResourceTypeObject {
		nameValidation = "oneof=viewer owner" //Todo- check with API
	}

	if err := c.validator.Var(pc, nameValidation); err != nil {
		return nil, err
	}

	return &pc, nil
}
