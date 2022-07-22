package provider

import "github.com/odpf/guardian/domain"

func GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	for _, r := range pc.Resources {
		if r.Type == resourceType {
			return r.Roles, nil
		}
	}

	return nil, ErrInvalidResourceType
}

type PermissionManager struct{}

func (m PermissionManager) GetPermissions(pc *domain.ProviderConfig, resourceType, role string) ([]interface{}, error) {
	for _, rc := range pc.Resources {
		if rc.Type != resourceType {
		        continue
		}
		for _, r := range rc.Roles {
			if r.ID == role {
				return r.Permissions, nil
			}
		}
		return nil, ErrInvalidRole
	}
	return nil, ErrInvalidResourceType
}
