package activity

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/pkg/log"
)

//go:generate mockery --name=repository --exported --with-expecter
type repository interface {
	GetOne(context.Context, string) (*domain.Activity, error)
	BulkUpsert(context.Context, []*domain.Activity) error
	Find(context.Context, domain.ListProviderActivitiesFilter) ([]*domain.Activity, error)
}

//go:generate mockery --name=providerService --exported --with-expecter
type providerService interface {
	ImportActivities(context.Context, domain.ListActivitiesFilter) ([]*domain.Activity, error)
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
		repo:            deps.Repository,
		providerService: deps.ProviderService,
		validator:       deps.Validator,
		logger:          deps.Logger,
		auditLogger:     deps.AuditLogger,
	}
}

func (s *Service) GetOne(ctx context.Context, id string) (*domain.Activity, error) {
	return s.repo.GetOne(ctx, id)
}

func (s *Service) Find(ctx context.Context, filter domain.ListProviderActivitiesFilter) ([]*domain.Activity, error) {
	return s.repo.Find(ctx, filter)
}

func (s *Service) Import(ctx context.Context, filter domain.ListActivitiesFilter) ([]*domain.Activity, error) {
	activities, err := s.providerService.ImportActivities(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err := s.repo.BulkUpsert(ctx, activities); err != nil {
		return nil, fmt.Errorf("inserting activities to db: %w", err)
	}

	return activities, nil
}
