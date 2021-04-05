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
	mockRepository *mocks.ApprovalRepository

	service domain.ApprovalService
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(mocks.ApprovalRepository)

	s.service = approval.NewService(s.mockRepository)
}

func (s *ServiceTestSuite) TestCreate() {
	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("repository error")
		s.mockRepository.On("BulkInsert", mock.Anything).Return(expectedError).Once()

		actualError := s.service.BulkInsert([]*domain.Approval{})

		s.EqualError(actualError, expectedError.Error())
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
