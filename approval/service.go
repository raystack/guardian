package approval

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mcuadros/go-lookup"
	"github.com/odpf/guardian/domain"
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

func (s *service) GetPendingApprovals(user string) ([]*domain.Approval, error) {
	return s.repo.GetPendingApprovals(user)
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
		if p == nil {
			return ErrPolicyNotFound
		}

		policy = p
	}

	stepNameIndex := map[string]int{}
	for i, s := range policy.Steps {
		stepNameIndex[s.Name] = i
	}

	for _, approval := range appeal.Approvals {
		if approval.Status == domain.ApprovalStatusRejected {
			break
		} else if approval.Status == domain.ApprovalStatusPending {
			if approval.IsManualApproval() {
				break
			}

			stepConfig := policy.Steps[approval.Index]

			hasSkippedDependencies := false
			for _, d := range stepConfig.Dependencies {
				dependencyApprovalStep := appeal.Approvals[stepNameIndex[d]]
				if dependencyApprovalStep == nil {
					return ErrDependencyApprovalStepNotFound
				}

				if dependencyApprovalStep.Status == domain.ApprovalStatusSkipped {
					hasSkippedDependencies = true
				}
			}
			if hasSkippedDependencies {
				approval.Status = domain.ApprovalStatusSkipped
				break
			}

			for _, c := range stepConfig.Conditions {
				if c == nil {
					return ErrApprovalStepConditionNotFound
				}

				passed, err := s.evalCondition(appeal, c)
				if err != nil {
					return err
				}

				if passed {
					approval.Status = domain.ApprovalStatusApproved
				} else {
					if stepConfig.AllowFailed {
						approval.Status = domain.ApprovalStatusSkipped
					} else {
						approval.Status = domain.ApprovalStatusRejected
						appeal.Status = domain.AppealStatusRejected
					}
				}
			}
		}
	}

	return nil
}

func (s *service) evalCondition(a *domain.Appeal, c *domain.Condition) (bool, error) {
	if strings.HasPrefix(c.Field, domain.ApproversKeyResource) {
		if a.Resource == nil {
			return false, ErrNilResourceInAppeal
		}
		resourceMap, err := structToMap(a.Resource)
		if err != nil {
			return false, err
		}

		path := strings.TrimPrefix(c.Field, fmt.Sprintf("%s.", domain.ApproversKeyResource))
		value, err := lookup.LookupString(resourceMap, path)
		if err != nil {
			return false, err
		}

		expectedValue := c.Match.Eq
		return value.Interface() == expectedValue, nil
	}

	return false, ErrInvalidConditionField
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
