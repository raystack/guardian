package provider_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/goto/guardian/core/provider"
	providermocks "github.com/goto/guardian/core/provider/mocks"
	"github.com/goto/guardian/core/resource"
	"github.com/goto/guardian/domain"
	"github.com/goto/salt/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	mockProviderType = "mock_provider_type"
	mockProvider     = "mock_provider"
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
		s.mockProviderRepository.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(expectedError).Once()

		actualError := s.service.Create(context.Background(), p)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should pass the model from the param and trigger fetch resources on success", func() {
		s.mockProvider.On("GetAccountTypes").Return([]string{"user"}).Once()
		s.mockProvider.On("CreateConfig", mock.Anything).Return(nil).Once()
		s.mockProviderRepository.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), p).Return(nil).Once()
		s.mockAuditLogger.On("Log", mock.Anything, provider.AuditKeyCreate, mock.Anything).Return(nil).Once()

		expectedResources := []*domain.Resource{}
		s.mockResourceService.On("Find", mock.Anything, domain.ListResourcesFilter{
			ProviderType: p.Type,
			ProviderURN:  p.URN,
		}).Return([]*domain.Resource{}, nil).Once()
		s.mockProvider.On("GetResources", p.Config).Return(expectedResources, nil).Once()
		s.mockResourceService.On("BulkUpsert", mock.Anything, expectedResources).Return(nil).Once()

		actualError := s.service.Create(context.Background(), p)

		s.Nil(actualError)
		s.mockProviderRepository.AssertExpectations(s.T())
		s.mockAuditLogger.AssertExpectations(s.T())
	})

	s.Run("with dryRun true", func() {
		s.Run("should not perform any changes", func() {
			s.mockProvider.On("GetAccountTypes").Return([]string{"user"}).Once()
			s.mockProvider.On("CreateConfig", mock.Anything).Return(nil).Once()

			expectedResources := []*domain.Resource{}
			s.mockResourceService.On("Find", mock.Anything, domain.ListResourcesFilter{
				ProviderType: p.Type,
				ProviderURN:  p.URN,
			}).Return([]*domain.Resource{}, nil).Once()
			s.mockProvider.On("GetResources", p.Config).Return(expectedResources, nil).Once()

			ctx := provider.WithDryRun(context.Background())

			actualError := s.service.Create(ctx, p)

			s.Nil(actualError)
			s.mockProviderRepository.AssertNotCalled(s.T(), "Create")
			s.mockAuditLogger.AssertNotCalled(s.T(), "Log")
			s.mockResourceService.AssertNotCalled(s.T(), "BulkUpsert")
		})
	})
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockProviderRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx")).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find(context.Background())

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedResult := []*domain.Provider{}
		s.mockProviderRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx")).Return(expectedResult, nil).Once()

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

			s.mockProviderRepository.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
				Return(&domain.Provider{}, nil).
				Once()
			s.mockProviderRepository.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), mock.Anything, mock.Anything).
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

			s.mockProviderRepository.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
				Return(&domain.Provider{}, nil).
				Once()
			s.mockProvider.On("GetAccountTypes").Return([]string{"user"}).Once()
			actualError := s.service.Update(context.Background(), p)

			s.Error(actualError)
		})
	})
}

