package policy_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/policy"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockPolicyRepository *mocks.PolicyRepository
	service              *policy.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockPolicyRepository = new(mocks.PolicyRepository)
	s.service = policy.NewService(s.mockPolicyRepository)
}

func (s *ServiceTestSuite) TestCreate() {
	p := &domain.Policy{
		ID:      "test",
		Version: 1,
	}

	s.Run("should return error if got error from the policy repository", func() {
		expectedError := errors.New("error from repository")
		s.mockPolicyRepository.On("Create", mock.Anything).Return(expectedError).Once()

		actualError := s.service.Create(&domain.Policy{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should pass the model from the param", func() {
		s.mockPolicyRepository.On("Create", p).Return(nil).Once()

		actualError := s.service.Create(p)

		s.Nil(actualError)
		s.mockPolicyRepository.AssertExpectations(s.T())
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
