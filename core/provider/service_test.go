package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/core/provider"
	providermocks "github.com/odpf/guardian/core/provider/mocks"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	mockProviderType = "mock_provider_type"
)

type ServiceTestSuite struct {
	suite.Suite
	mockProviderRepository *providermocks.Repository
	mockResourceService    *providermocks.ResourceService
	mockProvider           *providermocks.Client
	mockAuditLogger        *providermocks.AuditLogger
	service                *provider.Service
}

func (s *ServiceTestSuite) SetupTest() {
	logger := log.NewLogrus(log.LogrusWithLevel("info"))
	validator := validator.New()
	s.mockProviderRepository = new(providermocks.Repository)
	s.mockResourceService = new(providermocks.ResourceService)
	s.mockProvider = new(providermocks.Client)
	s.mockAuditLogger = new(providermocks.AuditLogger)
	s.mockProvider.On("GetType").Return(mockProviderType).Once()

	s.service = provider.NewService(provider.ServiceDeps{
		Repository:      s.mockProviderRepository,
		ResourceService: s.mockResourceService,
		Clients:         []provider.Client{s.mockProvider},
		Validator:       validator,
		Logger:          logger,
		AuditLogger:     s.mockAuditLogger,
	})
}

func (s *ServiceTestSuite) TestCreate() {
	config := &domain.ProviderConfig{}
	p := &domain.Provider{
		Type:   mockProviderType,
		Config: config,
	}

	s.Run("should return error if unable to retrieve provider", func() {
		expectedError := provider.ErrInvalidProviderType

		actualError := s.service.Create(context.Background(), &domain.Provider{
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

			actualError := s.service.Create(context.Background(), p)

			s.Error(actualError)
		})

		s.Run("should return error if got error from account types validation", func() {
			p := &domain.Provider{
				Type: mockProviderType,
				Config: &domain.ProviderConfig{
					AllowedAccountTypes: []string{"invalid-type"},
				},
			}

			expectedAccountTypes := []string{"non-user-only"}
			s.mockProvider.On("GetAccountTypes").Return(expectedAccountTypes).Once()

			actualError := s.service.Create(context.Background(), p)

			s.Error(actualError)
		})

		s.Run("should return error if got error from the provider config validation", func() {
			expectedError := errors.New("provider config validation error")
			s.mockProvider.On("GetAccountTypes").Return([]string{"user"}).Once()
			s.mockProvider.On("CreateConfig", mock.Anything).Return(expectedError).Once()

			actualError := s.service.Create(context.Background(), p)

			s.EqualError(actualError, expectedError.Error())
		})
	})

	s.Run("should return error if got error from the provider repository", func() {
		expectedError := errors.New("error from repository")
		s.mockProvider.On("GetAccountTypes").Return([]string{"user"}).Once()
		s.mockProvider.On("CreateConfig", mock.Anything).Return(nil).Once()
		s.mockProviderRepository.On("Create", mock.Anything).Return(expectedError).Once()

		actualError := s.service.Create(context.Background(), p)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should pass the model from the param and trigger fetch resources on success", func() {
		s.mockProvider.On("GetAccountTypes").Return([]string{"user"}).Once()
		s.mockProvider.On("CreateConfig", mock.Anything).Return(nil).Once()
		s.mockProviderRepository.On("Create", p).Return(nil).Once()
		s.mockAuditLogger.On("Log", mock.Anything, provider.AuditKeyCreate, mock.Anything).Return(nil).Once()

		expectedResources := []*domain.Resource{}
		s.mockResourceService.On("Find", mock.Anything, map[string]interface{}{
			"provider_type": p.Type,
			"provider_urn":  p.URN,
		}).Return([]*domain.Resource{}, nil).Once()
		s.mockProvider.On("GetResources", p.Config).Return(expectedResources, nil).Once()
		s.mockResourceService.On("BulkUpsert", mock.Anything, expectedResources).Return(nil).Once()

		actualError := s.service.Create(context.Background(), p)

		s.Nil(actualError)
		s.mockProviderRepository.AssertExpectations(s.T())
		s.mockAuditLogger.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockProviderRepository.On("Find").Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find(context.Background())

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedResult := []*domain.Provider{}
		s.mockProviderRepository.On("Find").Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.Find(context.Background())

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
		s.mockProviderRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestUpdateValidation() {
	s.Run("validation", func() {
		s.Run("should return error if got error on account types validation", func() {
			p := &domain.Provider{
				Type: mockProviderType,
				Config: &domain.ProviderConfig{
					AllowedAccountTypes: []string{"invalid-type"},
				},
			}

			s.mockProviderRepository.On("GetByID", mock.Anything).
				Return(&domain.Provider{}, nil).
				Once()
			s.mockProviderRepository.On("GetOne", mock.Anything, mock.Anything).
				Return(&domain.Provider{}, nil)
			s.mockProvider.On("GetAccountTypes").Return([]string{"non-user-only"}).Once()
			actualError := s.service.Update(context.Background(), p)

			s.Error(actualError)
		})

		s.Run("should return error if appeal config is invalid", func() {
			p := &domain.Provider{
				Type: mockProviderType,
				Config: &domain.ProviderConfig{
					Appeal: &domain.AppealConfig{
						AllowActiveAccessExtensionIn: "invalid-duration",
					},
				},
			}

			s.mockProviderRepository.On("GetByID", mock.Anything).
				Return(&domain.Provider{}, nil).
				Once()
			s.mockProvider.On("GetAccountTypes").Return([]string{"user"}).Once()
			actualError := s.service.Update(context.Background(), p)

			s.Error(actualError)
		})
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
				ID: "1",
			}
			expectedError := tc.expectedError
			s.mockProviderRepository.On("GetByID", expectedProvider.ID).Return(tc.expectedExistingProvider, tc.expectedRepositoryError).Once()

			actualError := s.service.Update(context.Background(), expectedProvider)

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
					ID: "1",
					Config: &domain.ProviderConfig{
						Labels: map[string]string{
							"foo": "bar",
						},
					},
				},
				existingProvider: &domain.Provider{
					ID:   "1",
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
					ID:   "1",
					Type: mockProviderType,
					Config: &domain.ProviderConfig{
						Appeal: &domain.AppealConfig{
							AllowPermanentAccess:         true,
							AllowActiveAccessExtensionIn: "1h",
						},
						AllowedAccountTypes: []string{"user"},
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
			s.mockProvider.On("GetAccountTypes").Return([]string{"user"}).Once()
			s.mockProvider.On("CreateConfig", mock.Anything).Return(nil).Once()
			s.mockProviderRepository.On("Update", tc.expectedNewProvider).Return(nil)
			s.mockAuditLogger.On("Log", mock.Anything, provider.AuditKeyUpdate, mock.Anything).Return(nil).Once()

			actualError := s.service.Update(context.Background(), tc.updatePayload)

			s.Nil(actualError)
		}
	})
}

func (s *ServiceTestSuite) TestFetchResources() {
	s.Run("should return error if got any from provider repository", func() {
		expectedError := errors.New("any error")
		s.mockProviderRepository.On("Find").Return(nil, expectedError).Once()

		actualError := s.service.FetchResources(context.Background())

		s.EqualError(actualError, expectedError.Error())
	})

	providers := []*domain.Provider{
		{
			ID:     "1",
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
		s.mockResourceService.On("BulkUpsert", mock.Anything, mock.Anything).Return(expectedError).Once()
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return([]*domain.Resource{}, nil).Once()
		actualError := s.service.FetchResources(context.Background())

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
		s.mockResourceService.On("BulkUpsert", mock.Anything, expectedResources).Return(nil).Once()
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return([]*domain.Resource{}, nil).Once()
		actualError := s.service.FetchResources(context.Background())

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
			actualError := s.service.GrantAccess(context.Background(), tc.appealParam)
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
		actualError := s.service.GrantAccess(context.Background(), appeal)
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

		actualError := s.service.GrantAccess(context.Background(), validAppeal)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if provider not found", func() {
		s.mockProviderRepository.On("GetOne", mock.Anything, mock.Anything).
			Return(nil, provider.ErrRecordNotFound).
			Once()
		expectedError := provider.ErrRecordNotFound

		actualError := s.service.GrantAccess(context.Background(), validAppeal)

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

		actualError := s.service.GrantAccess(context.Background(), validAppeal)

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

		actualError := s.service.GrantAccess(context.Background(), validAppeal)

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
			actualError := s.service.RevokeAccess(context.Background(), tc.appealParam)
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
		actualError := s.service.RevokeAccess(context.Background(), appeal)
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

		actualError := s.service.RevokeAccess(context.Background(), validAppeal)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if provider not found", func() {
		s.mockProviderRepository.On("GetOne", mock.Anything, mock.Anything).
			Return(nil, provider.ErrRecordNotFound).
			Once()
		expectedError := provider.ErrRecordNotFound

		actualError := s.service.RevokeAccess(context.Background(), validAppeal)

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

		actualError := s.service.RevokeAccess(context.Background(), validAppeal)

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

		actualError := s.service.RevokeAccess(context.Background(), validAppeal)

		s.Nil(actualError)
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
