package appeal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/core/access"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/pkg/evaluator"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/guardian/utils"
	"github.com/odpf/salt/log"
)

const (
	AuditKeyBulkInsert     = "appeal.bulkInsert"
	AuditKeyCancel         = "appeal.cancel"
	AuditKeyApprove        = "appeal.approve"
	AuditKeyReject         = "appeal.reject"
	AuditKeyRevoke         = "appeal.revoke"
	AuditKeyExtend         = "appeal.extend"
	AuditKeyAddApprover    = "appeal.addApprover"
	AuditKeyDeleteApprover = "appeal.deleteApprover"

	RevokeReasonForExtension = "Automatically revoked for access extension"
)

var TimeNow = time.Now

//go:generate mockery --name=repository --exported --with-expecter
type repository interface {
	BulkUpsert([]*domain.Appeal) error
	Find(*domain.ListAppealsFilter) ([]*domain.Appeal, error)
	GetByID(id string) (*domain.Appeal, error)
	Update(*domain.Appeal) error
}

//go:generate mockery --name=iamManager --exported --with-expecter
type iamManager interface {
	domain.IAMManager
}

//go:generate mockery --name=notifier --exported --with-expecter
type notifier interface {
	notifiers.Client
}

//go:generate mockery --name=policyService --exported --with-expecter
type policyService interface {
	Find(context.Context) ([]*domain.Policy, error)
	GetOne(context.Context, string, uint) (*domain.Policy, error)
}

//go:generate mockery --name=approvalService --exported --with-expecter
type approvalService interface {
	AdvanceApproval(context.Context, *domain.Appeal) error
	AddApprover(ctx context.Context, approvalID, email string) error
	DeleteApprover(ctx context.Context, approvalID, email string) error
}

//go:generate mockery --name=providerService --exported --with-expecter
type providerService interface {
	Find(context.Context) ([]*domain.Provider, error)
	GrantAccess(context.Context, domain.Access) error
	RevokeAccess(context.Context, domain.Access) error
	ValidateAppeal(context.Context, *domain.Appeal, *domain.Provider) error
	GetPermissions(context.Context, *domain.ProviderConfig, string, string) ([]interface{}, error)
}

//go:generate mockery --name=resourceService --exported --with-expecter
type resourceService interface {
	Find(context.Context, map[string]interface{}) ([]*domain.Resource, error)
	Get(context.Context, *domain.ResourceIdentifier) (*domain.Resource, error)
}

//go:generate mockery --name=accessService --exported --with-expecter
type accessService interface {
	List(context.Context, domain.ListAccessesFilter) ([]domain.Access, error)
	Prepare(context.Context, domain.Appeal) (*domain.Access, error)
	Revoke(ctx context.Context, id, actor, reason string, opts ...access.Option) (*domain.Access, error)
}

//go:generate mockery --name=auditLogger --exported --with-expecter
type auditLogger interface {
	Log(ctx context.Context, action string, data interface{}) error
}

type CreateAppealOption func(*createAppealOptions)

type createAppealOptions struct {
	IsAdditionalAppeal bool
}

func CreateWithAdditionalAppeal() CreateAppealOption {
	return func(opts *createAppealOptions) {
		opts.IsAdditionalAppeal = true
	}
}

type ServiceDeps struct {
	Repository      repository
	ApprovalService approvalService
	ResourceService resourceService
	ProviderService providerService
	PolicyService   policyService
	AccessService   accessService
	IAMManager      iamManager

	Notifier    notifier
	Validator   *validator.Validate
	Logger      log.Logger
	AuditLogger auditLogger
}

// Service handling the business logics
type Service struct {
	repo            repository
	approvalService approvalService
	resourceService resourceService
	providerService providerService
	policyService   policyService
	accessService   accessService
	iam             domain.IAMManager

	notifier    notifier
	validator   *validator.Validate
	logger      log.Logger
	auditLogger auditLogger

	TimeNow func() time.Time
}

