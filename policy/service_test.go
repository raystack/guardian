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

	s.Run("should set version to 1", func() {
		p := &domain.Policy{
			ID: "test",
		}

		expectedPolicy := &domain.Policy{
			ID:      p.ID,
			Version: 1,
		}
		s.mockPolicyRepository.On("Create", p).Return(nil).Once()

		actualError := s.service.Create(p)

		s.Nil(actualError)
		s.Equal(expectedPolicy, p)
		s.mockPolicyRepository.AssertExpectations(s.T())
	})

	s.Run("should pass the model from the param", func() {
		s.mockPolicyRepository.On("Create", p).Return(nil).Once()

		actualError := s.service.Create(p)

		s.Nil(actualError)
		s.mockPolicyRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockPolicyRepository.On("Find").Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find()

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedResult := []*domain.Policy{}
		s.mockPolicyRepository.On("Find").Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.Find()

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
		s.mockPolicyRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestGetOne() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockPolicyRepository.On("GetOne", mock.Anything, mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.GetOne("", 0)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedResult := &domain.Policy{}
		s.mockPolicyRepository.On("GetOne", mock.Anything, mock.Anything).Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.GetOne("", 0)

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
		s.mockPolicyRepository.AssertExpectations(s.T())
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
