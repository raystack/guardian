package provider_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/provider"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	mockProviderType = "mock_provider_type"
)

type ServiceTestSuite struct {
	suite.Suite
	mockProviderRepository *mocks.ProviderRepository
	mockResourceService    *mocks.ResourceService
	mockProvider           *mocks.ProviderInterface
	service                *provider.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockProviderRepository = new(mocks.ProviderRepository)
	s.mockResourceService = new(mocks.ResourceService)
	s.mockProvider = new(mocks.ProviderInterface)
	s.mockProvider.On("GetType").Return(mockProviderType).Once()

	s.service = provider.NewService(s.mockProviderRepository, s.mockResourceService, []domain.ProviderInterface{s.mockProvider})
}

func (s *ServiceTestSuite) TestCreate() {
	config := &domain.ProviderConfig{}
	p := &domain.Provider{
		Type:   mockProviderType,
		Config: config,
	}

	s.Run("should return error if unable to retrieve provider", func() {
		expectedError := provider.ErrInvalidProviderType

		actualError := s.service.Create(&domain.Provider{
			Type: "invalid-provider-type",
		})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from the provider config validation", func() {
		expectedError := errors.New("provider config validation error")
		s.mockProvider.On("CreateConfig", mock.Anything).Return(expectedError).Once()

		actualError := s.service.Create(p)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from the provider repository", func() {
		expectedError := errors.New("error from repository")
		s.mockProvider.On("CreateConfig", mock.Anything).Return(nil).Once()
		s.mockProviderRepository.On("Create", mock.Anything).Return(expectedError).Once()

		actualError := s.service.Create(p)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should pass the model from the param", func() {
		s.mockProvider.On("CreateConfig", mock.Anything).Return(nil).Once()
		s.mockProviderRepository.On("Create", p).Return(nil).Once()

		actualError := s.service.Create(p)

		s.Nil(actualError)
		s.mockProviderRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockProviderRepository.On("Find").Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find()

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedResult := []*domain.Provider{}
		s.mockProviderRepository.On("Find").Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.Find()

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
		s.mockProviderRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestUpdate() {
	s.Run("should return error if got error getting existing record", func() {
		testCases := []struct {
			expectedExistingProvider *domain.Provider
			expectedRepositoryError  error
			expectedError            error
		}{
			{
				expectedExistingProvider: nil,
				expectedRepositoryError:  nil,
				expectedError:            provider.ErrRecordNotFound,
			},
			{
				expectedExistingProvider: nil,
				expectedRepositoryError:  errors.New("repository error"),
				expectedError:            errors.New("repository error"),
			},
		}

		for _, tc := range testCases {
			expectedProvider := &domain.Provider{
				ID: 1,
			}
			expectedError := tc.expectedError
			s.mockProviderRepository.On("GetOne", expectedProvider.ID).Return(tc.expectedExistingProvider, tc.expectedRepositoryError).Once()

			actualError := s.service.Update(expectedProvider)

			s.EqualError(actualError, expectedError.Error())
		}
	})

	s.Run("should update only non-zero values", func() {
		testCases := []struct {
			updatePayload       *domain.Provider
			existingProvider    *domain.Provider
			expectedNewProvider *domain.Provider
		}{
			{
				updatePayload: &domain.Provider{
					ID: 1,
					Config: &domain.ProviderConfig{
						Labels: map[string]interface{}{
							"foo": "bar",
						},
					},
				},
				existingProvider: &domain.Provider{
					ID: 1,
					Config: &domain.ProviderConfig{
						Appeal: &domain.AppealConfig{
							AllowPermanentAccess:         true,
							AllowActiveAccessExtensionIn: "1d",
						},
						Type: "type",
						URN:  "urn",
					},
				},
				expectedNewProvider: &domain.Provider{
					ID: 1,
					Config: &domain.ProviderConfig{
						Appeal: &domain.AppealConfig{
							AllowPermanentAccess:         true,
							AllowActiveAccessExtensionIn: "1d",
						},
						Labels: map[string]interface{}{
							"foo": "bar",
						},
						Type: "type",
						URN:  "urn",
					},
				},
			},
		}

		for _, tc := range testCases {
			s.mockProviderRepository.On("GetOne", tc.updatePayload.ID).Return(tc.existingProvider, nil).Once()
			s.mockProviderRepository.On("Update", tc.expectedNewProvider).Return(nil)

			actualError := s.service.Update(tc.updatePayload)

			s.Nil(actualError)
		}
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
