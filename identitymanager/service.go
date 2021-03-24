package identitymanager

import "github.com/odpf/guardian/domain"

// Service handles business logic for identity manager
type Service struct {
	client domain.IdentityManagerClient
}

// NewService returns *identitymanager.Service
func NewService(client domain.IdentityManagerClient) *Service {
	return &Service{client}
}

// GetUserApproverEmails returns array of approver emails or error if any
func (s *Service) GetUserApproverEmails(userEmail string) ([]string, error) {
	if userEmail == "" {
		return nil, ErrEmptyUserEmailParam
	}

	q := map[string]string{
		"email": userEmail,
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
