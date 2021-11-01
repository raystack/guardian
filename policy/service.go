package policy

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/evaluator"
)

// Service handling the business logics
type Service struct {
	validator        *validator.Validate
	policyRepository domain.PolicyRepository
	resourceService  domain.ResourceService
	providerService  domain.ProviderService
}

// NewService returns service struct
func NewService(v *validator.Validate, pr domain.PolicyRepository, rs domain.ResourceService, ps domain.ProviderService) *Service {
	return &Service{v, pr, rs, ps}
}

// Create record
func (s *Service) Create(p *domain.Policy) error {
	p.Version = 1

	if err := s.validatePolicy(p); err != nil {
		return fmt.Errorf("policy validation: %w", err)
	}

	return s.policyRepository.Create(p)
}

// Find records
func (s *Service) Find() ([]*domain.Policy, error) {
	return s.policyRepository.Find()
}

// GetOne record
func (s *Service) GetOne(id string, version uint) (*domain.Policy, error) {
	p, err := s.policyRepository.GetOne(id, version)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Update a record
func (s *Service) Update(p *domain.Policy) error {
	if p.ID == "" {
		return ErrEmptyIDParam
	}

	if err := s.validatePolicy(p, "Version"); err != nil {
		return fmt.Errorf("policy validation: %w", err)
	}

	latestPolicy, err := s.GetOne(p.ID, p.Version)
	if err != nil {
		return err
	}

	p.Version = latestPolicy.Version + 1
	return s.policyRepository.Create(p)
}

func (s *Service) validatePolicy(p *domain.Policy, excludedFields ...string) error {
	if containsWhitespaces(p.ID) {
		return ErrIDContainsWhitespaces
	}

	if err := s.validator.StructExcept(p, excludedFields...); err != nil {
		return err
	}

	if err := s.validateSteps(p.Steps); err != nil {
		return err
	}

	if err := s.validateRequirements(p.Requirements); err != nil {
		return fmt.Errorf("invalid requirements: %w", err)
	}

	return nil
}

func (s *Service) validateRequirements(requirements []*domain.Requirement) error {
	for i, r := range requirements {
		for j, aa := range r.Appeals {
			resource, err := s.resourceService.Get(aa.Resource)
			if err != nil {
				return fmt.Errorf("requirement[%v].appeals[%v].resource: %w", i, j, err)
			}
			provider, err := s.providerService.GetOne(resource.ProviderType, resource.ProviderURN)
			if err != nil {
				return fmt.Errorf("requirement[%v].appeals[%v].resource: retrieving provider: %w", i, j, err)
			}

			appeal := &domain.Appeal{
				ResourceID: resource.ID,
				Resource:   resource,
				Role:       aa.Role,
				Options:    aa.Options,
			}
			if err := s.providerService.ValidateAppeal(appeal, provider); err != nil {
				return fmt.Errorf("requirement[%v].appeals[%v]: %w", i, j, err)
			}
		}
	}
	return nil
}

func (s *Service) validateSteps(steps []*domain.Step) error {
	for _, step := range steps {
		if containsWhitespaces(step.Name) {
			return fmt.Errorf(`%w: "%s"`, ErrStepNameContainsWhitespaces, step.Name)
		}

		if step.Approvers != nil {
			for _, approver := range step.Approvers {
				if err := s.validateApprover(approver); err != nil {
					return fmt.Errorf(`validating approver "%s": %w`, approver, err)
				}
			}
		}
	}

	return nil
}

func (s *Service) validateApprover(expr string) error {
	if err := s.validator.Var(expr, "email"); err == nil {
		return nil
	}

	// skip validation if expression is accessing resource details (interface{})
	if strings.Contains(expr, "$appeal.resource.details") {
		return nil
	}

	dummyAppeal := &domain.Appeal{
		Resource: &domain.Resource{},
	}
	dummyAppealMap, err := structToMap(dummyAppeal)
	if err != nil {
		return fmt.Errorf("parsing appeal to map: %w", err)
	}
	approvers, err := evaluator.Expression(expr).EvaluateWithVars(map[string]interface{}{
		"appeal": dummyAppealMap,
	})
	if err != nil {
		return fmt.Errorf("evaluating expression: %w", err)
	}

	// value type should be string or []string
	value := reflect.ValueOf(approvers)
	switch value.Type().Kind() {
	case reflect.String:
		return nil
	case reflect.Slice:
		elem := value.Type().Elem()
		switch elem.Kind() {
		case
			reflect.String,
			reflect.Interface: // can't determine exact type of interface{} elem
			return nil
		}
	}

	return fmt.Errorf(`invalid value type: "%s"`, expr)
}

func containsWhitespaces(s string) bool {
	r, _ := regexp.Compile(`\s`)
	return r.Match([]byte(s))
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
