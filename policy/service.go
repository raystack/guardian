package policy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/domain"
)

// Service handling the business logics
type Service struct {
	validator        *validator.Validate
	policyRepository domain.PolicyRepository
}

// NewService returns service struct
func NewService(v *validator.Validate, pr domain.PolicyRepository) *Service {
	return &Service{v, pr}
}

// Create record
func (s *Service) Create(p *domain.Policy) error {
	p.Version = 1

	if err := s.validatePolicy(p); err != nil {
		return fmt.Errorf("policy validation: %v", err)
	}

	return s.policyRepository.Create(p)
}

// Find records
func (s *Service) Find() ([]*domain.Policy, error) {
	return s.policyRepository.Find()
}

// GetOne record
func (s *Service) GetOne(id string, version uint) (*domain.Policy, error) {
	return s.policyRepository.GetOne(id, version)
}

// Update a record
func (s *Service) Update(p *domain.Policy) error {
	if p.ID == "" {
		return ErrEmptyIDParam
	}

	if err := s.validatePolicy(p, "Version"); err != nil {
		return fmt.Errorf("policy validation: %v", err)
	}

	latestPolicy, err := s.policyRepository.GetOne(p.ID, p.Version)
	if err != nil {
		return err
	}
	if latestPolicy == nil {
		return ErrPolicyDoesNotExists
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

	return nil
}

func (s *Service) validateSteps(steps []*domain.Step) error {
	validVariables := []string{
		domain.ApproversKeyResource,
		domain.ApproversKeyUserApprovers,
	}

	for i, step := range steps {
		if containsWhitespaces(step.Name) {
			return ErrStepNameContainsWhitespaces
		}

		// validate approvers
		if strings.HasPrefix(step.Approvers, "$") {
			isValidVariable := false
			for _, v := range validVariables {
				if strings.HasPrefix(step.Approvers, v) {
					isValidVariable = true
					break
				}
			}

			if !isValidVariable {
				return fmt.Errorf("%v: %v", ErrInvalidApprovers, step.Approvers)
			}
		}

		// validate dependencies
		for _, d := range step.Dependencies {
			isDependencyExists := false
			for j := 0; j < i; j++ {
				if steps[j].Name == d {
					isDependencyExists = true
					break
				}
			}

			if !isDependencyExists {
				return fmt.Errorf("%v: %v", ErrStepDependencyDoesNotExists, d)
			}
		}
	}

	return nil
}

func containsWhitespaces(s string) bool {
	r, _ := regexp.Compile(`\s`)
	return r.Match([]byte(s))
}
