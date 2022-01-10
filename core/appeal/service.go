package appeal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/evaluator"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/guardian/store"
	"github.com/odpf/guardian/utils"
	"github.com/odpf/salt/log"
)

var TimeNow = time.Now

// Service handling the business logics
type Service struct {
	repo store.AppealRepository

	approvalService domain.ApprovalService
	resourceService domain.ResourceService
	providerService domain.ProviderService
	policyService   domain.PolicyService
	iam             domain.IAMManager
	notifier        notifiers.Client
	logger          log.Logger

	validator *validator.Validate
	TimeNow   func() time.Time
}

// NewService returns service struct
func NewService(
	appealRepository store.AppealRepository,
	approvalService domain.ApprovalService,
	resourceService domain.ResourceService,
	providerService domain.ProviderService,
	policyService domain.PolicyService,
	iam domain.IAMManager,
	notifier notifiers.Client,
	logger log.Logger,
) *Service {
	return &Service{
		repo:            appealRepository,
		approvalService: approvalService,
		resourceService: resourceService,
		providerService: providerService,
		policyService:   policyService,
		iam:             iam,
		notifier:        notifier,
		validator:       validator.New(),
		logger:          logger,
		TimeNow:         time.Now,
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
	resources, err := s.getResourcesMap(resourceIDs)
	if err != nil {
		return err
	}
	providers, err := s.getProvidersMap()
	if err != nil {
		return err
	}
	policies, err := s.getPoliciesMap()
	if err != nil {
		return err
	}
	pendingAppeals, activeAppeals, err := s.getExistingAppealsMap()
	if err != nil {
		return err
	}

	notifications := []domain.Notification{}
	expiredAppeals := []*domain.Appeal{}

	for _, appeal := range appeals {
		appeal.SetDefaults()

		if err := validateAppeal(appeal, pendingAppeals); err != nil {
			return err
		}
		if err := addResource(appeal, resources); err != nil {
			return err
		}
		provider, err := getProvider(appeal, providers)
		if err != nil {
			return err
		}

		expiredAppeal, err := s.checkAppealExtension(appeal, provider, activeAppeals)
		if err != nil {
			return err
		}
		if expiredAppeal != nil {
			expiredAppeals = append(expiredAppeals, expiredAppeal)
		}

		if err := s.providerService.ValidateAppeal(appeal, provider); err != nil {
			return err
		}

		policy, err := getPolicy(appeal, provider, policies)
		if err != nil {
			return err
		}

		if err := s.addCreatorDetails(appeal, policy); err != nil {
			return err
		}

		if err := s.fillApprovals(appeal, policy); err != nil {
			return err
		}

		appeal.PolicyID = policy.ID
		appeal.PolicyVersion = policy.Version
		appeal.Status = domain.AppealStatusPending

		appeal.Policy = policy
		if err := s.approvalService.AdvanceApproval(appeal); err != nil {
			return err
		}
		appeal.Policy = nil

		for _, approval := range appeal.Approvals {
			if approval.Index == len(appeal.Approvals)-1 && approval.Status == domain.ApprovalStatusApproved {
				if err := s.createAccess(appeal); err != nil {
					return err
				}
				notifications = append(notifications, domain.Notification{
					User: appeal.CreatedBy,
					Message: domain.NotificationMessage{
						Type: domain.NotificationTypeAppealApproved,
						Variables: map[string]interface{}{
							"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
							"role":          appeal.Role,
						},
					},
				})
			}
		}
	}

	allAppeals := append(appeals, expiredAppeals...)
	if err := s.repo.BulkUpsert(allAppeals); err != nil {
		return err
	}

	for _, a := range appeals {
		notifications = append(notifications, getApprovalNotifications(a)...)
	}

	if len(notifications) > 0 {
		if err := s.notifier.Notify(notifications); err != nil {
			s.logger.Error(err.Error())
		}
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
			approval.Reason = approvalAction.Reason
			approval.UpdatedAt = TimeNow()

			if approvalAction.Action == domain.AppealActionNameApprove {
				approval.Status = domain.ApprovalStatusApproved
				if i+1 <= len(appeal.Approvals)-1 {
					appeal.Approvals[i+1].Status = domain.ApprovalStatusPending
				}
				if err := s.approvalService.AdvanceApproval(appeal); err != nil {
					return nil, err
				}

				if i == len(appeal.Approvals)-1 {
					if err := s.createAccess(appeal); err != nil {
						return nil, err
					}
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

			notifications := []domain.Notification{}
			if appeal.Status == domain.AppealStatusActive {
				notifications = append(notifications, domain.Notification{
					User: appeal.CreatedBy,
					Message: domain.NotificationMessage{
						Type: domain.NotificationTypeAppealApproved,
						Variables: map[string]interface{}{
							"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
							"role":          appeal.Role,
						},
					},
				})
			} else if appeal.Status == domain.AppealStatusRejected {
				notifications = append(notifications, domain.Notification{
					User: appeal.CreatedBy,
					Message: domain.NotificationMessage{
						Type: domain.NotificationTypeAppealRejected,
						Variables: map[string]interface{}{
							"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
							"role":          appeal.Role,
						},
					},
				})
			} else {
				notifications = append(notifications, getApprovalNotifications(appeal)...)
			}
			if len(notifications) > 0 {
				if err := s.notifier.Notify(notifications); err != nil {
					s.logger.Error(err.Error())
				}
			}

			return appeal, nil
		}
	}

	return nil, ErrApprovalNameNotFound
}

func (s *Service) Cancel(id uint) (*domain.Appeal, error) {
	appeal, err := s.GetByID(id)
	if err != nil {
		return nil, err
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

func (s *Service) Revoke(id uint, actor, reason string) (*domain.Appeal, error) {
	appeal, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	revokedAppeal := &domain.Appeal{}
	*revokedAppeal = *appeal
	revokedAppeal.Status = domain.AppealStatusTerminated
	revokedAppeal.RevokedAt = s.TimeNow()
	revokedAppeal.RevokedBy = actor
	revokedAppeal.RevokeReason = reason

	if err := s.repo.Update(revokedAppeal); err != nil {
		return nil, err
	}

	if err := s.providerService.RevokeAccess(appeal); err != nil {
		if err := s.repo.Update(appeal); err != nil {
			return nil, err
		}
		return nil, err
	}

	if err := s.notifier.Notify([]domain.Notification{{
		User: appeal.CreatedBy,
		Message: domain.NotificationMessage{
			Type: domain.NotificationTypeAccessRevoked,
			Variables: map[string]interface{}{
				"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
				"role":          appeal.Role,
			},
		},
	}}); err != nil {
		s.logger.Error(err.Error())
	}

	return revokedAppeal, nil
}

func (s *Service) getExistingAppealsMap() (map[string]map[uint]map[string]*domain.Appeal, map[string]map[uint]map[string]*domain.Appeal, error) {
	appeals, err := s.repo.Find(map[string]interface{}{
		"statuses": []string{domain.AppealStatusPending, domain.AppealStatusActive},
	})
	if err != nil {
		return nil, nil, err
	}

	pendingAppealsMap := map[string]map[uint]map[string]*domain.Appeal{}
	activeAppealsMap := map[string]map[uint]map[string]*domain.Appeal{}
	for _, a := range appeals {
		if a.Status == domain.AppealStatusPending {
			if pendingAppealsMap[a.AccountID] == nil {
				pendingAppealsMap[a.AccountID] = map[uint]map[string]*domain.Appeal{}
			}
			if pendingAppealsMap[a.AccountID][a.ResourceID] == nil {
				pendingAppealsMap[a.AccountID][a.ResourceID] = map[string]*domain.Appeal{}
			}
			pendingAppealsMap[a.AccountID][a.ResourceID][a.Role] = a
		} else if a.Status == domain.AppealStatusActive {
			if activeAppealsMap[a.AccountID] == nil {
				activeAppealsMap[a.AccountID] = map[uint]map[string]*domain.Appeal{}
			}
			if activeAppealsMap[a.AccountID][a.ResourceID] == nil {
				activeAppealsMap[a.AccountID][a.ResourceID] = map[string]*domain.Appeal{}
			}
			activeAppealsMap[a.AccountID][a.ResourceID][a.Role] = a
		}
	}

	return pendingAppealsMap, activeAppealsMap, nil
}

func (s *Service) getResourcesMap(ids []uint) (map[uint]*domain.Resource, error) {
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

func (s *Service) getProvidersMap() (map[string]map[string]*domain.Provider, error) {
	providers, err := s.providerService.Find()
	if err != nil {
		return nil, err
	}

	providersMap := map[string]map[string]*domain.Provider{}
	for _, p := range providers {
		providerType := p.Type
		providerURN := p.URN
		if providersMap[providerType] == nil {
			providersMap[providerType] = map[string]*domain.Provider{}
		}
		if providersMap[providerType][providerURN] == nil {
			providersMap[providerType][providerURN] = p
		}
	}

	return providersMap, nil
}

func (s *Service) getPoliciesMap() (map[string]map[uint]*domain.Policy, error) {
	policies, err := s.policyService.Find()
	if err != nil {
		return nil, err
	}
	policiesMap := map[string]map[uint]*domain.Policy{}
	for _, p := range policies {
		id := p.ID
		version := p.Version
		if policiesMap[id] == nil {
			policiesMap[id] = map[uint]*domain.Policy{}
		}
		policiesMap[id][version] = p
	}

	return policiesMap, nil
}

func (s *Service) resolveApprovers(expressions []string, appeal *domain.Appeal) ([]string, error) {
	var approvers []string

	// TODO: validate from policyService.Validate(policy)
	for _, expr := range expressions {
		if err := s.validator.Var(expr, "email"); err == nil {
			approvers = append(approvers, expr)
		} else {
			appealMap, err := structToMap(appeal)
			if err != nil {
				return nil, fmt.Errorf("parsing appeal to map: %w", err)
			}
			params := map[string]interface{}{
				"appeal": appealMap,
			}

			approversValue, err := evaluator.Expression(expr).EvaluateWithVars(params)
			if err != nil {
				return nil, fmt.Errorf("evaluating aprrovers expression: %w", err)
			}

			value := reflect.ValueOf(approversValue)
			switch value.Type().Kind() {
			case reflect.String:
				approvers = append(approvers, value.String())
			case reflect.Slice:
				for i := 0; i < value.Len(); i++ {
					itemValue := reflect.ValueOf(value.Index(i).Interface())
					switch itemValue.Type().Kind() {
					case reflect.String:
						approvers = append(approvers, itemValue.String())
					default:
						return nil, fmt.Errorf(`%w: %s`, ErrApproverInvalidType, itemValue.Type().Kind())
					}
				}
			default:
				return nil, fmt.Errorf(`%w: %s`, ErrApproverInvalidType, value.Type().Kind())
			}
		}
	}

	if err := s.validator.Var(approvers, "dive,email"); err != nil {
		return nil, err
	}
	return approvers, nil
}

func getApprovalNotifications(appeal *domain.Appeal) []domain.Notification {
	notifications := []domain.Notification{}
	approval := appeal.GetNextPendingApproval()
	if approval != nil {
		for _, approver := range approval.Approvers {
			notifications = append(notifications, domain.Notification{
				User: approver,
				Message: domain.NotificationMessage{
					Type: domain.NotificationTypeApproverNotification,
					Variables: map[string]interface{}{
						"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
						"role":          appeal.Role,
						"requestor":     appeal.CreatedBy,
						"appeal_id":     appeal.ID,
					},
				},
			})
		}
	}
	return notifications
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
	case domain.ApprovalStatusBlocked:
		err = ErrApprovalDependencyIsBlocked
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
	case domain.ApprovalStatusBlocked:
		err = ErrAppealStatusBlocked
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

func (s *Service) fillApprovals(a *domain.Appeal, p *domain.Policy) error {
	approvals := []*domain.Approval{}
	for i, step := range p.Steps { // TODO: move this logic to approvalService
		var approverEmails []string
		var err error
		if step.Strategy == domain.ApprovalStepStrategyManual {
			approverEmails, err = s.resolveApprovers(step.Approvers, a)
			if err != nil {
				return fmt.Errorf("resolving approvers `%s`: %w", step.Approvers, err)
			}
		}

		status := domain.ApprovalStatusPending
		if i > 0 {
			status = domain.ApprovalStatusBlocked
		}

		approvals = append(approvals, &domain.Approval{
			Name:          step.Name,
			Index:         i,
			Status:        status,
			PolicyID:      p.ID,
			PolicyVersion: p.Version,
			Approvers:     approverEmails,
		})
	}

	a.Approvals = approvals
	return nil
}

func (s *Service) handleAppealRequirements(a *domain.Appeal, p *domain.Policy) error {
	additionalAppeals := []*domain.Appeal{}
	if p.Requirements != nil && len(p.Requirements) > 0 {
		for reqIndex, r := range p.Requirements {
			isAppealMatchesRequirement, err := r.On.IsMatch(a)
			if err != nil {
				return fmt.Errorf("evaluating requirements[%v]: %v", reqIndex, err)
			}
			if !isAppealMatchesRequirement {
				continue
			}

			for _, aa := range r.Appeals {
				// TODO: populate resource data from policyService
				resource, err := s.resourceService.Get(aa.Resource)
				if err != nil {
					return fmt.Errorf("retrieving resource: %v", err)
				}

				additionalAppeal := &domain.Appeal{
					AccountID:   a.AccountID,
					AccountType: a.AccountType,
					CreatedBy:   a.CreatedBy,
					Role:        aa.Role,
					ResourceID:  resource.ID,
				}
				if aa.Options != nil {
					additionalAppeal.Options = aa.Options
				}
				if aa.Policy != nil {
					additionalAppeal.PolicyID = aa.Policy.ID
					additionalAppeal.PolicyVersion = uint(aa.Policy.Version)
				}
				additionalAppeals = append(additionalAppeals, additionalAppeal)
			}
		}
	}
	if len(additionalAppeals) > 0 {
		if err := s.Create(additionalAppeals); err != nil {
			return fmt.Errorf("creating additional appeals: %v", err)
		}
	}
	return nil
}

func (s *Service) createAccess(a *domain.Appeal) error {
	policy := a.Policy
	if policy == nil {
		p, err := s.policyService.GetOne(a.PolicyID, a.PolicyVersion)
		if err != nil {
			return fmt.Errorf("retrieving policy: %v", err)
		}
		policy = p
	}

	if err := s.handleAppealRequirements(a, policy); err != nil {
		return err
	}

	if err := s.providerService.GrantAccess(a); err != nil {
		return err
	}

	if err := a.Activate(); err != nil {
		return fmt.Errorf("activating appeal: %v", err)
	}

	return nil
}

func (s *Service) checkAppealExtension(a *domain.Appeal, p *domain.Provider, activeAppealsMap map[string]map[uint]map[string]*domain.Appeal) (*domain.Appeal, error) {
	if activeAppealsMap[a.AccountID] != nil &&
		activeAppealsMap[a.AccountID][a.ResourceID] != nil &&
		activeAppealsMap[a.AccountID][a.ResourceID][a.Role] != nil {
		if p.Config.Appeal != nil {
			if p.Config.Appeal.AllowActiveAccessExtensionIn == "" {
				return nil, ErrAppealFoundActiveAccess
			}

			duration, err := time.ParseDuration(p.Config.Appeal.AllowActiveAccessExtensionIn)
			if err != nil {
				return nil, fmt.Errorf("%v: %v: %v", ErrAppealInvalidExtensionDuration, p.Config.Appeal.AllowActiveAccessExtensionIn, err)
			}

			now := s.TimeNow()
			activeAppealExpDate := activeAppealsMap[a.AccountID][a.ResourceID][a.Role].Options.ExpirationDate
			isEligibleForExtension := activeAppealExpDate.Sub(now) <= duration
			if isEligibleForExtension {
				oldAppeal := &domain.Appeal{}
				*oldAppeal = *activeAppealsMap[a.AccountID][a.ResourceID][a.Role]
				oldAppeal.Terminate()
				return oldAppeal, nil
			} else {
				return nil, fmt.Errorf("%v: the extension policy for this resource is %v before current access expiration", ErrAppealNotEligibleForExtension, duration)
			}
		}
	}
	return nil, nil
}

func getPolicy(a *domain.Appeal, p *domain.Provider, policiesMap map[string]map[uint]*domain.Policy) (*domain.Policy, error) {
	var resourceConfig *domain.ResourceConfig
	for _, rc := range p.Config.Resources {
		if rc.Type == a.Resource.Type {
			resourceConfig = rc
			break
		}
	}
	if resourceConfig == nil {
		return nil, ErrResourceTypeNotFound
	}

	policyConfig := resourceConfig.Policy
	if policiesMap[policyConfig.ID] == nil {
		return nil, ErrPolicyIDNotFound
	} else if policiesMap[policyConfig.ID][uint(policyConfig.Version)] == nil {
		return nil, ErrPolicyVersionNotFound
	}

	return policiesMap[policyConfig.ID][uint(policyConfig.Version)], nil
}

func (s *Service) addCreatorDetails(a *domain.Appeal, p *domain.Policy) error {
	if p.IAM != nil {
		iamConfig, err := s.iam.ParseConfig(p.IAM)
		if err != nil {
			return fmt.Errorf("parsing iam config: %w", err)
		}
		iamClient, err := s.iam.GetClient(iamConfig)
		if err != nil {
			return fmt.Errorf("getting iam client: %w", err)
		}

		userProfile, err := iamClient.GetUser(a.CreatedBy)
		if err != nil {
			return fmt.Errorf("fetching creator's user iam: %w", err)
		}

		var creator map[string]interface{}
		if userProfileMap, ok := userProfile.(map[string]interface{}); ok {
			if p.IAM.ProfileSchema != nil {
				creator = map[string]interface{}{}
				for profileKey, targetKey := range p.IAM.ProfileSchema {
					creator[profileKey] = userProfileMap[targetKey]
				}
			} else {
				creator = userProfileMap
			}
		}

		a.Creator = creator
	}

	return nil
}

func addResource(a *domain.Appeal, resourcesMap map[uint]*domain.Resource) error {
	r := resourcesMap[a.ResourceID]
	if r == nil {
		return ErrResourceNotFound
	} else if r.IsDeleted {
		return ErrResourceIsDeleted
	}

	a.Resource = r
	return nil
}

func getProvider(a *domain.Appeal, providersMap map[string]map[string]*domain.Provider) (*domain.Provider, error) {
	if providersMap[a.Resource.ProviderType] == nil {
		return nil, ErrProviderTypeNotFound
	} else if providersMap[a.Resource.ProviderType][a.Resource.ProviderURN] == nil {
		return nil, ErrProviderURNNotFound
	}

	return providersMap[a.Resource.ProviderType][a.Resource.ProviderURN], nil
}

func validateAppeal(a *domain.Appeal, pendingAppealsMap map[string]map[uint]map[string]*domain.Appeal) error {
	if a.AccountType == domain.DefaultAppealAccountType && a.AccountID != a.CreatedBy {
		return ErrCannotCreateAppealForOtherUser
	}

	if pendingAppealsMap[a.AccountID] != nil &&
		pendingAppealsMap[a.AccountID][a.ResourceID] != nil &&
		pendingAppealsMap[a.AccountID][a.ResourceID][a.Role] != nil {
		return ErrAppealDuplicate
	}

	return nil
}
