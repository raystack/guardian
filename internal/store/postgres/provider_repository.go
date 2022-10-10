package postgres

import (
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres/model"
	"gorm.io/gorm"
)

// ProviderRepository talks to the store to read or insert data
type ProviderRepository struct {
	db    *gorm.DB
	sqldb *sql.DB
}

// NewProviderRepository returns repository struct
func NewProviderRepository(db *gorm.DB) *ProviderRepository {
	sqldb, _ := db.DB() // TODO: replace gormDB with sql.DB
	return &ProviderRepository{db, sqldb}
}

// Create new record to database
func (r *ProviderRepository) Create(p *domain.Provider) error {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	p.ID = uuid.NewString()

	m := new(model.Provider)
	if err := m.FromDomain(p); err != nil {
		return err
	}

	_, err := sq.
		Insert("providers").
		Columns("id", "type", "urn", "config", "created_at", "updated_at").
		Values(m.ID.String(), m.Type, m.URN, m.Config, m.CreatedAt, m.UpdatedAt).
		RunWith(r.sqldb).
		PlaceholderFormat(sq.Dollar).
		Exec()

	if err != nil {
		return err
	}

	newProvider, err := m.ToDomain()
	if err != nil {
		return err
	}

	*p = *newProvider

	return nil
}

// Find records based on filters
func (r *ProviderRepository) Find() ([]*domain.Provider, error) {
	rows, err := sq.
		Select("id", "type", "urn", "config", "created_at", "updated_at").
		From("providers").
		Where(sq.Eq{"deleted_at": nil}).
		RunWith(r.sqldb).
		Query()

	if err != nil {
		return nil, err
	}

	providers := []*domain.Provider{}

	for rows.Next() {
		m := &model.Provider{}
		if err := rows.Scan(&m.ID, &m.Type, &m.URN, &m.Config, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}

		p, err := m.ToDomain()
		if err != nil {
			return nil, err
		}

		providers = append(providers, p)
	}
	return providers, nil
}

// GetByID record by ID
func (r *ProviderRepository) GetByID(id string) (*domain.Provider, error) {
	if id == "" {
		return nil, provider.ErrEmptyIDParam
	}

	m := &model.Provider{}
	err := sq.
		Select("id", "type", "urn", "config", "created_at", "updated_at").
		From("providers").
		Where(sq.Eq{"id": id, "deleted_at": nil}).
		RunWith(r.sqldb).
		PlaceholderFormat(sq.Dollar).
		QueryRow().
		Scan(&m.ID, &m.Type, &m.URN, &m.Config, &m.CreatedAt, &m.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, provider.ErrRecordNotFound
	}

	if err != nil {
		return nil, err
	}

	return m.ToDomain()
}

func (r *ProviderRepository) GetTypes() ([]domain.ProviderType, error) {
	var results []struct {
		ProviderType string
		ResourceType string
	}

	rows, err := r.sqldb.Query("SELECT DISTINCT provider_type, type AS resource_type FROM resources")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var result struct {
			ProviderType string
			ResourceType string
		}
		err := rows.Scan(&result.ProviderType, &result.ResourceType)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, errors.New("no provider types found")
	}

	providerTypesMap := make(map[string][]string)
	for _, res := range results {
		if val, ok := providerTypesMap[res.ProviderType]; ok {
			providerTypesMap[res.ProviderType] = append(val, res.ResourceType)
		} else {
			providerTypesMap[res.ProviderType] = []string{res.ResourceType}
		}
	}

	var providerTypes []domain.ProviderType
	for providerType, resourceTypes := range providerTypesMap {
		providerTypes = append(providerTypes, domain.ProviderType{
			Name:          providerType,
			ResourceTypes: resourceTypes,
		})
	}

	return providerTypes, nil
}

// GetOne returns provider by type and urn
func (r *ProviderRepository) GetOne(pType, urn string) (*domain.Provider, error) {
	if pType == "" {
		return nil, provider.ErrEmptyProviderType
	}
	if urn == "" {
		return nil, provider.ErrEmptyProviderURN
	}

	m := &model.Provider{}

	err := sq.
		Select("id", "type", "urn", "config", "created_at", "updated_at").
		From("providers").
		Where(sq.Eq{"type": pType, "urn": urn, "deleted_at": nil}).
		RunWith(r.sqldb).
		PlaceholderFormat(sq.Dollar).
		QueryRow().
		Scan(&m.ID, &m.Type, &m.URN, &m.Config, &m.CreatedAt, &m.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, provider.ErrRecordNotFound
	}

	if err != nil {
		return nil, err
	}

	return m.ToDomain()
}

// Update record by ID
func (r *ProviderRepository) Update(p *domain.Provider) error {
	if p.ID == "" {
		return provider.ErrEmptyIDParam
	}

	p.UpdatedAt = time.Now()

	m := new(model.Provider)
	if err := m.FromDomain(p); err != nil {
		return err
	}

	_, err := sq.
		Update("providers").
		SetMap(map[string]interface{}{
			"type":       m.Type,
			"urn":        m.URN,
			"config":     m.Config,
			"updated_at": m.UpdatedAt,
		}).
		Where(sq.Eq{"id": m.ID, "deleted_at": nil}).
		RunWith(r.sqldb).
		PlaceholderFormat(sq.Dollar).
		Exec()

	if err != nil {
		return err
	}

	newRecord, err := m.ToDomain()
	if err != nil {
		return err
	}

	*p = *newRecord

	return nil
}

// Delete record by ID
func (r *ProviderRepository) Delete(id string) error {
	if id == "" {
		return provider.ErrEmptyIDParam
	}

	rows, err := sq.Delete("providers").
		Where(sq.Eq{"id": id, "deleted_at": nil}).
		RunWith(r.sqldb).
		PlaceholderFormat(sq.Dollar).
		Exec()

	if err != nil {
		return err
	}

	affected, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return provider.ErrRecordNotFound
	}

	return nil
}