// NewService returns service struct
func NewService(deps ServiceDeps) *Service {
	return &Service{
		deps.Repository,
		deps.ApprovalService,
		deps.ResourceService,
		deps.ProviderService,
		deps.PolicyService,
		deps.AccessService,
		deps.IAMManager,

		deps.Notifier,
		deps.Validator,
		deps.Logger,
		deps.AuditLogger,
		time.Now,
	}
}

// GetByID returns one record by id
func (s *Service) GetByID(ctx context.Context, id string) (*domain.Appeal, error) {
	if id == "" {
		return nil, ErrAppealIDEmptyParam
	}

	return s.repo.GetByID(id)
}

// Find appeals by filters
func (s *Service) Find(ctx context.Context, filters *domain.ListAppealsFilter) ([]*domain.Appeal, error) {
	for i, v := range filters.Statuses {
		if v == domain.AppealStatusApproved {
			filters.Statuses[i] = domain.AppealStatusActive
		}
	}
	return s.repo.Find(filters)
}

// Create record
func (s *Service) Create(ctx context.Context, appeals []*domain.Appeal, opts ...CreateAppealOption) error {
	createAppealOpts := &createAppealOptions{}
	for _, opt := range opts {
		opt(createAppealOpts)
	}
	isAdditionalAppealCreation := createAppealOpts.IsAdditionalAppeal

	resourceIDs := []string{}
	for _, a := range appeals {
		resourceIDs = append(resourceIDs, a.ResourceID)
	}
	resources, err := s.getResourcesMap(ctx, resourceIDs)
	if err != nil {
		return err
	}
	providers, err := s.getProvidersMap(ctx)
	if err != nil {
		return err
	}
	policies, err := s.getPoliciesMap(ctx)
	if err != nil {
		return err
	}

	pendingAppeals, err := s.getPendingAppealsMap()
	if err != nil {
		return fmt.Errorf("listing pending appeals: %w", err)
	}
	activeAccesses, err := s.getActiveAccessesMap(ctx)
	if err != nil {
		return fmt.Errorf("listing active accesses: %w", err)
	}

	notifications := []domain.Notification{}

	for _, appeal := range appeals {
		appeal.SetDefaults()

		if err := validateAppeal(appeal, pendingAppeals); err != nil {
			return err
		}
		if err := addResource(appeal, resources); err != nil {
			return fmt.Errorf("retrieving resource details for %s: %w", appeal.ResourceID, err)
		}
		provider, err := getProvider(appeal, providers)
		if err != nil {
			return fmt.Errorf("retrieving provider: %w", err)
		}

		if err := s.checkExtensionEligibility(appeal, provider, activeAccesses); err != nil {
			return fmt.Errorf("checking access extension eligibility: %w", err)
		}

		if err := s.providerService.ValidateAppeal(ctx, appeal, provider); err != nil {
			return fmt.Errorf("validating appeal based on provider: %w", err)
		}

		strPermissions, err := s.getPermissions(ctx, provider.Config, appeal.Resource.Type, appeal.Role)
		if err != nil {
			return fmt.Errorf("getting permissions list: %w", err)
		}
		appeal.Permissions = strPermissions

		var policy *domain.Policy
		if isAdditionalAppealCreation && appeal.PolicyID != "" && appeal.PolicyVersion != 0 {
			policy = policies[appeal.PolicyID][appeal.PolicyVersion]
		} else {
			var err error
			policy, err = getPolicy(appeal, provider, policies)
			if err != nil {
				return fmt.Errorf("retrieving policy: %w", err)
			}
		}

		if err := validateAppealDurationConfig(appeal, policy); err != nil {
			return fmt.Errorf("validating appeal duration: %w", err)
		}

		if err := s.addCreatorDetails(appeal, policy); err != nil {
			return fmt.Errorf("retrieving creator details: %w", err)
		}

		if err := s.fillApprovals(appeal, policy); err != nil {
			return fmt.Errorf("populating approvals: %w", err)
		}

		appeal.Init(policy)

		appeal.Policy = policy
		if err := s.approvalService.AdvanceApproval(ctx, appeal); err != nil {
			return fmt.Errorf("initializing approval step statuses: %w", err)
		}
		appeal.Policy = nil

		for _, approval := range appeal.Approvals {
			if approval.Index == len(appeal.Approvals)-1 && approval.Status == domain.ApprovalStatusApproved {
				newAccess, revokedAccess, err := s.prepareAccess(ctx, appeal)
				if err != nil {
					return fmt.Errorf("preparing access: %w", err)
				}
				newAccess.Resource = appeal.Resource
				appeal.Access = newAccess
				if revokedAccess != nil {
					if _, err := s.accessService.Revoke(ctx, revokedAccess.ID, domain.SystemActorName, RevokeReasonForExtension,
						access.SkipNotifications(),
						access.SkipRevokeAccessInProvider(),
					); err != nil {
						return fmt.Errorf("revoking previous access: %w", err)
					}
				} else {
					if err := s.CreateAccess(ctx, appeal); err != nil {
						return fmt.Errorf("granting access: %w", err)
					}
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

	if err := s.repo.BulkUpsert(appeals); err != nil {
		return fmt.Errorf("inserting appeals into db: %w", err)
	}

	if err := s.auditLogger.Log(ctx, AuditKeyBulkInsert, appeals); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	for _, a := range appeals {
		notifications = append(notifications, getApprovalNotifications(a)...)
	}

	if len(notifications) > 0 {
		if errs := s.notifier.Notify(notifications); errs != nil {
			for _, err1 := range errs {
				s.logger.Error("failed to send notifications", "error", err1.Error())
			}
		}
	}

	return nil
}

func validateAppealDurationConfig(appeal *domain.Appeal, policy *domain.Policy) error {
	// return nil if duration options are not configured for this policy
	if policy.AppealConfig == nil || policy.AppealConfig.DurationOptions == nil {
		return nil
	}
	for _, durationOption := range policy.AppealConfig.DurationOptions {
		if appeal.Options.Duration == durationOption.Value {
			return nil
		}
	}

	return fmt.Errorf("%w: %s", ErrOptionsDurationNotFound, appeal.Options.Duration)
}

// MakeAction Approve an approval step
func (s *Service) MakeAction(ctx context.Context, approvalAction domain.ApprovalAction) (*domain.Appeal, error) {
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
				approval.Approve()
				if i+1 <= len(appeal.Approvals)-1 {
					appeal.Approvals[i+1].Status = domain.ApprovalStatusPending
				}
				if err := s.approvalService.AdvanceApproval(ctx, appeal); err != nil {
					return nil, err
				}
			} else if approvalAction.Action == domain.AppealActionNameReject {
				approval.Reject()
				appeal.Reject()

				if i < len(appeal.Approvals)-1 {
					for j := i + 1; j < len(appeal.Approvals); j++ {
						appeal.Approvals[j].Skip()
						appeal.Approvals[j].UpdatedAt = TimeNow()
					}
				}
			} else {
				return nil, ErrActionInvalidValue
			}

			if appeal.Status == domain.AppealStatusActive {
				newAccess, revokedAccess, err := s.prepareAccess(ctx, appeal)
				if err != nil {
					return nil, fmt.Errorf("preparing access: %w", err)
				}
				newAccess.Resource = appeal.Resource
				appeal.Access = newAccess
				if revokedAccess != nil {
					if _, err := s.accessService.Revoke(ctx, revokedAccess.ID, domain.SystemActorName, RevokeReasonForExtension,
						access.SkipNotifications(),
						access.SkipRevokeAccessInProvider(),
					); err != nil {
						return nil, fmt.Errorf("revoking previous access: %w", err)
					}
				} else {
					if err := s.CreateAccess(ctx, appeal); err != nil {
						return nil, fmt.Errorf("granting access: %w", err)
					}
				}
			}

			if err := s.repo.Update(appeal); err != nil {
				if err := s.providerService.RevokeAccess(ctx, *appeal.Access); err != nil {
					return nil, fmt.Errorf("revoking access: %w", err)
				}
				return nil, fmt.Errorf("updating appeal: %w", err)
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
				if errs := s.notifier.Notify(notifications); errs != nil {
					for _, err1 := range errs {
						s.logger.Error("failed to send notifications", "error", err1.Error())
					}
				}
			}

			var auditKey string
			if approvalAction.Action == string(domain.ApprovalActionReject) {
				auditKey = AuditKeyReject
			} else if approvalAction.Action == string(domain.ApprovalActionApprove) {
				auditKey = AuditKeyApprove
			}
			if auditKey != "" {
				if err := s.auditLogger.Log(ctx, auditKey, approvalAction); err != nil {
					s.logger.Error("failed to record audit log", "error", err)
				}
			}

			return appeal, nil
		}
	}

	return nil, ErrApprovalNotFound
}

func (s *Service) Cancel(ctx context.Context, id string) (*domain.Appeal, error) {
	appeal, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: check only appeal creator who is allowed to cancel the appeal

	if err := checkIfAppealStatusStillPending(appeal.Status); err != nil {
		return nil, err
	}

	appeal.Cancel()
	if err := s.repo.Update(appeal); err != nil {
		return nil, err
	}

	if err := s.auditLogger.Log(ctx, AuditKeyCancel, map[string]interface{}{
		"appeal_id": id,
	}); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return appeal, nil
}

func (s *Service) Revoke(ctx context.Context, id string, actor, reason string) (*domain.Appeal, error) {
	appeal, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	revokedAppeal := &domain.Appeal{}
	*revokedAppeal = *appeal
	if err := revokedAppeal.Revoke(actor, reason); err != nil {
		return nil, err
	}
	appeal.Access.Resource = appeal.Resource
	if err := appeal.Access.Revoke(actor, reason); err != nil {
		return nil, fmt.Errorf("updating access status: %s", err)
	}

	if err := s.repo.Update(revokedAppeal); err != nil {
		return nil, err
	}

	if err := s.providerService.RevokeAccess(ctx, *appeal.Access); err != nil {
		if err := s.repo.Update(appeal); err != nil {
			return nil, err
		}
		return nil, err
	}

	if errs := s.notifier.Notify([]domain.Notification{{
		User: appeal.CreatedBy,
		Message: domain.NotificationMessage{
			Type: domain.NotificationTypeAccessRevoked,
			Variables: map[string]interface{}{
				"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
				"role":          appeal.Role,
				"account_type":  appeal.AccountType,
				"account_id":    appeal.AccountID,
			},
		},
	}}); errs != nil {
		for _, err1 := range errs {
			s.logger.Error("failed to send notifications", "error", err1.Error())
		}
	}

	if err := s.auditLogger.Log(ctx, AuditKeyRevoke, map[string]interface{}{
		"appeal_id": id,
		"reason":    reason,
	}); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return revokedAppeal, nil
}

func (s *Service) BulkRevoke(ctx context.Context, filters *domain.RevokeAppealsFilter, actor, reason string) ([]*domain.Appeal, error) {
	if filters.AccountIDs == nil || len(filters.AccountIDs) == 0 {
		return nil, fmt.Errorf("account_ids is required")
	}

	result := make([]*domain.Appeal, 0)
	// TODO: list access
	appeals, err := s.Find(ctx, &domain.ListAppealsFilter{
		Statuses:      []string{domain.AppealStatusActive},
		AccountIDs:    filters.AccountIDs,
		ProviderTypes: filters.ProviderTypes,
		ProviderURNs:  filters.ProviderURNs,
		ResourceTypes: filters.ResourceTypes,
		ResourceURNs:  filters.ResourceURNs,
	})
	if err != nil {
		return nil, err
	}

	if len(appeals) == 0 {
		return nil, nil
	}

	batchSize := 10
	timeLimiter := make(chan int, batchSize)

	for i := 1; i <= batchSize; i++ {
		timeLimiter <- i
	}

	go func() {
		for range time.Tick(1 * time.Second) {
			for i := 1; i <= batchSize; i++ {
				timeLimiter <- i
			}
		}
	}()

	totalRequests := len(appeals)
	done := make(chan *domain.Appeal, totalRequests)
	resourceAppealMap := make(map[string][]*domain.Appeal, 0)

	for _, appeal := range appeals {
		var resourceAppeals []*domain.Appeal
		var ok bool
		if resourceAppeals, ok = resourceAppealMap[appeal.ResourceID]; ok {
			resourceAppeals = append(resourceAppeals, appeal)
		} else {
			resourceAppeals = []*domain.Appeal{appeal}
		}
		resourceAppealMap[appeal.ResourceID] = resourceAppeals
	}

	for _, resourceAppeals := range resourceAppealMap {
		go s.expiredInActiveUserAppeal(ctx, timeLimiter, done, actor, reason, resourceAppeals)
	}

	var successRevoke []string
	var failedRevoke []string
	for {
		select {
		case appeal := <-done:
			if appeal.Status == domain.AppealStatusTerminated {
				successRevoke = append(successRevoke, appeal.ID)
			} else {
				failedRevoke = append(failedRevoke, appeal.ID)
			}
			result = append(result, appeal)
			if len(result) == totalRequests {
				s.logger.Info("successful appeal revocation", "count", len(successRevoke), "ids", successRevoke)
				s.logger.Info("failed appeal revocation", "count", len(failedRevoke), "ids", failedRevoke)
				return result, nil
			}
		}
	}
}

func (s *Service) expiredInActiveUserAppeal(ctx context.Context, timeLimiter chan int, done chan *domain.Appeal, actor string, reason string, appeals []*domain.Appeal) {
	for _, appeal := range appeals {
		<-timeLimiter

		revokedAppeal := &domain.Appeal{}
		*revokedAppeal = *appeal
		if err := revokedAppeal.Revoke(actor, reason); err != nil {
			s.logger.Error("failed to update appeal status", "id", appeal.ID, "error", err)
			return
		}
		revokedAppeal.Access.Resource = appeal.Resource
		if err := revokedAppeal.Access.Revoke(actor, reason); err != nil {
			s.logger.Error("failed to update access status", "id", revokedAppeal.Access.ID, "error", err)
			return
		}

		if err := s.providerService.RevokeAccess(ctx, *appeal.Access); err != nil {
			done <- appeal
			s.logger.Error("failed to revoke appeal-access in provider", "id", appeal.ID, "error", err)
			return
		}

		if err := s.repo.Update(revokedAppeal); err != nil {
			done <- appeal
			s.logger.Error("failed to update appeal-revoke status", "id", appeal.ID, "error", err)
			return
		} else {
			done <- revokedAppeal
			s.logger.Info("appeal revoked", "id", appeal.ID)
		}
	}
}

func (s *Service) AddApprover(ctx context.Context, appealID, approvalID, email string) (*domain.Appeal, error) {
	if err := s.validator.Var(email, "email"); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrApproverEmail, err)
	}

	appeal, approval, err := s.getApproval(appealID, approvalID)
	if err != nil {
		return nil, err
	}
	if appeal.Status != domain.AppealStatusPending {
		return nil, fmt.Errorf("%w: can't add new approver to appeal with %q status", ErrUnableToAddApprover, appeal.Status)
	}

	switch approval.Status {
	case domain.ApprovalStatusPending:
		break
	case domain.ApprovalStatusBlocked:
		// check if approval type is auto
		// this approach is the quickest way to assume that approval is auto, otherwise need to fetch the policy details and lookup the approval type which takes more time
		if approval.Approvers == nil || len(approval.Approvers) == 0 {
			// approval is automatic (strategy: auto) that is still on blocked
			return nil, fmt.Errorf("%w: can't modify approvers for approval with strategy auto", ErrUnableToAddApprover)
		}
	default:
		return nil, fmt.Errorf("%w: can't add approver to approval with %q status", ErrUnableToAddApprover, approval.Status)
	}

	if err := s.approvalService.AddApprover(ctx, approval.ID, email); err != nil {
		return nil, fmt.Errorf("adding new approver: %w", err)
	}
	approval.Approvers = append(approval.Approvers, email)

	if err := s.auditLogger.Log(ctx, AuditKeyAddApprover, approval); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	if errs := s.notifier.Notify([]domain.Notification{
		{
			User: email,
			Message: domain.NotificationMessage{
				Type: domain.NotificationTypeApproverNotification,
				Variables: map[string]interface{}{
					"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
					"role":          appeal.Role,
					"requestor":     appeal.CreatedBy,
					"appeal_id":     appeal.ID,
				},
			},
		},
	}); errs != nil {
		for _, err1 := range errs {
			s.logger.Error("failed to send notifications", "error", err1.Error())
		}
	}

	return appeal, nil
}

