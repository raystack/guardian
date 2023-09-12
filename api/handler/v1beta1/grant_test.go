package v1beta1_test

import (
	"context"
	"errors"
	"time"

	guardianv1beta1 "github.com/goto/guardian/api/proto/gotocompany/guardian/v1beta1"
	"github.com/goto/guardian/core/grant"
	"github.com/goto/guardian/domain"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *GrpcHandlersSuite) TestListGrants() {
	s.Run("should return list of grants on success", func() {
		s.setup()
		timeNow := time.Now()

		dummyGrants := []domain.Grant{
			{
				ID:             "test-id",
				Status:         "test-status",
				AccountID:      "test-account-id",
				AccountType:    "test-account-type",
				ResourceID:     "test-resource-id",
				Permissions:    []string{"test-permission"},
				ExpirationDate: &timeNow,
				AppealID:       "test-appeal-id",
				RevokedBy:      "test-revoked-by",
				RevokedAt:      &timeNow,
				RevokeReason:   "test-revoke-reason",
				CreatedAt:      timeNow,
				UpdatedAt:      timeNow,
				Resource: &domain.Resource{
					ID: "test-resource-id",
				},
				Appeal: &domain.Appeal{
					ID: "test-appeal-id",
				},
			},
		}
		expectedResponse := &guardianv1beta1.ListGrantsResponse{
			Grants: []*guardianv1beta1.Grant{
				{
					Id:             "test-id",
					Status:         "test-status",
					AccountId:      "test-account-id",
					AccountType:    "test-account-type",
					ResourceId:     "test-resource-id",
					Permissions:    []string{"test-permission"},
					ExpirationDate: timestamppb.New(timeNow),
					AppealId:       "test-appeal-id",
					RevokedBy:      "test-revoked-by",
					RevokedAt:      timestamppb.New(timeNow),
					RevokeReason:   "test-revoke-reason",
					CreatedAt:      timestamppb.New(timeNow),
					UpdatedAt:      timestamppb.New(timeNow),
					Resource: &guardianv1beta1.Resource{
						Id: "test-resource-id",
					},
					Appeal: &guardianv1beta1.Appeal{
						Id: "test-appeal-id",
					},
				},
			},
			Total: 1,
		}
		expectedFilter := domain.ListGrantsFilter{
			Statuses:     []string{"test-status"},
			AccountIDs:   []string{"test-account-id"},
			AccountTypes: []string{"test-account-type"},
			ResourceIDs:  []string{"test-resource-id"},
		}
		s.grantService.EXPECT().
			List(mock.AnythingOfType("*context.cancelCtx"), expectedFilter).
			Return(dummyGrants, nil).Once()
		s.grantService.EXPECT().
			GetGrantsTotalCount(mock.AnythingOfType("*context.cancelCtx"), expectedFilter).
			Return(int64(1), nil).Once()

		req := &guardianv1beta1.ListGrantsRequest{
			Statuses:     expectedFilter.Statuses,
			AccountIds:   expectedFilter.AccountIDs,
			AccountTypes: expectedFilter.AccountTypes,
			ResourceIds:  expectedFilter.ResourceIDs,
		}
		res, err := s.grpcServer.ListGrants(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.grantService.AssertExpectations(s.T())
	})

	s.Run("should return error if service returns an error", func() {
		s.setup()

		expectedError := errors.New("unexpected error")
		s.grantService.EXPECT().
			List(mock.AnythingOfType("*context.cancelCtx"), mock.AnythingOfType("domain.ListGrantsFilter")).
			Return(nil, expectedError).Once()
		s.grantService.EXPECT().
			GetGrantsTotalCount(mock.AnythingOfType("*context.cancelCtx"), mock.AnythingOfType("domain.ListGrantsFilter")).
			Return(int64(0), nil).Once()

		req := &guardianv1beta1.ListGrantsRequest{}
		res, err := s.grpcServer.ListGrants(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.grantService.AssertExpectations(s.T())
	})

	s.Run("should return error if there is an error when parsing the grant", func() {
		s.setup()

		expectedGrants := []domain.Grant{
			{
				Resource: &domain.Resource{
					Details: map[string]interface{}{
						"foo": make(chan int), // invalid value
					},
				},
			},
		}
		s.grantService.EXPECT().
			List(mock.AnythingOfType("*context.cancelCtx"), mock.AnythingOfType("domain.ListGrantsFilter")).
			Return(expectedGrants, nil).Once()
		s.grantService.EXPECT().
			GetGrantsTotalCount(mock.AnythingOfType("*context.cancelCtx"), mock.AnythingOfType("domain.ListGrantsFilter")).
			Return(int64(1), nil).Once()
		req := &guardianv1beta1.ListGrantsRequest{}
		res, err := s.grpcServer.ListGrants(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.grantService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestGetGrant() {
	s.Run("should return grant details on succes", func() {
		s.setup()
		timeNow := time.Now()

		grantID := "test-id"
		dummyGrant := &domain.Grant{
			ID:             grantID,
			Status:         "test-status",
			AccountID:      "test-account-id",
			AccountType:    "test-account-type",
			ResourceID:     "test-resource-id",
			Permissions:    []string{"test-permission"},
			ExpirationDate: &timeNow,
			AppealID:       "test-appeal-id",
			RevokedBy:      "test-revoked-by",
			RevokedAt:      &timeNow,
			RevokeReason:   "test-revoke-reason",
			CreatedAt:      timeNow,
			UpdatedAt:      timeNow,
			Resource: &domain.Resource{
				ID: "test-resource-id",
			},
			Appeal: &domain.Appeal{
				ID: "test-appeal-id",
			},
		}
		expectedResponse := &guardianv1beta1.GetGrantResponse{
			Grant: &guardianv1beta1.Grant{
				Id:             grantID,
				Status:         "test-status",
				AccountId:      "test-account-id",
				AccountType:    "test-account-type",
				ResourceId:     "test-resource-id",
				Permissions:    []string{"test-permission"},
				ExpirationDate: timestamppb.New(timeNow),
				AppealId:       "test-appeal-id",
				RevokedBy:      "test-revoked-by",
				RevokedAt:      timestamppb.New(timeNow),
				RevokeReason:   "test-revoke-reason",
				CreatedAt:      timestamppb.New(timeNow),
				UpdatedAt:      timestamppb.New(timeNow),
				Resource: &guardianv1beta1.Resource{
					Id: "test-resource-id",
				},
				Appeal: &guardianv1beta1.Appeal{
					Id: "test-appeal-id",
				},
			},
		}
		s.grantService.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), grantID).
			Return(dummyGrant, nil).Once()

		req := &guardianv1beta1.GetGrantRequest{Id: grantID}
		res, err := s.grpcServer.GetGrant(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.grantService.AssertExpectations(s.T())
	})

	s.Run("should return error if grant service returns an error", func() {
		testCases := []struct {
			name          string
			expectedError error
			expectedCode  codes.Code
		}{
			{
				"should return not found error if record not found",
				grant.ErrGrantNotFound,
				codes.NotFound,
			},
			{
				"should return internal error if there's an unexpected error",
				errors.New("unexpected error"),
				codes.Internal,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.setup()

				s.grantService.EXPECT().
					GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string")).
					Return(nil, tc.expectedError).Once()

				req := &guardianv1beta1.GetGrantRequest{Id: "test-id"}
				res, err := s.grpcServer.GetGrant(context.Background(), req)

				s.Equal(tc.expectedCode, status.Code(err))
				s.Nil(res)
				s.grantService.AssertExpectations(s.T())
			})
		}
	})

	s.Run("should return error if there is an error when parsing the grant", func() {
		s.setup()

		expectedGrant := &domain.Grant{
			Resource: &domain.Resource{
				Details: map[string]interface{}{
					"foo": make(chan int), // invalid value
				},
			},
		}
		s.grantService.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string")).
			Return(expectedGrant, nil).Once()

		req := &guardianv1beta1.GetGrantRequest{Id: "test-id"}
		res, err := s.grpcServer.GetGrant(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.grantService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestListUserRoles() {
	s.Run("should return roles", func() {
		s.setup()
		expectedResponse := &guardianv1beta1.ListUserRolesResponse{
			Roles: []string{
				"viewer",
			},
		}
		expectedUser := "test-user"
		s.grantService.EXPECT().
			ListUserRoles(mock.AnythingOfType("*context.valueCtx"), "test-user").
			Return(expectedResponse.Roles, nil).Once()

		ctx := context.WithValue(context.Background(), authEmailTestContextKey{}, expectedUser)
		req := &guardianv1beta1.ListUserRolesRequest{}
		res, err := s.grpcServer.ListUserRoles(ctx, req)

		s.Nil(err) // Check that there are no errors.
		s.grantService.AssertExpectations(s.T())
		s.Equal(expectedResponse.Roles, res.Roles)
		s.Equal(codes.OK, status.Code(err))

	})
	s.Run("should return unauthenticated user", func() {
		s.setup()

		s.grantService.EXPECT().
			ListUserRoles(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string")).
			Return(nil, nil).Once()

		req := &guardianv1beta1.ListUserRolesRequest{}
		res, err := s.grpcServer.ListUserRoles(context.Background(), req)

		s.Equal(codes.Unauthenticated, status.Code(err))
		s.Nil(res)

	})
	s.Run("should return internal error if listroles returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.grantService.EXPECT().
			ListUserRoles(mock.AnythingOfType("*context.valueCtx"), mock.Anything).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.ListUserRolesRequest{}
		ctx := context.WithValue(context.Background(), authEmailTestContextKey{}, "test-user")
		res, err := s.grpcServer.ListUserRoles(ctx, req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.grantService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestUpdateGrant() {
	s.Run("should return grant details on succes", func() {
		s.setup()

		expectedGrant := &domain.Grant{
			ID:    "test-id",
			Owner: "test-owner",
		}
		now := time.Now()
		s.grantService.EXPECT().
			Update(mock.AnythingOfType("*context.emptyCtx"), expectedGrant).
			Run(func(_a0 context.Context, g *domain.Grant) {
				g.UpdatedAt = now
			}).
			Return(nil).Once()

		req := &guardianv1beta1.UpdateGrantRequest{
			Id:    "test-id",
			Owner: "test-owner",
		}
		res, err := s.grpcServer.UpdateGrant(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedGrant.ID, res.Grant.Id)
		s.Equal(expectedGrant.Owner, res.Grant.Owner)
		s.Equal(timestamppb.New(now), res.Grant.UpdatedAt)
	})

	s.Run("should return error if grant service returns an error", func() {
		testCases := []struct {
			name          string
			expectedError error
			expectedCode  codes.Code
		}{
			{
				"should return not found error if record not found",
				grant.ErrGrantNotFound,
				codes.NotFound,
			},
			{
				"should return invalid argument error if owner is empty",
				grant.ErrEmptyOwner,
				codes.InvalidArgument,
			},
			{
				"should return internal error if there's an unexpected error",
				errors.New("unexpected error"),
				codes.Internal,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.setup()

				s.grantService.EXPECT().
					Update(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*domain.Grant")).
					Return(tc.expectedError).Once()

				req := &guardianv1beta1.UpdateGrantRequest{
					Id:    "test-id",
					Owner: "test-owner",
				}
				res, err := s.grpcServer.UpdateGrant(context.Background(), req)

				s.Equal(tc.expectedCode, status.Code(err))
				s.Nil(res)
				s.grantService.AssertExpectations(s.T())
			})
		}
	})
}

func (s *GrpcHandlersSuite) TestImportFromProvider() {
	s.Run("should return grants on success", func() {
		s.setup()
		timeNow := time.Now()

		dummyGrant := &domain.Grant{
			ID:             "test-id",
			Status:         "test-status",
			AccountID:      "test-account-id",
			AccountType:    "test-account-type",
			ResourceID:     "test-resource-id",
			Permissions:    []string{"test-permission"},
			ExpirationDate: &timeNow,
			AppealID:       "test-appeal-id",
			RevokedBy:      "test-revoked-by",
			RevokedAt:      &timeNow,
			RevokeReason:   "test-revoke-reason",
			CreatedAt:      timeNow,
			UpdatedAt:      timeNow,
			Resource: &domain.Resource{
				ID:   "test-resource-id",
				Type: "test-resource-type",
				URN:  "test-resource-urn",
			},
			Appeal: &domain.Appeal{
				ID: "test-appeal-id",
			},
		}
		expectedResponse := &guardianv1beta1.ImportGrantsFromProviderResponse{
			Grants: []*guardianv1beta1.Grant{
				{
					Id:             "test-id",
					Status:         "test-status",
					AccountId:      "test-account-id",
					AccountType:    "test-account-type",
					ResourceId:     "test-resource-id",
					Permissions:    []string{"test-permission"},
					ExpirationDate: timestamppb.New(timeNow),
					AppealId:       "test-appeal-id",
					RevokedBy:      "test-revoked-by",
					RevokedAt:      timestamppb.New(timeNow),
					RevokeReason:   "test-revoke-reason",
					CreatedAt:      timestamppb.New(timeNow),
					UpdatedAt:      timestamppb.New(timeNow),
					Resource: &guardianv1beta1.Resource{
						Id:   "test-resource-id",
						Type: "test-resource-type",
						Urn:  "test-resource-urn",
					},
					Appeal: &guardianv1beta1.Appeal{
						Id: "test-appeal-id",
					},
				},
			},
		}
		s.grantService.EXPECT().
			ImportFromProvider(mock.AnythingOfType("*context.emptyCtx"), grant.ImportFromProviderCriteria{
				ProviderID:    "test-provider-id",
				ResourceIDs:   []string{"test-resource-id"},
				ResourceTypes: []string{"test-resource-type"},
				ResourceURNs:  []string{"test-resource-urn"},
			}).
			Return([]*domain.Grant{dummyGrant}, nil).Once()

		req := &guardianv1beta1.ImportGrantsFromProviderRequest{
			ProviderId:    "test-provider-id",
			ResourceIds:   []string{"test-resource-id"},
			ResourceTypes: []string{"test-resource-type"},
			ResourceUrns:  []string{"test-resource-urn"},
		}

		res, err := s.grpcServer.ImportGrantsFromProvider(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
	})
}
