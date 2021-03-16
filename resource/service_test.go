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

func (s *ServiceTestSuite) TestUpdate() {
	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockRepository.On("Update", mock.Anything).Return(expectedError).Once()

		actualError := s.service.Update(&domain.Resource{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should only allows details and labels to be edited", func() {
		testCases := []struct {
			resourceUpdatePayload *domain.Resource
			expectedUpdatedValues *domain.Resource
		}{
			{
				resourceUpdatePayload: &domain.Resource{
					ID: 1,
					Labels: map[string]interface{}{
						"key": "value",
					},
				},
				expectedUpdatedValues: &domain.Resource{
					ID: 1,
					Labels: map[string]interface{}{
						"key": "value",
					},
				},
			},
			{
				resourceUpdatePayload: &domain.Resource{
					ID: 2,
					Details: map[string]interface{}{
						"key": "value",
					},
				},
				expectedUpdatedValues: &domain.Resource{
					ID: 2,
					Details: map[string]interface{}{
						"key": "value",
					},
				},
			},
			{
				resourceUpdatePayload: &domain.Resource{
					ID:   2,
					Type: "test",
				},
				expectedUpdatedValues: &domain.Resource{
					ID: 2,
				},
			},
		}

		for _, tc := range testCases {
			s.mockRepository.On("Update", tc.expectedUpdatedValues).Return(nil)

			actualError := s.service.Update(tc.resourceUpdatePayload)

			s.Nil(actualError)
			s.mockRepository.AssertExpectations(s.T())
		}
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
