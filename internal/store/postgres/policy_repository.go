package postgres

import (
	"context"
	"fmt"

	"github.com/raystack/guardian/core/policy"
	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/internal/store/postgres/model"
	"gorm.io/gorm"
)

// PolicyRepository talks to the store to read or insert data
type PolicyRepository struct {
	db *gorm.DB
}

// NewPolicyRepository returns repository struct
func NewPolicyRepository(db *gorm.DB) *PolicyRepository {
	return &PolicyRepository{db}
}

// Create new record to database
func (r *PolicyRepository) Create(ctx context.Context, p *domain.Policy) error {
	m := new(model.Policy)
	if err := m.FromDomain(p); err != nil {
		return fmt.Errorf("serializing policy: %w", err)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if result := tx.Create(m); result.Error != nil {
			return result.Error
		}

		newPolicy, err := m.ToDomain()
		if err != nil {
			return fmt.Errorf("deserializing policy: %w", err)
		}

		*p = *newPolicy

		return nil
	})
}

// Find records based on filters
func (r *PolicyRepository) Find(ctx context.Context) ([]*domain.Policy, error) {
	policies := []*domain.Policy{}

	var models []*model.Policy
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	for _, m := range models {
		p, err := m.ToDomain()
		if err != nil {
			return nil, err
		}

		policies = append(policies, p)
	}

	return policies, nil
}

// GetOne returns a policy record based on the id and version params.
// If version is 0, the latest version will be returned
func (r *PolicyRepository) GetOne(ctx context.Context, id string, version uint) (*domain.Policy, error) {
	m := &model.Policy{}
	condition := "id = ?"
	args := []interface{}{id}
	if version != 0 {
		condition = "id = ? AND version = ?"
		args = append(args, version)
	}

	conds := append([]interface{}{condition}, args...)
	if err := r.db.WithContext(ctx).Order("version desc").First(m, conds...).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, policy.ErrPolicyNotFound
		}
		return nil, err
	}

	p, err := m.ToDomain()
	if err != nil {
		return nil, err
	}

	return p, nil
}
