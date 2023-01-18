package newpoc

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
)

var (
	ErrShouldHaveOneResource = errors.New("gcloud_iam should have one resource")
	ErrInvalidCredentials    = errors.New("invalid credentials type")
	ErrRolesShouldNotBeEmpty = errors.New("gcloud_iam provider should not have empty roles")

	resourceTypeValidation = fmt.Sprintf("oneof=%s %s", ResourceTypeProject, ResourceTypeOrganization)
)

// type ConfigManager interface {
// 	Validate(context.Context, *domain.Provider) error
// 	Encrypt(context.Context, *domain.Provider) error
// 	Decrypt(context.Context, *domain.Provider) error
// }

type credentials struct {
	ServiceAccountKey string `mapstructure:"service_account_key" json:"service_account_key" validate:"required"`
	ResourceName      string `mapstructure:"resource_name" json:"resource_name" validate:"startswith=projects/|startswith=organizations/"`
}

func (c *credentials) Decode(v interface{}) error {
	return mapstructure.Decode(v, c)
}

func (c *credentials) Validate(validator *validator.Validate) error {
	if err := validator.Struct(c); err != nil {
		return err
	}
	return nil
}

type ConfigManager struct {
	validator *validator.Validate
	crypto    domain.Crypto
}

func (m ConfigManager) Validate(ctx context.Context, p *domain.Provider) error {
	// TODO: validate p is not nil

	// validate credentials
	creds := new(credentials)
	if err := creds.Decode(p.Config.Credentials); err != nil {
		return fmt.Errorf("decoding credentials: %w", err)
	}
	if err := creds.Validate(m.validator); err != nil {
		return fmt.Errorf("validating credentials: %w", err)
	}

	// validate resource config
	if len(p.Config.Resources) != 1 {
		return ErrShouldHaveOneResource
	}
	rc := p.Config.Resources[0]
	if err := m.validator.Var(rc.Type, resourceTypeValidation); err != nil {
		return fmt.Errorf("validating resource type %q: %w", rc.Type, err)
	}
	if len(rc.Roles) == 0 {
		return ErrRolesShouldNotBeEmpty
	}

	return nil
}

func (m ConfigManager) Encrypt(ctx context.Context, p *domain.Provider) error {
	credentials := new(credentials)
	if err := credentials.Decode(p.Config.Credentials); err != nil {
		return ErrInvalidCredentials
	}

	// TODO: check if creds value is the decrypted one

	encryptedSA, err := m.crypto.Encrypt(credentials.ServiceAccountKey)
	if err != nil {
		return err
	}

	credentials.ServiceAccountKey = encryptedSA
	p.Config.Credentials = credentials

	return nil
}

func (m ConfigManager) Decrypt(ctx context.Context, p *domain.Provider) error {
	credentials := new(credentials)
	if err := credentials.Decode(p.Config.Credentials); err != nil {
		return ErrInvalidCredentials
	}

	// TODO: check if creds value is the encrypted one

	decryptedSA, err := m.crypto.Decrypt(credentials.ServiceAccountKey)
	if err != nil {
		return err
	}

	credentials.ServiceAccountKey = decryptedSA
	p.Config.Credentials = credentials

	return nil
}