func (s *ServiceTestSuite) TestUpdate() {
	s.Run("should update record on success", func() {
		testCases := []struct {
			updatePayload       *domain.Provider
			existingProvider    *domain.Provider
			expectedNewProvider *domain.Provider
		}{
			{
				updatePayload: &domain.Provider{
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
			s.mockProvider.On("GetAccountTypes").Return([]string{"user"}).Once()
			s.mockProvider.On("CreateConfig", mock.Anything).Return(nil).Once()
			s.mockProviderRepository.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), tc.expectedNewProvider).Return(nil)
			s.mockAuditLogger.On("Log", mock.Anything, provider.AuditKeyUpdate, mock.Anything).Return(nil).Once()

			expectedResources := []*domain.Resource{}
			s.mockResourceService.On("Find", mock.Anything, domain.ListResourcesFilter{
				ProviderType: tc.updatePayload.Type,
				ProviderURN:  tc.updatePayload.URN,
			}).Return([]*domain.Resource{}, nil).Once()
			s.mockProvider.On("GetResources", tc.updatePayload.Config).Return(expectedResources, nil).Once()
			s.mockResourceService.On("BulkUpsert", mock.Anything, expectedResources).Return(nil).Once()

			actualError := s.service.Update(context.Background(), tc.updatePayload)

			s.Nil(actualError)
		}
	})

	s.Run("with dryRun true", func() {
		s.Run("should not perform any changes", func() {
			p := &domain.Provider{
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
			}

			s.mockProvider.On("GetAccountTypes").Return([]string{"user"}).Once()
			s.mockProvider.On("CreateConfig", mock.Anything).Return(nil).Once()

			ctx := provider.WithDryRun(context.Background())

			expectedResources := []*domain.Resource{}
			s.mockResourceService.On("Find", mock.Anything, domain.ListResourcesFilter{
				ProviderType: p.Type,
				ProviderURN:  p.URN,
			}).Return([]*domain.Resource{}, nil).Once()
			s.mockProvider.On("GetResources", p.Config).Return(expectedResources, nil).Once()

			actualError := s.service.Update(ctx, p)

			s.Nil(actualError)
			s.mockProviderRepository.AssertNotCalled(s.T(), "Update")
			s.mockAuditLogger.AssertNotCalled(s.T(), "Log")
		})
	})
}