func (s *Service) DeleteApprover(ctx context.Context, appealID, approvalID, email string) (*domain.Appeal, error) {
	if err := s.validator.Var(email, "email"); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrApproverEmail, err)
	}

	appeal, approval, err := s.getApproval(appealID, approvalID)
	if err != nil {
		return nil, err
	}
	if appeal.Status != domain.AppealStatusPending {
		return nil, fmt.Errorf("%w: can't delete approver to appeal with %q status", ErrUnableToDeleteApprover, appeal.Status)
	}

	switch approval.Status {
	case domain.ApprovalStatusPending:
		break
	case domain.ApprovalStatusBlocked:
		// check if approval type is auto
		// this approach is the quickest way to assume that approval is auto, otherwise need to fetch the policy details and lookup the approval type which takes more time
		if approval.Approvers == nil || len(approval.Approvers) == 0 {
			// approval is automatic (strategy: auto) that is still on blocked
			return nil, fmt.Errorf("%w: can't modify approvers for approval with strategy auto", ErrUnableToDeleteApprover)
		}
	default:
		return nil, fmt.Errorf("%w: can't delete approver to approval with %q status", ErrUnableToDeleteApprover, approval.Status)
	}

	if len(approval.Approvers) == 1 {
		return nil, fmt.Errorf("%w: can't delete if there's only one approver", ErrUnableToDeleteApprover)
	}

	if err := s.approvalService.DeleteApprover(ctx, approvalID, email); err != nil {
		return nil, err
	}

	var newApprovers []string
	for _, a := range approval.Approvers {
		if a != email {
			newApprovers = append(newApprovers, a)
		}
	}
	approval.Approvers = newApprovers

	if err := s.auditLogger.Log(ctx, AuditKeyDeleteApprover, approval); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return appeal, nil
}

