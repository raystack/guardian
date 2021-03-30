package resource

import (
	"github.com/imdario/mergo"
	"github.com/odpf/guardian/domain"
)

// Service handles the business logic for resource
type Service struct {
	repo domain.ResourceRepository
}

// NewService returns *Service
func NewService(repo domain.ResourceRepository) *Service {
	return &Service{repo}
}

// Find records based on filters
func (s *Service) Find(filters map[string]interface{}) ([]*domain.Resource, error) {
	return s.repo.Find(filters)
}

// BulkUpsert inserts or updates records
func (s *Service) BulkUpsert(resources []*domain.Resource) error {
	return s.repo.BulkUpsert(resources)
}

// Update updates only details and labels of a resource by ID
func (s *Service) Update(r *domain.Resource) error {
	existingResource, err := s.repo.GetOne(r.ID)
	if err != nil {
		return err
	}
	if existingResource == nil {
		return ErrRecordNotFound
	}

	if err := mergo.Merge(r, existingResource); err != nil {
		return err
	}

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
