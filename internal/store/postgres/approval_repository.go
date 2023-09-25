package postgres

import (
	"context"
	"fmt"

	"github.com/raystack/guardian/core/appeal"
	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/internal/store/postgres/model"
	"github.com/raystack/guardian/utils"
	"gorm.io/gorm"
)

var (
	ApprovalStatusDefaultSort = []string{
		domain.ApprovalStatusPending,
		domain.ApprovalStatusApproved,
		domain.ApprovalStatusRejected,
		domain.ApprovalStatusBlocked,
		domain.ApprovalStatusSkipped,
	}
)

type ApprovalRepository struct {
	store *Store
}

func NewApprovalRepository(db *Store) *ApprovalRepository {
	return &ApprovalRepository{db}
}

func (r *ApprovalRepository) ListApprovals(ctx context.Context, conditions *domain.ListApprovalsFilter) ([]*domain.Approval, error) {
	if err := utils.ValidateStruct(conditions); err != nil {
		return nil, err
	}

	var models []*model.Approval
	err := r.store.Tx(ctx, func(tx *gorm.DB) error {
		tx = tx.Preload("Appeal.Resource")
		tx = tx.Joins("Appeal")
		tx = tx.Joins(`JOIN "approvers" ON "approvals"."id" = "approvers"."approval_id"`)

		if conditions.CreatedBy != "" {
			tx = tx.Where(`"approvers"."email" = ?`, conditions.CreatedBy)
		}
		if conditions.Statuses != nil {
			tx = tx.Where(`"approvals"."status" IN ?`, conditions.Statuses)
		}
		if conditions.AccountID != "" {
			tx = tx.Where(`"Appeal"."account_id" = ?`, conditions.AccountID)
		}

		if len(conditions.AppealStatuses) == 0 {
			tx = tx.Where(`"Appeal"."status" != ?`, domain.AppealStatusCanceled)
		} else {
			tx = tx.Where(`"Appeal"."status" IN ?`, conditions.AppealStatuses)
		}

		if conditions.OrderBy != nil {
			tx = addOrderByClause(tx, conditions.OrderBy, addOrderByClauseOptions{
				statusColumnName: `"approvals"."status"`,
				statusesOrder:    AppealStatusDefaultSort,
			})
		}

		if conditions.Size > 0 {
			tx = tx.Limit(conditions.Size)
		}

		if conditions.Offset > 0 {
			tx = tx.Offset(conditions.Offset)
		}

		return tx.Find(&models).Error
	})
	if err != nil {
		return nil, err
	}

	records := []*domain.Approval{}
	for _, m := range models {
		approval, err := m.ToDomain()
		if err != nil {
			return nil, err
		}

		records = append(records, approval)
	}

	return records, nil
}

func (r *ApprovalRepository) GetApprovalsTotalCount(ctx context.Context, filter *domain.ListApprovalsFilter) (int64, error) {
	db := r.store.db.WithContext(ctx)
	db = applyFilter(db, filter)

	var count int64
	if err := db.Model(&model.Approval{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *ApprovalRepository) BulkInsert(ctx context.Context, approvals []*domain.Approval) error {
	models := []*model.Approval{}
	for _, a := range approvals {
		m := new(model.Approval)
		if err := m.FromDomain(a); err != nil {
			return err
		}
		m.NamespaceID = namespaceFromContext(ctx)
		models = append(models, m)
	}

	return r.store.Tx(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(models).Error; err != nil {
			return err
		}

		for i, m := range models {
			newApproval, err := m.ToDomain()
			if err != nil {
				return err
			}

			*approvals[i] = *newApproval
		}

		return nil
	})
}

func (r *ApprovalRepository) AddApprover(ctx context.Context, approver *domain.Approver) error {
	m := new(model.Approver)
	if err := m.FromDomain(approver); err != nil {
		return fmt.Errorf("parsing approver: %w", err)
	}
	m.NamespaceID = namespaceFromContext(ctx)
	err := r.store.Tx(ctx, func(tx *gorm.DB) error {
		return tx.Create(m).Error
	})
	if err != nil {
		return fmt.Errorf("inserting new approver: %w", err)
	}

	newApprover := m.ToDomain()
	*approver = *newApprover
	return nil
}

func (r *ApprovalRepository) DeleteApprover(ctx context.Context, approvalID, email string) error {
	var result *gorm.DB
	_ = r.store.Tx(ctx, func(tx *gorm.DB) error {
		result = tx.Where("approval_id = ?", approvalID).
			Where("email = ?", email).
			Delete(&model.Approver{})
		return nil
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appeal.ErrApproverNotFound
	}

	return nil
}

func applyFilter(db *gorm.DB, filter *domain.ListApprovalsFilter) *gorm.DB {
	db = db.Joins("Appeal").
		Joins("Appeal.Resource").
		Joins(`JOIN "approvers" ON "approvals"."id" = "approvers"."approval_id"`)

	if filter.Q != "" {
		// NOTE: avoid adding conditions before this grouped where clause.
		// Otherwise, it will be wrapped in parentheses and the query will be invalid.
		db = db.Where(db.
			Where(`"Appeal"."account_id" LIKE ?`, fmt.Sprintf("%%%s%%", filter.Q)).
			Or(`"Appeal"."role" LIKE ?`, fmt.Sprintf("%%%s%%", filter.Q)).
			Or(`"Appeal__Resource"."urn" LIKE ?`, fmt.Sprintf("%%%s%%", filter.Q)),
		)
	}
	if filter.CreatedBy != "" {
		db = db.Where(`"approvers"."email" = ?`, filter.CreatedBy)
	}
	if filter.Statuses != nil {
		db = db.Where(`"approvals"."status" IN ?`, filter.Statuses)
	}
	if filter.AccountID != "" {
		db = db.Where(`"Appeal"."account_id" = ?`, filter.AccountID)
	}
	if filter.AccountTypes != nil {
		db = db.Where(`"Appeal"."account_type" IN ?`, filter.AccountTypes)
	}
	if filter.ResourceTypes != nil {
		db = db.Where(`"Appeal__Resource"."type" IN ?`, filter.ResourceTypes)
	}

	if len(filter.AppealStatuses) == 0 {
		db = db.Where(`"Appeal"."status" != ?`, domain.AppealStatusCanceled)
	} else {
		db = db.Where(`"Appeal"."status" IN ?`, filter.AppealStatuses)
	}

	if filter.OrderBy != nil {
		db = addOrderByClause(db, filter.OrderBy, addOrderByClauseOptions{
			statusColumnName: `"approvals"."status"`,
			statusesOrder:    AppealStatusDefaultSort,
		})
	}

	return db
}
