package resource_test

import (
	"context"
	"errors"
	"testing"

	"github.com/odpf/guardian/core/resource"
	"github.com/odpf/guardian/core/resource/mocks"
	"github.com/odpf/guardian/domain"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository  *mocks.Repository
	mockAuditLogger *mocks.AuditLogger
	service         *resource.Service

	authenticatedUserEmail string
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(mocks.Repository)
	s.mockAuditLogger = new(mocks.AuditLogger)
	s.service = resource.NewService(resource.ServiceDeps{
		Repository:  s.mockRepository,
		AuditLogger: s.mockAuditLogger,
	})
	s.authenticatedUserEmail = "user@example.com"
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find(context.Background(), domain.ListResourcesFilter{})

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedFilters := domain.ListResourcesFilter{}
		expectedResult := []*domain.Resource{}
		s.mockRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx"), expectedFilters).Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.Find(context.Background(), expectedFilters)

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
		s.mockRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestBulkUpsert() {
	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockRepository.EXPECT().BulkUpsert(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(expectedError).Once()

		actualError := s.service.BulkUpsert(context.Background(), []*domain.Resource{})

		s.EqualError(actualError, expectedError.Error())
	})
}

func (s *ServiceTestSuite) TestUpdate() {
	s.Run("should return error if got error getting existing record", func() {
		testCases := []struct {
			expectedExistingResource *domain.Resource
			expectedRepositoryError  error
			expectedError            error
		}{
			{
				expectedExistingResource: nil,
				expectedRepositoryError:  resource.ErrRecordNotFound,
				expectedError:            resource.ErrRecordNotFound,
			},
			{
				expectedExistingResource: nil,
				expectedRepositoryError:  errors.New("repository error"),
				expectedError:            errors.New("repository error"),
			},
		}

		for _, tc := range testCases {
			expectedResource := &domain.Resource{
				ID: "1",
			}
			expectedError := tc.expectedError
			s.mockRepository.EXPECT().
				GetOne(mock.AnythingOfType("*context.emptyCtx"), expectedResource.ID).
				Return(tc.expectedExistingResource, tc.expectedRepositoryError).Once()

			actualError := s.service.Update(context.Background(), expectedResource)

			s.EqualError(actualError, expectedError.Error())
		}
	})

	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockRepository.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(&domain.Resource{}, nil).Once()
		s.mockRepository.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedError).Once()

		actualError := s.service.Update(context.Background(), &domain.Resource{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should only allows details and labels to be edited", func() {
		testCases := []struct {
			resourceUpdatePayload *domain.Resource
			existingResource      *domain.Resource
			expectedUpdatedValues *domain.Resource
		}{
			{
				resourceUpdatePayload: &domain.Resource{
					ID: "1",
					Labels: map[string]string{
						"key": "value",
					},
				},
				existingResource: &domain.Resource{
					ID: "1",
				},
				expectedUpdatedValues: &domain.Resource{
					ID: "1",
					Labels: map[string]string{
						"key": "value",
					},
				},
			},
			{
				resourceUpdatePayload: &domain.Resource{
					ID: "2",
					Details: map[string]interface{}{
						"key": "value",
					},
				},
				existingResource: &domain.Resource{
					ID: "2",
				},
				expectedUpdatedValues: &domain.Resource{
					ID: "2",
					Details: map[string]interface{}{
						"key": "value",
					},
				},
			},
			{
				resourceUpdatePayload: &domain.Resource{
					ID:   "2",
					Type: "test",
				},
				existingResource: &domain.Resource{
					ID: "2",
				},
				expectedUpdatedValues: &domain.Resource{
					ID: "2",
				},
			},
		}

		for _, tc := range testCases {
			s.mockRepository.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), tc.resourceUpdatePayload.ID).Return(tc.existingResource, nil).Once()
			s.mockRepository.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), tc.expectedUpdatedValues).Return(nil).Once()
			s.mockAuditLogger.EXPECT().Log(mock.Anything, resource.AuditKeyResourceUpdate, mock.Anything).Return(nil)

			actualError := s.service.Update(context.Background(), tc.resourceUpdatePayload)

			s.Nil(actualError)
			s.mockRepository.AssertExpectations(s.T())
		}
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
