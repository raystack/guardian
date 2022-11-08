package bigquery

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
)

const (
	AccountTypeUser           = "user"
	AccountTypeServiceAccount = "serviceAccount"
)

// Credentials is the authentication configuration used by the bigquery client
type Credentials struct {
	ServiceAccountKey string `mapstructure:"service_account_key" json:"service_account_key" validate:"required,base64"`
	ResourceName      string `mapstructure:"resource_name" json:"resource_name" validate:"startswith=projects/"`
}

// Encrypt encrypts BigQuery credentials
func (c *Credentials) Encrypt(encryptor domain.Encryptor) error {
	if c == nil {
		return ErrUnableToEncryptNilCredentials
	}

	encryptedCredentials, err := encryptor.Encrypt(c.ServiceAccountKey)
	if err != nil {
		return err
	}

	c.ServiceAccountKey = encryptedCredentials
	return nil
}

// Decrypt decrypts BigQuery credentials
func (c *Credentials) Decrypt(decryptor domain.Decryptor) error {
	if c == nil {
		return ErrUnableToDecryptNilCredentials
	}

	decryptedCredentials, err := decryptor.Decrypt(c.ServiceAccountKey)
	if err != nil {
		return err
	}

	c.ServiceAccountKey = decryptedCredentials
	return nil
}

// Permission is for mapping role into bigquery permissions
type Permission string

// Config for bigquery provider
type Config struct {
	ProviderConfig *domain.ProviderConfig
	valid          bool

	crypto    domain.Crypto
	validator *validator.Validate
}

// NewConfig returns bigquery config struct
func NewConfig(pc *domain.ProviderConfig, crypto domain.Crypto) *Config {
	return &Config{
		ProviderConfig: pc,
		validator:      validator.New(),
		crypto:         crypto,
	}
}

// ParseAndValidate validates bigquery config within provider config and make the interface{} config value castable into the expected bigquery config value
func (c *Config) ParseAndValidate() error {
	return c.parseAndValidate()
}

// EncryptCredentials encrypts the bigquery credentials config
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

func (c *Config) parseAndValidate() error {
	if c.valid {
		return nil
	}

	credentials, err := c.validateCredentials(c.ProviderConfig.Credentials)
	if err != nil {
		return err
	} else {
		c.ProviderConfig.Credentials = credentials
	}

	projectID := strings.Replace(credentials.ResourceName, "projects/", "", 1)
	client, err := newBigQueryClient(projectID, []byte(credentials.ServiceAccountKey))
	if err != nil {
		return err
	}

	permissionValidationErrors := []error{}

	for _, resource := range c.ProviderConfig.Resources {
		for _, role := range resource.Roles {
			for i, permission := range role.Permissions {
				if permissionConfig, err := c.validatePermission(permission, resource.Type, client); err != nil {
					permissionValidationErrors = append(permissionValidationErrors, err)
				} else {
					role.Permissions[i] = permissionConfig
				}
			}
		}
	}

	if len(permissionValidationErrors) > 0 {
		errorStrings := []string{}
		for _, err := range permissionValidationErrors {
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

func (c *Config) validatePermission(value interface{}, resourceType string, client *bigQueryClient) (*Permission, error) {
	permision, ok := value.(string)
	if !ok {
		return nil, ErrInvalidPermissionConfig
	}

	if resourceType == ResourceTypeDataset {
		roles, err := client.getGrantableRolesForDataset()
		if err != nil {
			return nil, ErrCannotFetchDatabasePermission
		}
		findMatchingRole := false
		for _, role := range roles {
			if string(role) == permision {
				findMatchingRole = true
				break
			}
		}
		if !findMatchingRole {
			return nil, fmt.Errorf("%v: %v", ErrInvalidDatasetPermission, permision)
		}
	} else if resourceType == ResourceTypeTable {
		roles, err := client.getGrantableRolesForTables()
		if err != nil {
			if err == ErrEmptyResource {
				return nil, ErrCannotVerifyTablePermission
			}
			return nil, err
		}

		if !utils.ContainsString(roles, permision) {
			return nil, fmt.Errorf("%v: %v", ErrInvalidTablePermission, permision)
		}
	} else {
		return nil, ErrInvalidResourceType
	}

	configValue := Permission(permision)
	return &configValue, nil
}
