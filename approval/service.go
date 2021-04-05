package approval

import "github.com/odpf/guardian/domain"

type service struct {
	repo domain.ApprovalRepository
}

func NewService(ar domain.ApprovalRepository) *service {
	return &service{ar}
}

func (s *service) BulkInsert(approvals []*domain.Approval) error {
	return s.repo.BulkInsert(approvals)
}
