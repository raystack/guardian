package policytag

import (
	"github.com/odpf/guardian/domain"
)

const (
	ResourceTypeTag               = "tag"
	FineGrainReaderPermission     = "roles/datacatalog.categoryFineGrainedReader"
	FineGrainReaderPermissionRole = "fineGrainReader"
	US                            = "us"
	PageSize                      = 100
)

// Policy is a reference to a Data Catalog Policy Tag
type Policy struct {
	Name        string
	DisplayName string
	Description string
}

func (p *Policy) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeTag,
		Name: p.DisplayName,
		URN:  p.Name,
		Details: map[string]interface{}{
			"description": p.Description,
		},
	}
}

func (p *Policy) FromDomain(r *domain.Resource) {
	p.Name = r.URN
}