func (s *ServiceTestSuite) TestFetchResources() {
	s.Run("should return error if got any from provider repository", func() {
		expectedError := errors.New("any error")
		s.mockProviderRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx")).Return(nil, expectedError).Once()

		actualError := s.service.FetchResources(context.Background())

		s.EqualError(actualError, expectedError.Error())
	})

	providers := []*domain.Provider{
		{
			ID:     "1",
			Type:   mockProviderType,
			URN:    mockProvider,
			Config: &domain.ProviderConfig{},
		},
	}

	s.Run("should return error if got any from resource service", func() {
		s.mockProviderRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx")).Return(providers, nil).Once()
		for _, p := range providers {
			s.mockProvider.On("GetResources", p.Config).Return([]*domain.Resource{}, nil).Once()
		}
		expectedError := errors.New("failed to add resources providers - [mock_provider]")
		s.mockResourceService.On("BulkUpsert", mock.Anything, mock.Anything).Return(expectedError).Once()
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return([]*domain.Resource{}, nil).Once()
		actualError := s.service.FetchResources(context.Background())

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should upsert all resources on success", func() {
		existingResources := []*domain.Resource{
			{
				ID:           "1",
				ProviderType: mockProviderType,
				ProviderURN:  mockProvider,
				Type:         "test-resource-type",
				URN:          "test-resource-urn-1",
				Details: map[string]interface{}{
					"owner": "test-owner",
					resource.ReservedDetailsKeyMetadata: map[string]interface{}{
						"labels": map[string]string{
							"foo": "bar",
							"baz": "qux",
						},
						"x": "y",
					},
				},
			},
		}
		newResources := []*domain.Resource{
			{
				ProviderType: mockProviderType,
				ProviderURN:  mockProvider,
				Type:         "test-resource-type",
				URN:          "test-resource-urn-1",
				Details: map[string]interface{}{
					resource.ReservedDetailsKeyMetadata: map[string]interface{}{
						"labels": map[string]string{
							"new-key": "new-value",
						},
					},
				},
			},
			{
				ProviderType: mockProviderType,
				ProviderURN:  mockProvider,
				Type:         "test-resource-type",
				URN:          "test-resource-urn-2",
			},
		}
		expectedResources := []*domain.Resource{
			{
				ProviderType: mockProviderType,
				ProviderURN:  mockProvider,
				Type:         "test-resource-type",
				URN:          "test-resource-urn-1",
				Details: map[string]interface{}{
					"owner": "test-owner", // owner not changed
					resource.ReservedDetailsKeyMetadata: map[string]interface{}{ // metadata updated
						"labels": map[string]string{
							"new-key": "new-value",
						},
					},
				},
			},
			{
				ProviderType: mockProviderType,
				ProviderURN:  mockProvider,
				Type:         "test-resource-type",
				URN:          "test-resource-urn-2",
			},
		}

		expectedProvider := providers[0]
		s.mockProviderRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx")).Return([]*domain.Provider{expectedProvider}, nil).Once()
		s.mockProvider.EXPECT().GetResources(expectedProvider.Config).Return(newResources, nil).Once()
		s.mockResourceService.EXPECT().BulkUpsert(mock.Anything, mock.AnythingOfType("[]*domain.Resource")).
			Run(func(_a0 context.Context, resources []*domain.Resource) {
				s.Empty(cmp.Diff(expectedResources, resources, cmpopts.IgnoreFields(domain.Resource{}, "ID", "CreatedAt", "UpdatedAt")))
			}).Return(nil).Once()
		s.mockResourceService.EXPECT().Find(mock.Anything, mock.Anything).Return(existingResources, nil).Once()
		actualError := s.service.FetchResources(context.Background())

		s.Nil(actualError)
	})

	s.Run("should upsert filter resources on success", func() {
		providersWithResourceFilter := []*domain.Provider{
			{
				ID:   "1",
				Type: mockProviderType,
				URN:  mockProvider,
				Config: &domain.ProviderConfig{Resources: []*domain.ResourceConfig{
					{Type: "dataset", Filter: "$urn == 'resource2'"},
				}},
			},
		}
		s.mockProviderRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx")).Return(providersWithResourceFilter, nil).Once()
		expectedResources := []*domain.Resource{}
		for _, p := range providersWithResourceFilter {
			resources := []*domain.Resource{
				{
					ProviderType: p.Type,
					ProviderURN:  p.URN,
					Type:         "dataset",
					URN:          "resource1",
				}, {
					ProviderType: p.Type,
					ProviderURN:  p.URN,
					Type:         "dataset",
					URN:          "resource2",
				},
			}
			s.mockProvider.On("GetResources", p.Config).Return(resources, nil).Once()
			expectedResources = append(expectedResources, resources[1])
		}
		s.mockResourceService.On("BulkUpsert", mock.Anything, expectedResources).Return(nil)
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return([]*domain.Resource{}, nil).Once()
		actualError := s.service.FetchResources(context.Background())

		s.Nil(actualError)
	})

	s.Run("should upsert filter resources ends with `transaction` on success", func() {
		providersWithResourceFilter := []*domain.Provider{
			{
				ID:   "1",
				Type: mockProviderType,
				URN:  mockProvider,
				Config: &domain.ProviderConfig{Resources: []*domain.ResourceConfig{
					{Type: "dataset", Filter: "$urn endsWith 'transaction' && $details.category == 'transaction'"},
				}},
			},
		}
		s.mockProviderRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx")).Return(providersWithResourceFilter, nil).Once()
		expectedResources := []*domain.Resource{}
		for _, p := range providersWithResourceFilter {
			resources := []*domain.Resource{
				{
					ProviderType: p.Type,
					ProviderURN:  p.URN,
					Type:         "dataset",
					URN:          "resource1",
				}, {
					ProviderType: p.Type,
					ProviderURN:  p.URN,
					Type:         "dataset",
					URN:          "order_transaction",
					Details:      map[string]interface{}{"category": "transaction"},
				},
			}
			s.mockProvider.On("GetResources", p.Config).Return(resources, nil).Once()
			expectedResources = append(expectedResources, resources[1])
		}
		s.mockResourceService.On("BulkUpsert", mock.Anything, expectedResources).Return(nil)
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return([]*domain.Resource{}, nil).Once()
		actualError := s.service.FetchResources(context.Background())

		s.Nil(actualError)
	})
}