func (s *Service) getApproval(appealID, approvalID string) (*domain.Appeal, *domain.Approval, error) {
	if appealID == "" {
		return nil, nil, ErrAppealIDEmptyParam
	}
	if approvalID == "" {
		return nil, nil, ErrApprovalIDEmptyParam
	}

	appeal, err := s.repo.GetByID(appealID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting appeal details: %w", err)
	}

	approval := appeal.GetApproval(approvalID)
	if approval == nil {
		return nil, nil, ErrApprovalNotFound
	}

	return appeal, approval, nil
}

// getPendingAppealsMap returns map[account_id]map[resource_id]map[role]*domain.Appeal, error
func (s *Service) getPendingAppealsMap() (map[string]map[string]map[string]*domain.Appeal, error) {
	appeals, err := s.repo.Find(&domain.ListAppealsFilter{
		Statuses: []string{domain.AppealStatusPending},
	})
	if err != nil {
		return nil, err
	}

	appealsMap := map[string]map[string]map[string]*domain.Appeal{}
	for _, a := range appeals {
		if appealsMap[a.AccountID] == nil {
			appealsMap[a.AccountID] = map[string]map[string]*domain.Appeal{}
		}
		if appealsMap[a.AccountID][a.ResourceID] == nil {
			appealsMap[a.AccountID][a.ResourceID] = map[string]*domain.Appeal{}
		}
		appealsMap[a.AccountID][a.ResourceID][a.Role] = a
	}

	return appealsMap, nil
}

