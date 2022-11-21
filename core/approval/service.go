package approval

import (
	"context"

	"github.com/odpf/guardian/domain"
)

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

type ServiceDeps struct {
	Repository    repository
	PolicyService policyService
}
type Service struct {
	repo          repository
	policyService policyService
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		deps.Repository,
		deps.PolicyService,
	}
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