func (s *ServiceTestSuite) TestGrantAccess() {
	s.Run("should return error if got error on appeal param validation", func() {
		testCases := []struct {
			appealParam   domain.Grant
			expectedError error
		}{
			{
				appealParam:   domain.Grant{},
				expectedError: provider.ErrNilResource,
			},
		}
		for _, tc := range testCases {
			actualError := s.service.GrantAccess(context.Background(), tc.appealParam)
			s.EqualError(actualError, tc.expectedError.Error())
		}
	})

	s.Run("should return error if provider is not exists", func() {
		appeal := domain.Grant{
			Resource: &domain.Resource{
				ProviderType: "invalid-provider-type",
			},
		}
		expectedError := provider.ErrInvalidProviderType
		actualError := s.service.GrantAccess(context.Background(), appeal)
		s.EqualError(actualError, expectedError.Error())
	})

	validAppeal := domain.Grant{
		Resource: &domain.Resource{
			ProviderType: mockProviderType,
			ProviderURN:  "urn",
		},
	}

	s.Run("should return error if got any from provider repository", func() {
		expectedError := errors.New("any error")
		s.mockProviderRepository.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), mock.Anything, mock.Anything).
			Return(nil, expectedError).
			Once()

		actualError := s.service.GrantAccess(context.Background(), validAppeal)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if provider not found", func() {
		s.mockProviderRepository.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), mock.Anything, mock.Anything).
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
			EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), validAppeal.Resource.ProviderType, validAppeal.Resource.ProviderURN).
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
			EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), validAppeal.Resource.ProviderType, validAppeal.Resource.ProviderURN).
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
			appealParam   domain.Grant
			expectedError error
		}{
			{
				appealParam:   domain.Grant{},
				expectedError: provider.ErrNilResource,
			},
		}
		for _, tc := range testCases {
			actualError := s.service.RevokeAccess(context.Background(), tc.appealParam)
			s.EqualError(actualError, tc.expectedError.Error())
		}
	})

	s.Run("should return error if provider is not exists", func() {
		appeal := domain.Grant{
			Resource: &domain.Resource{
				ProviderType: "invalid-provider-type",
			},
		}
		expectedError := provider.ErrInvalidProviderType
		actualError := s.service.RevokeAccess(context.Background(), appeal)
		s.EqualError(actualError, expectedError.Error())
	})

	validAppeal := domain.Grant{
		Resource: &domain.Resource{
			ProviderType: mockProviderType,
			ProviderURN:  "urn",
		},
	}

	s.Run("should return error if got any from provider repository", func() {
		expectedError := errors.New("any error")
		s.mockProviderRepository.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), mock.Anything, mock.Anything).
			Return(nil, expectedError).
			Once()

		actualError := s.service.RevokeAccess(context.Background(), validAppeal)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if provider not found", func() {
		s.mockProviderRepository.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), mock.Anything, mock.Anything).
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
			EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), validAppeal.Resource.ProviderType, validAppeal.Resource.ProviderURN).
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
			EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), validAppeal.Resource.ProviderType, validAppeal.Resource.ProviderURN).
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

