package postgres

import (
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/store/postgres/model"
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

type approvalRepository struct {
	db *gorm.DB
}

func NewApprovalRepository(db *gorm.DB) *approvalRepository {
	return &approvalRepository{db}
}

func (r *approvalRepository) ListApprovals(conditions *domain.ListApprovalsFilter) ([]*domain.Approval, error) {
	if err := utils.ValidateStruct(conditions); err != nil {
		return nil, err
	}

	db := r.db
	if conditions.CreatedBy != "" {
		db = db.Where("email = ?", conditions.CreatedBy)
	}

	var approverModels []*model.Approver
	if err := db.Find(&approverModels).Error; err != nil {
		return nil, err
	}

	records := []*domain.Approval{}

	if len(approverModels) > 0 {
		var approvalIDs []string
		for _, a := range approverModels {
			approvalIDs = append(approvalIDs, a.ApprovalID)
		}

		db = r.db.Preload("Appeal.Resource")
		db = db.Joins("Appeal")
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
		if err := db.
			Find(&models, approvalIDs).
			Error; err != nil {
			return nil, err
		}

		for _, m := range models {
			appeal, err := m.ToDomain()
			if err != nil {
				return nil, err
			}

			records = append(records, appeal)
		}
	}

	return records, nil
}

func (r *approvalRepository) BulkInsert(approvals []*domain.Approval) error {
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
