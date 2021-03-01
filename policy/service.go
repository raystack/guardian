package policy

import "github.com/odpf/guardian/domain"

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
	return s.policyRepository.Create(p)
}

// Find records
func (s *Service) Find() ([]*domain.Policy, error) {
	return s.policyRepository.Find()
}
