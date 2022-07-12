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

func (s *GrpcHandlersSuite) TestListUserApprovals() {
	s.Run("should return list of approvals on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedUser := "test-user"
		expectedFilters := &domain.ListApprovalsFilter{
			CreatedBy: expectedUser,
			AccountID: "test-account-id",
			Statuses:  []string{"active", "pending"},
			OrderBy:   []string{"test-order"},
		}
		expectedApprovals := []*domain.Approval{
			{
				ID:        "test-approval-id",
				Name:      "test-approval-name",
				Status:    "pending",
				Approvers: []string{"test-approver"},
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
				Appeal: &domain.Appeal{
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
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				},
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
		expectedResponse := &guardianv1beta1.ListUserApprovalsResponse{
			Approvals: []*guardianv1beta1.Approval{
				{
					Id:        "test-approval-id",
					Name:      "test-approval-name",
					Status:    "pending",
					Approvers: []string{"test-approver"},
					CreatedAt: timestamppb.New(timeNow),
					UpdatedAt: timestamppb.New(timeNow),
					Appeal: &guardianv1beta1.Appeal{
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
						Details:   expectedDetails,
						CreatedAt: timestamppb.New(timeNow),
						UpdatedAt: timestamppb.New(timeNow),
					},
				},
			},
		}
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.valueCtx"), expectedFilters).
			Return(expectedApprovals, nil).Once()

		req := &guardianv1beta1.ListUserApprovalsRequest{
			AccountId: "test-account-id",
			Statuses:  []string{"active", "pending"},
			OrderBy:   []string{"test-order"},
		}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: expectedUser,
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.ListUserApprovals(ctx, req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.approvalService.AssertExpectations(s.T())
	})

	s.Run("should return unathenticated error if request is not authenticated", func() {
		s.setup()

		req := &guardianv1beta1.ListUserApprovalsRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.ListUserApprovals(ctx, req)

		s.Equal(codes.Unauthenticated, status.Code(err))
		s.Nil(res)
		s.approvalService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if approval service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.valueCtx"), mock.Anything).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.ListUserApprovalsRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: "test-user",
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.ListUserApprovals(ctx, req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.approvalService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there is an error when parsing approval", func() {
		s.setup()

		invalidApprovals := []*domain.Approval{
			{
				Appeal: &domain.Appeal{
					Creator: map[string]interface{}{
						"foo": make(chan int),
					},
				},
			},
		}
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.valueCtx"), mock.Anything).
			Return(invalidApprovals, nil).Once()

		req := &guardianv1beta1.ListUserApprovalsRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: "test-user",
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.ListUserApprovals(ctx, req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.approvalService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestListApprovals() {
	s.Run("should return list of approvals on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedUser := "test-user"
		expectedFilters := &domain.ListApprovalsFilter{
			CreatedBy: expectedUser,
			AccountID: "test-account-id",
			Statuses:  []string{"active", "pending"},
			OrderBy:   []string{"test-order"},
		}
		expectedApprovals := []*domain.Approval{
			{
				ID:        "test-approval-id",
				Name:      "test-approval-name",
				Status:    "pending",
				Approvers: []string{"test-approver"},
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
				Appeal: &domain.Appeal{
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
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				},
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
		expectedResponse := &guardianv1beta1.ListApprovalsResponse{
			Approvals: []*guardianv1beta1.Approval{
				{
					Id:        "test-approval-id",
					Name:      "test-approval-name",
					Status:    "pending",
					Approvers: []string{"test-approver"},
					CreatedAt: timestamppb.New(timeNow),
					UpdatedAt: timestamppb.New(timeNow),
					Appeal: &guardianv1beta1.Appeal{
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
						Details:   expectedDetails,
						CreatedAt: timestamppb.New(timeNow),
						UpdatedAt: timestamppb.New(timeNow),
					},
				},
			},
		}
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.emptyCtx"), expectedFilters).
			Return(expectedApprovals, nil).Once()

		req := &guardianv1beta1.ListApprovalsRequest{
			AccountId: "test-account-id",
			CreatedBy: expectedUser,
			Statuses:  []string{"active", "pending"},
			OrderBy:   []string{"test-order"},
		}
		res, err := s.grpcServer.ListApprovals(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.approvalService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if approval service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.ListApprovalsRequest{}
		res, err := s.grpcServer.ListApprovals(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.approvalService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there is an error when parsing approval", func() {
		s.setup()

		invalidApprovals := []*domain.Approval{
			{
				Appeal: &domain.Appeal{
					Creator: map[string]interface{}{
						"foo": make(chan int),
					},
				},
			},
		}
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(invalidApprovals, nil).Once()

		req := &guardianv1beta1.ListApprovalsRequest{}
		res, err := s.grpcServer.ListApprovals(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.approvalService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestUpdateApproval() {
	s.Run("should return appeal details on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedUser := "user@example.com"
		expectedID := "test-appeal-id"
		expectedApprovalName := "test-approval-name"
		expectedAction := "approve"
		expectedReason := "test-reason"
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
					Name:      expectedApprovalName,
					Status:    "approved",
					Actor:     &expectedUser,
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
		expectedResponse := &guardianv1beta1.UpdateApprovalResponse{
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
						Status:    "approved",
						Actor:     expectedUser,
						Approvers: []string{"test-approver"},
						CreatedAt: timestamppb.New(timeNow),
						UpdatedAt: timestamppb.New(timeNow),
					},
				},
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}
		expectedApprovalAction := domain.ApprovalAction{
			AppealID:     expectedID,
			ApprovalName: expectedApprovalName,
			Actor:        expectedUser,
			Action:       expectedAction,
			Reason:       expectedReason,
		}
		s.appealService.EXPECT().MakeAction(mock.AnythingOfType("*context.valueCtx"), expectedApprovalAction).Return(expectedAppeal, nil).Once()

		req := &guardianv1beta1.UpdateApprovalRequest{
			Id:           expectedID,
			ApprovalName: expectedApprovalName,
			Action: &guardianv1beta1.UpdateApprovalRequest_Action{
				Action: expectedAction,
				Reason: expectedReason,
			},
		}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: expectedUser,
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.UpdateApproval(ctx, req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return unathenticated error if request is not authenticated", func() {
		s.setup()

		req := &guardianv1beta1.UpdateApprovalRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.UpdateApproval(ctx, req)

		s.Equal(codes.Unauthenticated, status.Code(err))
		s.Nil(res)
		s.approvalService.AssertExpectations(s.T())
	})

	s.Run("should return error if appeal service returns an error", func() {
		testCases := []struct {
			name               string
			expectedError      error
			expectedStatusCode codes.Code
		}{

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
				"should return invalid error if approval dependency is still pending",
				appeal.ErrApprovalDependencyIsPending,
				codes.InvalidArgument,
			},
			{
				"should return invalid error if approval status unrecognized",
				appeal.ErrApprovalStatusUnrecognized,
				codes.InvalidArgument,
			},
			{
				"should return invalid error if approval status already approved",
				appeal.ErrApprovalStatusApproved,
				codes.InvalidArgument,
			},
			{
				"should return invalid error if approval status already rejected",
				appeal.ErrApprovalStatusRejected,
				codes.InvalidArgument,
			},
			{
				"should return invalid error if approval status already skipped",
				appeal.ErrApprovalStatusSkipped,
				codes.InvalidArgument,
			},
			{
				"should return invalid error if action name is invalid",
				appeal.ErrActionInvalidValue,
				codes.InvalidArgument,
			},

			{
				"should return not found error if appeal not found",
				appeal.ErrActionForbidden,
				codes.PermissionDenied,
			},
			{
				"should return not found error if appeal not found",
				appeal.ErrApprovalNameNotFound,
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

				expectedUser := "user@example.com"
				s.appealService.EXPECT().MakeAction(mock.AnythingOfType("*context.valueCtx"), mock.Anything).
					Return(nil, tc.expectedError).Once()

				req := &guardianv1beta1.UpdateApprovalRequest{}
				ctx := context.Background()
				md := metadata.New(map[string]string{
					s.authenticatedUserHeaderKey: expectedUser,
				})
				ctx = metadata.NewIncomingContext(ctx, md)
				res, err := s.grpcServer.UpdateApproval(ctx, req)

				s.Equal(tc.expectedStatusCode, status.Code(err))
				s.Nil(res)
				s.appealService.AssertExpectations(s.T())
			})
		}
	})

	s.Run("should return unathenticated error if request is not authenticated", func() {
		s.setup()

		invalidAppeal := &domain.Appeal{
			Creator: map[string]interface{}{
				"foo": make(chan int), // invalid json
			},
		}
		s.appealService.EXPECT().MakeAction(mock.AnythingOfType("*context.valueCtx"), mock.Anything).
			Return(invalidAppeal, nil).Once()

		req := &guardianv1beta1.UpdateApprovalRequest{}
		ctx := context.Background()
		md := metadata.New(map[string]string{
			s.authenticatedUserHeaderKey: "user@example.com",
		})
		ctx = metadata.NewIncomingContext(ctx, md)
		res, err := s.grpcServer.UpdateApproval(ctx, req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.approvalService.AssertExpectations(s.T())
	})
}