func (s *ServiceTestSuite) TestDelete() {
	s.Run("should return error if provider repository returns error", func() {
		expectedError := errors.New("random error")
		s.mockProviderRepository.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(nil, expectedError).Once()

		err := s.service.Delete(context.Background(), "test-provider")

		s.ErrorIs(err, expectedError)
	})

	s.Run("should return error if resourceService.Find returns error", func() {
		s.mockProviderRepository.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(&domain.Provider{}, nil).Once()
		expectedError := errors.New("random error")
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return(nil, expectedError).Once()

		err := s.service.Delete(context.Background(), "test-provider")

		s.ErrorIs(err, expectedError)
	})

	s.Run("should return error if resourceService.BatchDelete returns error", func() {
		s.mockProviderRepository.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(&domain.Provider{}, nil).Once()
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return([]*domain.Resource{}, nil).Once()
		expectedError := errors.New("random error")
		s.mockResourceService.On("BatchDelete", mock.Anything, mock.Anything).Return(expectedError).Once()

		err := s.service.Delete(context.Background(), "test-provider")

		s.ErrorIs(err, expectedError)
	})

	s.Run("should return error if providerRepository.Delete returns error", func() {
		s.mockProviderRepository.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(&domain.Provider{}, nil).Once()
		s.mockResourceService.On("Find", mock.Anything, mock.Anything).Return([]*domain.Resource{}, nil).Once()
		s.mockResourceService.On("BatchDelete", mock.Anything, mock.Anything).Return(nil).Once()
		expectedError := errors.New("random error")
		s.mockProviderRepository.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(expectedError).Once()

		err := s.service.Delete(context.Background(), "test-provider")

		s.ErrorIs(err, expectedError)
	})

	s.Run("should return nil on success", func() {
		testID := "test-provider"
		dummyProvider := &domain.Provider{
			Type: "test-type",
			URN:  "test-urn",
		}
		dummyResources := []*domain.Resource{{ID: "a"}, {ID: "b"}}

		s.mockProviderRepository.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), testID).Return(dummyProvider, nil).Once()
		s.mockResourceService.On("Find", mock.Anything, domain.ListResourcesFilter{
			ProviderType: dummyProvider.Type,
			ProviderURN:  dummyProvider.URN,
		}).Return(dummyResources, nil).Once()
		s.mockResourceService.On("BatchDelete", mock.Anything, []string{"a", "b"}).Return(nil).Once()
		s.mockProviderRepository.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), testID).Return(nil).Once()
		s.mockAuditLogger.On("Log", mock.Anything, provider.AuditKeyDelete, dummyProvider).Return(nil).Once()

		err := s.service.Delete(context.Background(), "test-provider")

		s.NoError(err)
	})
}

