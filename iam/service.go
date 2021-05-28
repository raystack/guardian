package iam

import "github.com/odpf/guardian/domain"

// Service handles business logic for identity manager
type Service struct {
	client domain.IAMClient
}

// NewService returns *iam.Service
func NewService(client domain.IAMClient) *Service {
	return &Service{client}
}

// GetUserApproverEmails returns array of approver emails or error if any
func (s *Service) GetUserApproverEmails(user string) ([]string, error) {
	if user == "" {
		return nil, ErrEmptyUserEmailParam
	}

	q := map[string]string{
		"user": user,
	}
	approverEmails, err := s.client.GetUserApproverEmails(q)
	if err != nil {
		return nil, err
	}
	if len(approverEmails) == 0 {
		return nil, ErrEmptyApprovers
	}

	return approverEmails, nil
}
