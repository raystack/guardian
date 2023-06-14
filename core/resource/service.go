package resource

import (
	"context"

	"github.com/imdario/mergo"
	"github.com/raystack/guardian/domain"
	"github.com/raystack/salt/log"
)

const (
	AuditKeyResoruceBulkUpsert  = "resource.bulkUpsert"
	AuditKeyResourceUpdate      = "resource.update"
	AuditKeyResourceDelete      = "resource.delete"
	AuditKeyResourceBatchDelete = "resource.batchDelete"

	ReservedDetailsKeyMetadata = "__metadata"
)

//go:generate mockery --name=repository --exported --with-expecter
type repository interface {
	Find(context.Context, domain.ListResourcesFilter) ([]*domain.Resource, error)
	GetOne(ctx context.Context, id string) (*domain.Resource, error)
	BulkUpsert(context.Context, []*domain.Resource) error
	Update(context.Context, *domain.Resource) error
	Delete(ctx context.Context, id string) error
	BatchDelete(context.Context, []string) error
}

//go:generate mockery --name=auditLogger --exported --with-expecter
type auditLogger interface {
	Log(ctx context.Context, action string, data interface{}) error
}

// Service handles the business logic for resource
type Service struct {
	repo repository

	logger      log.Logger
	auditLogger auditLogger
}

type ServiceDeps struct {
	Repository repository

	Logger      log.Logger
	AuditLogger auditLogger
}

// NewService returns *Service
func NewService(deps ServiceDeps) *Service {
	return &Service{
		deps.Repository,

		deps.Logger,
		deps.AuditLogger,
	}
}

// Find records based on filters
func (s *Service) Find(ctx context.Context, filter domain.ListResourcesFilter) ([]*domain.Resource, error) {
	return s.repo.Find(ctx, filter)
}

func (s *Service) GetOne(ctx context.Context, id string) (*domain.Resource, error) {
	r, err := s.repo.GetOne(ctx, id)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// BulkUpsert inserts or updates records
func (s *Service) BulkUpsert(ctx context.Context, resources []*domain.Resource) error {
	if err := s.repo.BulkUpsert(ctx, resources); err != nil {
		return err
	}

	if err := s.auditLogger.Log(ctx, AuditKeyResoruceBulkUpsert, resources); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return nil
}

// Update updates only details and labels of a resource by ID
func (s *Service) Update(ctx context.Context, r *domain.Resource) error {
	existingResource, err := s.GetOne(ctx, r.ID)
	if err != nil {
		return err
	}

	// Details[ReservedDetailsKeyMetadata] is not allowed to be updated by users
	// value for this field should only set by the provider on FetchResources
	delete(r.Details, ReservedDetailsKeyMetadata)

	if err := mergo.Merge(r, existingResource); err != nil {
		return err
	}

	res := &domain.Resource{
		ID:      r.ID,
		Details: r.Details,
		Labels:  r.Labels,
	}
	if err := s.repo.Update(ctx, res); err != nil {
		return err
	}

	r.UpdatedAt = res.UpdatedAt

	if err := s.auditLogger.Log(ctx, AuditKeyResourceUpdate, r); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return nil
}

func (s *Service) Get(ctx context.Context, ri *domain.ResourceIdentifier) (*domain.Resource, error) {
	var resource *domain.Resource
	if ri.ID != "" {
		if r, err := s.GetOne(ctx, ri.ID); err != nil {
			return nil, err
		} else {
			resource = r
		}
	} else {
		if resources, err := s.Find(ctx, domain.ListResourcesFilter{
			ProviderType: ri.ProviderType,
			ProviderURN:  ri.ProviderURN,
			ResourceType: ri.Type,
			ResourceURN:  ri.URN,
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

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.auditLogger.Log(ctx, AuditKeyResourceDelete, map[string]interface{}{"id": id}); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return nil
}

func (s *Service) BatchDelete(ctx context.Context, ids []string) error {
	if err := s.repo.BatchDelete(ctx, ids); err != nil {
		return err
	}

	if err := s.auditLogger.Log(ctx, AuditKeyResourceBatchDelete, map[string]interface{}{"ids": ids}); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return nil
}
