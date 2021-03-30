package policy

import (
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/model"
	"gorm.io/gorm"
)

// Repository talks to the store to read or insert data
type Repository struct {
	db *gorm.DB
}

// NewRepository returns repository struct
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db}
}

// Create new record to database
func (r *Repository) Create(p *domain.Policy) error {
	m := new(model.Policy)
	if err := m.FromDomain(p); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if result := tx.Create(m); result.Error != nil {
			return result.Error
		}

		newPolicy, err := m.ToDomain()
		if err != nil {
			return err
		}

		*p = *newPolicy

		return nil
	})
}

// Find records based on filters
func (r *Repository) Find() ([]*domain.Policy, error) {
	policies := []*domain.Policy{}

	var models []*model.Policy
	latestPoliciesQuery := r.db.Model(&model.Policy{}).Select("id, max(version)").Group("id")
	if err := r.db.Where("(id,version) IN (?)", latestPoliciesQuery).Find(&models).Error; err != nil {
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
func (r *Repository) GetOne(id string, version uint) (*domain.Policy, error) {
	m := &model.Policy{}
	condition := "id = ?"
	args := []interface{}{id}
	if version != 0 {
		condition = "id = ? AND version = ?"
		args = append(args, version)
	}

	conds := append([]interface{}{condition}, args...)
	if err := r.db.Order("version desc").First(m, conds...).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	p, err := m.ToDomain()
	if err != nil {
		return nil, err
	}

	return p, nil
}
