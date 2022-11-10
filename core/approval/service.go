package approval

import (
	"context"
	"fmt"
	"time"

	"github.com/odpf/guardian/core/grant"
	"github.com/odpf/guardian/domain"
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

	RevokeReasonForExtension = "Automatically revoked for grant extension"
)

var TimeNow = time.Now

//go:generate mockery --name=repository --exported --with-expecter
type repository interface {
	BulkInsert(context.Context, []*domain.Approval) error
	ListApprovals(context.Context, *domain.ListApprovalsFilter) ([]*domain.Approval, error)
	AddApprover(context.Context, *domain.Approver) error
	DeleteApprover(ctx context.Context, approvalID, email string) error
}

//go:generate mockery --name=policyService --exported --with-expecter
type policyService interface {
	GetOne(context.Context, string, uint) (*domain.Policy, error)
}

//go:generate mockery --name=grantService --exported --with-expecter
type grantService interface {
	List(context.Context, domain.ListGrantsFilter) ([]domain.Grant, error)
	Prepare(context.Context, domain.Appeal) (*domain.Grant, error)
	Revoke(ctx context.Context, id, actor, reason string, opts ...grant.Option) (*domain.Grant, error)
}

//go:generate mockery --name=appealService --exported --with-expecter
type appealService interface {
	GetByID(ctx context.Context, id string) (*domain.Appeal, error)
	GrantAccessToProvider(ctx context.Context, a *domain.Appeal) error
	Update(ctx context.Context, appeal *domain.Appeal) error
}

//go:generate mockery --name=providerService --exported --with-expecter
type providerService interface {
	Find(context.Context) ([]*domain.Provider, error)
	GrantAccess(context.Context, domain.Grant) error
	RevokeAccess(context.Context, domain.Grant) error
	ValidateAppeal(context.Context, *domain.Appeal, *domain.Provider, *domain.Policy) error
	GetPermissions(context.Context, *domain.ProviderConfig, string, string) ([]interface{}, error)
}

//go:generate mockery --name=auditLogger --exported --with-expecter
type auditLogger interface {
	Log(ctx context.Context, action string, data interface{}) error
}

//go:generate mockery --name=notifier --exported --with-expecter
type notifier interface {
	notifiers.Client
}

type ServiceDeps struct {
	Repository      repository
	PolicyService   policyService
	GrantService    grantService
	ProviderService providerService

	Notifier    notifier
	Logger      log.Logger
	AuditLogger auditLogger
}
type Service struct {
	repo            repository
	policyService   policyService
	grantService    grantService
	providerService providerService
	appealService   appealService

	notifier    notifier
	logger      log.Logger
	auditLogger auditLogger

	TimeNow func() time.Time
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		deps.Repository,
		deps.PolicyService,
		deps.GrantService,
		deps.ProviderService,
		nil,
		deps.Notifier,
		deps.Logger,
		deps.AuditLogger,
		time.Now,
	}
}

func (s *Service) SetAppealService(appealService appealService) {
	s.appealService = appealService
}

func (s *Service) ListApprovals(ctx context.Context, filters *domain.ListApprovalsFilter) ([]*domain.Approval, error) {
	return s.repo.ListApprovals(ctx, filters)
}

func (s *Service) BulkInsert(ctx context.Context, approvals []*domain.Approval) error {
	return s.repo.BulkInsert(ctx, approvals)
}

func (s *Service) AddApprover(ctx context.Context, approvalID, email string) error {
	return s.repo.AddApprover(ctx, &domain.Approver{
		ApprovalID: approvalID,
		Email:      email,
	})
}

func (s *Service) DeleteApprover(ctx context.Context, approvalID, email string) error {
	return s.repo.DeleteApprover(ctx, approvalID, email)
}

// UpdateApproval updates an approval.
func (s *Service) UpdateApproval(ctx context.Context, approvalAction domain.ApprovalAction) (*domain.Appeal, error) {
	if err := utils.ValidateStruct(approvalAction); err != nil {
		return nil, err
	}

	appeal, err := s.appealService.GetByID(ctx, approvalAction.AppealID)
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
			if err := appeal.AdvanceApproval(ctx); err != nil {
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
				if err := s.appealService.GrantAccessToProvider(ctx, appeal); err != nil {
					return nil, fmt.Errorf("granting access: %w", err)
				}
			}
		}

		if err := s.appealService.Update(ctx, appeal); err != nil {
			if err := s.providerService.RevokeAccess(ctx, *appeal.Grant); err != nil {
				return nil, fmt.Errorf("revoking access: %w", err)
			}
			return nil, fmt.Errorf("updating appeal: %w", err)
		}

		notifications := []domain.Notification{}
		if appeal.Status == domain.AppealStatusApproved {
			notifications = append(notifications, domain.Notification{
				User: appeal.CreatedBy,
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

	return nil, ErrApprovalNotFound
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

func addOnBehalfApprovedNotification(appeal *domain.Appeal, notifications []domain.Notification) []domain.Notification {
	if appeal.AccountType == domain.DefaultAppealAccountType && appeal.AccountID != appeal.CreatedBy {
		notifications = append(notifications, domain.Notification{
			User: appeal.AccountID,
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
