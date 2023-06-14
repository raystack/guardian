package approval_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/guardian/core/approval"
	approvalmocks "github.com/raystack/guardian/core/approval/mocks"
	"github.com/raystack/guardian/domain"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository    *approvalmocks.Repository
	mockPolicyService *approvalmocks.PolicyService

	service *approval.Service
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(approvalmocks.Repository)
	s.mockPolicyService = new(approvalmocks.PolicyService)

	s.service = approval.NewService(approval.ServiceDeps{
		s.mockRepository,
		s.mockPolicyService,
	})
}

func (s *ServiceTestSuite) TestBulkInsert() {
	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("repository error")
		s.mockRepository.EXPECT().
			BulkInsert(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedError).Once()

		actualError := s.service.BulkInsert(context.Background(), []*domain.Approval{})

		s.EqualError(actualError, expectedError.Error())
	})
}

func (s *ServiceTestSuite) TestAddApprover() {
	s.Run("should return nil error on success", func() {
		expectedApprover := &domain.Approver{
			ApprovalID: uuid.New().String(),
			Email:      "user@example.com",
		}
		s.mockRepository.EXPECT().AddApprover(mock.AnythingOfType("*context.emptyCtx"), expectedApprover).Return(nil)

		err := s.service.AddApprover(context.Background(), expectedApprover.ApprovalID, expectedApprover.Email)

		s.NoError(err)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if repository returns an error", func() {
		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().AddApprover(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(expectedError)

		err := s.service.AddApprover(context.Background(), "", "")

		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestDeleteApprover() {
	s.Run("should return nil error on success", func() {
		approvalID := uuid.New().String()
		approverEmail := "user@example.com"

		s.mockRepository.EXPECT().DeleteApprover(mock.AnythingOfType("*context.emptyCtx"), approvalID, approverEmail).Return(nil)

		err := s.service.DeleteApprover(context.Background(), approvalID, approverEmail)

		s.NoError(err)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if repository returns an error", func() {
		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().DeleteApprover(mock.AnythingOfType("*context.emptyCtx"), mock.Anything, mock.Anything).Return(expectedError)

		err := s.service.DeleteApprover(context.Background(), "", "")

		s.ErrorIs(err, expectedError)
		s.mockRepository.AssertExpectations(s.T())
	})
}
