package postgres

import (
	"fmt"
	"strings"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/store/model"
	"github.com/odpf/guardian/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	if conditions.AccountID != "" {
		db = db.Where("email = ?", conditions.AccountID)
	}

	var approverModels []*model.Approver
	if err := db.Find(&approverModels).Error; err != nil {
		return nil, err
	}

	var approvalIDs []uint
	for _, a := range approverModels {
		approvalIDs = append(approvalIDs, a.ApprovalID)
	}

	db = r.db.Preload("Appeal.Resource")
	db = db.Joins("Appeal")
	if conditions.Statuses != nil {
		db = db.Where(`"approvals"."status" IN ?`, conditions.Statuses)
	}
	db = db.Where(`"Appeal"."status" != ?`, domain.AppealStatusCanceled)

	if conditions.OrderBy != nil {
		var orderByClauses []string
		var vars []interface{}

		for _, orderBy := range conditions.OrderBy {
			if strings.Contains(orderBy, "status") {
				orderByClauses = append(orderByClauses, `ARRAY_POSITION(ARRAY[?], "approvals"."status")`)
				vars = append(vars, ApprovalStatusDefaultSort)
			} else {
				columnOrder := strings.Split(orderBy, ":")
				column := columnOrder[0]
				if utils.ContainsString([]string{"updated_at", "created_at"}, column) {
					if len(columnOrder) == 1 {
						orderByClauses = append(orderByClauses, fmt.Sprintf(`"%s"`, column))
					} else if len(columnOrder) == 2 {
						order := columnOrder[1]
						if utils.ContainsString([]string{"asc", "desc"}, order) {
							orderByClauses = append(orderByClauses, fmt.Sprintf(`"%s" %s`, column, order))
						}
					}
				}
			}
		}

		db = db.Clauses(clause.OrderBy{
			Expression: clause.Expr{
				SQL:                strings.Join(orderByClauses, ", "),
				Vars:               vars,
				WithoutParentheses: true,
			},
		})
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
