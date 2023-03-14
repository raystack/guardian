package noop

import (
	"errors"

	"github.com/goto/guardian/core/provider"
	"github.com/goto/guardian/domain"
	"github.com/goto/salt/log"
)

var (
	ErrInvalidProviderType         = errors.New("provider type not equal to noop")
	ErrInvalidAllowedAccountTypes  = errors.New("allowed account types for noop provider is only \"user\"")
	ErrInvalidCredentials          = errors.New("credentials should be empty")
	ErrInvalidResourceConfigLength = errors.New("resource config length should be 1")
	ErrInvalidResourceConfigType   = errors.New("resource config type should be \"noop\"")
	ErrInvalidRolePermissions      = errors.New("permissions should be empty")
)

type Provider struct {
	provider.UnimplementedClient
	provider.PermissionManager

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
	if cfg.Type != domain.ProviderTypeNoOp {
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

	if cfg.Resources[0].Type != ResourceTypeNoOp {
		return ErrInvalidResourceConfigType
	}

	for _, r := range cfg.Resources[0].Roles {
		if len(r.Permissions) != 0 {
			return ErrInvalidRolePermissions
		}
	}

	return nil
}

func (p *Provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	return []*domain.Resource{
		{
			ProviderType: domain.ProviderTypeNoOp,
			ProviderURN:  pc.URN,
			Type:         ResourceTypeNoOp,
			URN:          pc.URN,
			Name:         pc.URN,
		},
	}, nil
}

func (p *Provider) GrantAccess(*domain.ProviderConfig, domain.Grant) error {
	return nil
}

func (p *Provider) RevokeAccess(*domain.ProviderConfig, domain.Grant) error {
	return nil
}

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	return provider.GetRoles(pc, resourceType)
}

func (p *Provider) GetAccountTypes() []string {
	return []string{"user"}
}
