package tableau

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
)

type Credentials struct {
	Host     string `json:"host" mapstructure:"host" validate:"required,url"`
	Username string `json:"username" mapstructure:"username" validate:"required"`
	Password string `json:"password" mapstructure:"password" validate:"required"`
}

func (c *Credentials) Encrypt(encryptor domain.Encryptor) error {
	if c == nil {
		return ErrUnableToEncryptNilCredentials
	}

	encryptedPassword, err := encryptor.Encrypt(c.Password)
	if err != nil {
		return err
	}

	c.Password = encryptedPassword
	return nil
}

func (c *Credentials) Decrypt(decryptor domain.Decryptor) error {
	if c == nil {
		return ErrUnableToDecryptNilCredentials
	}

	encryptedPassword, err := decryptor.Decrypt(c.Password)
	if err != nil {
		return err
	}

	c.Password = encryptedPassword
	return nil
}

type PermissionConfig struct {
	Name string `json:"name" mapstructure:"name" validate:"required"`
}

type Config struct {
	ProviderConfig *domain.ProviderConfig
	valid          bool

	crypto    domain.Crypto
	validator *validator.Validate
}

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

func (c *Config) EncryptCredentials() error {
	if err := c.parseAndValidate(); err != nil {
		return err
	}

	credentials, ok := c.ProviderConfig.Credentials.(*Credentials)
	if !ok {
		return ErrInvalidCredentials
	}

	if err := credentials.Encrypt(c.crypto); err != nil {
		return err
	}

	c.ProviderConfig.Credentials = credentials
	return nil
}

func (c *Config) parseAndValidate() error {
	if c.valid {
		return nil
	}

	validationErrors := []error{}

	if credentials, err := c.validateCredentials(c.ProviderConfig.Credentials); err != nil {
		validationErrors = append(validationErrors, err)
	} else {
		c.ProviderConfig.Credentials = credentials
	}

	for _, r := range c.ProviderConfig.Resources {
		if err := c.validateResourceConfig(r); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	if len(validationErrors) > 0 {
		errorStrings := []string{}
		for _, err := range validationErrors {
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

func (c *Config) validateResourceConfig(resource *domain.ResourceConfig) error {
	resourceTypeValidation := fmt.Sprintf("oneof=%s %s", ResourceTypeWorkbook, ResourceTypeFlow)
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

func (c *Config) validatePermission(resourceType string, value interface{}) (*PermissionConfig, error) {
	permissionConfig, ok := value.(map[string]interface{})
	if !ok {
		return nil, ErrInvalidPermissionConfig
	}

	var pc PermissionConfig
	if err := mapstructure.Decode(permissionConfig, &pc); err != nil {
		return nil, err
	}

	var nameValidation string
	if resourceType == ResourceTypeWorkbook {
		nameValidation = "oneof=AddComment:Allow ChangeHierarchy:Allow ChangePermissions:Allow Delete:Allow ExportData:Allow ExportImage:Allow ExportXml:Allow Filter:Allow Read:Allow ShareView:Allow ViewComments:Allow ViewUnderlyingData:Allow WebAuthoring:Allow Write:Allow AddComment:Deny ChangeHierarchy:Deny ChangePermissions:Deny Delete:Deny ExportData:Deny ExportImage:Deny ExportXml:Deny Filter:Deny Read:Deny ShareView:Deny ViewComments:Deny ViewUnderlyingData:Deny WebAuthoring:Deny Write:Deny"
	} else if resourceType == ResourceTypeFlow {
		nameValidation = "oneof=ChangeHierarchy:Allow ChangePermissions:Allow Delete:Allow Execute:Allow ExportXml:Allow Read:Allow Write:Allow ChangeHierarchy:Deny ChangePermissions:Deny Delete:Deny Execute:Deny ExportXml:Deny Read:Deny Write:Deny"
	}

	if err := c.validator.Var(pc.Name, nameValidation); err != nil {
		return nil, err
	}

	return &pc, nil
}
