package gcloudiam

import "github.com/odpf/guardian/domain"

const (
	ResourceTypeRole = "role"
)

type Role struct {
	Name        string
	Title       string
	Description string
}

func (r *Role) toDomain() *domain.Resource {
	var details map[string]interface{}
	if r.Description != "" {
		details = map[string]interface{}{
			"description": r.Description,
		}
	}

	return &domain.Resource{
		Type:    ResourceTypeRole,
		Name:    r.Title,
		URN:     r.Name,
		Details: details,
	}
}

func (r *Role) fromDomain(res *domain.Resource) error {
	if res.Type != ResourceTypeRole {
		return ErrInvalidResourceType
	}

	r.Name = res.URN
	r.Title = res.Name
	if res.Details != nil && res.Details["description"] != "" {
		r.Description = res.Details["description"].(string)
	}

	return nil
}
