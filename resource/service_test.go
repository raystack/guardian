package resource_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository *mocks.ResourceRepository
	service        *resource.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(mocks.ResourceRepository)
	s.service = resource.NewService(s.mockRepository)
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockRepository.On("Find").Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find()

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedResult := []*domain.Resource{}
		s.mockRepository.On("Find").Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.Find()

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
		s.mockRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestBulkUpsert() {
	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockRepository.On("BulkUpsert", mock.Anything).Return(expectedError).Once()

		actualError := s.service.BulkUpsert([]*domain.Resource{})

		s.EqualError(actualError, expectedError.Error())
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
