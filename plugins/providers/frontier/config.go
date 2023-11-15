package frontier

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/raystack/guardian/domain"
)

const (
	RoleOrgViewer        = "app_organization_viewer"
	RoleOrgManager       = "app_organization_manager"
	RoleOrgOwner         = "app_organization_owner"
	RoleOrgAccessManager = "app_organization_accessmanager"
	RoleProjectOwner     = "app_project_owner"
	RoleProjectManager   = "app_project_manager"
	RoleProjectViewer    = "app_project_viewer"
	RoleGroupOwner       = "app_group_owner"
	RoleGroupMember      = "app_group_member"

	AccountTypeUser = "user"
)

type Credentials struct {
	Host       string `json:"host" mapstructure:"host" validate:"required"`
	AuthHeader string `json:"auth_header" mapstructure:"auth_header" validate:"required"`
	AuthEmail  string `json:"auth_email" mapstructure:"auth_email" validate:"required"`
}

type Permission string

type Config struct {
	ProviderConfig *domain.ProviderConfig
	valid          bool
	validator      *validator.Validate
}

func NewConfig(pc *domain.ProviderConfig) *Config {
	return &Config{
		ProviderConfig: pc,
		validator:      validator.New(),
	}
}

func (c *Config) ParseAndValidate() error {
	return c.parseAndValidate()
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
	resourceTypeValidation := fmt.Sprintf("oneof=%s %s %s", ResourceTypeTeam, ResourceTypeProject, ResourceTypeOrganization)
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
	if resourceType == ResourceTypeTeam {
		nameValidation = fmt.Sprintf("oneof=%s %s", RoleGroupOwner, RoleGroupMember)
	} else if resourceType == ResourceTypeProject {
		nameValidation = fmt.Sprintf("oneof=%s %s %s", RoleProjectManager, RoleProjectOwner, RoleProjectViewer)
	} else if resourceType == ResourceTypeOrganization {
		nameValidation = fmt.Sprintf("oneof=%s %s %s %s", RoleOrgViewer, RoleOrgManager, RoleOrgOwner, RoleOrgAccessManager)
	}

	if err := c.validator.Var(pc, nameValidation); err != nil {
		return nil, err
	}

	return &pc, nil
}
