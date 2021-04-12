package approval

import (
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/model"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db}
}

func (r *repository) GetPendingApprovals(approverEmail string) ([]*domain.Approval, error) {
	var approverModels []*model.Approver
	if err := r.db.Find(&approverModels, "email = ?", approverEmail).Error; err != nil {
		return nil, err
	}

	var approvalIDs []uint
	for _, a := range approverModels {
		approvalIDs = append(approvalIDs, a.ApprovalID)
	}

	earliestPendingApprovalQuery := r.db.Model(&model.Approval{}).
		Select(`appeal_id, min("index")`).
		Where("status = ?", domain.ApprovalStatusPending).
		Group("appeal_id")
	var models []*model.Approval
	if err := r.db.
		Preload("Appeal").
		Where(`("appeal_id","index") IN (?)`, earliestPendingApprovalQuery).
		Find(&models, approvalIDs).
		Error; err != nil {
		return nil, err
	}

	records := []*domain.Approval{}
	for _, m := range models {
		appeal, err := m.ToDomain()
		if err != nil {
			return nil, err
		}

		records = append(records, appeal)
	}

	return records, nil
}

func (r *repository) BulkInsert(approvals []*domain.Approval) error {
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
