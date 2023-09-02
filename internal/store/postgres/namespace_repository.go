package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/internal/store/postgres/model"
	"gorm.io/gorm"
)

type NamespaceRepository struct {
	store *Store
}

func NewNamespaceRepository(store *Store) *NamespaceRepository {
	return &NamespaceRepository{
		store: store,
	}
}

func (r *NamespaceRepository) List(ctx context.Context) ([]*domain.Namespace, error) {
	var namespaces []*model.Namespace

	if err := r.store.Tx(ctx, func(tx *gorm.DB) error {
		return tx.Find(&namespaces).Error
	}); err != nil {
		return nil, err
	}

	var results []*domain.Namespace
	for _, activity := range namespaces {
		a := &domain.Namespace{}
		if err := activity.ToDomain(a); err != nil {
			return nil, fmt.Errorf("failed to convert model %q to domain: %w", activity.ID, err)
		}
		results = append(results, a)
	}
	return results, nil
}

func (r *NamespaceRepository) GetOne(ctx context.Context, id string) (*domain.Namespace, error) {
	var m model.Namespace
	if err := r.store.Tx(ctx, func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).First(&m).Error
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("namespace not found")
		}
		return nil, err
	}
	a := &domain.Namespace{}
	if err := m.ToDomain(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (r *NamespaceRepository) BulkUpsert(ctx context.Context, namespaces []*domain.Namespace) error {
	if len(namespaces) == 0 {
		return nil
	}

	namespaceModels := make([]*model.Namespace, len(namespaces))
	for i, a := range namespaces {
		namespaceModels[i] = &model.Namespace{}
		if err := namespaceModels[i].FromDomain(a); err != nil {
			return fmt.Errorf("failed to convert domain to model: %w", err)
		}
	}

	return r.store.Tx(ctx, func(tx *gorm.DB) error {
		if err := tx.Save(namespaceModels).Error; err != nil {
			return fmt.Errorf("failed to upsert provider namespaces: %w", err)
		}
		return nil
	})
}
