package resource

import (
	"context"

	"github.com/imdario/mergo"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/store"
)

type auditLogger interface {
	Log(ctx context.Context, actor, action string, data interface{}) error
}

// Service handles the business logic for resource
type Service struct {
	repo        store.ResourceRepository
	auditLogger auditLogger
}

// NewService returns *Service
func NewService(repo store.ResourceRepository, auditLogger auditLogger) *Service {
	return &Service{repo, auditLogger}
}

// Find records based on filters
func (s *Service) Find(filters map[string]interface{}) ([]*domain.Resource, error) {
	return s.repo.Find(filters)
}

func (s *Service) GetOne(id string) (*domain.Resource, error) {
	r, err := s.repo.GetOne(id)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// BulkUpsert inserts or updates records
func (s *Service) BulkUpsert(resources []*domain.Resource) error {
	return s.repo.BulkUpsert(resources)
}

// Update updates only details and labels of a resource by ID
func (s *Service) Update(ctx context.Context, r *domain.Resource) error {
	existingResource, err := s.GetOne(r.ID)
	if err != nil {
		return err
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
	s.auditLogger.Log(ctx, "// TODO", "resource.update", r)

	return nil
}

func (s *Service) Get(ri *domain.ResourceIdentifier) (*domain.Resource, error) {
	var resource *domain.Resource
	if ri.ID != "" {
		if r, err := s.GetOne(ri.ID); err != nil {
			return nil, err
		} else {
			resource = r
		}
	} else {
		if resources, err := s.Find(map[string]interface{}{
			"provider_type": ri.ProviderType,
			"provider_urn":  ri.ProviderURN,
			"type":          ri.Type,
			"urn":           ri.URN,
		}); err != nil {
			return nil, err
		} else {
			if len(resources) == 0 {
				return nil, ErrRecordNotFound
			} else {
				resource = resources[0]
			}
		}
	}
	return resource, nil
}