func (s *Service) getActiveAccessesMap(ctx context.Context) (map[string]map[string]map[string]*domain.Access, error) {
	accesses, err := s.accessService.List(ctx, domain.ListAccessesFilter{
		Statuses: []string{string(domain.AccessStatusActive)},
	})
	if err != nil {
		return nil, err
	}

	accessesMap := map[string]map[string]map[string]*domain.Access{}
	for i, a := range accesses {
		if accessesMap[a.AccountID] == nil {
			accessesMap[a.AccountID] = map[string]map[string]*domain.Access{}
		}
		if accessesMap[a.AccountID][a.ResourceID] == nil {
			accessesMap[a.AccountID][a.ResourceID] = map[string]*domain.Access{}
		}
		accessesMap[a.AccountID][a.ResourceID][a.Role] = &accesses[i]
	}

	return accessesMap, nil
}

func (s *Service) getResourcesMap(ctx context.Context, ids []string) (map[string]*domain.Resource, error) {
	filters := map[string]interface{}{"ids": ids}
	resources, err := s.resourceService.Find(ctx, filters)
	if err != nil {
		return nil, err
	}

	result := map[string]*domain.Resource{}
	for _, r := range resources {
		result[r.ID] = r
	}

	return result, nil
}

func (s *Service) getProvidersMap(ctx context.Context) (map[string]map[string]*domain.Provider, error) {
	providers, err := s.providerService.Find(ctx)
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

func (s *Service) getPoliciesMap(ctx context.Context) (map[string]map[uint]*domain.Policy, error) {
	policies, err := s.policyService.Find(ctx)
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

		approval := &domain.Approval{}
		if err := approval.Init(p, i, approverEmails); err != nil {
			return fmt.Errorf(`initializing approval "%s": %w`, step.Name, err)
		}
		approvals = append(approvals, approval)
	}

	a.Approvals = approvals
	return nil
}

