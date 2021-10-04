package provider_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/provider"
	"github.com/odpf/salt/log"
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
	logger := log.NewLogrus(log.LogrusWithLevel("info"))
	s.mockProviderRepository = new(mocks.ProviderRepository)
	s.mockResourceService = new(mocks.ResourceService)
	s.mockProvider = new(mocks.ProviderInterface)
	s.mockProvider.On("GetType").Return(mockProviderType).Once()

	s.service = provider.NewService(logger, s.mockProviderRepository, s.mockResourceService, []domain.ProviderInterface{s.mockProvider})
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

	s.Run("error validation", func() {
		s.Run("should return error if appeal config is invalid", func() {
			p := &domain.Provider{
				Config: &domain.ProviderConfig{
					Appeal: &domain.AppealConfig{
						AllowActiveAccessExtensionIn: "invalid-duration",
					},
				},
			}

			actualError := s.service.Create(p)

			s.Error(actualError)
		})

		s.Run("should return error if got error from the provider config validation", func() {
			expectedError := errors.New("provider config validation error")
			s.mockProvider.On("CreateConfig", mock.Anything).Return(expectedError).Once()

			actualError := s.service.Create(p)

			s.EqualError(actualError, expectedError.Error())
		})
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
	s.Run("should return error if appeal config is invalid", func() {
		p := &domain.Provider{
			Config: &domain.ProviderConfig{
				Appeal: &domain.AppealConfig{
					AllowActiveAccessExtensionIn: "invalid-duration",
				},
			},
		}

		s.mockProviderRepository.On("GetByID", mock.Anything).
			Return(&domain.Provider{}, nil).
			Once()
		actualError := s.service.Update(p)

		s.Error(actualError)
	})

	s.Run("should return error if got error getting existing record", func() {
		testCases := []struct {
			expectedExistingProvider *domain.Provider
			expectedRepositoryError  error
			expectedError            error
		}{
			{
				expectedExistingProvider: nil,
				expectedRepositoryError:  provider.ErrRecordNotFound,
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
						Labels: map[string]string{
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
							AllowActiveAccessExtensionIn: "1h",
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
							AllowActiveAccessExtensionIn: "1h",
						},
						Labels: map[string]string{
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

func (s *ServiceTestSuite) TestGrantAccess() {
	s.Run("should return error if got error on appeal param validation", func() {
		testCases := []struct {
			appealParam   *domain.Appeal
			expectedError error
		}{
			{
				appealParam:   nil,
				expectedError: provider.ErrNilAppeal,
			},
			{
				appealParam:   &domain.Appeal{},
				expectedError: provider.ErrNilResource,
			},
		}
		for _, tc := range testCases {
			actualError := s.service.GrantAccess(tc.appealParam)
			s.EqualError(actualError, tc.expectedError.Error())
		}
	})

	s.Run("should return error if provider is not exists", func() {
		appeal := &domain.Appeal{
			Resource: &domain.Resource{
				ProviderType: "invalid-provider-type",
			},
		}
		expectedError := provider.ErrInvalidProviderType
		actualError := s.service.GrantAccess(appeal)
		s.EqualError(actualError, expectedError.Error())
	})

	validAppeal := &domain.Appeal{
		Resource: &domain.Resource{
			ProviderType: mockProviderType,
			ProviderURN:  "urn",
		},
	}

	s.Run("should return error if got any from provider repository", func() {
		expectedError := errors.New("any error")
		s.mockProviderRepository.On("GetOne", mock.Anything, mock.Anything).
			Return(nil, expectedError).
			Once()

		actualError := s.service.GrantAccess(validAppeal)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if provider not found", func() {
		s.mockProviderRepository.On("GetOne", mock.Anything, mock.Anything).
			Return(nil, provider.ErrRecordNotFound).
			Once()
		expectedError := provider.ErrRecordNotFound

		actualError := s.service.GrantAccess(validAppeal)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if error if got error from provider.GrantAccess", func() {
		provider := &domain.Provider{
			Config: &domain.ProviderConfig{},
		}
		s.mockProviderRepository.
			On("GetOne", validAppeal.Resource.ProviderType, validAppeal.Resource.ProviderURN).
			Return(provider, nil).
			Once()
		expectedError := errors.New("any error")
		s.mockProvider.On("GrantAccess", mock.Anything, mock.Anything).
			Return(expectedError).
			Once()

		actualError := s.service.GrantAccess(validAppeal)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should grant access to the provider on success", func() {
		provider := &domain.Provider{
			Config: &domain.ProviderConfig{},
		}
		s.mockProviderRepository.
			On("GetOne", validAppeal.Resource.ProviderType, validAppeal.Resource.ProviderURN).
			Return(provider, nil).
			Once()
		s.mockProvider.
			On("GrantAccess", provider.Config, validAppeal).
			Return(nil).
			Once()

		actualError := s.service.GrantAccess(validAppeal)

		s.Nil(actualError)
	})
}

func (s *ServiceTestSuite) TestRevokeAccess() {
	s.Run("should return error if got error on appeal param validation", func() {
		testCases := []struct {
			appealParam   *domain.Appeal
			expectedError error
		}{
			{
				appealParam:   nil,
				expectedError: provider.ErrNilAppeal,
			},
			{
				appealParam:   &domain.Appeal{},
				expectedError: provider.ErrNilResource,
			},
		}
		for _, tc := range testCases {
			actualError := s.service.RevokeAccess(tc.appealParam)
			s.EqualError(actualError, tc.expectedError.Error())
		}
	})

	s.Run("should return error if provider is not exists", func() {
		appeal := &domain.Appeal{
			Resource: &domain.Resource{
				ProviderType: "invalid-provider-type",
			},
		}
		expectedError := provider.ErrInvalidProviderType
		actualError := s.service.RevokeAccess(appeal)
		s.EqualError(actualError, expectedError.Error())
	})

	validAppeal := &domain.Appeal{
		Resource: &domain.Resource{
			ProviderType: mockProviderType,
			ProviderURN:  "urn",
		},
	}

	s.Run("should return error if got any from provider repository", func() {
		expectedError := errors.New("any error")
		s.mockProviderRepository.On("GetOne", mock.Anything, mock.Anything).
			Return(nil, expectedError).
			Once()

		actualError := s.service.RevokeAccess(validAppeal)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if provider not found", func() {
		s.mockProviderRepository.On("GetOne", mock.Anything, mock.Anything).
			Return(nil, provider.ErrRecordNotFound).
			Once()
		expectedError := provider.ErrRecordNotFound

		actualError := s.service.RevokeAccess(validAppeal)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if error if got error from provider.RevokeAccess", func() {
		provider := &domain.Provider{
			Config: &domain.ProviderConfig{},
		}
		s.mockProviderRepository.
			On("GetOne", validAppeal.Resource.ProviderType, validAppeal.Resource.ProviderURN).
			Return(provider, nil).
			Once()
		expectedError := errors.New("any error")
		s.mockProvider.On("RevokeAccess", mock.Anything, mock.Anything).
			Return(expectedError).
			Once()

		actualError := s.service.RevokeAccess(validAppeal)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should grant access to the provider on success", func() {
		provider := &domain.Provider{
			Config: &domain.ProviderConfig{},
		}
		s.mockProviderRepository.
			On("GetOne", validAppeal.Resource.ProviderType, validAppeal.Resource.ProviderURN).
			Return(provider, nil).
			Once()
		s.mockProvider.
			On("RevokeAccess", provider.Config, validAppeal).
			Return(nil).
			Once()

		actualError := s.service.RevokeAccess(validAppeal)

		s.Nil(actualError)
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
