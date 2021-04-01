package appeal

import (
	"github.com/odpf/guardian/domain"
)

// Service handling the business logics
type Service struct {
	repo domain.AppealRepository

	resourceService domain.ResourceService
	providerService domain.ProviderService
	policyService   domain.PolicyService
}

// NewService returns service struct
func NewService(ar domain.AppealRepository, rs domain.ResourceService, ps domain.ProviderService, policyService domain.PolicyService) *Service {
	return &Service{ar, rs, ps, policyService}
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

	policies, err := s.policyService.Find()
	if err != nil {
		return nil, err
	}
	approvalSteps := map[string]map[uint][]*domain.Step{}
	for _, p := range policies {
		id := p.ID
		version := p.Version
		if approvalSteps[id] == nil {
			approvalSteps[id] = map[uint][]*domain.Step{}
		}
		approvalSteps[id][version] = p.Steps
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
		policyConfig := policyConfigs[r.ProviderType][r.ProviderURN][r.Type]

		if approvalSteps[policyConfig.ID] == nil {
			return nil, ErrPolicyIDNotFound
		} else if approvalSteps[policyConfig.ID][uint(policyConfig.Version)] == nil {
			return nil, ErrPolicyVersionNotFound
		}
		steps := approvalSteps[policyConfig.ID][uint(policyConfig.Version)]

		approvals := []*domain.Approval{}
		for _, s := range steps {
			approvals = append(approvals, &domain.Approval{
				Name:          s.Name,
				Status:        domain.ApprovalStatusPending,
				PolicyID:      policyConfig.ID,
				PolicyVersion: uint(policyConfig.Version),
				// TODO: retrieve approvers based on the approval flow config
			})
		}

		appeals = append(appeals, &domain.Appeal{
			ResourceID:    r.ID,
			PolicyID:      policyConfig.ID,
			PolicyVersion: uint(policyConfig.Version),
			Email:         email,
			Status:        domain.AppealStatusPending,
			Approvals:     approvals,
		})
	}

	if err := s.repo.BulkInsert(appeals); err != nil {
		return nil, err
	}

	return appeals, nil
}
