package approval

import (
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/model"
	"github.com/odpf/guardian/utils"
	"gorm.io/gorm"
)

type listFilters struct {
	User     string   `mapstructure:"user" validate:"omitempty,required"`
	Statuses []string `mapstructure:"statuses" validate:"omitempty,min=1"`
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db}
}

func (r *repository) ListApprovals(conditions *domain.ListApprovalsFilter) ([]*domain.Approval, error) {
	if err := utils.ValidateStruct(conditions); err != nil {
		return nil, err
	}

	db := r.db
	if conditions.User != "" {
		db = db.Where("email = ?", conditions.User)
	}

	var approverModels []*model.Approver
	if err := db.Find(&approverModels).Error; err != nil {
		return nil, err
	}

	var approvalIDs []uint
	for _, a := range approverModels {
		approvalIDs = append(approvalIDs, a.ApprovalID)
	}

	db = r.db.Joins("Appeal")
	if conditions.Statuses != nil {
		db = db.Where(`"approvals"."status" IN ?`, conditions.Statuses)
	}

	var models []*model.Approval
	if err := db.
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
