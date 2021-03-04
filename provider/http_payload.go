package provider

import "github.com/odpf/guardian/domain"

type updatePayload struct {
	Labels      map[string]interface{}   `yaml:"labels"`
	Credentials interface{}              `yaml:"credentials"`
	Appeal      *domain.AppealConfig     `yaml:"appeal"`
	Resources   []*domain.ResourceConfig `yaml:"resources"`
}

func (p *updatePayload) toDomain() *domain.Provider {
	return &domain.Provider{
		Config: &domain.ProviderConfig{
			Labels:      p.Labels,
			Credentials: p.Credentials,
			Appeal:      p.Appeal,
			Resources:   p.Resources,
		},
	}
}
