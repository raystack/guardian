package appeal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/goto/guardian/core/grant"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/pkg/evaluator"
	"github.com/goto/guardian/plugins/notifiers"
	"github.com/goto/guardian/utils"
	"github.com/goto/salt/log"
	"golang.org/x/sync/errgroup"
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

	RevokeReasonForExtension = "Automatically revoked for grant extension"
)

var TimeNow = time.Now

//go:generate mockery --name=repository --exported --with-expecter
type repository interface {
	BulkUpsert(context.Context, []*domain.Appeal) error
	Find(context.Context, *domain.ListAppealsFilter) ([]*domain.Appeal, error)
	GetByID(ctx context.Context, id string) (*domain.Appeal, error)
	Update(context.Context, *domain.Appeal) error
	GetAppealsTotalCount(context.Context, *domain.ListAppealsFilter) (int64, error)
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
	AddApprover(ctx context.Context, approvalID, email string) error
	DeleteApprover(ctx context.Context, approvalID, email string) error
}

//go:generate mockery --name=providerService --exported --with-expecter
type providerService interface {
	Find(context.Context) ([]*domain.Provider, error)
	GrantAccess(context.Context, domain.Grant) error
	RevokeAccess(context.Context, domain.Grant) error
	ValidateAppeal(context.Context, *domain.Appeal, *domain.Provider, *domain.Policy) error
	GetPermissions(context.Context, *domain.ProviderConfig, string, string) ([]interface{}, error)
}

//go:generate mockery --name=resourceService --exported --with-expecter
type resourceService interface {
	Find(context.Context, domain.ListResourcesFilter) ([]*domain.Resource, error)
	Get(context.Context, *domain.ResourceIdentifier) (*domain.Resource, error)
}

//go:generate mockery --name=grantService --exported --with-expecter
type grantService interface {
	List(context.Context, domain.ListGrantsFilter) ([]domain.Grant, error)
	Prepare(context.Context, domain.Appeal) (*domain.Grant, error)
	Revoke(ctx context.Context, id, actor, reason string, opts ...grant.Option) (*domain.Grant, error)
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
	GrantService    grantService
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
	grantService    grantService
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
		deps.GrantService,
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

	if !utils.IsValidUUID(id) {
		return nil, InvalidError{AppealID: id}
	}

	return s.repo.GetByID(ctx, id)
}

// Find appeals by filters
func (s *Service) Find(ctx context.Context, filters *domain.ListAppealsFilter) ([]*domain.Appeal, error) {
	return s.repo.Find(ctx, filters)
}

