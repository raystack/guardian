package appeal

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mcuadros/go-lookup"
	"github.com/odpf/guardian/domain"
)

// Service handling the business logics
type Service struct {
	repo domain.AppealRepository

	resourceService        domain.ResourceService
	providerService        domain.ProviderService
	policyService          domain.PolicyService
	identityManagerService domain.IdentityManagerService

	validator *validator.Validate
}

// NewService returns service struct
func NewService(
	appealRepository domain.AppealRepository,
	resourceService domain.ResourceService,
	providerService domain.ProviderService,
	policyService domain.PolicyService,
	identityManagerService domain.IdentityManagerService,
) *Service {
	return &Service{
		repo:                   appealRepository,
		resourceService:        resourceService,
		providerService:        providerService,
		policyService:          policyService,
		identityManagerService: identityManagerService,
		validator:              validator.New(),
	}
}

// Create record
func (s *Service) Create(appeals []*domain.Appeal) error {
	resourceIDs := []uint{}
	for _, a := range appeals {
		resourceIDs = append(resourceIDs, a.ResourceID)
	}
	resources, err := s.getResourceMap(resourceIDs)
	if err != nil {
		return err
	}
	policyConfigs, err := s.getPolicyConfigMap()
	if err != nil {
		return err
	}
	approvalSteps, err := s.getApprovalSteps()
	if err != nil {
		return err
	}

	for _, a := range appeals {
		r := resources[a.ResourceID]
		if r == nil {
			return errors.New("resource doesn't exists")
		}

		if policyConfigs[r.ProviderType] == nil {
			return ErrProviderTypeNotFound
		} else if policyConfigs[r.ProviderType][r.ProviderURN] == nil {
			return ErrProviderURNNotFound
		} else if policyConfigs[r.ProviderType][r.ProviderURN][r.Type] == nil {
			return ErrPolicyConfigNotFound
		}
		policyConfig := policyConfigs[r.ProviderType][r.ProviderURN][r.Type]

		if approvalSteps[policyConfig.ID] == nil {
			return ErrPolicyIDNotFound
		} else if approvalSteps[policyConfig.ID][uint(policyConfig.Version)] == nil {
			return ErrPolicyVersionNotFound
		}
		steps := approvalSteps[policyConfig.ID][uint(policyConfig.Version)]

		approvals := []*domain.Approval{}
		for _, step := range steps {
			var approvers []string
			if step.Approvers != "" {
				approvers, err = s.resolveApprovers(a.User, r, step.Approvers)
				if err != nil {
					return err
				}
			}

			approvals = append(approvals, &domain.Approval{
				Name:          step.Name,
				Status:        domain.ApprovalStatusPending,
				PolicyID:      policyConfig.ID,
				PolicyVersion: uint(policyConfig.Version),
				Approvers:     approvers,
			})
		}

		a.PolicyID = policyConfig.ID
		a.PolicyVersion = uint(policyConfig.Version)
		a.Status = domain.AppealStatusPending
		a.Approvals = approvals
	}

	if err := s.repo.BulkInsert(appeals); err != nil {
		return err
	}

	return nil
}

func (s *Service) getResourceMap(ids []uint) (map[uint]*domain.Resource, error) {
	filters := map[string]interface{}{"ids": ids}
	resources, err := s.resourceService.Find(filters)
	if err != nil {
		return nil, err
	}

	result := map[uint]*domain.Resource{}
	for _, r := range resources {
		result[r.ID] = r
	}

	return result, nil
}

func (s *Service) getPolicyConfigMap() (map[string]map[string]map[string]*domain.PolicyConfig, error) {
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

	return policyConfigs, nil
}

func (s *Service) getApprovalSteps() (map[string]map[uint][]*domain.Step, error) {
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

	return approvalSteps, nil
}

func (s *Service) resolveApprovers(user string, resource *domain.Resource, approversKey string) ([]string, error) {
	var approvers []string

	if strings.HasPrefix(approversKey, domain.ApproversKeyResource) {
		mapResource, err := structToMap(resource)
		if err != nil {
			return nil, err
		}

		path := strings.TrimPrefix(approversKey, fmt.Sprintf("%s.", domain.ApproversKeyResource))
		approversReflectValue, err := lookup.LookupString(mapResource, path)
		if err != nil {
			return nil, err
		}

		email, ok := approversReflectValue.Interface().(string)
		if !ok {
			emails, ok := approversReflectValue.Interface().([]interface{})
			if !ok {
				return nil, ErrApproverInvalidType
			}

			for _, e := range emails {
				emailString, ok := e.(string)
				if !ok {
					return nil, ErrApproverInvalidType
				}
				approvers = append(approvers, emailString)
			}
		} else {
			approvers = append(approvers, email)
		}
	} else if strings.HasPrefix(approversKey, domain.ApproversKeyUserApprovers) {
		approverEmails, err := s.identityManagerService.GetUserApproverEmails(user)
		if err != nil {
			return nil, err
		}
		approvers = approverEmails
	} else {
		return nil, ErrApproverKeyNotRecognized
	}

	if err := s.validator.Var(approvers, "dive,email"); err != nil {
		return nil, err
	}
	return approvers, nil
}

func structToMap(item interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	if item != nil {
		jsonString, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(jsonString, &result); err != nil {
			return nil, err
		}
	}

	return result, nil
}