func (s *Service) handleAppealRequirements(ctx context.Context, a *domain.Appeal, p *domain.Policy) error {
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
				resource, err := s.resourceService.Get(ctx, aa.Resource)
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
				if err := s.Create(ctx, []*domain.Appeal{additionalAppeal}, CreateWithAdditionalAppeal()); err != nil {
					if errors.Is(err, ErrAppealDuplicate) {
						continue
					}
					return fmt.Errorf("creating additional appeals: %w", err)
				}
			}
		}
	}
	return nil
}

func (s *Service) CreateAccess(ctx context.Context, a *domain.Appeal, opts ...CreateAppealOption) error {
	// TODO: rename to GrantAccess
	policy := a.Policy
	if policy == nil {
		p, err := s.policyService.GetOne(ctx, a.PolicyID, a.PolicyVersion)
		if err != nil {
			return fmt.Errorf("retrieving policy: %w", err)
		}
		policy = p
	}

	createAppealOpts := &createAppealOptions{}
	for _, opt := range opts {
		opt(createAppealOpts)
	}

	isAdditionalAppealCreation := createAppealOpts.IsAdditionalAppeal
	if !isAdditionalAppealCreation {
		if err := s.handleAppealRequirements(ctx, a, policy); err != nil {
			return fmt.Errorf("handling appeal requirements: %w", err)
		}
	}

	if err := s.providerService.GrantAccess(ctx, *a.Access); err != nil {
		return fmt.Errorf("granting access: %w", err)
	}

	return nil
}

