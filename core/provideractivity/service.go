package provideractivity

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/log"
)

//go:generate mockery --name=repository --exported --with-expecter
type repository interface {
	GetOne(context.Context, string) (*domain.ProviderActivity, error)
	BulkInsert(context.Context, []*domain.ProviderActivity) error
	Find(context.Context, domain.ListProviderActivitiesFilter) ([]*domain.ProviderActivity, error)
}

//go:generate mockery --name=providerService --exported --with-expecter
type providerService interface {
	ImportActivities(context.Context, domain.ImportActivitiesFilter) ([]*domain.ProviderActivity, error)
}

//go:generate mockery --name=auditLogger --exported --with-expecter
type auditLogger interface {
	Log(ctx context.Context, action string, data interface{}) error
}

type ServiceDeps struct {
	Repository      repository
	ProviderService providerService

	Validator   *validator.Validate
	Logger      log.Logger
	AuditLogger auditLogger
}

type Service struct {
	repo            repository
	providerService providerService
	validator       *validator.Validate
	logger          log.Logger
	auditLogger     auditLogger
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		repo:        deps.Repository,
		validator:   deps.Validator,
		logger:      deps.Logger,
		auditLogger: deps.AuditLogger,
	}
}

func (s *Service) BulkInsert(ctx context.Context, activities []*domain.ProviderActivity) error {
	return s.repo.BulkInsert(ctx, activities)
}

func (s *Service) GetOne(ctx context.Context, id string) (*domain.ProviderActivity, error) {
	return s.repo.GetOne(ctx, id)
}

func (s *Service) Find(ctx context.Context, filter domain.ListProviderActivitiesFilter) ([]*domain.ProviderActivity, error) {
	return s.repo.Find(ctx, filter)
}

func (s *Service) Import(ctx context.Context, filter domain.ImportActivitiesFilter) ([]*domain.ProviderActivity, error) {
	activities, err := s.providerService.ImportActivities(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err := s.repo.BulkInsert(ctx, activities); err != nil {
		return nil, fmt.Errorf("inserting activities to db: %w", err)
	}

	return activities, nil
}
