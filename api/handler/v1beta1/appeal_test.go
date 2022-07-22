package v1beta1_test

import (
	"context"
	"errors"
	"time"

	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *GrpcHandlersSuite) TestListUserAppeals() {
	s.Run("should return list of user appeals on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedUser := "test-user"
		expectedFilters := &domain.ListAppealsFilter{
			CreatedBy:     expectedUser,
			Statuses:      []string{"active", "pending"},
			Role:          "test-role",
			ProviderTypes: []string{"test-provider-type"},
			ProviderURNs:  []string{"test-provider-urn"},
			ResourceTypes: []string{"test-resource-type"},
			ResourceURNs:  []string{"test-resource-urn"},
			OrderBy:       []string{"test-order"},
		}
		expectedAppeals := []*domain.Appeal{
			{
				ID:         "test-id",
				ResourceID: "test-resource-id",
				Resource: &domain.Resource{
					ID: "test-resource-id",
				},
				PolicyID:      "test-policy-id",
				PolicyVersion: 1,
				Status:        "active",
				AccountID:     "test-account-id",
				AccountType:   "test-account-type",
				CreatedBy:     expectedUser,
				Creator: map[string]interface{}{
					"foo": "bar",
				},
				Role: "test-role",
				Options: &domain.AppealOptions{
					Duration:       "24h",
					ExpirationDate: &timeNow,
				},
				Details: map[string]interface{}{
					"foo": "bar",
				},
				Approvals: []*domain.Approval{
					{
						ID:        "test-approval-id",
						Name:      "test-approval-name",
						Status:    "pending",
						Approvers: []string{"test-approver"},
						CreatedAt: timeNow,
						UpdatedAt: timeNow,
					},
				},
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			},
		}
		expectedCreator, err := structpb.NewValue(map[string]interface{}{
			"foo": "bar",
		})
		s.Require().NoError(err)
		expectedDetails, err := structpb.NewStruct(map[string]interface{}{
			"foo": "bar",
		})
		s.Require().NoError(err)
		expectedResponse := &guardianv1beta1.ListUserAppealsResponse{
			Appeals: []*guardianv1beta1.Appeal{
				{
					Id:         "test-id",
					ResourceId: "test-resource-id",
					Resource: &guardianv1beta1.Resource{
						Id: "test-resource-id",
					},
					PolicyId:      "test-policy-id",
					PolicyVersion: 1,
					Status:        "active",
					AccountId:     "test-account-id",
					AccountType:   "test-account-type",
					CreatedBy:     expectedUser,
					Creator:       expectedCreator,
					Role:          "test-role",
					Options: &guardianv1beta1.AppealOptions{
						Duration:       "24h",
						ExpirationDate: timestamppb.New(timeNow),
					},
					Details: expectedDetails,
					Approvals: []*guardianv1beta1.Approval{
						{
							Id:        "test-approval-id",
							Name:      "test-approval-name",
							Status:    "pending",
							Approvers: []string{"test-approver"},
							CreatedAt: timestamppb.New(timeNow),
							UpdatedAt: timestamppb.New(timeNow),
						},
					},
					CreatedAt: timestamppb.New(timeNow),
					UpdatedAt: timestamppb.New(timeNow),
				},
			},
		}
		s.appealService.EXPECT().Find(mock.AnythingOfType("*context.valueCtx"), expectedFilters).
			Return(expectedAppeals, nil).Once()

		req := &guardianv1beta1.ListUserAppealsRequest{
			Statuses:      []string{"active", "pending"},
			Role:          "test-role",
			ProviderTypes: []string{"test-provider-type"},
			ProviderUrns:  []string{"test-provider-urn"},
			ResourceTypes: []string{"test-resource-type"},
			ResourceUrns:  []string{"test-resource-urn"},
			OrderBy:       []string{"test-order"},
		}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: expectedUser,
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.ListUserAppeals(ctx, req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return unathenticated error if request is not authenticated", func() {
		s.setup()

		req := &guardianv1beta1.ListUserAppealsRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.ListUserAppeals(ctx, req)

		s.Equal(codes.Unauthenticated, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if appeal service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.appealService.EXPECT().Find(mock.AnythingOfType("*context.valueCtx"), mock.Anything).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.ListUserAppealsRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: "test-user",
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.ListUserAppeals(ctx, req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there is an error when parsing appeal", func() {
		s.setup()

		invalidAppeals := []*domain.Appeal{
			{
				Creator: map[string]interface{}{
					"foo": make(chan int),
				},
			},
		}
		s.appealService.EXPECT().Find(mock.AnythingOfType("*context.valueCtx"), mock.Anything).
			Return(invalidAppeals, nil).Once()

		req := &guardianv1beta1.ListUserAppealsRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: "test-user",
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.ListUserAppeals(ctx, req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestListAppeals() {
	s.Run("should return list of appeals on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedUser := "user@example.com"
		expectedFilters := &domain.ListAppealsFilter{
			AccountID:     []string{expectedUser},
			Statuses:      []string{"active", "pending"},
			Role:          "test-role",
			ProviderTypes: []string{"test-provider-type"},
			ProviderURNs:  []string{"test-provider-urn"},
			ResourceTypes: []string{"test-resource-type"},
			ResourceURNs:  []string{"test-resource-urn"},
			OrderBy:       []string{"test-order"},
		}
		expectedAppeals := []*domain.Appeal{
			{
				ID:         "test-id",
				ResourceID: "test-resource-id",
				Resource: &domain.Resource{
					ID: "test-resource-id",
				},
				PolicyID:      "test-policy-id",
				PolicyVersion: 1,
				Status:        "active",
				AccountID:     expectedUser,
				AccountType:   "test-account-type",
				CreatedBy:     expectedUser,
				Creator: map[string]interface{}{
					"foo": "bar",
				},
				Role: "test-role",
				Options: &domain.AppealOptions{
					Duration:       "24h",
					ExpirationDate: &timeNow,
				},
				Details: map[string]interface{}{
					"foo": "bar",
				},
				Approvals: []*domain.Approval{
					{
						ID:        "test-approval-id",
						Name:      "test-approval-name",
						Status:    "pending",
						Approvers: []string{"test-approver"},
						CreatedAt: timeNow,
						UpdatedAt: timeNow,
					},
				},
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			},
		}
		expectedCreator, err := structpb.NewValue(map[string]interface{}{
			"foo": "bar",
		})
		s.Require().NoError(err)
		expectedDetails, err := structpb.NewStruct(map[string]interface{}{
			"foo": "bar",
		})
		s.Require().NoError(err)
		expectedResponse := &guardianv1beta1.ListAppealsResponse{
			Appeals: []*guardianv1beta1.Appeal{
				{
					Id:         "test-id",
					ResourceId: "test-resource-id",
					Resource: &guardianv1beta1.Resource{
						Id: "test-resource-id",
					},
					PolicyId:      "test-policy-id",
					PolicyVersion: 1,
					Status:        "active",
					AccountId:     expectedUser,
					AccountType:   "test-account-type",
					CreatedBy:     expectedUser,
					Creator:       expectedCreator,
					Role:          "test-role",
					Options: &guardianv1beta1.AppealOptions{
						Duration:       "24h",
						ExpirationDate: timestamppb.New(timeNow),
					},
					Details: expectedDetails,
					Approvals: []*guardianv1beta1.Approval{
						{
							Id:        "test-approval-id",
							Name:      "test-approval-name",
							Status:    "pending",
							Approvers: []string{"test-approver"},
							CreatedAt: timestamppb.New(timeNow),
							UpdatedAt: timestamppb.New(timeNow),
						},
					},
					CreatedAt: timestamppb.New(timeNow),
					UpdatedAt: timestamppb.New(timeNow),
				},
			},
		}
		s.appealService.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx"), expectedFilters).
			Return(expectedAppeals, nil).Once()

		req := &guardianv1beta1.ListAppealsRequest{
			AccountId:     expectedUser,
			Statuses:      []string{"active", "pending"},
			Role:          "test-role",
			ProviderTypes: []string{"test-provider-type"},
			ProviderUrns:  []string{"test-provider-urn"},
			ResourceTypes: []string{"test-resource-type"},
			ResourceUrns:  []string{"test-resource-urn"},
			OrderBy:       []string{"test-order"},
		}
		res, err := s.grpcServer.ListAppeals(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if appeal service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.appealService.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.ListAppealsRequest{}
		res, err := s.grpcServer.ListAppeals(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there is an error when parsing appeal", func() {
		s.setup()

		invalidAppeals := []*domain.Appeal{
			{
				Creator: map[string]interface{}{
					"foo": make(chan int),
				},
			},
		}
		s.appealService.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(invalidAppeals, nil).Once()

		req := &guardianv1beta1.ListAppealsRequest{}
		res, err := s.grpcServer.ListAppeals(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestCreateAppeal() {
	s.Run("should return list of appeal on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedUser := "user@example.com"
		expectedResource := &domain.Resource{
			ID:           "test-resource-id",
			ProviderType: "test-provider-type",
			ProviderURN:  "test-provider-urn",
			Type:         "test-resource-type",
			URN:          "test-resource-urn",
			Name:         "test-name",
		}
		expectedApproval := &domain.Approval{
			ID:            "test-approval-id",
			Name:          "test-approval-step",
			Status:        "pending",
			AppealID:      "test-id",
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,
			Approvers:     []string{"approver@example.com"},
			CreatedAt:     timeNow,
			UpdatedAt:     timeNow,
		}
		expectedOptions := &domain.AppealOptions{
			ExpirationDate: &timeNow,
			Duration:       "24h",
		}
		expectedAppeals := []*domain.Appeal{
			{
				AccountID:   expectedUser,
				AccountType: "user",
				CreatedBy:   expectedUser,
				ResourceID:  "test-resource-id",
				Role:        "test-role",
				Options: &domain.AppealOptions{
					Duration: "24h",
				},
				Details: map[string]interface{}{
					"foo": "bar",
				},
			},
		}
		expectedDetails, err := structpb.NewStruct(map[string]interface{}{
			"foo": "bar",
		})
		s.Require().NoError(err)
		expectedResponse := &guardianv1beta1.CreateAppealResponse{
			Appeals: []*guardianv1beta1.Appeal{
				{
					Id:            "test-id",
					ResourceId:    "test-resource-id",
					AccountId:     expectedUser,
					AccountType:   "user",
					CreatedBy:     expectedUser,
					Role:          "test-role",
					PolicyId:      "test-policy-id",
					PolicyVersion: 1,
					Status:        "pending",
					Resource: &guardianv1beta1.Resource{
						Id:           "test-resource-id",
						ProviderType: "test-provider-type",
						ProviderUrn:  "test-provider-urn",
						Type:         "test-resource-type",
						Urn:          "test-resource-urn",
						Name:         "test-name",
					},
					Approvals: []*guardianv1beta1.Approval{
						{
							Id:            "test-approval-id",
							Name:          "test-approval-step",
							Status:        "pending",
							AppealId:      "test-id",
							PolicyId:      "test-policy-id",
							PolicyVersion: 1,
							Approvers:     []string{"approver@example.com"},
							CreatedAt:     timestamppb.New(timeNow),
							UpdatedAt:     timestamppb.New(timeNow),
						},
					},
					Options: &guardianv1beta1.AppealOptions{
						ExpirationDate: timestamppb.New(timeNow),
						Duration:       "24h",
					},
					Details:   expectedDetails,
					CreatedAt: timestamppb.New(timeNow),
					UpdatedAt: timestamppb.New(timeNow),
				},
			},
		}
		s.appealService.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), expectedAppeals).
			Run(func(_a0 context.Context, _a1 []*domain.Appeal, _a2 ...appeal.CreateAppealOption) {
				for _, a := range _a1 {
					a.ID = "test-id"
					a.Resource = expectedResource
					a.PolicyID = "test-policy-id"
					a.PolicyVersion = 1
					a.Status = "pending"
					a.Approvals = []*domain.Approval{expectedApproval}
					a.CreatedAt = timeNow
					a.UpdatedAt = timeNow
					a.Options = expectedOptions
				}
			}).
			Return(nil).Once()

		reqOptions, err := structpb.NewStruct(map[string]interface{}{
			"duration": "24h",
		})
		s.Require().NoError(err)

		req := &guardianv1beta1.CreateAppealRequest{
			AccountId:   expectedUser,
			AccountType: "user",
			Resources: []*guardianv1beta1.CreateAppealRequest_Resource{
				{
					Id:      "test-resource-id",
					Role:    "test-role",
					Options: reqOptions,
					Details: expectedDetails,
				},
			},
		}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: expectedUser,
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.CreateAppeal(ctx, req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return unauthenticated error if request is unauthenticated", func() {
		s.setup()

		req := &guardianv1beta1.CreateAppealRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.CreateAppeal(ctx, req)

		s.Equal(codes.Unauthenticated, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if payload is invalid", func() {
		// TODO: find a way to simulate invalid request payload
	})

	s.Run("should return duplicate error if appeal already exists", func() {
		s.setup()

		s.appealService.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), mock.Anything).Return(appeal.ErrAppealDuplicate).Once()

		req := &guardianv1beta1.CreateAppealRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: "user@example.com",
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.CreateAppeal(ctx, req)

		s.Equal(codes.AlreadyExists, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if appeal service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.appealService.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), mock.Anything).Return(expectedError).Once()

		req := &guardianv1beta1.CreateAppealRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: "user@example.com",
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.CreateAppeal(ctx, req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if failed to parse appeal", func() {
		s.setup()

		invalidAppeal := &domain.Appeal{

			Creator: map[string]interface{}{
				"foo": make(chan int),
			},
		}
		s.appealService.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), mock.Anything).
			Run(func(_a0 context.Context, _a1 []*domain.Appeal, _a2 ...appeal.CreateAppealOption) {
				*_a1[0] = *invalidAppeal
			}).
			Return(nil).Once()

		req := &guardianv1beta1.CreateAppealRequest{Resources: make([]*guardianv1beta1.CreateAppealRequest_Resource, 1)}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: "user@example.com",
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.CreateAppeal(ctx, req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestGetAppeal() {
	s.Run("should return appeal details on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedID := "test-appeal-id"
		expectedAppeal := &domain.Appeal{
			ID:         expectedID,
			ResourceID: "test-resource-id",
			Resource: &domain.Resource{
				ID: "test-resource-id",
			},
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,
			Status:        "active",
			AccountID:     "test-account-id",
			AccountType:   "test-account-type",
			CreatedBy:     "test-created-by",
			Creator: map[string]interface{}{
				"foo": "bar",
			},
			Role: "test-role",
			Options: &domain.AppealOptions{
				Duration:       "24h",
				ExpirationDate: &timeNow,
			},
			Details: map[string]interface{}{
				"foo": "bar",
			},
			Approvals: []*domain.Approval{
				{
					ID:        "test-approval-id",
					Name:      "test-approval-name",
					Status:    "pending",
					Approvers: []string{"test-approver"},
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				},
			},
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		expectedCreator, err := structpb.NewValue(map[string]interface{}{
			"foo": "bar",
		})
		s.Require().NoError(err)
		expectedDetails, err := structpb.NewStruct(map[string]interface{}{
			"foo": "bar",
		})
		s.Require().NoError(err)
		expectedResponse := &guardianv1beta1.GetAppealResponse{
			Appeal: &guardianv1beta1.Appeal{
				Id:         expectedID,
				ResourceId: "test-resource-id",
				Resource: &guardianv1beta1.Resource{
					Id: "test-resource-id",
				},
				PolicyId:      "test-policy-id",
				PolicyVersion: 1,
				Status:        "active",
				AccountId:     "test-account-id",
				AccountType:   "test-account-type",
				CreatedBy:     "test-created-by",
				Creator:       expectedCreator,
				Role:          "test-role",
				Options: &guardianv1beta1.AppealOptions{
					Duration:       "24h",
					ExpirationDate: timestamppb.New(timeNow),
				},
				Details: expectedDetails,
				Approvals: []*guardianv1beta1.Approval{
					{
						Id:        "test-approval-id",
						Name:      "test-approval-name",
						Status:    "pending",
						Approvers: []string{"test-approver"},
						CreatedAt: timestamppb.New(timeNow),
						UpdatedAt: timestamppb.New(timeNow),
					},
				},
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}
		s.appealService.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), expectedID).Return(expectedAppeal, nil).Once()

		req := &guardianv1beta1.GetAppealRequest{
			Id: expectedID,
		}
		res, err := s.grpcServer.GetAppeal(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if appeal service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.appealService.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.GetAppealRequest{}
		res, err := s.grpcServer.GetAppeal(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return not found error if appeal not found", func() {
		s.setup()

		s.appealService.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(nil, nil).Once()

		req := &guardianv1beta1.GetAppealRequest{}
		res, err := s.grpcServer.GetAppeal(context.Background(), req)

		s.Equal(codes.NotFound, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if failed to parse appeal", func() {
		s.setup()

		invalidAppeal := &domain.Appeal{
			Creator: map[string]interface{}{
				"foo": make(chan int),
			},
		}
		s.appealService.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(invalidAppeal, nil).Once()

		req := &guardianv1beta1.GetAppealRequest{}
		res, err := s.grpcServer.GetAppeal(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestCancelAppeal() {
	s.Run("should return appeal details on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedID := "test-appeal-id"
		expectedAppeal := &domain.Appeal{
			ID:         expectedID,
			ResourceID: "test-resource-id",
			Resource: &domain.Resource{
				ID: "test-resource-id",
			},
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,
			Status:        "active",
			AccountID:     "test-account-id",
			AccountType:   "test-account-type",
			CreatedBy:     "test-created-by",
			Creator: map[string]interface{}{
				"foo": "bar",
			},
			Role: "test-role",
			Options: &domain.AppealOptions{
				Duration:       "24h",
				ExpirationDate: &timeNow,
			},
			Details: map[string]interface{}{
				"foo": "bar",
			},
			Approvals: []*domain.Approval{
				{
					ID:        "test-approval-id",
					Name:      "test-approval-name",
					Status:    "pending",
					Approvers: []string{"test-approver"},
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				},
			},
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		expectedCreator, err := structpb.NewValue(map[string]interface{}{
			"foo": "bar",
		})
		s.Require().NoError(err)
		expectedDetails, err := structpb.NewStruct(map[string]interface{}{
			"foo": "bar",
		})
		s.Require().NoError(err)
		expectedResponse := &guardianv1beta1.CancelAppealResponse{
			Appeal: &guardianv1beta1.Appeal{
				Id:         expectedID,
				ResourceId: "test-resource-id",
				Resource: &guardianv1beta1.Resource{
					Id: "test-resource-id",
				},
				PolicyId:      "test-policy-id",
				PolicyVersion: 1,
				Status:        "active",
				AccountId:     "test-account-id",
				AccountType:   "test-account-type",
				CreatedBy:     "test-created-by",
				Creator:       expectedCreator,
				Role:          "test-role",
				Options: &guardianv1beta1.AppealOptions{
					Duration:       "24h",
					ExpirationDate: timestamppb.New(timeNow),
				},
				Details: expectedDetails,
				Approvals: []*guardianv1beta1.Approval{
					{
						Id:        "test-approval-id",
						Name:      "test-approval-name",
						Status:    "pending",
						Approvers: []string{"test-approver"},
						CreatedAt: timestamppb.New(timeNow),
						UpdatedAt: timestamppb.New(timeNow),
					},
				},
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}
		s.appealService.EXPECT().Cancel(mock.AnythingOfType("*context.emptyCtx"), expectedID).Return(expectedAppeal, nil).Once()

		req := &guardianv1beta1.CancelAppealRequest{
			Id: expectedID,
		}
		res, err := s.grpcServer.CancelAppeal(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return error if appeal service returns an error", func() {
		testCases := []struct {
			name               string
			expectedError      error
			expectedStatusCode codes.Code
		}{
			{
				"should return not found error if appeal not found",
				appeal.ErrAppealNotFound,
				codes.NotFound,
			},
			{
				"should return invalid error if appeal status already cancelled",
				appeal.ErrAppealStatusCanceled,
				codes.InvalidArgument,
			},
			{
				"should return invalid error if appeal status already approved",
				appeal.ErrAppealStatusApproved,
				codes.InvalidArgument,
			},
			{
				"should return invalid error if appeal status already rejected",
				appeal.ErrAppealStatusRejected,
				codes.InvalidArgument,
			},
			{
				"should return invalid error if appeal status already terminated",
				appeal.ErrAppealStatusTerminated,
				codes.InvalidArgument,
			},
			{
				"should return invalid error if appeal status unrecognized",
				appeal.ErrAppealStatusUnrecognized,
				codes.InvalidArgument,
			},
			{
				"should return internal error if appeal service returns a unexpected error",
				errors.New("unexpected error"),
				codes.Internal,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.setup()

				s.appealService.EXPECT().Cancel(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
					Return(nil, tc.expectedError).Once()

				req := &guardianv1beta1.CancelAppealRequest{}
				res, err := s.grpcServer.CancelAppeal(context.Background(), req)

				s.Equal(tc.expectedStatusCode, status.Code(err))
				s.Nil(res)
				s.appealService.AssertExpectations(s.T())
			})
		}
	})

	s.Run("should return internal error if failed to parse appeal", func() {
		s.setup()

		invalidAppeal := &domain.Appeal{
			Creator: map[string]interface{}{
				"foo": make(chan int),
			},
		}
		s.appealService.EXPECT().Cancel(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(invalidAppeal, nil).Once()

		req := &guardianv1beta1.CancelAppealRequest{}
		res, err := s.grpcServer.CancelAppeal(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestRevokeAppeal() {
	s.Run("should return appeal details on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedUser := "user@example.com"
		expectedID := "test-appeal-id"
		expectedReason := "test-reason"
		expectedAppeal := &domain.Appeal{
			ID:         expectedID,
			ResourceID: "test-resource-id",
			Resource: &domain.Resource{
				ID: "test-resource-id",
			},
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,
			Status:        "terminated",
			AccountID:     "test-account-id",
			AccountType:   "test-account-type",
			CreatedBy:     "test-created-by",
			Creator: map[string]interface{}{
				"foo": "bar",
			},
			Role: "test-role",
			Options: &domain.AppealOptions{
				Duration:       "24h",
				ExpirationDate: &timeNow,
			},
			Details: map[string]interface{}{
				"foo": "bar",
			},
			Approvals: []*domain.Approval{
				{
					ID:        "test-approval-id",
					Name:      "test-approval-name",
					Status:    "pending",
					Approvers: []string{"test-approver"},
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				},
			},
			RevokedBy:    expectedUser,
			RevokeReason: expectedReason,
			CreatedAt:    timeNow,
			UpdatedAt:    timeNow,
			RevokedAt:    timeNow,
		}
		expectedCreator, err := structpb.NewValue(map[string]interface{}{
			"foo": "bar",
		})
		s.Require().NoError(err)
		expectedDetails, err := structpb.NewStruct(map[string]interface{}{
			"foo": "bar",
		})
		s.Require().NoError(err)
		expectedResponse := &guardianv1beta1.RevokeAppealResponse{
			Appeal: &guardianv1beta1.Appeal{
				Id:         expectedID,
				ResourceId: "test-resource-id",
				Resource: &guardianv1beta1.Resource{
					Id: "test-resource-id",
				},
				PolicyId:      "test-policy-id",
				PolicyVersion: 1,
				Status:        "terminated",
				AccountId:     "test-account-id",
				AccountType:   "test-account-type",
				CreatedBy:     "test-created-by",
				Creator:       expectedCreator,
				Role:          "test-role",
				Options: &guardianv1beta1.AppealOptions{
					Duration:       "24h",
					ExpirationDate: timestamppb.New(timeNow),
				},
				Details: expectedDetails,
				Approvals: []*guardianv1beta1.Approval{
					{
						Id:        "test-approval-id",
						Name:      "test-approval-name",
						Status:    "pending",
						Approvers: []string{"test-approver"},
						CreatedAt: timestamppb.New(timeNow),
						UpdatedAt: timestamppb.New(timeNow),
					},
				},
				RevokedBy:    expectedUser,
				RevokeReason: expectedReason,
				CreatedAt:    timestamppb.New(timeNow),
				UpdatedAt:    timestamppb.New(timeNow),
				RevokedAt:    timestamppb.New(timeNow),
			},
		}
		s.appealService.EXPECT().Revoke(mock.AnythingOfType("*context.valueCtx"), expectedID, expectedUser, expectedReason).Return(expectedAppeal, nil).Once()

		req := &guardianv1beta1.RevokeAppealRequest{
			Id: expectedID,
			Reason: &guardianv1beta1.RevokeAppealRequest_Reason{
				Reason: expectedReason,
			},
		}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: expectedUser,
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.RevokeAppeal(ctx, req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return unathenticated error if request is not authenticated", func() {
		s.setup()

		req := &guardianv1beta1.RevokeAppealRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.RevokeAppeal(ctx, req)

		s.Equal(codes.Unauthenticated, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return error if appeal service returns an error", func() {
		testCases := []struct {
			name               string
			expectedError      error
			expectedStatusCode codes.Code
		}{
			{
				"should return not found error if appeal not found",
				appeal.ErrAppealNotFound,
				codes.NotFound,
			},
			{
				"should return internal error if appeal service returns a unexpected error",
				errors.New("unexpected error"),
				codes.Internal,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.setup()

				s.appealService.EXPECT().Revoke(mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(nil, tc.expectedError).Once()

				req := &guardianv1beta1.RevokeAppealRequest{}
				ctx := context.Background()
				md := metadata.New(map[string]string{
					s.authenticatedUserHeaderKey: "user@example.com",
				})
				ctx = metadata.NewIncomingContext(ctx, md)
				res, err := s.grpcServer.RevokeAppeal(ctx, req)

				s.Equal(tc.expectedStatusCode, status.Code(err))
				s.Nil(res)
				s.appealService.AssertExpectations(s.T())
			})
		}
	})

	s.Run("should return internal error if failed to parse appeal", func() {
		s.setup()

		invalidAppeal := &domain.Appeal{
			Creator: map[string]interface{}{
				"foo": make(chan int),
			},
		}
		s.appealService.EXPECT().Revoke(mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
			Return(invalidAppeal, nil).Once()

		req := &guardianv1beta1.RevokeAppealRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: "user@example.com",
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.RevokeAppeal(ctx, req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})
}