func (s *Service) checkExtensionEligibility(a *domain.Appeal, p *domain.Provider, activeAccesses map[string]map[string]map[string]*domain.Access) error {
	if activeAccesses[a.AccountID] != nil &&
		activeAccesses[a.AccountID][a.ResourceID] != nil &&
		activeAccesses[a.AccountID][a.ResourceID][a.Role] != nil {
		if p.Config.Appeal != nil {
			if p.Config.Appeal.AllowActiveAccessExtensionIn == "" {
				return ErrAppealFoundActiveAccess
			}

			extensionDurationRule, err := time.ParseDuration(p.Config.Appeal.AllowActiveAccessExtensionIn)
			if err != nil {
				return fmt.Errorf("%w: %v: %v", ErrAppealInvalidExtensionDuration, p.Config.Appeal.AllowActiveAccessExtensionIn, err)
			}

			access := activeAccesses[a.AccountID][a.ResourceID][a.Role]
			if !access.IsEligibleForExtension(extensionDurationRule) {
				return ErrAccessNotEligibleForExtension
			}
		}
	}
	return nil
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

		userDetails, err := iamClient.GetUser(a.CreatedBy)
		if err != nil {
			return fmt.Errorf("fetching creator's user iam: %w", err)
		}

		var creator map[string]interface{}
		if userDetailsMap, ok := userDetails.(map[string]interface{}); ok {
			if p.IAM.Schema != nil {
				creator = map[string]interface{}{}
				for schemaKey, targetKey := range p.IAM.Schema {
					creator[schemaKey] = userDetailsMap[targetKey]
				}
			} else {
				creator = userDetailsMap
			}
		}

		a.Creator = creator
	}

	return nil
}

