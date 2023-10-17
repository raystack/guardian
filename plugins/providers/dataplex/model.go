package dataplex

import (
	"github.com/raystack/guardian/domain"
)

const (
	ResourceTypeTag               = "tag"
	FineGrainReaderPermission     = "roles/datacatalog.categoryFineGrainedReader"
	FineGrainReaderPermissionRole = "fineGrainReader"
	PageSize                      = 100
)

// Policy is a reference to a Dataplex Policy Tag
type Policy struct {
	Name                string
	DisplayName         string
	Description         string
	TaxonomyDisplayName string
}

func (p *Policy) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeTag,
		Name: p.DisplayName,
		URN:  p.Name,
		Details: map[string]interface{}{
			"description": p.Description,
			"taxonomy":    p.TaxonomyDisplayName,
		},
	}
}

func (p *Policy) FromDomain(r *domain.Resource) {
	p.Name = r.URN
}
