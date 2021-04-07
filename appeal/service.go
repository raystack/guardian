package appeal

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mcuadros/go-lookup"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
)

var TimeNow = time.Now

// Service handling the business logics
type Service struct {
	repo domain.AppealRepository

	approvalService        domain.ApprovalService
	resourceService        domain.ResourceService
	providerService        domain.ProviderService
	policyService          domain.PolicyService
	identityManagerService domain.IdentityManagerService

	validator *validator.Validate
}

// NewService returns service struct
func NewService(
	appealRepository domain.AppealRepository,
	approvalService domain.ApprovalService,
	resourceService domain.ResourceService,
	providerService domain.ProviderService,
	policyService domain.PolicyService,
	identityManagerService domain.IdentityManagerService,
) *Service {
	return &Service{
		repo:                   appealRepository,
		approvalService:        approvalService,
		resourceService:        resourceService,
		providerService:        providerService,
		policyService:          policyService,
		identityManagerService: identityManagerService,
		validator:              validator.New(),
	}
}

// GetByID returns one record by id
func (s *Service) GetByID(id uint) (*domain.Appeal, error) {
	if id == 0 {
		return nil, ErrAppealIDEmptyParam
	}

	return s.repo.GetByID(id)
}

// Find appeals by filters
func (s *Service) Find(filters map[string]interface{}) ([]*domain.Appeal, error) {
	return s.repo.Find(filters)
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
			return ErrResourceNotFound
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
		for i, step := range steps {
			var approvers []string
			if step.Approvers != "" {
				approvers, err = s.resolveApprovers(a.User, r, step.Approvers)
				if err != nil {
					return err
				}
			}

			approvals = append(approvals, &domain.Approval{
				Name:          step.Name,
				Index:         i,
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

// Approve an approval step
func (s *Service) MakeAction(approvalAction domain.ApprovalAction) (*domain.Appeal, error) {
	if err := utils.ValidateStruct(approvalAction); err != nil {
		return nil, err
	}
	appeal, err := s.repo.GetByID(approvalAction.AppealID)
	if err != nil {
		return nil, err
	}
	if appeal == nil {
		return nil, nil
	}

	if appeal.Status != domain.AppealStatusPending {
		switch appeal.Status {
		case domain.AppealStatusActive:
			err = ErrAppealStatusApproved
		case domain.AppealStatusRejected:
			err = ErrAppealStatusRejected
		case domain.AppealStatusTerminated:
			err = ErrAppealStatusTerminated
		default:
			err = ErrAppealStatusUnrecognized
		}
		return nil, err
	}

	for i, approval := range appeal.Approvals {
		if approval.Name != approvalAction.ApprovalName {
			switch approval.Status {
			case domain.ApprovalStatusApproved:
			case domain.ApprovalStatusSkipped:
				continue
			case domain.ApprovalStatusPending:
				return nil, ErrApprovalDependencyIsPending
			case domain.ApprovalStatusRejected:
				return nil, ErrAppealStatusRejected
			default:
				return nil, ErrApprovalStatusUnrecognized
			}
		} else {
			if approval.Status != domain.ApprovalStatusPending {
				switch approval.Status {
				case domain.ApprovalStatusApproved:
					err = ErrApprovalStatusApproved
				case domain.ApprovalStatusRejected:
					err = ErrApprovalStatusRejected
				case domain.ApprovalStatusSkipped:
					err = ErrApprovalStatusSkipped
				default:
					err = ErrApprovalStatusUnrecognized
				}
				return nil, err
			}

			if !utils.ContainsString(approval.Approvers, approvalAction.Actor) {
				return nil, ErrActionForbidden
			}

			approval.Actor = &approvalAction.Actor
			approval.UpdatedAt = TimeNow()
			if approvalAction.Action == domain.AppealActionNameApprove {
				approval.Status = domain.ApprovalStatusApproved
				if i == len(appeal.Approvals)-1 {
					// TODO: grant access to the actual provider
					appeal.Status = domain.AppealStatusActive
				}
			} else if approvalAction.Action == domain.AppealActionNameReject {
				approval.Status = domain.ApprovalStatusRejected
				appeal.Status = domain.AppealStatusRejected
				if i < len(appeal.Approvals)-1 {
					for j := i + 1; j < len(appeal.Approvals); j++ {
						appeal.Approvals[j].Status = domain.ApprovalStatusSkipped
						appeal.Approvals[j].UpdatedAt = TimeNow()
					}
				}
			} else {
				return nil, ErrActionInvalidValue
			}

			if err := s.repo.Update(appeal); err != nil {
				// TODO: rollback granted access in the actual provider
				return nil, err
			}

			return appeal, nil
		}
	}

	return nil, ErrApprovalNameNotFound
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

func (s *Service) GetPendingApprovals(user string) ([]*domain.Approval, error) {
	return s.approvalService.GetPendingApprovals(user)
}
