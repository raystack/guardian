package providers

import "github.com/odpf/guardian/domain"

type Client interface {
	GetType() string
	CreateConfig(*domain.ProviderConfig) error
	GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error)
	GrantAccess(*domain.ProviderConfig, domain.Access) error
	RevokeAccess(*domain.ProviderConfig, domain.Access) error
	GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error)
	GetAccountTypes() []string
}

type PermissionManager interface {
	GetPermissions(p *domain.ProviderConfig, resourceType, role string) ([]interface{}, error)
}
