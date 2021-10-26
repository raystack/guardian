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

// GetUser fetches user details from external IAM
func (s *Service) GetUser(id string) (interface{}, error) {
	if id == "" {
		return nil, ErrEmptyUserEmailParam
	}

	return s.client.GetUser(id)
}