func (s *ServiceTestSuite) TestValidateAppeal() {
	s.Run("should return error if got error if appeal is nil", func() {
		expectedError := provider.ErrNilAppeal

		var appeal *domain.Appeal
		provider := &domain.Provider{}
		policy := &domain.Policy{}

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error if resource is nil", func() {
		expectedError := provider.ErrNilResource

		appeal := &domain.Appeal{
			Resource: nil,
		}
		provider := &domain.Provider{}
		policy := &domain.Policy{}

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if provider type is invalid", func() {
		expectedError := provider.ErrInvalidProviderType

		appeal := &domain.Appeal{
			AccountType: "invalid",
			Resource: &domain.Resource{
				ProviderType: mockProviderType,
			},
		}
		provider := &domain.Provider{
			Config: &domain.ProviderConfig{
				AllowedAccountTypes: []string{"test"},
			},
			Type: "invalid-provider-type",
		}
		policy := &domain.Policy{}

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if account type is not allowed", func() {
		appeal := &domain.Appeal{
			AccountType: "invalid",
			Resource: &domain.Resource{
				ProviderType: mockProviderType,
			},
		}
		provider := &domain.Provider{
			Config: &domain.ProviderConfig{
				AllowedAccountTypes: []string{"test"},
			},
			Type: mockProviderType,
		}
		policy := &domain.Policy{}

		expectedError := fmt.Errorf("invalid account type: %v. allowed account types for %v: %v", appeal.AccountType, mockProviderType, provider.Config.AllowedAccountTypes[0])

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if get roles failed", func() {
		expectedError := fmt.Errorf("mock error")

		appeal := &domain.Appeal{
			AccountType: "test",
			Resource: &domain.Resource{
				ProviderType: mockProviderType,
			},
		}
		provider := &domain.Provider{
			Config: &domain.ProviderConfig{
				AllowedAccountTypes: []string{"test"},
			},
			Type: mockProviderType,
		}
		policy := &domain.Policy{}

		s.mockProvider.On("GetRoles", mock.Anything, mock.Anything).Return(nil, expectedError).Once()

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if get roles not exist", func() {
		expectedError := provider.ErrInvalidRole

		appeal := &domain.Appeal{
			AccountType: "test",
			Resource: &domain.Resource{
				ProviderType: mockProviderType,
			},
			Role: "invalid-role",
		}
		provider := &domain.Provider{
			Config: &domain.ProviderConfig{
				AllowedAccountTypes: []string{"test"},
			},
			Type: mockProviderType,
		}
		policy := &domain.Policy{}

		role1 := &domain.Role{ID: "role-1"}
		s.mockProvider.On("GetRoles", mock.Anything, mock.Anything).Return([]*domain.Role{role1}, nil).Once()

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if not allow permanent access and duration option not found", func() {
		expectedError := provider.ErrOptionsDurationNotFound
		role1 := &domain.Role{ID: "role-1"}

		appeal := &domain.Appeal{
			AccountType: "test",
			Resource: &domain.Resource{
				ProviderType: mockProviderType,
			},
			Role: role1.ID,
		}
		provider := &domain.Provider{
			Config: &domain.ProviderConfig{
				AllowedAccountTypes: []string{"test"},
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess: false,
				},
			},
			Type: mockProviderType,
		}
		policy := &domain.Policy{}

		s.mockProvider.On("GetRoles", mock.Anything, mock.Anything).Return([]*domain.Role{role1}, nil).Once()

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if not allow permanent access and duration is empty", func() {
		expectedError := provider.ErrDurationIsRequired
		role1 := &domain.Role{ID: "role-1"}

		appeal := &domain.Appeal{
			AccountType: "test",
			Resource: &domain.Resource{
				ProviderType: mockProviderType,
			},
			Role: role1.ID,
			Options: &domain.AppealOptions{
				Duration: "",
			},
		}
		provider := &domain.Provider{
			Config: &domain.ProviderConfig{
				AllowedAccountTypes: []string{"test"},
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess: false,
				},
			},
			Type: mockProviderType,
		}
		policy := &domain.Policy{}

		s.mockProvider.On("GetRoles", mock.Anything, mock.Anything).Return([]*domain.Role{role1}, nil).Once()

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if not allow permanent access and duration is empty", func() {
		role1 := &domain.Role{ID: "role-1"}
		appeal := &domain.Appeal{
			AccountType: "test",
			Resource: &domain.Resource{
				ProviderType: mockProviderType,
			},
			Role: role1.ID,
			Options: &domain.AppealOptions{
				Duration: "invalid-duration",
			},
		}
		expectedError := fmt.Errorf("invalid duration: parsing duration: time: invalid duration \"invalid-duration\"")

		provider := &domain.Provider{
			Config: &domain.ProviderConfig{
				AllowedAccountTypes: []string{"test"},
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess: false,
				},
			},
			Type: mockProviderType,
		}
		policy := &domain.Policy{}

		s.mockProvider.On("GetRoles", mock.Anything, mock.Anything).Return([]*domain.Role{role1}, nil).Once()

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if not allow permanent access and duration is empty", func() {
		role1 := &domain.Role{ID: "role-1"}
		appeal := &domain.Appeal{
			AccountType: "test",
			Resource: &domain.Resource{
				ProviderType: mockProviderType,
			},
			Role: role1.ID,
			Options: &domain.AppealOptions{
				Duration: "invalid-duration",
			},
		}
		expectedError := fmt.Errorf("invalid duration: parsing duration: time: invalid duration \"invalid-duration\"")

		provider := &domain.Provider{
			Config: &domain.ProviderConfig{
				AllowedAccountTypes: []string{"test"},
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess: true,
				},
			},
			Type: mockProviderType,
		}
		policy := &domain.Policy{
			AppealConfig: &domain.PolicyAppealConfig{
				AllowPermanentAccess: false,
			},
		}

		s.mockProvider.On("GetRoles", mock.Anything, mock.Anything).Return([]*domain.Role{role1}, nil).Once()

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error when required provider parameter not present", func() {
		role1 := &domain.Role{ID: "role-1"}
		appeal := &domain.Appeal{
			AccountType: "test",
			Resource: &domain.Resource{
				ProviderType: mockProviderType,
			},
			Role: role1.ID,
			Options: &domain.AppealOptions{
				Duration: "24h",
			},
		}

		expectedError := fmt.Errorf(`parameter "%s" is required`, "username")

		provider := &domain.Provider{
			Config: &domain.ProviderConfig{
				AllowedAccountTypes: []string{"test"},
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess: false,
				},
				Parameters: []*domain.ProviderParameter{
					{
						Key:         "username",
						Label:       "Username",
						Required:    true,
						Description: "Please enter your username",
					},
				},
			},
			Type: mockProviderType,
		}
		policy := &domain.Policy{}

		s.mockProvider.On("GetRoles", mock.Anything, mock.Anything).Return([]*domain.Role{role1}, nil).Once()

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error when required policy question not present", func() {
		role1 := &domain.Role{ID: "role-1"}
		appeal := &domain.Appeal{
			AccountType: "test",
			Resource: &domain.Resource{
				ProviderType: mockProviderType,
			},
			Role: role1.ID,
			Options: &domain.AppealOptions{
				Duration: "24h",
			},
		}

		expectedError := fmt.Errorf(`question "%s" is required`, "team")

		provider := &domain.Provider{
			Config: &domain.ProviderConfig{
				AllowedAccountTypes: []string{"test"},
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess: false,
				},
			},
			Type: mockProviderType,
		}
		policy := &domain.Policy{
			AppealConfig: &domain.PolicyAppealConfig{
				Questions: []domain.Question{
					{
						Key:         "team",
						Question:    "What team are you in?",
						Required:    true,
						Description: "Please provide the name of the team you are in",
					},
				},
			},
		}

		s.mockProvider.On("GetRoles", mock.Anything, mock.Anything).Return([]*domain.Role{role1}, nil).Once()

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return nil when all valids", func() {
		role1 := &domain.Role{ID: "role-1"}
		appeal := &domain.Appeal{
			AccountType: "test",
			Resource: &domain.Resource{
				ProviderType: mockProviderType,
			},
			Role: role1.ID,
			Options: &domain.AppealOptions{
				Duration: "24h",
			},
			Details: map[string]interface{}{
				provider.ReservedDetailsKeyProviderParameters: map[string]interface{}{
					"username": "john.doe",
				},
				provider.ReservedDetailsKeyPolicyQuestions: map[string]interface{}{
					"team": "green",
				},
			},
		}

		provider := &domain.Provider{
			Config: &domain.ProviderConfig{
				AllowedAccountTypes: []string{"test"},
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess: false,
				},
				Parameters: []*domain.ProviderParameter{
					{
						Key:         "username",
						Label:       "Username",
						Required:    true,
						Description: "Please enter your username",
					},
				},
			},
			Type: mockProviderType,
		}
		policy := &domain.Policy{
			AppealConfig: &domain.PolicyAppealConfig{
				Questions: []domain.Question{
					{
						Key:         "team",
						Question:    "What team are you in?",
						Required:    true,
						Description: "Please provide the name of the team you are in",
					},
				},
			},
		}

		s.mockProvider.On("GetRoles", mock.Anything, mock.Anything).Return([]*domain.Role{role1}, nil).Once()

		actualError := s.service.ValidateAppeal(context.Background(), appeal, provider, policy)

		s.NoError(actualError)
	})
}

func (s *ServiceTestSuite) TestListAccess() {
	p := &domain.Provider{
		Type: mockProviderType,
		Config: &domain.ProviderConfig{
			AllowedAccountTypes: []string{"user"},
		},
	}
	resources := []*domain.Resource{}
	returnedAccess := domain.MapResourceAccess{
		"resource-1": []domain.AccessEntry{
			{
				AccountID:   "user@example.com",
				AccountType: "user",
				Permission:  "read",
			},
			{
				AccountID:   "user@example-sa.com",
				AccountType: "serviceAccount",
				Permission:  "read",
			},
		},
	}
	expectedAccess := domain.MapResourceAccess{
		"resource-1": []domain.AccessEntry{
			{
				AccountID:   "user@example.com",
				AccountType: "user",
				Permission:  "read",
			},
		},
	}
	s.mockProvider.EXPECT().
		ListAccess(mock.AnythingOfType("*context.emptyCtx"), *p.Config, resources).
		Return(returnedAccess, nil).Once()

	actualAccess, err := s.service.ListAccess(context.Background(), *p, resources)

	s.NoError(err)
	s.Equal(expectedAccess, actualAccess)
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
