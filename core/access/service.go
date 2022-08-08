package access

import (
	"context"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/log"
)

//go:generate mockery --name=repository --exported --with-expecter
type repository interface {
	List(context.Context, domain.ListAccessesFilter) ([]domain.Access, error)
	GetByID(context.Context, string) (*domain.Access, error)
}

type Service struct {
	repo   repository
	logger log.Logger
}

type ServiceDeps struct {
	Repository repository
	Logger     log.Logger
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		repo:   deps.Repository,
		logger: deps.Logger,
	}
}

func (s *Service) List(ctx context.Context, filter domain.ListAccessesFilter) ([]domain.Access, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Access, error) {
	if id == "" {
		return nil, ErrEmptyIDParam
	}
	return s.repo.GetByID(ctx, id)
}
