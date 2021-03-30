package appeal

import "github.com/odpf/guardian/domain"

type resourceCreatePayload struct {
	ID      uint        `json:"id" validate:"required"`
	Options interface{} `json:"options"`
}

type createPayload struct {
	Email     string                  `json:"email" validate:"required"`
	Resources []resourceCreatePayload `json:"resources" validate:"required,min=1"`
}

func (p *createPayload) toDomain() (string, []*domain.Resource) {
	// TODO: add options
	resources := []*domain.Resource{}
	for _, r := range p.Resources {
		resources = append(resources, &domain.Resource{
			ID: r.ID,
		})
	}

	return p.Email, resources
}
