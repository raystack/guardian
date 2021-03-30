package appeal

import (
	"github.com/odpf/guardian/domain"
)

// Service handling the business logics
type Service struct {
	repo domain.AppealRepository

	resourceService domain.ResourceService
	providerService domain.ProviderService
}

// NewService returns service struct
func NewService(ar domain.AppealRepository, rs domain.ResourceService, ps domain.ProviderService) *Service {
	return &Service{ar, rs, ps}
}

// Create record
func (s *Service) Create(email string, resourceIDs []uint) ([]*domain.Appeal, error) {
	resources, err := s.resourceService.Find(map[string]interface{}{"ids": resourceIDs})
	if err != nil {
		return nil, err
	}

	providers, err := s.providerService.Find()
	if err != nil {
		return nil, err
	}
	policyConfigs := map[string]map[string]map[string]*domain.PolicyConfig{}
	for _, p := range providers {
		providerType := p.Type
		providerURN := p.URN
		if policyConfigs[providerType] == nil {
			policyConfigs[providerType] = map[string]map[string]*domain.PolicyConfig{}
		}
		if policyConfigs[providerType][providerURN] == nil {
			policyConfigs[providerType][providerURN] = map[string]*domain.PolicyConfig{}
		}
		for _, r := range p.Config.Resources {
			resourceType := r.Type
			policyConfigs[providerType][providerURN][resourceType] = r.Policy
		}
	}

	appeals := []*domain.Appeal{}
	for _, r := range resources {
		if policyConfigs[r.ProviderType] == nil {
			return nil, ErrProviderTypeNotFound
		} else if policyConfigs[r.ProviderType][r.ProviderURN] == nil {
			return nil, ErrProviderURNNotFound
		} else if policyConfigs[r.ProviderType][r.ProviderURN][r.Type] == nil {
			return nil, ErrPolicyConfigNotFound
		}
		pc := policyConfigs[r.ProviderType][r.ProviderURN][r.Type]

		appeals = append(appeals, &domain.Appeal{
			ResourceID:    r.ID,
			PolicyID:      pc.ID,
			PolicyVersion: uint(pc.Version),
			Email:         email,
			Status:        domain.AppealStatusPending,
		})
	}

	if err := s.repo.BulkInsert(appeals); err != nil {
		return nil, err
	}

	return appeals, nil
}
