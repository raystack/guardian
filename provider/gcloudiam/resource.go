package gcloudiam

import "github.com/odpf/guardian/domain"

const (
	ResourceTypeGcloudIam = "gcloud_iam"
	ResourceTypeRole      = "role"
)

type Role struct {
	Name        string
	Title       string
	Description string
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
