package v1beta1_test

import (
	"context"
	"errors"
	"time"

	guardianv1beta1 "github.com/raystack/guardian/api/proto/raystack/guardian/v1beta1"
	"github.com/raystack/guardian/core/provider"
	"github.com/raystack/guardian/domain"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *GrpcHandlersSuite) TestListProvider() {
	s.Run("should return list of providers on success", func() {
		s.setup()
		timeNow := time.Now()

		dummyProviders := []*domain.Provider{
			{
				ID:   "test-id",
				Type: "test-type",
				URN:  "test-urn",
				Config: &domain.ProviderConfig{
					Type:                "test-type",
					URN:                 "test-urn",
					AllowedAccountTypes: []string{"user"},
					Appeal: &domain.AppealConfig{
						AllowPermanentAccess:         true,
						AllowActiveAccessExtensionIn: "24h",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: "test-resource-type",
							Policy: &domain.PolicyConfig{
								ID:      "test-policy",
								Version: 1,
							},
							Roles: []*domain.Role{
								{
									ID:   "test-role-id",
									Name: "test-name",
								},
							},
						},
					},
				},
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			},
		}
		expectedResponse := &guardianv1beta1.ListProvidersResponse{
			Providers: []*guardianv1beta1.Provider{
				{
					Id:   "test-id",
					Type: "test-type",
					Urn:  "test-urn",
					Config: &guardianv1beta1.ProviderConfig{
						Type:                "test-type",
						Urn:                 "test-urn",
						AllowedAccountTypes: []string{"user"},
						Appeal: &guardianv1beta1.ProviderConfig_AppealConfig{
							AllowPermanentAccess:         true,
							AllowActiveAccessExtensionIn: "24h",
						},
						Resources: []*guardianv1beta1.ProviderConfig_ResourceConfig{
							{
								Type: "test-resource-type",
								Policy: &guardianv1beta1.PolicyConfig{
									Id:      "test-policy",
									Version: 1,
								},
								Roles: []*guardianv1beta1.Role{
									{
										Id:   "test-role-id",
										Name: "test-name",
									},
								},
							},
						},
					},
					CreatedAt: timestamppb.New(timeNow),
					UpdatedAt: timestamppb.New(timeNow),
				},
			},
		}
		s.providerService.EXPECT().Find(mock.AnythingOfType("context.backgroundCtx")).
			Return(dummyProviders, nil).Once()

		req := &guardianv1beta1.ListProvidersRequest{}
		res, err := s.grpcServer.ListProviders(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return error if provider service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.providerService.EXPECT().Find(mock.AnythingOfType("context.backgroundCtx")).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.ListProvidersRequest{}
		res, err := s.grpcServer.ListProviders(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there's an error when parsing the provider", func() {
		s.setup()

		expectedProviders := []*domain.Provider{
			{
				Config: &domain.ProviderConfig{
					Resources: []*domain.ResourceConfig{
						{
							Roles: []*domain.Role{
								{
									Permissions: []interface{}{make(chan int)}, // invalid json
								},
							},
						},
					},
				},
			},
		}
		s.providerService.EXPECT().Find(mock.AnythingOfType("context.backgroundCtx")).
			Return(expectedProviders, nil).Once()

		req := &guardianv1beta1.ListProvidersRequest{}
		res, err := s.grpcServer.ListProviders(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestGetProvider() {
	s.Run("should return provider details on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedProvider := &domain.Provider{
			ID:   "test-id",
			Type: "test-type",
			URN:  "test-urn",
			Config: &domain.ProviderConfig{
				Type:                "test-type",
				URN:                 "test-urn",
				AllowedAccountTypes: []string{"user"},
				Resources: []*domain.ResourceConfig{
					{
						Type: "test-resource-type",
						Policy: &domain.PolicyConfig{
							ID:      "test-policy",
							Version: 1,
						},
						Roles: []*domain.Role{
							{
								ID:   "test-role-id",
								Name: "test-name",
							},
						},
					},
				},
			},
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		expectedResponse := &guardianv1beta1.GetProviderResponse{
			Provider: &guardianv1beta1.Provider{
				Id:   "test-id",
				Type: "test-type",
				Urn:  "test-urn",
				Config: &guardianv1beta1.ProviderConfig{
					Type:                "test-type",
					Urn:                 "test-urn",
					AllowedAccountTypes: []string{"user"},
					Resources: []*guardianv1beta1.ProviderConfig_ResourceConfig{
						{
							Type: "test-resource-type",
							Policy: &guardianv1beta1.PolicyConfig{
								Id:      "test-policy",
								Version: 1,
							},
							Roles: []*guardianv1beta1.Role{
								{
									Id:   "test-role-id",
									Name: "test-name",
								},
							},
						},
					},
				},
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}
		s.providerService.EXPECT().GetByID(mock.AnythingOfType("context.backgroundCtx"), expectedProvider.ID).Return(expectedProvider, nil).Once()

		req := &guardianv1beta1.GetProviderRequest{
			Id: expectedProvider.ID,
		}
		res, err := s.grpcServer.GetProvider(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return not found error if provider not found", func() {
		s.setup()

		expectedError := provider.ErrRecordNotFound
		s.providerService.EXPECT().GetByID(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.GetProviderRequest{}
		res, err := s.grpcServer.GetProvider(context.Background(), req)

		s.Equal(codes.NotFound, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if provider service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.providerService.EXPECT().GetByID(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.GetProviderRequest{}
		res, err := s.grpcServer.GetProvider(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there's an error when parsing the provider", func() {
		s.setup()

		expectedProvider := &domain.Provider{
			Config: &domain.ProviderConfig{
				Credentials: make(chan int), // invalid json
			},
		}
		s.providerService.EXPECT().GetByID(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
			Return(expectedProvider, nil).Once()

		req := &guardianv1beta1.GetProviderRequest{}
		res, err := s.grpcServer.GetProvider(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestGetProviderTypes() {
	s.Run("should return provider types on success", func() {
		s.setup()

		expectedProviderTypes := []domain.ProviderType{
			{
				Name:          "test-name",
				ResourceTypes: []string{"test-type-1"},
			},
		}
		expectedResponse := &guardianv1beta1.GetProviderTypesResponse{
			ProviderTypes: []*guardianv1beta1.ProviderType{
				{
					Name:          "test-name",
					ResourceTypes: []string{"test-type-1"},
				},
			},
		}
		s.providerService.EXPECT().GetTypes(mock.AnythingOfType("context.backgroundCtx")).
			Return(expectedProviderTypes, nil).Once()

		req := &guardianv1beta1.GetProviderTypesRequest{}
		res, err := s.grpcServer.GetProviderTypes(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if provider service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.providerService.EXPECT().GetTypes(mock.AnythingOfType("context.backgroundCtx")).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.GetProviderTypesRequest{}
		res, err := s.grpcServer.GetProviderTypes(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestCreateProvider() {
	s.Run("should return newly created provider on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedID := "test-id"
		expectedProvider := &domain.Provider{
			Type: "test-type",
			URN:  "test-urn",
			Config: &domain.ProviderConfig{
				Type:                "test-type",
				URN:                 "test-urn",
				AllowedAccountTypes: []string{"user"},
				Resources: []*domain.ResourceConfig{
					{
						Type: "test-resource-type",
						Policy: &domain.PolicyConfig{
							ID:      "test-policy",
							Version: 1,
						},
						Roles: []*domain.Role{
							{
								ID:   "test-role-id",
								Name: "test-name",
							},
						},
					},
				},
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess:         true,
					AllowActiveAccessExtensionIn: "24h",
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
		}
		expectedResponse := &guardianv1beta1.CreateProviderResponse{
			Provider: &guardianv1beta1.Provider{
				Id:   expectedID,
				Type: "test-type",
				Urn:  "test-urn",
				Config: &guardianv1beta1.ProviderConfig{
					Type:                "test-type",
					Urn:                 "test-urn",
					AllowedAccountTypes: []string{"user"},
					Resources: []*guardianv1beta1.ProviderConfig_ResourceConfig{
						{
							Type: "test-resource-type",
							Policy: &guardianv1beta1.PolicyConfig{
								Id:      "test-policy",
								Version: 1,
							},
							Roles: []*guardianv1beta1.Role{
								{
									Id:   "test-role-id",
									Name: "test-name",
								},
							},
						},
					},
					Appeal: &guardianv1beta1.ProviderConfig_AppealConfig{
						AllowPermanentAccess:         true,
						AllowActiveAccessExtensionIn: "24h",
					},
					Parameters: []*guardianv1beta1.ProviderConfig_ProviderParameter{
						{
							Key:         "username",
							Label:       "Username",
							Required:    true,
							Description: "Please enter your username",
						},
					},
				},
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}
		s.providerService.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), expectedProvider).Return(nil).
			Run(func(_a0 context.Context, _a1 *domain.Provider) {
				_a1.ID = expectedID
				_a1.CreatedAt = timeNow
				_a1.UpdatedAt = timeNow
			}).Once()

		req := &guardianv1beta1.CreateProviderRequest{
			Config: &guardianv1beta1.ProviderConfig{
				Type:                "test-type",
				Urn:                 "test-urn",
				AllowedAccountTypes: []string{"user"},
				Resources: []*guardianv1beta1.ProviderConfig_ResourceConfig{
					{
						Type: "test-resource-type",
						Policy: &guardianv1beta1.PolicyConfig{
							Id:      "test-policy",
							Version: 1,
						},
						Roles: []*guardianv1beta1.Role{
							{
								Id:   "test-role-id",
								Name: "test-name",
							},
						},
					},
				},
				Appeal: &guardianv1beta1.ProviderConfig_AppealConfig{
					AllowPermanentAccess:         true,
					AllowActiveAccessExtensionIn: "24h",
				},
				Parameters: []*guardianv1beta1.ProviderConfig_ProviderParameter{
					{
						Key:         "username",
						Label:       "Username",
						Required:    true,
						Description: "Please enter your username",
					},
				},
			},
		}
		res, err := s.grpcServer.CreateProvider(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if provider service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.providerService.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*domain.Provider")).Return(expectedError).Once()

		req := &guardianv1beta1.CreateProviderRequest{}
		res, err := s.grpcServer.CreateProvider(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there's an error when parsing the provider", func() {
		s.setup()

		expectedProvider := &domain.Provider{
			Config: &domain.ProviderConfig{
				Credentials: make(chan int), // invalid json
			},
		}
		s.providerService.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*domain.Provider")).Return(nil).
			Run(func(_a0 context.Context, _a1 *domain.Provider) {
				*_a1 = *expectedProvider
			}).Once()

		req := &guardianv1beta1.CreateProviderRequest{}
		res, err := s.grpcServer.CreateProvider(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestUpdatedProvider() {
	s.Run("should return newly updated provider on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedID := "test-id"
		expectedProvider := &domain.Provider{
			ID:   expectedID,
			Type: "test-type",
			URN:  "test-urn",
			Config: &domain.ProviderConfig{
				Type:                "test-type",
				URN:                 "test-urn",
				AllowedAccountTypes: []string{"user"},
				Resources: []*domain.ResourceConfig{
					{
						Type: "test-resource-type",
						Policy: &domain.PolicyConfig{
							ID:      "test-policy",
							Version: 1,
						},
						Roles: []*domain.Role{
							{
								ID:   "test-role-id",
								Name: "test-name",
							},
						},
					},
				},
			},
		}
		expectedResponse := &guardianv1beta1.UpdateProviderResponse{
			Provider: &guardianv1beta1.Provider{
				Id:   expectedID,
				Type: "test-type",
				Urn:  "test-urn",
				Config: &guardianv1beta1.ProviderConfig{
					Type:                "test-type",
					Urn:                 "test-urn",
					AllowedAccountTypes: []string{"user"},
					Resources: []*guardianv1beta1.ProviderConfig_ResourceConfig{
						{
							Type: "test-resource-type",
							Policy: &guardianv1beta1.PolicyConfig{
								Id:      "test-policy",
								Version: 1,
							},
							Roles: []*guardianv1beta1.Role{
								{
									Id:   "test-role-id",
									Name: "test-name",
								},
							},
						},
					},
				},
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}
		s.providerService.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), expectedProvider).Return(nil).
			Run(func(_a0 context.Context, _a1 *domain.Provider) {
				_a1.CreatedAt = timeNow
				_a1.UpdatedAt = timeNow
			}).Once()

		req := &guardianv1beta1.UpdateProviderRequest{
			Id: expectedID,
			Config: &guardianv1beta1.ProviderConfig{
				Type:                "test-type",
				Urn:                 "test-urn",
				AllowedAccountTypes: []string{"user"},
				Resources: []*guardianv1beta1.ProviderConfig_ResourceConfig{
					{
						Type: "test-resource-type",
						Policy: &guardianv1beta1.PolicyConfig{
							Id:      "test-policy",
							Version: 1,
						},
						Roles: []*guardianv1beta1.Role{
							{
								Id:   "test-role-id",
								Name: "test-name",
							},
						},
					},
				},
			},
		}
		res, err := s.grpcServer.UpdateProvider(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if provider service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.providerService.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*domain.Provider")).Return(expectedError).Once()

		req := &guardianv1beta1.UpdateProviderRequest{}
		res, err := s.grpcServer.UpdateProvider(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there's an error when parsing the provider", func() {
		s.setup()

		expectedProvider := &domain.Provider{
			Config: &domain.ProviderConfig{
				Credentials: make(chan int), // invalid json
			},
		}
		s.providerService.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*domain.Provider")).Return(nil).
			Run(func(_a0 context.Context, _a1 *domain.Provider) {
				*_a1 = *expectedProvider
			}).Once()

		req := &guardianv1beta1.UpdateProviderRequest{}
		res, err := s.grpcServer.UpdateProvider(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestDeleteProvider() {
	s.Run("should return no error on success", func() {
		s.setup()

		expectedResponse := &guardianv1beta1.DeleteProviderResponse{}
		expectedID := "test-id"
		s.providerService.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), expectedID).Return(nil).Once()

		req := &guardianv1beta1.DeleteProviderRequest{
			Id: expectedID,
		}
		res, err := s.grpcServer.DeleteProvider(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return not found error if provider not found", func() {
		s.setup()

		expectedError := provider.ErrRecordNotFound
		s.providerService.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
			Return(expectedError).Once()

		req := &guardianv1beta1.DeleteProviderRequest{}
		res, err := s.grpcServer.DeleteProvider(context.Background(), req)

		s.Equal(codes.NotFound, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if provider service returns an unknown error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.providerService.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
			Return(expectedError).Once()

		req := &guardianv1beta1.DeleteProviderRequest{}
		res, err := s.grpcServer.DeleteProvider(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestListRoles() {
	s.Run("should return list of roles on success", func() {
		s.setup()

		expectedProviderID := "test-provider"
		expectedResourceType := "test-resource-type"
		expectedRoles := []*domain.Role{
			{
				ID:          "test-id",
				Name:        "test-name",
				Description: "test-description",
				Permissions: []interface{}{"test-permission"},
			},
		}
		expectedPermission, err := structpb.NewValue("test-permission")
		s.Require().NoError(err)
		expectedResponse := &guardianv1beta1.ListRolesResponse{
			Roles: []*guardianv1beta1.Role{
				{
					Id:          "test-id",
					Name:        "test-name",
					Description: "test-description",
					Permissions: []*structpb.Value{expectedPermission},
				},
			},
		}
		s.providerService.EXPECT().GetRoles(mock.AnythingOfType("context.backgroundCtx"), expectedProviderID, expectedResourceType).
			Return(expectedRoles, nil).Once()

		req := &guardianv1beta1.ListRolesRequest{
			Id:           expectedProviderID,
			ResourceType: expectedResourceType,
		}
		res, err := s.grpcServer.ListRoles(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if provider service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.providerService.EXPECT().GetRoles(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.ListRolesRequest{}
		res, err := s.grpcServer.ListRoles(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there's an error when parsing the role", func() {
		s.setup()

		invalidRoles := []*domain.Role{
			{
				Permissions: []interface{}{
					make(chan int), //int
				},
			},
		}
		s.providerService.EXPECT().GetRoles(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
			Return(invalidRoles, nil).Once()

		req := &guardianv1beta1.ListRolesRequest{}
		res, err := s.grpcServer.ListRoles(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.providerService.AssertExpectations(s.T())
	})
}
