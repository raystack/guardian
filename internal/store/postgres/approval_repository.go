package postgres

import (
	"fmt"

	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres/model"
	"github.com/odpf/guardian/utils"
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
	db *gorm.DB
}

func NewApprovalRepository(db *gorm.DB) *ApprovalRepository {
	return &ApprovalRepository{db}
}

func (r *ApprovalRepository) ListApprovals(conditions *domain.ListApprovalsFilter) ([]*domain.Approval, error) {
	if err := utils.ValidateStruct(conditions); err != nil {
		return nil, err
	}

	records := []*domain.Approval{}

	db := r.db.Preload("Appeal.Resource")
	db = db.Joins("Appeal")
	db = db.Joins(`JOIN "approvers" ON "approvals"."id" = "approvers"."approval_id"`)

	if conditions.CreatedBy != "" {
		db = db.Where(`"approvers"."email" = ?`, conditions.CreatedBy)
	}
	if conditions.Statuses != nil {
		db = db.Where(`"approvals"."status" IN ?`, conditions.Statuses)
	}
	if conditions.AccountID != "" {
		db = db.Where(`"Appeal"."account_id" = ?`, conditions.AccountID)
	}
	db = db.Where(`"Appeal"."status" != ?`, domain.AppealStatusCanceled)

	if conditions.OrderBy != nil {
		db = addOrderByClause(db, conditions.OrderBy, addOrderByClauseOptions{
			statusColumnName: `"approvals"."status"`,
		})
	}

	var models []*model.Approval
	if err := db.Find(&models).Error; err != nil {
		return nil, err
	}

	for _, m := range models {
		approval, err := m.ToDomain()
		if err != nil {
			return nil, err
		}

		records = append(records, approval)
	}

	return records, nil
}

func (r *ApprovalRepository) BulkInsert(approvals []*domain.Approval) error {
	models := []*model.Approval{}
	for _, a := range approvals {
		m := new(model.Approval)
		if err := m.FromDomain(a); err != nil {
			return err
		}

		models = append(models, m)
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
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

func (r *ApprovalRepository) AddApprover(approver *domain.Approver) error {
	m := new(model.Approver)
	if err := m.FromDomain(approver); err != nil {
		return fmt.Errorf("parsing approver: %w", err)
	}

	result := r.db.Create(m)
	if result.Error != nil {
		return fmt.Errorf("inserting new approver: %w", result.Error)
	}

	newApprover := m.ToDomain()
	*approver = *newApprover
	return nil
}

func (r *ApprovalRepository) DeleteApprover(approvalID, email string) error {
	result := r.db.
		Where("approval_id = ?", approvalID).
		Where("email = ?", email).
		Delete(&model.Approver{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appeal.ErrApproverNotFound
	}

	return nil
}
