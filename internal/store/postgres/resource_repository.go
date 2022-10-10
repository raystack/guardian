package postgres

import (
	"database/sql"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/odpf/guardian/core/resource"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres/model"
	"github.com/odpf/guardian/utils"
	"gorm.io/gorm"
)

// ResourceRepository talks to the store/database to read/insert data
type ResourceRepository struct {
	db    *gorm.DB
	sqldb *sql.DB
}

// NewResourceRepository returns *Repository
func NewResourceRepository(db *gorm.DB) *ResourceRepository {
	sqldb, _ := db.DB() // TODO: replace gormDB with sql.DB
	return &ResourceRepository{db, sqldb}
}

// Find records based on filters
func (r *ResourceRepository) Find(filter domain.ListResourcesFilter) ([]*domain.Resource, error) {
	if err := utils.ValidateStruct(filter); err != nil {
		return nil, err
	}

	queryBuilder := sq.Select("id", "provider_type", "provider_urn", "type", "urn", "name", "details", "labels", "created_at", "updated_at").
		From("resources").
		Where(sq.Eq{"deleted_at": nil})

	if filter.IDs != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"id": filter.IDs})
	}
	if !filter.IsDeleted {
		queryBuilder = queryBuilder.Where(sq.Eq{"is_deleted": filter.IsDeleted})
	}
	if filter.ResourceType != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"type": filter.ResourceType})
	}
	if filter.Name != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"name": filter.Name})
	}
	if filter.ProviderType != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"provider_type": filter.ProviderType})
	}
	if filter.ProviderURN != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"provider_urn": filter.ProviderURN})
	}
	if filter.ResourceURN != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"urn": filter.ResourceURN})
	}
	if filter.ResourceURNs != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"urn": filter.ResourceURNs})
	}
	if filter.ResourceTypes != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"type": filter.ResourceTypes})
	}
	for path, v := range filter.Details {
		pathArr := "{" + strings.Join(strings.Split(path, "."), ",") + "}"
		queryBuilder = queryBuilder.Where(sq.Expr(`details #>> ? = ?`, pathArr, v))
	}

	var models []*model.Resource

	rows, err := queryBuilder.RunWith(r.sqldb).PlaceholderFormat(sq.Dollar).Query()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m model.Resource
		err := rows.Scan(&m.ID, &m.ProviderType, &m.ProviderURN, &m.Type, &m.URN, &m.Name, &m.Details, &m.Labels, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		models = append(models, &m)
	}

	records := []*domain.Resource{}
	for _, m := range models {
		r, err := m.ToDomain()
		if err != nil {
			return nil, err
		}

		records = append(records, r)
	}

	return records, nil
}

// GetOne record by ID
func (r *ResourceRepository) GetOne(id string) (*domain.Resource, error) {
	if id == "" {
		return nil, resource.ErrEmptyIDParam
	}

	var m model.Resource

	err := sq.
		Select("id", "provider_type", "provider_urn", "type", "urn", "name", "details", "labels", "created_at", "updated_at").
		From("resources").
		Where(sq.Eq{"id": id, "deleted_at": nil}).
		RunWith(r.sqldb).
		PlaceholderFormat(sq.Dollar).
		QueryRow().
		Scan(&m.ID, &m.ProviderType, &m.ProviderURN, &m.Type, &m.URN, &m.Name, &m.Details, &m.Labels, &m.CreatedAt, &m.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, resource.ErrRecordNotFound
	}

	if err != nil {
		return nil, err
	}

	return m.ToDomain()
}

// BulkUpsert inserts records if the records are not exist, or updates the records if they are already exist
func (r *ResourceRepository) BulkUpsert(resources []*domain.Resource) error {
	var models []*model.Resource
	for _, r := range resources {
		r.ID = uuid.NewString()
		r.CreatedAt = time.Now()
		r.UpdatedAt = time.Now()

		m := new(model.Resource)
		if err := m.FromDomain(r); err != nil {
			return err
		}

		models = append(models, m)
	}

	if len(models) > 0 {
		tx, err := r.sqldb.Begin()
		if err != nil {
			return err
		}

		for i, m := range models {
			_, err := sq.
				Insert("resources").
				Columns("id", "provider_type", "provider_urn", "type", "urn", "name", "details", "labels", "created_at", "updated_at", "is_deleted").
				Values(m.ID, m.ProviderType, m.ProviderURN, m.Type, m.URN, m.Name, m.Details, m.Labels, m.CreatedAt, m.UpdatedAt, m.IsDeleted).
				Suffix(`
					ON CONFLICT (provider_type, provider_urn, type, urn) 
					DO UPDATE 
					SET name = EXCLUDED.name, 
						details = EXCLUDED.details, 
						updated_at = EXCLUDED.updated_at, 
						is_deleted = EXCLUDED.is_deleted
					`).
				RunWith(tx).
				PlaceholderFormat(sq.Dollar).
				Exec()

			if err != nil {
				tx.Rollback()
				return err
			}

			r, err := m.ToDomain()
			if err != nil {
				return err
			}
			*resources[i] = *r
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// Update record by ID
func (r *ResourceRepository) Update(res *domain.Resource) error {
	if res.ID == "" {
		return resource.ErrEmptyIDParam
	}

	res.UpdatedAt = time.Now()

	m := new(model.Resource)
	if err := m.FromDomain(res); err != nil {
		return err
	}

	queryBuilder := sq.
		Update("resources").
		SetMap(sq.Eq{
			"updated_at": m.UpdatedAt,
		}).
		Where(sq.Eq{"id": m.ID, "deleted_at": nil})

	if m.ProviderType != "" {
		queryBuilder = queryBuilder.Set("provider_type", m.ProviderType)
	}

	if m.ProviderURN != "" {
		queryBuilder = queryBuilder.Set("provider_urn", m.ProviderURN)
	}

	if m.Type != "" {
		queryBuilder = queryBuilder.Set("type", m.Type)
	}

	if m.URN != "" {
		queryBuilder = queryBuilder.Set("urn", m.URN)
	}

	if m.Name != "" {
		queryBuilder = queryBuilder.Set("name", m.Name)
	}

	if m.Details != nil {
		queryBuilder = queryBuilder.Set("details", m.Details)
	}

	if m.Labels != nil {
		queryBuilder = queryBuilder.Set("labels", m.Labels)
	}

	result, err := queryBuilder.RunWith(r.sqldb).PlaceholderFormat(sq.Dollar).Exec()

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return resource.ErrRecordNotFound
	}

	newRecord, err := m.ToDomain()
	if err != nil {
		return err
	}

	*res = *newRecord

	return nil
}

func (r *ResourceRepository) Delete(id string) error {
	if id == "" {
		return resource.ErrEmptyIDParam
	}

	rows, err := sq.Update("resources").
		Set("deleted_at", time.Now()).
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
		return resource.ErrRecordNotFound
	}

	return nil
}

func (r *ResourceRepository) BatchDelete(ids []string) error {
	if ids == nil {
		return resource.ErrEmptyIDParam
	}

	rows, err := sq.Update("resources").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"id": ids, "deleted_at": nil}).
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
		return resource.ErrRecordNotFound
	}

	return nil
}