func addResource(a *domain.Appeal, resourcesMap map[string]*domain.Resource) error {
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

func validateAppeal(a *domain.Appeal, pendingAppealsMap map[string]map[string]map[string]*domain.Appeal) error {
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

func (s *Service) getPermissions(ctx context.Context, pc *domain.ProviderConfig, resourceType, role string) ([]string, error) {
	permissions, err := s.providerService.GetPermissions(ctx, pc, resourceType, role)
	if err != nil {
		return nil, err
	}

	if permissions == nil {
		return nil, nil
	}

	strPermissions := []string{}
	for _, p := range permissions {
		strPermissions = append(strPermissions, fmt.Sprintf("%s", p))
	}
	return strPermissions, nil
}

func (s *Service) prepareAccess(ctx context.Context, appeal *domain.Appeal) (newAccess *domain.Access, deactivatedAccess *domain.Access, err error) {
	activeAccesses, err := s.accessService.List(ctx, domain.ListAccessesFilter{
		AccountIDs:  []string{appeal.AccountID},
		ResourceIDs: []string{appeal.ResourceID},
		Statuses:    []string{string(domain.AccessStatusActive)},
		Permissions: appeal.Permissions,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve existing active accesses: %w", err)
	}

	if len(activeAccesses) > 0 {
		deactivatedAccess = &activeAccesses[0]
		if err := deactivatedAccess.Revoke(domain.SystemActorName, "Extended to a new access"); err != nil {
			return nil, nil, fmt.Errorf("revoking previous access: %w", err)
		}
	}

	if err := appeal.Activate(); err != nil {
		return nil, nil, fmt.Errorf("activating appeal: %w", err)
	}

	access, err := s.accessService.Prepare(ctx, *appeal)
	if err != nil {
		return nil, nil, err
	}

	return access, deactivatedAccess, nil
}
