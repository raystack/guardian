package providers

import (
	"context"

	"github.com/goto/guardian/domain"
)

type Client interface {
	GetType() string
	CreateConfig(*domain.ProviderConfig) error
	GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error)
	GrantAccess(*domain.ProviderConfig, domain.Grant) error
	RevokeAccess(*domain.ProviderConfig, domain.Grant) error
	GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error)
	GetAccountTypes() []string
	ListAccess(context.Context, domain.ProviderConfig, []*domain.Resource) (domain.MapResourceAccess, error)
}

type PermissionManager interface {
	GetPermissions(p *domain.ProviderConfig, resourceType, role string) ([]interface{}, error)
}
