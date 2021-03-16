package resource

import "github.com/odpf/guardian/domain"

type updatePayload struct {
	Details map[string]interface{} `json:"details"`
	Labels  map[string]interface{} `json:"labels"`
}

func (p *updatePayload) toDomain() *domain.Resource {
	return &domain.Resource{
		Details: p.Details,
		Labels:  p.Labels,
	}
}
