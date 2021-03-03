package policy

import (
	"github.com/odpf/guardian/domain"
)

// Service handling the business logics
type Service struct {
	policyRepository domain.PolicyRepository
}

// NewService returns service struct
func NewService(pr domain.PolicyRepository) *Service {
	return &Service{pr}
}

// Create record
func (s *Service) Create(p *domain.Policy) error {
	p.Version = 1
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
