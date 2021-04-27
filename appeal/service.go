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

type resourceConfig struct {
	policy           *domain.PolicyConfig
	availableRoleIDs []string
}
type providerConfig struct {
	appeal    *domain.AppealConfig
	resources map[string]*resourceConfig
}

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
	providerConfigs, err := s.getProviderConfigs()
	if err != nil {
		return err
	}
	approvalSteps, err := s.getApprovalSteps()
	if err != nil {
		return err
	}

	for _, a := range appeals {
		resource := resources[a.ResourceID]
		if resource == nil {
			return ErrResourceNotFound
		}

		if providerConfigs[resource.ProviderType] == nil {
			return ErrProviderTypeNotFound
		} else if providerConfigs[resource.ProviderType][resource.ProviderURN] == nil {
			return ErrProviderURNNotFound
		}
		providerConfig := providerConfigs[resource.ProviderType][resource.ProviderURN]

		if providerConfig.resources[resource.Type] == nil {
			return ErrResourceTypeNotFound
		}

		appealConfig := providerConfig.appeal
		if !appealConfig.AllowPermanentAccess {
			if a.Options == nil || a.Options.ExpirationDate == nil {
				return ErrOptionsExpirationDateOptionNotFound
			} else if a.Options.ExpirationDate.IsZero() {
				return ErrExpirationDateIsRequired
			}
		}

		resourceConfig := providerConfig.resources[resource.Type]
		if !utils.ContainsString(resourceConfig.availableRoleIDs, a.Role) {
			return ErrInvalidRole
		}

		policyConfig := resourceConfig.policy
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
				approvers, err = s.resolveApprovers(a.User, resource, step.Approvers)
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

	if err := checkIfAppealStatusStillPending(appeal.Status); err != nil {
		return nil, err
	}

	for i, approval := range appeal.Approvals {
		if approval.Name != approvalAction.ApprovalName {
			if err := checkPreviousApprovalStatus(approval.Status); err != nil {
				return nil, err
			}
			continue
		} else {
			if approval.Status != domain.ApprovalStatusPending {
				if err := checkApprovalStatus(approval.Status); err != nil {
					return nil, err
				}
			}

			if !utils.ContainsString(approval.Approvers, approvalAction.Actor) {
				return nil, ErrActionForbidden
			}

			approval.Actor = &approvalAction.Actor
			approval.UpdatedAt = TimeNow()

			if approvalAction.Action == domain.AppealActionNameApprove {
				approval.Status = domain.ApprovalStatusApproved

				if i == len(appeal.Approvals)-1 {
					if err := s.providerService.GrantAccess(appeal); err != nil {
						return nil, err
					}
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
				if err := s.providerService.RevokeAccess(appeal); err != nil {
					return nil, err
				}
				return nil, err
			}

			return appeal, nil
		}
	}

	return nil, ErrApprovalNameNotFound
}

func (s *Service) GetPendingApprovals(user string) ([]*domain.Approval, error) {
	return s.approvalService.GetPendingApprovals(user)
}

func (s *Service) Cancel(id uint) (*domain.Appeal, error) {
	appeal, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if appeal == nil {
		return nil, nil
	}

	// TODO: check only appeal creator who is allowed to cancel the appeal

	if err := checkIfAppealStatusStillPending(appeal.Status); err != nil {
		return nil, err
	}

	appeal.Status = domain.AppealStatusCanceled
	if err := s.repo.Update(appeal); err != nil {
		return nil, err
	}

	return appeal, nil
}
func (s *Service) Revoke(id uint, actor string) (*domain.Appeal, error) {
	appeal, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if appeal == nil {
		return nil, ErrAppealNotFound
	}

	if actor != domain.SystemActorName {
		lastApprovalStep := appeal.Approvals[len(appeal.Approvals)-1]
		if !utils.ContainsString(lastApprovalStep.Approvers, actor) {
			return nil, ErrRevokeAppealForbidden
		}
	}

	revokedAppeal := &domain.Appeal{}
	*revokedAppeal = *appeal
	revokedAppeal.Status = domain.AppealStatusTerminated

	if err := s.repo.Update(revokedAppeal); err != nil {
		return nil, err
	}

	if err := s.providerService.RevokeAccess(appeal); err != nil {
		if err := s.repo.Update(appeal); err != nil {
			return nil, err
		}
		return nil, err
	}

	return revokedAppeal, nil
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

func (s *Service) getProviderConfigs() (map[string]map[string]*providerConfig, error) {
	providers, err := s.providerService.Find()
	if err != nil {
		return nil, err
	}

	providerConfigs := map[string]map[string]*providerConfig{}
	for _, p := range providers {
		providerType := p.Type
		providerURN := p.URN
		if providerConfigs[providerType] == nil {
			providerConfigs[providerType] = map[string]*providerConfig{}
		}
		if providerConfigs[providerType][providerURN] == nil {
			providerConfigs[providerType][providerURN] = &providerConfig{
				appeal:    p.Config.Appeal,
				resources: map[string]*resourceConfig{},
			}
		}
		for _, r := range p.Config.Resources {
			resourceType := r.Type

			availableRoleIDs := []string{}
			for _, role := range r.Roles {
				availableRoleIDs = append(availableRoleIDs, role.ID)
			}
			providerConfigs[providerType][providerURN].resources[resourceType] = &resourceConfig{
				policy:           r.Policy,
				availableRoleIDs: availableRoleIDs,
			}
		}
	}

	return providerConfigs, nil
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

func checkIfAppealStatusStillPending(status string) error {
	if status == domain.AppealStatusPending {
		return nil
	}

	var err error
	switch status {
	case domain.AppealStatusCanceled:
		err = ErrAppealStatusCanceled
	case domain.AppealStatusActive:
		err = ErrAppealStatusApproved
	case domain.AppealStatusRejected:
		err = ErrAppealStatusRejected
	case domain.AppealStatusTerminated:
		err = ErrAppealStatusTerminated
	default:
		err = ErrAppealStatusUnrecognized
	}
	return err
}

func checkPreviousApprovalStatus(status string) error {
	var err error
	switch status {
	case domain.ApprovalStatusApproved,
		domain.ApprovalStatusSkipped:
		err = nil
	case domain.ApprovalStatusPending:
		err = ErrApprovalDependencyIsPending
	case domain.ApprovalStatusRejected:
		err = ErrAppealStatusRejected
	default:
		err = ErrApprovalStatusUnrecognized
	}
	return err
}

func checkApprovalStatus(status string) error {
	var err error
	switch status {
	case domain.ApprovalStatusApproved:
		err = ErrApprovalStatusApproved
	case domain.ApprovalStatusRejected:
		err = ErrApprovalStatusRejected
	case domain.ApprovalStatusSkipped:
		err = ErrApprovalStatusSkipped
	default:
		err = ErrApprovalStatusUnrecognized
	}
	return err
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
