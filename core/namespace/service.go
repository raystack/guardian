package namespace

import (
	"context"

	"github.com/raystack/guardian/domain"
)

type Repository interface {
	BulkUpsert(ctx context.Context, namespaces []*domain.Namespace) error
	List(ctx context.Context) ([]*domain.Namespace, error)
	GetOne(ctx context.Context, id string) (*domain.Namespace, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, flt domain.NamespaceFilter) ([]*domain.Namespace, error) {
	return s.repository.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Namespace, error) {
	return s.repository.GetOne(ctx, id)
}

func (s *Service) Create(ctx context.Context, namespaces *domain.Namespace) error {
	return s.repository.BulkUpsert(ctx, []*domain.Namespace{namespaces})
}

func (s *Service) Update(ctx context.Context, namespaces *domain.Namespace) error {
	return s.repository.BulkUpsert(ctx, []*domain.Namespace{namespaces})
}