// Create record
func (s *Service) Create(ctx context.Context, appeals []*domain.Appeal, opts ...CreateAppealOption) error {
	createAppealOpts := &createAppealOptions{}
	for _, opt := range opts {
		opt(createAppealOpts)
	}
	isAdditionalAppealCreation := createAppealOpts.IsAdditionalAppeal

	resourceIDs := []string{}
	accountIDs := []string{}
	for _, a := range appeals {
		resourceIDs = append(resourceIDs, a.ResourceID)
		accountIDs = append(accountIDs, a.AccountID)
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

	pendingAppeals, err := s.getAppealsMap(ctx, &domain.ListAppealsFilter{
		Statuses:   []string{domain.AppealStatusPending},
		AccountIDs: accountIDs,
	})
	if err != nil {
		return fmt.Errorf("listing pending appeals: %w", err)
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

		var policy *domain.Policy
		if isAdditionalAppealCreation && appeal.PolicyID != "" && appeal.PolicyVersion != 0 {
			policy = policies[appeal.PolicyID][appeal.PolicyVersion]
		} else {
			policy, err = getPolicy(appeal, provider, policies)
			if err != nil {
				return fmt.Errorf("retrieving policy: %w", err)
			}
		}

		activeGrant, err := s.findActiveGrant(ctx, appeal)
		if err != nil && err != ErrGrantNotFound {
			return err
		}

		if activeGrant != nil {
			if err := s.checkExtensionEligibility(appeal, provider, policy, activeGrant); err != nil {
				return fmt.Errorf("checking grant extension eligibility: %w", err)
			}
		}

		if err := s.providerService.ValidateAppeal(ctx, appeal, provider, policy); err != nil {
			return fmt.Errorf("validating appeal based on provider: %w", err)
		}

		strPermissions, err := s.getPermissions(ctx, provider.Config, appeal.Resource.Type, appeal.Role)
		if err != nil {
			return fmt.Errorf("getting permissions list: %w", err)
		}
		appeal.Permissions = strPermissions

		if err := validateAppealDurationConfig(appeal, policy); err != nil {
			return fmt.Errorf("validating appeal duration: %w", err)
		}

		if err := validateAppealOnBehalf(appeal, policy); err != nil {
			return fmt.Errorf("validating cross-individual appeal: %w", err)
		}

		if err := s.addCreatorDetails(appeal, policy); err != nil {
			return fmt.Errorf("retrieving creator details: %w", err)
		}

		if err := appeal.ApplyPolicy(policy); err != nil {
			return fmt.Errorf("populating approvals: %w", err)
		}

		if err := appeal.AdvanceApproval(policy); err != nil {
			return fmt.Errorf("initializing approval step statuses: %w", err)
		}
		appeal.Policy = nil

		for _, approval := range appeal.Approvals {
			if approval.Index == len(appeal.Approvals)-1 && (approval.Status == domain.ApprovalStatusApproved || appeal.Status == domain.AppealStatusApproved) {
				newGrant, revokedGrant, err := s.prepareGrant(ctx, appeal)
				if err != nil {
					return fmt.Errorf("preparing grant: %w", err)
				}
				newGrant.Resource = appeal.Resource
				appeal.Grant = newGrant
				if revokedGrant != nil {
					if _, err := s.grantService.Revoke(ctx, revokedGrant.ID, domain.SystemActorName, RevokeReasonForExtension,
						grant.SkipNotifications(),
						grant.SkipRevokeAccessInProvider(),
					); err != nil {
						return fmt.Errorf("revoking previous grant: %w", err)
					}
				} else {
					if err := s.GrantAccessToProvider(ctx, appeal, opts...); err != nil {
						return fmt.Errorf("granting access: %w", err)
					}
				}

				notifications = append(notifications, domain.Notification{
					User: appeal.CreatedBy,
					Labels: map[string]string{
						"appeal_id": appeal.ID,
					},
					Message: domain.NotificationMessage{
						Type: domain.NotificationTypeAppealApproved,
						Variables: map[string]interface{}{
							"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
							"role":          appeal.Role,
							"account_id":    appeal.AccountID,
							"appeal_id":     appeal.ID,
							"requestor":     appeal.CreatedBy,
						},
					},
				})

				notifications = addOnBehalfApprovedNotification(appeal, notifications)
			}
		}
	}

	if err := s.repo.BulkUpsert(ctx, appeals); err != nil {
		return fmt.Errorf("inserting appeals into db: %w", err)
	}

	if err := s.auditLogger.Log(ctx, AuditKeyBulkInsert, appeals); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	for _, a := range appeals {
		if a.Status == domain.AppealStatusRejected {
			var reason string
			for _, approval := range a.Approvals {
				if approval.Status == domain.ApprovalStatusRejected {
					reason = approval.Reason
					break
				}
			}

			notifications = append(notifications, domain.Notification{
				User: a.CreatedBy,
				Labels: map[string]string{
					"appeal_id": a.ID,
				},
				Message: domain.NotificationMessage{
					Type: domain.NotificationTypeAppealRejected,
					Variables: map[string]interface{}{
						"resource_name": fmt.Sprintf("%s (%s: %s)", a.Resource.Name, a.Resource.ProviderType, a.Resource.URN),
						"role":          a.Role,
						"account_id":    a.AccountID,
						"appeal_id":     a.ID,
						"requestor":     a.CreatedBy,
						"reason":        reason,
					},
				},
			})
		}

		notifications = append(notifications, s.getApprovalNotifications(a)...)
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

func (s *Service) findActiveGrant(ctx context.Context, a *domain.Appeal) (*domain.Grant, error) {
	grants, err := s.grantService.List(ctx, domain.ListGrantsFilter{
		Statuses:    []string{string(domain.GrantStatusActive)},
		AccountIDs:  []string{a.AccountID},
		ResourceIDs: []string{a.ResourceID},
		Roles:       []string{a.Role},
		OrderBy:     []string{"updated_at:desc"},
	})

	if err != nil {
		return nil, fmt.Errorf("listing active grants: %w", err)
	}

	if len(grants) == 0 {
		return nil, ErrGrantNotFound
	}

	return &grants[0], nil
}

func addOnBehalfApprovedNotification(appeal *domain.Appeal, notifications []domain.Notification) []domain.Notification {
	if appeal.AccountType == domain.DefaultAppealAccountType && appeal.AccountID != appeal.CreatedBy {
		notifications = append(notifications, domain.Notification{
			User: appeal.AccountID,
			Labels: map[string]string{
				"appeal_id": appeal.ID,
			},
			Message: domain.NotificationMessage{
				Type: domain.NotificationTypeOnBehalfAppealApproved,
				Variables: map[string]interface{}{
					"appeal_id":     appeal.ID,
					"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
					"role":          appeal.Role,
					"account_id":    appeal.AccountID,
					"requestor":     appeal.CreatedBy,
				},
			},
		})
	}
	return notifications
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

func validateAppealOnBehalf(a *domain.Appeal, policy *domain.Policy) error {
	if a.AccountType == domain.DefaultAppealAccountType {
		if policy.AppealConfig != nil && policy.AppealConfig.AllowOnBehalf {
			return nil
		}
		if a.AccountID != a.CreatedBy {
			return ErrCannotCreateAppealForOtherUser
		}
	}
	return nil
}

// UpdateApproval Approve an approval step
func (s *Service) UpdateApproval(ctx context.Context, approvalAction domain.ApprovalAction) (*domain.Appeal, error) {
	if err := utils.ValidateStruct(approvalAction); err != nil {
		return nil, err
	}

	appeal, err := s.GetByID(ctx, approvalAction.AppealID)
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
		}

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
			if appeal.Policy == nil {
				appeal.Policy, err = s.policyService.GetOne(ctx, appeal.PolicyID, appeal.PolicyVersion)
				if err != nil {
					return nil, err
				}
			}
			if err := appeal.AdvanceApproval(appeal.Policy); err != nil {
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

		if appeal.Status == domain.AppealStatusApproved {
			newGrant, revokedGrant, err := s.prepareGrant(ctx, appeal)
			if err != nil {
				return nil, fmt.Errorf("preparing grant: %w", err)
			}
			newGrant.Resource = appeal.Resource
			appeal.Grant = newGrant
			if revokedGrant != nil {
				if _, err := s.grantService.Revoke(ctx, revokedGrant.ID, domain.SystemActorName, RevokeReasonForExtension,
					grant.SkipNotifications(),
					grant.SkipRevokeAccessInProvider(),
				); err != nil {
					return nil, fmt.Errorf("revoking previous grant: %w", err)
				}
			} else {
				if err := s.GrantAccessToProvider(ctx, appeal); err != nil {
					return nil, fmt.Errorf("granting access: %w", err)
				}
			}
		}

		if err := s.Update(ctx, appeal); err != nil {
			if err := s.providerService.RevokeAccess(ctx, *appeal.Grant); err != nil {
				return nil, fmt.Errorf("revoking access: %w", err)
			}
			return nil, fmt.Errorf("updating appeal: %w", err)
		}

		notifications := []domain.Notification{}
		if appeal.Status == domain.AppealStatusApproved {
			notifications = append(notifications, domain.Notification{
				User: appeal.CreatedBy,
				Labels: map[string]string{
					"appeal_id": appeal.ID,
				},
				Message: domain.NotificationMessage{
					Type: domain.NotificationTypeAppealApproved,
					Variables: map[string]interface{}{
						"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
						"role":          appeal.Role,
						"account_id":    appeal.AccountID,
						"appeal_id":     appeal.ID,
						"requestor":     appeal.CreatedBy,
					},
				},
			})
			notifications = addOnBehalfApprovedNotification(appeal, notifications)
		} else if appeal.Status == domain.AppealStatusRejected {
			notifications = append(notifications, domain.Notification{
				User: appeal.CreatedBy,
				Labels: map[string]string{
					"appeal_id": appeal.ID,
				},
				Message: domain.NotificationMessage{
					Type: domain.NotificationTypeAppealRejected,
					Variables: map[string]interface{}{
						"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
						"role":          appeal.Role,
						"account_id":    appeal.AccountID,
						"appeal_id":     appeal.ID,
						"requestor":     appeal.CreatedBy,
					},
				},
			})
		} else {
			notifications = append(notifications, s.getApprovalNotifications(appeal)...)
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

	return nil, ErrApprovalNotFound
}

func (s *Service) Update(ctx context.Context, appeal *domain.Appeal) error {
	return s.repo.Update(ctx, appeal)
}

func (s *Service) Cancel(ctx context.Context, id string) (*domain.Appeal, error) {
	if id == "" {
		return nil, ErrAppealIDEmptyParam
	}

	if !utils.IsValidUUID(id) {
		return nil, InvalidError{AppealID: id}
	}

	appeal, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: check only appeal creator who is allowed to cancel the appeal

	if err := checkIfAppealStatusStillPending(appeal.Status); err != nil {
		return nil, err
	}

	appeal.Cancel()
	if err := s.repo.Update(ctx, appeal); err != nil {
		return nil, err
	}

	if err := s.auditLogger.Log(ctx, AuditKeyCancel, map[string]interface{}{
		"appeal_id": id,
	}); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return appeal, nil
}

func (s *Service) AddApprover(ctx context.Context, appealID, approvalID, email string) (*domain.Appeal, error) {
	if err := s.validator.Var(email, "email"); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrApproverEmail, err)
	}

	appeal, approval, err := s.getApproval(ctx, appealID, approvalID)
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

	duration := domain.PermanentDurationLabel
	if !appeal.IsDurationEmpty() {
		duration, err = utils.GetReadableDuration(appeal.Options.Duration)
		if err != nil {
			s.logger.Error("failed to get readable duration", "error", err, "appeal_id", appeal.ID)
		}
	}

	if errs := s.notifier.Notify([]domain.Notification{
		{
			User: email,
			Labels: map[string]string{
				"appeal_id": appeal.ID,
			},
			Message: domain.NotificationMessage{
				Type: domain.NotificationTypeApproverNotification,
				Variables: map[string]interface{}{
					"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
					"role":          appeal.Role,
					"requestor":     appeal.CreatedBy,
					"appeal_id":     appeal.ID,
					"account_id":    appeal.AccountID,
					"account_type":  appeal.AccountType,
					"provider_type": appeal.Resource.ProviderType,
					"resource_type": appeal.Resource.Type,
					"created_at":    appeal.CreatedAt,
					"approval_step": approval.Name,
					"actor":         email,
					"details":       appeal.Details,
					"duration":      duration,
					"creator":       appeal.Creator,
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

	appeal, approval, err := s.getApproval(ctx, appealID, approvalID)
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

func (s *Service) getApproval(ctx context.Context, appealID, approvalID string) (*domain.Appeal, *domain.Approval, error) {
	if appealID == "" {
		return nil, nil, ErrAppealIDEmptyParam
	}
	if approvalID == "" {
		return nil, nil, ErrApprovalIDEmptyParam
	}

	appeal, err := s.repo.GetByID(ctx, appealID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting appeal details: %w", err)
	}

	approval := appeal.GetApproval(approvalID)
	if approval == nil {
		return nil, nil, ErrApprovalNotFound
	}

	return appeal, approval, nil
}

// getAppealsMap returns map[account_id]map[resource_id]map[role]*domain.Appeal, error
func (s *Service) getAppealsMap(ctx context.Context, filters *domain.ListAppealsFilter) (map[string]map[string]map[string]*domain.Appeal, error) {
	appeals, err := s.repo.Find(ctx, filters)
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

func (s *Service) getResourcesMap(ctx context.Context, ids []string) (map[string]*domain.Resource, error) {
	filters := domain.ListResourcesFilter{IDs: ids}
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

func (s *Service) getApprovalNotifications(appeal *domain.Appeal) []domain.Notification {
	notifications := []domain.Notification{}
	approval := appeal.GetNextPendingApproval()

	duration := domain.PermanentDurationLabel
	var err error
	if !appeal.IsDurationEmpty() {
		duration, err = utils.GetReadableDuration(appeal.Options.Duration)
		if err != nil {
			s.logger.Error("failed to get readable duration", "error", err, "appeal_id", appeal.ID)
		}
	}

	if approval != nil {
		for _, approver := range approval.Approvers {
			notifications = append(notifications, domain.Notification{
				User: approver,
				Labels: map[string]string{
					"appeal_id": appeal.ID,
				},
				Message: domain.NotificationMessage{
					Type: domain.NotificationTypeApproverNotification,
					Variables: map[string]interface{}{
						"resource_name": fmt.Sprintf("%s (%s: %s)", appeal.Resource.Name, appeal.Resource.ProviderType, appeal.Resource.URN),
						"role":          appeal.Role,
						"requestor":     appeal.CreatedBy,
						"appeal_id":     appeal.ID,
						"account_id":    appeal.AccountID,
						"account_type":  appeal.AccountType,
						"provider_type": appeal.Resource.ProviderType,
						"resource_type": appeal.Resource.Type,
						"created_at":    appeal.CreatedAt,
						"approval_step": approval.Name,
						"actor":         approver,
						"details":       appeal.Details,
						"duration":      duration,
						"creator":       appeal.Creator,
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
	case domain.AppealStatusApproved:
		err = ErrAppealStatusApproved
	case domain.AppealStatusRejected:
		err = ErrAppealStatusRejected
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

func (s *Service) handleAppealRequirements(ctx context.Context, a *domain.Appeal, p *domain.Policy) error {
	if p.Requirements != nil && len(p.Requirements) > 0 {
		g, ctx := errgroup.WithContext(ctx)

		for reqIndex, r := range p.Requirements {
			isAppealMatchesRequirement, err := r.On.IsMatch(a)
			if err != nil {
				return fmt.Errorf("evaluating requirements[%v]: %v", reqIndex, err)
			}
			if !isAppealMatchesRequirement {
				continue
			}

			for _, aa := range r.Appeals {
				aa := aa // https://golang.org/doc/faq#closures_and_goroutines
				g.Go(func() error {
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
							return nil
						}
						return fmt.Errorf("creating additional appeals: %w", err)
					}
					return nil
				})
			}
		}
		if err := g.Wait(); err == nil {
			return err
		}
	}
	return nil
}

func (s *Service) GrantAccessToProvider(ctx context.Context, a *domain.Appeal, opts ...CreateAppealOption) error {
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

	if err := s.providerService.GrantAccess(ctx, *a.Grant); err != nil {
		return fmt.Errorf("granting access: %w", err)
	}

	return nil
}

func (s *Service) checkExtensionEligibility(a *domain.Appeal, p *domain.Provider, policy *domain.Policy, activeGrant *domain.Grant) error {
	AllowActiveAccessExtensionIn := ""

	// Default to use provider config if policy config is not set
	if p.Config.Appeal != nil {
		AllowActiveAccessExtensionIn = p.Config.Appeal.AllowActiveAccessExtensionIn
	}

	// Use policy config if set
	if policy != nil &&
		policy.AppealConfig != nil &&
		policy.AppealConfig.AllowActiveAccessExtensionIn != "" {
		AllowActiveAccessExtensionIn = policy.AppealConfig.AllowActiveAccessExtensionIn
	}

	if AllowActiveAccessExtensionIn == "" {
		return ErrAppealFoundActiveGrant
	}

	extensionDurationRule, err := time.ParseDuration(AllowActiveAccessExtensionIn)
	if err != nil {
		return fmt.Errorf("%w: %v: %v", ErrAppealInvalidExtensionDuration, AllowActiveAccessExtensionIn, err)
	}

	if !activeGrant.IsEligibleForExtension(extensionDurationRule) {
		return ErrGrantNotEligibleForExtension
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
	if p.IAM == nil {
		return nil
	}

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
		if p.AppealConfig != nil && p.AppealConfig.AllowCreatorDetailsFailure {
			s.logger.Warn("fetching creator's user iam", "error", err)
			return nil
		}
		return fmt.Errorf("fetching creator's user iam: %w", err)
	}

	userDetailsMap, ok := userDetails.(map[string]interface{})
	if !ok {
		return nil
	}

	if p.IAM.Schema == nil {
		a.Creator = userDetailsMap
		return nil
	}

	creator := map[string]interface{}{}
	for schemaKey, targetKey := range p.IAM.Schema {
		if strings.Contains(targetKey, "$response") {
			params := map[string]interface{}{
				"response": userDetailsMap,
			}
			v, err := evaluator.Expression(targetKey).EvaluateWithVars(params)
			if err != nil {
				return fmt.Errorf("evaluating expression: %w", err)
			}
			creator[schemaKey] = v
		} else {
			creator[schemaKey] = userDetailsMap[targetKey]
		}
	}

	a.Creator = creator
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

func (s *Service) prepareGrant(ctx context.Context, appeal *domain.Appeal) (newGrant *domain.Grant, deactivatedGrant *domain.Grant, err error) {
	activeGrants, err := s.grantService.List(ctx, domain.ListGrantsFilter{
		AccountIDs:  []string{appeal.AccountID},
		ResourceIDs: []string{appeal.ResourceID},
		Statuses:    []string{string(domain.GrantStatusActive)},
		Permissions: appeal.Permissions,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve existing active grants: %w", err)
	}

	if len(activeGrants) > 0 {
		deactivatedGrant = &activeGrants[0]
		if err := deactivatedGrant.Revoke(domain.SystemActorName, "Extended to a new grant"); err != nil {
			return nil, nil, fmt.Errorf("revoking previous grant: %w", err)
		}
	}

	if err := appeal.Approve(); err != nil {
		return nil, nil, fmt.Errorf("activating appeal: %w", err)
	}

	grant, err := s.grantService.Prepare(ctx, *appeal)
	if err != nil {
		return nil, nil, err
	}

	return grant, deactivatedGrant, nil
}

func (s *Service) GetAppealsTotalCount(ctx context.Context, filters *domain.ListAppealsFilter) (int64, error) {
	return s.repo.GetAppealsTotalCount(ctx, filters)
}
