package approval

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/pkg/evaluator"
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

func (s *Service) AdvanceApproval(ctx context.Context, appeal *domain.Appeal) error {
	policy := appeal.Policy
	if policy == nil {
		p, err := s.policyService.GetOne(ctx, appeal.PolicyID, appeal.PolicyVersion)
		if err != nil {
			return err
		}
		policy = p
	}

	stepNameIndex := map[string]int{}
	for i, s := range policy.Steps {
		stepNameIndex[s.Name] = i
	}

	for i, approval := range appeal.Approvals {
		if approval.Status == domain.ApprovalStatusRejected {
			break
		}
		if approval.Status == domain.ApprovalStatusPending {
			stepConfig := policy.Steps[approval.Index]

			appealMap, err := structToMap(appeal)
			if err != nil {
				return fmt.Errorf("parsing appeal struct to map: %w", err)
			}

			if stepConfig.When != "" {
				v, err := evaluator.Expression(stepConfig.When).EvaluateWithVars(map[string]interface{}{
					"appeal": appealMap,
				})
				if err != nil {
					return err
				}

				isFalsy := reflect.ValueOf(v).IsZero()
				if isFalsy {
					approval.Status = domain.ApprovalStatusSkipped
					if i < len(appeal.Approvals)-1 {
						appeal.Approvals[i+1].Status = domain.ApprovalStatusPending
					}
				}
			}

			if approval.Status != domain.ApprovalStatusSkipped && stepConfig.Strategy == domain.ApprovalStepStrategyAuto {
				v, err := evaluator.Expression(stepConfig.ApproveIf).EvaluateWithVars(map[string]interface{}{
					"appeal": appealMap,
				})
				if err != nil {
					return err
				}

				isFalsy := reflect.ValueOf(v).IsZero()
				if isFalsy {
					if stepConfig.AllowFailed {
						approval.Status = domain.ApprovalStatusSkipped
						if i+1 <= len(appeal.Approvals)-1 {
							appeal.Approvals[i+1].Status = domain.ApprovalStatusPending
						}
					} else {
						approval.Status = domain.ApprovalStatusRejected
						approval.Reason = stepConfig.RejectionReason
						appeal.Status = domain.AppealStatusRejected
					}
				} else {
					approval.Status = domain.ApprovalStatusApproved
					if i+1 <= len(appeal.Approvals)-1 {
						appeal.Approvals[i+1].Status = domain.ApprovalStatusPending
					}
				}
			}
		}
		if i == len(appeal.Approvals)-1 && (approval.Status == domain.ApprovalStatusSkipped || approval.Status == domain.ApprovalStatusApproved) {
			appeal.Status = domain.AppealStatusApproved
		}
	}

	return nil
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
