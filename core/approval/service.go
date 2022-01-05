package approval

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/evaluator"
)

type service struct {
	repo          domain.ApprovalRepository
	policyService domain.PolicyService
}

func NewService(
	ar domain.ApprovalRepository,
	ps domain.PolicyService,
) *service {
	return &service{ar, ps}
}

func (s *service) ListApprovals(filters *domain.ListApprovalsFilter) ([]*domain.Approval, error) {
	return s.repo.ListApprovals(filters)
}

func (s *service) BulkInsert(approvals []*domain.Approval) error {
	return s.repo.BulkInsert(approvals)
}

func (s *service) AdvanceApproval(appeal *domain.Appeal) error {
	policy := appeal.Policy
	if policy == nil {
		p, err := s.policyService.GetOne(appeal.PolicyID, appeal.PolicyVersion)
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
		} else if approval.Status == domain.ApprovalStatusPending {
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
					break
				}
			}

			if approval.IsManualApproval() {
				break
			}

			if stepConfig.Strategy == domain.ApprovalStepStrategyAuto {
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
	}

	return nil
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
