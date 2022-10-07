package postgres

import (
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/odpf/guardian/core/policy"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres/model"
	"gorm.io/gorm"
)

// PolicyRepository talks to the store to read or insert data
type PolicyRepository struct {
	db    *gorm.DB
	sqldb *sql.DB
}

// NewPolicyRepository returns repository struct
func NewPolicyRepository(db *gorm.DB) *PolicyRepository {
	sqldb, _ := db.DB() // TODO: replace gormDB with sql.DB
	return &PolicyRepository{db, sqldb}
}

// Create new record to database
func (r *PolicyRepository) Create(p *domain.Policy) error {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	m := new(model.Policy)
	if err := m.FromDomain(p); err != nil {
		return fmt.Errorf("serializing policy: %w", err)
	}

	_, err := sq.Insert("policies").
		Columns("id", "version", "description", "steps", "appeal_config", "labels", "requirements", "iam", "created_at", "updated_at").
		Values(m.ID, m.Version, m.Description, m.Steps, m.AppealConfig, m.Labels, m.Requirements, m.IAM, m.CreatedAt, m.UpdatedAt).
		RunWith(r.sqldb).
		PlaceholderFormat(sq.Dollar).
		Exec()
	return err
}

// Find records based on filters
func (r *PolicyRepository) Find() ([]*domain.Policy, error) {
	policies := []*domain.Policy{}

	var models []*model.Policy
	if err := r.db.Find(&models).Error; err != nil {
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
func (r *PolicyRepository) GetOne(id string, version uint) (*domain.Policy, error) {
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
