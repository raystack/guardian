package bigquery

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
)

var (
	// ErrInvalidAuthConfig is the error value for invalid auth config
	ErrInvalidAuthConfig = errors.New("invalid auth config type")
	// ErrInvalidPermissionConfig is the error value for invalid permission config
	ErrInvalidPermissionConfig = errors.New("invalid permission config type")
)

// AuthConfig is authentication configuration used by the bigquery client
type AuthConfig string

// PermissionConfig is for mapping role into bigquery permissions
type PermissionConfig struct {
	Name   string `json:"name" mapstructure:"name" validate:"required"`
	Target string `json:"target,omitempty" mapstructure:"target"`
}

// Config for bigquery provider
type Config struct {
	ProviderConfig *domain.ProviderConfig

	validate *validator.Validate
}

// NewConfig returns bigquery config struct
func NewConfig(pc *domain.ProviderConfig) *Config {
	return &Config{
		ProviderConfig: pc,
		validate:       validator.New(),
	}
}

// Validate validates bigquery config within provider config and make the interface{} config value castable into the expected bigquery config value
func (c *Config) Validate() error {
	validationErrors := []error{}

	if auth, err := c.validateAuth(c.ProviderConfig.Auth); err != nil {
		validationErrors = append(validationErrors, err)
	} else {
		c.ProviderConfig.Auth = auth
	}

	for _, resource := range c.ProviderConfig.Resources {
		for _, role := range resource.Roles {
			for i, permission := range role.Permissions {
				if permissionConfig, err := c.validatePermission(permission); err != nil {
					validationErrors = append(validationErrors, err)
				} else {
					role.Permissions[i] = permissionConfig
				}
			}
		}
	}

	if len(validationErrors) > 0 {
		errorStrings := []string{}
		for _, err := range validationErrors {
			errorStrings = append(errorStrings, err.Error())
		}
		return errors.New(strings.Join(errorStrings, "\n"))
	}

	return nil
}

func (c *Config) validateAuth(value interface{}) (*AuthConfig, error) {
	auth, ok := value.(string)
	if !ok {
		return nil, ErrInvalidAuthConfig
	}

	configValue := AuthConfig(auth)
	return &configValue, c.validate.Var(configValue, "required,base64")
}

func (c *Config) validatePermission(value interface{}) (*PermissionConfig, error) {
	permissionConfig, ok := value.(map[string]interface{})
	if !ok {
		return nil, ErrInvalidPermissionConfig
	}

	var pc PermissionConfig
	if err := mapstructure.Decode(permissionConfig, &pc); err != nil {
		return nil, err
	}

	return &pc, utils.ValidateStruct(pc)
}
