package gcloudiam

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/utils"
	"github.com/mitchellh/mapstructure"
)

const (
	AccountTypeUser           = "user"
	AccountTypeServiceAccount = "serviceAccount"
	AccountTypeGroup          = "group"
)

type Credentials struct {
	ServiceAccountKey string `mapstructure:"service_account_key" json:"service_account_key" validate:"required,base64"`
	ResourceName      string `mapstructure:"resource_name" json:"resource_name" validate:"startswith=projects/|startswith=organizations/"`
}

func (c *Credentials) Encrypt(encryptor domain.Encryptor) error {
	if c == nil {
		return ErrUnableToEncryptNilCredentials
	}

	encryptedSAKey, err := encryptor.Encrypt(c.ServiceAccountKey)
	if err != nil {
		return err
	}

	c.ServiceAccountKey = encryptedSAKey
	return nil
}

func (c *Credentials) Decrypt(decryptor domain.Decryptor) error {
	if c == nil {
		return ErrUnableToDecryptNilCredentials
	}

	decryptedSAKey, err := decryptor.Decrypt(c.ServiceAccountKey)
	if err != nil {
		return err
	}

	c.ServiceAccountKey = decryptedSAKey
	return nil
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

	credentials, err := c.validateCredentials(c.ProviderConfig.Credentials)
	if err != nil {
		return err
	} else {
		c.ProviderConfig.Credentials = credentials
	}

	if c.ProviderConfig.Resources == nil || len(c.ProviderConfig.Resources) == 0 {
		return errors.New("empty resource config")
	}
	uniqueResourceTypes := make(map[string]bool)
	for _, rc := range c.ProviderConfig.Resources {
		if _, ok := uniqueResourceTypes[rc.Type]; ok {
			validationErrors = append(validationErrors, fmt.Errorf("duplicate resource type: %q", rc.Type))
		}
		uniqueResourceTypes[rc.Type] = true

		allowedResourceTypes := []string{}
		if strings.HasPrefix(credentials.ResourceName, ResourceNameOrganizationPrefix) {
			allowedResourceTypes = []string{ResourceTypeOrganization}
		} else if strings.HasPrefix(credentials.ResourceName, ResourceNameProjectPrefix) {
			allowedResourceTypes = []string{ResourceTypeProject, ResourceTypeServiceAccount}
		}
		if !utils.ContainsString(allowedResourceTypes, rc.Type) {
			validationErrors = append(validationErrors, fmt.Errorf("invalid resource type: %q", rc.Type))
		}

		if len(rc.Roles) == 0 {
			validationErrors = append(validationErrors, ErrRolesShouldNotBeEmpty)
		}

		// check for duplicates in roles
		rolesMap := make(map[string]bool, 0)
		for _, role := range rc.Roles {
			if val, ok := rolesMap[role.ID]; ok && val {
				validationErrors = append(validationErrors, fmt.Errorf("duplicate role: %q", role.ID))
			} else {
				rolesMap[role.ID] = true
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

func (c *Config) validatePermissions(resource *domain.ResourceConfig, client GcloudIamClient) error {
	iamRoles, err := client.GetGrantableRoles(context.TODO(), resource.Type)
	if err != nil {
		return err
	}
	rolesMap := make(map[string]bool)
	for _, role := range iamRoles {
		rolesMap[role.Name] = true
	}

	for _, ro := range resource.Roles {
		for _, p := range ro.Permissions {
			permission := fmt.Sprint(p)
			if _, ok := rolesMap[permission]; !ok {
				return fmt.Errorf("%w: %v", ErrInvalidProjectRole, permission)
			}
		}
	}

	return nil
}
