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
	Host       string `json:"host" mapstructure:"host" validate:"required,url"`
	Username   string `json:"username" mapstructure:"username" validate:"required"`
	Password   string `json:"password" mapstructure:"password" validate:"required"`
	ContentURL string `json:"content_url" mapstructure:"content_url" validate:"required"`
}

var permissionNames = map[string][]string{
	ResourceTypeWorkbook: {"AddComment", "ChangeHierarchy", "ChangePermissions", "Delete", "ExportData", "ExportImage", "ExportXml", "Filter", "Read", "ShareView", "ViewComments", "ViewUnderlyingData", "WebAuthoring", "Write"},
	ResourceTypeFlow:     {"ChangeHierarchy", "ChangePermissions", "Delete", "Execute", "ExportXml", "Read", "Write"},
}

var permissionModes = []string{"Allow", "Deny"}

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
	Name   string `json:"name" mapstructure:"name" validate:"required"`
	Target string `json:"target,omitempty" mapstructure:"target"`
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

func (c *Config) getValidationString(resource string) string {
	validation := "oneof="

	for _, mode := range permissionModes {
		for _, permission := range permissionNames[resource] {
			validation = fmt.Sprintf("%v%v:%v ", validation, permission, mode)
		}
	}
	return validation
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
		nameValidation = c.getValidationString(ResourceTypeWorkbook)
	} else if resourceType == ResourceTypeFlow {
		nameValidation = c.getValidationString(ResourceTypeFlow)
	}

	if err := c.validator.Var(pc.Name, nameValidation); err != nil {
		return nil, err
	}

	return &pc, nil
}
