package approval_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/approval"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository    *mocks.ApprovalRepository
	mockPolicyService *mocks.PolicyService

	service domain.ApprovalService
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(mocks.ApprovalRepository)
	s.mockPolicyService = new(mocks.PolicyService)

	s.service = approval.NewService(s.mockRepository, s.mockPolicyService)
}

func (s *ServiceTestSuite) TestGetPendingApprovals() {
	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("repository error")
		s.mockRepository.On("GetPendingApprovals", mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.GetPendingApprovals("user@email.com")

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})
}

func (s *ServiceTestSuite) TestBulkInsert() {
	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("repository error")
		s.mockRepository.On("BulkInsert", mock.Anything).Return(expectedError).Once()

		actualError := s.service.BulkInsert([]*domain.Approval{})

		s.EqualError(actualError, expectedError.Error())
	})
}

func (s *ServiceTestSuite) TestAdvanceApproval() {
	// TODO: test
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
