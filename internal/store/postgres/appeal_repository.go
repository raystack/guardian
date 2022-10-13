package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres/model"
	"github.com/odpf/guardian/utils"
	"gorm.io/gorm"
)

var (
	AppealStatusDefaultSort = []string{
		domain.AppealStatusPending,
		domain.AppealStatusApproved,
		domain.AppealStatusRejected,
		domain.AppealStatusCanceled,
	}
)

// AppealRepository talks to the store to read or insert data
type AppealRepository struct {
	db     *gorm.DB
	sqldb  *sql.DB
	sqlxdb *sqlx.DB
}

// NewAppealRepository returns repository struct
func NewAppealRepository(db *gorm.DB) *AppealRepository {
	sqldb, _ := db.DB() // TODO: replace gormDB with sql.DB
	sqlxdb := sqlx.NewDb(sqldb, "postgres")
	return &AppealRepository{db, sqldb, sqlxdb}
}

// GetByID returns appeal record by id along with the approvals and the approvers
func (r *AppealRepository) GetByID(id string) (*domain.Appeal, error) {
	m := new(model.Appeal)
	if err := r.db.
		Preload("Approvals", func(db *gorm.DB) *gorm.DB {
			return db.Order("Approvals.index ASC")
		}).
		Preload("Approvals.Approvers").
		Preload("Resource").
		Preload("Grant").
		First(&m, "id = ?", id).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appeal.ErrAppealNotFound
		}
		return nil, err
	}

	a, err := m.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("parsing appeal: %w", err)
	}

	return a, nil
}

func (r *AppealRepository) Find(filters *domain.ListAppealsFilter) ([]*domain.Appeal, error) {
	if err := utils.ValidateStruct(filters); err != nil {
		return nil, err
	}

	var filtersMap map[string]interface{}
	if err := mapstructure.Decode(filters, &filtersMap); err != nil {
		return nil, fmt.Errorf("decoding filters: %w", err)
	}

	columns := []string{}
	columns = append(columns, model.AppealColumns.WithTableName("appeals")...)
	columns = append(columns, model.ResourceColumns...)
	columns = append(columns, model.GrantColumns...)
	queryBuilder := sq.Select(columns...).From("appeals").
		Where(sq.Eq{"appeals.deleted_at": nil}).
		LeftJoin("resources ON resources.id = appeals.resource_id").
		Join("grants ON grants.appeal_id = appeals.id").
		PlaceholderFormat(sq.Dollar)

	if filters.CreatedBy != "" {
		queryBuilder = queryBuilder.Where("created_by = ?", filters.CreatedBy)
	}
	var accounts []string
	if filters.AccountID != "" {
		accounts = append(accounts, filters.AccountID)
	}
	if filters.AccountIDs != nil {
		accounts = append(accounts, filters.AccountIDs...)
	}
	if len(accounts) > 0 {
		queryBuilder = queryBuilder.Where(sq.Eq{"appeals.account_id": accounts})
	}
	if filters.Statuses != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"appeals.status": filters.Statuses})
	}
	if filters.ResourceID != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"appeals.resource_id": filters.ResourceID})
	}
	if filters.Role != "" {
		queryBuilder = queryBuilder.Where("appeals.role = ?", filters.Role)
	}
	if !filters.ExpirationDateLessThan.IsZero() {
		queryBuilder = queryBuilder.Where("options->'expiration_date' < ?", filters.ExpirationDateLessThan)
	}
	if !filters.ExpirationDateGreaterThan.IsZero() {
		queryBuilder = queryBuilder.Where("options->'expiration_date' > ?", filters.ExpirationDateGreaterThan)
	}
	if filters.OrderBy != nil {
		queryBuilder = getOrderByClauses(queryBuilder, filters.OrderBy, addOrderByClauseOptions{
			statusColumnName: `"appeals"."status"`,
			statusesOrder:    AppealStatusDefaultSort,
		})
	}

	if filters.ProviderTypes != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"resources.provider_type": filters.ProviderTypes})
	}
	if filters.ProviderURNs != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"resources.provider_urn": filters.ProviderURNs})
	}
	if filters.ResourceTypes != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"resources.type": filters.ResourceTypes})
	}
	if filters.ResourceURNs != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"resources.urn": filters.ResourceURNs})
	}

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building sql: %w", err)
	}

	var models []*model.Appeal
	if err := r.sqlxdb.Select(&models, sql, args...); err != nil {
		return nil, err
	}

	var result []*domain.Appeal
	for _, m := range models {
		a, err := m.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("parsing appeal: %w", err)
		}
		result = append(result, a)
	}

	return result, nil
}

// BulkUpsert new record to database
func (r *AppealRepository) BulkUpsert(appeals []*domain.Appeal) error {
	queryBuilder := sq.Insert("appeals").Columns(model.AppealColumns...).
		PlaceholderFormat(sq.Dollar)

	var onConflictUpdatedFields []string
	for _, c := range model.AppealColumns {
		switch c {
		case "id", "created_at":
			continue
		default:
			onConflictUpdatedFields = append(onConflictUpdatedFields, fmt.Sprintf("%s = excluded.%s", c, c))
		}
	}
	queryBuilder = queryBuilder.Suffix("ON CONFLICT (id) DO UPDATE SET " + strings.Join(onConflictUpdatedFields, ", "))

	models := []*model.Appeal{}
	for _, a := range appeals {
		if a.ID == "" {
			a.ID = uuid.New().String()
			a.CreatedAt = time.Now()
		}
		a.UpdatedAt = time.Now()
		m := new(model.Appeal)
		if err := m.FromDomain(a); err != nil {
			return fmt.Errorf("parsing appeal: %w", err)
		}
		queryBuilder = queryBuilder.Values(m.Values()...)
		models = append(models, m)
	}

	if _, err := queryBuilder.RunWith(r.sqlxdb).Exec(); err != nil {
		return fmt.Errorf("inserting appeals: %w", err)
	}

	for i, m := range models {
		a, err := m.ToDomain()
		if err != nil {
			return fmt.Errorf("parsing appeal: %w", err)
		}
		*appeals[i] = *a
	}

	return nil
}

// Update an approval step
func (r *AppealRepository) Update(a *domain.Appeal) error {
	m := new(model.Appeal)
	if err := m.FromDomain(a); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Approvals.Approvers").Session(&gorm.Session{FullSaveAssociations: true}).Save(&m).Error; err != nil {
			return err
		}

		newRecord, err := m.ToDomain()
		if err != nil {
			return fmt.Errorf("parsing appeal: %w", err)
		}

		*a = *newRecord

		return nil
	})
}
