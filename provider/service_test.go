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
			s.mockProviderRepository.On("GetByID", expectedProvider.ID).Return(tc.expectedExistingProvider, tc.expectedRepositoryError).Once()

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
					ID:   1,
					Type: mockProviderType,
					Config: &domain.ProviderConfig{
						Appeal: &domain.AppealConfig{
							AllowPermanentAccess:         true,
							AllowActiveAccessExtensionIn: "1d",
						},
						Type: mockProviderType,
						URN:  "urn",
					},
				},
				expectedNewProvider: &domain.Provider{
					ID:   1,
					Type: mockProviderType,
					Config: &domain.ProviderConfig{
						Appeal: &domain.AppealConfig{
							AllowPermanentAccess:         true,
							AllowActiveAccessExtensionIn: "1d",
						},
						Labels: map[string]interface{}{
							"foo": "bar",
						},
						Type: mockProviderType,
						URN:  "urn",
					},
				},
			},
		}

		for _, tc := range testCases {
			s.mockProviderRepository.On("GetByID", tc.updatePayload.ID).Return(tc.existingProvider, nil).Once()
			s.mockProvider.On("CreateConfig", mock.Anything).Return(nil).Once()
			s.mockProviderRepository.On("Update", tc.expectedNewProvider).Return(nil)

			actualError := s.service.Update(tc.updatePayload)

			s.Nil(actualError)
		}
	})
}

func (s *ServiceTestSuite) TestFetchResources() {
	s.Run("should return error if got any from provider respository", func() {
		expectedError := errors.New("any error")
		s.mockProviderRepository.On("Find").Return(nil, expectedError).Once()

		actualError := s.service.FetchResources()

		s.EqualError(actualError, expectedError.Error())
	})

	providers := []*domain.Provider{
		{
			ID:     1,
			Type:   mockProviderType,
			Config: &domain.ProviderConfig{},
		},
	}

	s.Run("should return error if got any from provider's GetResources", func() {
		s.mockProviderRepository.On("Find").Return(providers, nil).Once()
		expectedError := errors.New("any error")
		s.mockProvider.On("GetResources", mock.Anything).Return(nil, expectedError).Once()

		actualError := s.service.FetchResources()

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got any from resource service", func() {
		s.mockProviderRepository.On("Find").Return(providers, nil).Once()
		for _, p := range providers {
			s.mockProvider.On("GetResources", p.Config).Return([]*domain.Resource{}, nil).Once()
		}
		expectedError := errors.New("any error")
		s.mockResourceService.On("BulkUpsert", mock.Anything).Return(expectedError).Once()

		actualError := s.service.FetchResources()

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should upsert all resources on success", func() {
		s.mockProviderRepository.On("Find").Return(providers, nil).Once()
		expectedResources := []*domain.Resource{}
		for _, p := range providers {
			resources := []*domain.Resource{
				{
					ProviderType: p.Type,
					ProviderURN:  p.URN,
				},
			}
			s.mockProvider.On("GetResources", p.Config).Return(resources, nil).Once()
			expectedResources = append(expectedResources, resources...)
		}
		s.mockResourceService.On("BulkUpsert", expectedResources).Return(nil).Once()

		actualError := s.service.FetchResources()

		s.Nil(actualError)
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
