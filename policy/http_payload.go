package policy

import "github.com/odpf/guardian/domain"

type createPayload struct {
	ID          string                 `yaml:"id" validate:"required"`
	Description string                 `yaml:"description"`
	Steps       []*domain.Step         `yaml:"steps" validate:"required"`
	Labels      map[string]interface{} `yaml:"labels"`
}

func (p *createPayload) toDomain() *domain.Policy {
	return &domain.Policy{
		ID:          p.ID,
		Description: p.Description,
		Steps:       p.Steps,
		Labels:      p.Labels,
	}
}
