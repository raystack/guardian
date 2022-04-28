//go:generate mockery --name=repository --exported
//go:generate mockery --name=auditLogger --exported
//go:generate mockery --name=helper --exported

package resource

import (
	"context"
	"fmt"

	"github.com/imdario/mergo"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/store"
	"github.com/odpf/salt/log"
)

const (
	AuditKeyResoruceBulkUpsert = "resource.bulkUpsert"
	AuditKeyResourceUpdate     = "resource.update"
)

type repository interface {
	store.ResourceRepository
}

type auditLogger interface {
	Log(ctx context.Context, actor, action string, data interface{}) error
}

type helper interface {
	GetAuthenticatedUser(context.Context) (string, error)
}

// Service handles the business logic for resource
type Service struct {
	repo repository

	logger      log.Logger
	auditLogger auditLogger
	helper      helper
}

type ServiceOptions struct {
	Repository repository

	Logger      log.Logger
	AuditLogger auditLogger
	Helper      helper
}

// NewService returns *Service
func NewService(opts ServiceOptions) *Service {
	return &Service{
		opts.Repository,

		opts.Logger,
		opts.AuditLogger,
		opts.Helper,
	}
}

// Find records based on filters
func (s *Service) Find(_ context.Context, filters map[string]interface{}) ([]*domain.Resource, error) {
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
func (s *Service) BulkUpsert(ctx context.Context, resources []*domain.Resource) error {
	if err := s.repo.BulkUpsert(resources); err != nil {
		return err
	}

	actor, err := s.helper.GetAuthenticatedUser(ctx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("unable to get authenticated user: %s", err))
	}
	if err := s.auditLogger.Log(ctx, actor, AuditKeyResoruceBulkUpsert, map[string]interface{}{
		"created_resource_ids": []string{}, // TODO: add inserted ids
		"updated_resource_ids": []string{}, // TODO: add modified ids
		"removed_resource_ids": []string{}, // TODO: add removed ids
	}); err != nil {
		s.logger.Error(fmt.Sprintf("failed to record audit log: %s", err))
	}

	return nil
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

	actor, err := s.helper.GetAuthenticatedUser(ctx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("unable to get authenticated user: %s", err))
	}
	if err := s.auditLogger.Log(ctx, actor, AuditKeyResourceUpdate, r); err != nil {
		s.logger.Error(fmt.Sprintf("failed to record audit log: %s", err))
	}

	return nil
}

func (s *Service) Get(ctx context.Context, ri *domain.ResourceIdentifier) (*domain.Resource, error) {
	var resource *domain.Resource
	if ri.ID != "" {
		if r, err := s.GetOne(ri.ID); err != nil {
			return nil, err
		} else {
			resource = r
		}
	} else {
		if resources, err := s.Find(ctx, map[string]interface{}{
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
