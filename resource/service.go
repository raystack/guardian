package resource

import "github.com/odpf/guardian/domain"

// Service handles the business logic for resource
type Service struct {
	repo domain.ResourceRepository
}

// NewService returns *Service
func NewService(repo domain.ResourceRepository) *Service {
	return &Service{repo}
}

// Find records
func (s *Service) Find() ([]*domain.Resource, error) {
	return s.repo.Find()
}

// BulkUpsert inserts or updates records
func (s *Service) BulkUpsert(resources []*domain.Resource) error {
	return s.repo.BulkUpsert(resources)
}

// Update updates only details and labels of a resource by ID
func (s *Service) Update(r *domain.Resource) error {
	res := &domain.Resource{
		ID:      r.ID,
		Details: r.Details,
		Labels:  r.Labels,
	}
	if err := s.repo.Update(res); err != nil {
		return err
	}

	r.UpdatedAt = res.UpdatedAt
	return nil
}
