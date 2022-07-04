package noop

import (
	"errors"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/log"
)

var (
	ErrInvalidProviderType         = errors.New("provider type not equal to no_op")
	ErrInvalidAllowedAccountTypes  = errors.New("allowed account types for no_op provider is only \"user\"")
	ErrInvalidCredentials          = errors.New("credentials should be empty")
	ErrInvalidResourceConfigLength = errors.New("resource config length should be 1")
	ErrInvalidResourceConfigType   = errors.New("resouce config type should be \"no_op\"")
	ErrInvalidRolePermissions      = errors.New("permissions should be empty")
)

type Provider struct {
	typeName string

	logger log.Logger
}

func NewProvider(typeName string, logger log.Logger) *Provider {
	return &Provider{
		typeName: typeName,

		logger: logger,
	}
}

func (p *Provider) GetType() string {
	return p.typeName
}

func (p *Provider) CreateConfig(cfg *domain.ProviderConfig) error {
	if cfg.Type != "no_op" {
		return ErrInvalidProviderType
	}

	if len(cfg.AllowedAccountTypes) != 1 || cfg.AllowedAccountTypes[0] != "user" {
		return ErrInvalidAllowedAccountTypes
	}

	if cfg.Credentials != nil {
		return ErrInvalidCredentials
	}

	if len(cfg.Resources) != 1 {
		return ErrInvalidResourceConfigLength
	}
	if cfg.Resources[0].Type != "no_op" {
		return ErrInvalidResourceConfigType
	}

	for _, r := range cfg.Resources[0].Roles {
		if r.Permissions != nil || len(r.Permissions) != 0 {
			return ErrInvalidRolePermissions
		}
	}

	return nil
}

func (p *Provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	return []*domain.Resource{
		{
			ProviderType: "no_op",
			ProviderURN:  pc.URN,
			Type:         "no_op",
			URN:          pc.URN,
			Name:         pc.URN,
		},
	}, nil
}

func (p *Provider) GrantAccess(*domain.ProviderConfig, *domain.Appeal) error {
	return nil
}

func (p *Provider) RevokeAccess(*domain.ProviderConfig, *domain.Appeal) error {
	return nil
}

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	// TODO
	return nil, nil
}

func (p *Provider) GetAccountTypes() []string {
	// TODO
	return nil
}
