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
	query := `
	SELECT
		appeals.id,
		appeals.resource_id,
		appeals.policy_id,
		appeals.policy_version,
		appeals.status,
		appeals.account_id,
		appeals.account_type,
		appeals.created_by,
		appeals.creator,
		appeals.role,
		appeals.permissions,
		appeals.options,
		appeals.labels,
		appeals.details,
		appeals.created_at,
		appeals.updated_at,

		COALESCE(grants.id, NULL) AS "grant.id",
		COALESCE(grants.status, '') AS "grant.status",
		COALESCE(grants.status_in_provider, '') AS "grant.status_in_provider",
		COALESCE(grants.account_id, '') AS "grant.account_id",
		COALESCE(grants.account_type, '') AS "grant.account_type",
		COALESCE(grants.resource_id, '00000000-0000-0000-0000-000000000000') AS "grant.resource_id",
		COALESCE(grants.role, '') AS "grant.role",
		COALESCE(grants.permissions, '{}') AS "grant.permissions",
		COALESCE(grants.is_permanent, FALSE) AS "grant.is_permanent",
		COALESCE(grants.expiration_date, NULL) AS "grant.expiration_date",
		COALESCE(grants.appeal_id, '00000000-0000-0000-0000-000000000000') AS "grant.appeal_id",
		COALESCE(grants.source, '') AS "grant.source",
		COALESCE(grants.revoked_by, '') AS "grant.revoked_by",
		COALESCE(grants.revoked_at, NULL) AS "grant.revoked_at",
		COALESCE(grants.revoke_reason, '') AS "grant.revoke_reason",
		COALESCE(grants.owner, '') AS "grant.owner",
		COALESCE(grants.created_at, '0001-01-01 00:00:00+07') AS "grant.created_at",
		COALESCE(grants.updated_at, '0001-01-01 00:00:00+07') AS "grant.updated_at",

		COALESCE(resources.id, NULL) AS "resource.id",
		COALESCE(resources.provider_type, '') AS "resource.provider_type",
		COALESCE(resources.provider_urn, '') AS "resource.provider_urn",
		COALESCE(resources.type, '') AS "resource.type",
		COALESCE(resources.urn, '') AS "resource.urn",
		COALESCE(resources.name, '') AS "resource.name",
		COALESCE(resources.details, 'null') AS "resource.details",
		COALESCE(resources.labels, 'null') AS "resource.labels",
		COALESCE(resources.created_at, '0001-01-01 00:00:00+07') AS "resource.created_at",
		COALESCE(resources.updated_at, '0001-01-01 00:00:00+07') AS "resource.updated_at",
		COALESCE(resources.is_deleted, FALSE) AS "resource.is_deleted"
	FROM
		appeals
		LEFT JOIN resources ON appeals.resource_id = resources.id
		LEFT JOIN grants ON grants.appeal_id = appeals.id
	WHERE appeals.id = $1 LIMIT 1;
	`

	var m model.Appeal
	err := r.sqlxdb.Get(&m, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appeal.ErrAppealNotFound
		}
		return nil, fmt.Errorf("querying appeal: %w", err)
	}
	if m.Grant.ID == uuid.Nil {
		m.Grant = nil
	}
	if m.Resource.ID == uuid.Nil {
		m.Resource = nil
	}

	approvalColumns := []string{}
	approvalColumns = append(approvalColumns, model.ApprovalColumns.WithTableAliases("approvals", "approval")...)
	approvalColumns = append(approvalColumns, model.ApproverColumns.WithTableAliases("approvers", "approver")...)
	query, args, err := sq.Select(approvalColumns...).From("approvals").
		Where(sq.Eq{"approvals.appeal_id": id}).
		LeftJoin("approvers ON approvers.approval_id = approvals.id").
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("building query: %w", err)
	}
	type approvalWithApprover struct {
		model.Approval `db:"approval"`
		model.Approver `db:"approver"`
	}
	var approvalWithApprovers []*approvalWithApprover
	if err := r.sqlxdb.Select(&approvalWithApprovers, query, args...); err != nil {
		return nil, fmt.Errorf("querying approvals: %w", err)
	}
	fmt.Printf("len(approvalWithApprovers): %v\n", len(approvalWithApprovers))
	approvalsMap := make(map[string]*model.Approval)
	for _, aa := range approvalWithApprovers {
		approval, ok := approvalsMap[aa.Approval.ID.String()]
		if !ok {
			approval = &aa.Approval
		}
		approval.Approvers = append(approval.Approvers, aa.Approver)
	}
	for _, approval := range approvalsMap {
		m.Approvals = append(m.Approvals, approval)
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
	columns = append(columns, model.ResourceColumns.WithTableAliases("resources", "resource")...)
	columns = append(columns, model.GrantColumns.WithTableAliases("grants", "grant")...)
	queryBuilder := sq.Select(columns...).From("appeals").
		Where(sq.Eq{"appeals.deleted_at": nil}).
		LeftJoin("resources ON resources.id = appeals.resource_id").
		LeftJoin("grants ON grants.appeal_id = appeals.id").
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
