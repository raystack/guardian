package v1beta1_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	guardianv1beta1 "github.com/raystack/guardian/api/proto/raystack/guardian/v1beta1"
	"github.com/raystack/guardian/core/appeal"
	"github.com/raystack/guardian/domain"
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
			Total: 1,
		}
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.cancelCtx"), expectedFilters).
			Return(expectedApprovals, nil).Once()
		s.approvalService.EXPECT().GetApprovalsTotalCount(mock.AnythingOfType("*context.cancelCtx"), expectedFilters).
			Return(int64(1), nil).Once()

		req := &guardianv1beta1.ListUserApprovalsRequest{
			AccountId: "test-account-id",
			Statuses:  []string{"active", "pending"},
			OrderBy:   []string{"test-order"},
		}
		ctx := context.WithValue(context.Background(), authEmailTestContextKey{}, expectedUser)
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

	s.Run("should return internal error if approvalService.ListApprovals returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.cancelCtx"), mock.Anything).
			Return(nil, expectedError).Once()
		s.approvalService.EXPECT().GetApprovalsTotalCount(mock.AnythingOfType("*context.cancelCtx"), mock.Anything).
			Return(int64(0), nil).Once()

		req := &guardianv1beta1.ListUserApprovalsRequest{}
		ctx := context.WithValue(context.Background(), authEmailTestContextKey{}, "test-user")
		res, err := s.grpcServer.ListUserApprovals(ctx, req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.approvalService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if approvalService.GetApprovalsTotalCount returns an error", func() {
		s.setup()

		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.cancelCtx"), mock.Anything).
			Return([]*domain.Approval{}, nil).Once()
		expectedError := errors.New("random error")
		s.approvalService.EXPECT().GetApprovalsTotalCount(mock.AnythingOfType("*context.cancelCtx"), mock.Anything).
			Return(int64(0), expectedError).Once()

		req := &guardianv1beta1.ListUserApprovalsRequest{}
		ctx := context.WithValue(context.Background(), authEmailTestContextKey{}, "test-user")
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
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.cancelCtx"), mock.Anything).
			Return(invalidApprovals, nil).Once()
		s.approvalService.EXPECT().GetApprovalsTotalCount(mock.AnythingOfType("*context.cancelCtx"), mock.Anything).
			Return(int64(1), nil).Once()

		req := &guardianv1beta1.ListUserApprovalsRequest{}
		ctx := context.WithValue(context.Background(), authEmailTestContextKey{}, "test-user")
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
			Total: 1,
		}
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.cancelCtx"), expectedFilters).
			Return(expectedApprovals, nil).Once()
		s.approvalService.EXPECT().GetApprovalsTotalCount(mock.AnythingOfType("*context.cancelCtx"), expectedFilters).
			Return(int64(1), nil).Once()

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
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.cancelCtx"), mock.Anything).
			Return(nil, expectedError).Once()
		s.approvalService.EXPECT().GetApprovalsTotalCount(mock.AnythingOfType("*context.cancelCtx"), mock.Anything).
			Return(int64(0), nil).Once()

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
		s.approvalService.EXPECT().ListApprovals(mock.AnythingOfType("*context.cancelCtx"), mock.Anything).
			Return(invalidApprovals, nil).Once()
		s.approvalService.EXPECT().GetApprovalsTotalCount(mock.AnythingOfType("*context.cancelCtx"), mock.Anything).
			Return(int64(1), nil).Once()

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
		s.appealService.EXPECT().UpdateApproval(mock.AnythingOfType("*context.valueCtx"), expectedApprovalAction).Return(expectedAppeal, nil).Once()

		req := &guardianv1beta1.UpdateApprovalRequest{
			Id:           expectedID,
			ApprovalName: expectedApprovalName,
			Action: &guardianv1beta1.UpdateApprovalRequest_Action{
				Action: expectedAction,
				Reason: expectedReason,
			},
		}
		ctx := context.WithValue(context.Background(), authEmailTestContextKey{}, expectedUser)
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
				appeal.ErrApprovalNotFound,
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
				s.appealService.EXPECT().UpdateApproval(mock.AnythingOfType("*context.valueCtx"), mock.Anything).
					Return(nil, tc.expectedError).Once()

				req := &guardianv1beta1.UpdateApprovalRequest{}
				ctx := context.WithValue(context.Background(), authEmailTestContextKey{}, expectedUser)
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
		s.appealService.EXPECT().UpdateApproval(mock.AnythingOfType("*context.valueCtx"), mock.Anything).
			Return(invalidAppeal, nil).Once()

		req := &guardianv1beta1.UpdateApprovalRequest{}
		ctx := context.WithValue(context.Background(), authEmailTestContextKey{}, "user@example.com")
		res, err := s.grpcServer.UpdateApproval(ctx, req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.approvalService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestAddApprover() {
	s.Run("should return appeal details on success", func() {
		s.setup()
		timeNow := time.Now()

		appealID := uuid.New().String()
		approvalID := uuid.New().String()
		email := "user@example.com"
		expectedAppeal := &domain.Appeal{
			ID:            appealID,
			ResourceID:    uuid.New().String(),
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,
			Status:        domain.AppealStatusPending,
			AccountID:     "test-account-id",
			AccountType:   "test-account-type",
			CreatedBy:     "test-created-by",
			Role:          "test-role",
			Options: &domain.AppealOptions{
				Duration: "24h",
			},
			Resource: &domain.Resource{},
			Approvals: []*domain.Approval{
				{
					ID:        approvalID,
					Name:      "test-name",
					Status:    domain.ApprovalStatusPending,
					Approvers: []string{"approver1@example.com", email},
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				},
			},
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		s.appealService.EXPECT().AddApprover(mock.AnythingOfType("*context.emptyCtx"), appealID, approvalID, email).Return(expectedAppeal, nil).Once()
		expectedResponse := &guardianv1beta1.AddApproverResponse{
			Appeal: &guardianv1beta1.Appeal{
				Id:            expectedAppeal.ID,
				ResourceId:    expectedAppeal.ResourceID,
				PolicyId:      expectedAppeal.PolicyID,
				PolicyVersion: uint32(expectedAppeal.PolicyVersion),
				Status:        expectedAppeal.Status,
				AccountId:     expectedAppeal.AccountID,
				Role:          expectedAppeal.Role,
				Options: &guardianv1beta1.AppealOptions{
					Duration: expectedAppeal.Options.Duration,
				},
				Resource:    &guardianv1beta1.Resource{},
				AccountType: expectedAppeal.AccountType,
				CreatedBy:   expectedAppeal.CreatedBy,
				Approvals: []*guardianv1beta1.Approval{
					{
						Id:        expectedAppeal.Approvals[0].ID,
						Name:      expectedAppeal.Approvals[0].Name,
						Status:    expectedAppeal.Approvals[0].Status,
						Approvers: expectedAppeal.Approvals[0].Approvers,
						CreatedAt: timestamppb.New(timeNow),
						UpdatedAt: timestamppb.New(timeNow),
					},
				},
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}

		req := &guardianv1beta1.AddApproverRequest{
			AppealId:   appealID,
			ApprovalId: approvalID,
			Email:      email,
		}
		res, err := s.grpcServer.AddApprover(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return error if appeal service returns error", func() {
		testCases := []struct {
			expectedError      error
			expectedStatusCode codes.Code
		}{
			{fmt.Errorf("err message: %w", appeal.ErrAppealIDEmptyParam), codes.InvalidArgument},
			{fmt.Errorf("err message: %w", appeal.ErrApprovalIDEmptyParam), codes.InvalidArgument},
			{fmt.Errorf("err message: %w", appeal.ErrApproverEmail), codes.InvalidArgument},
			{fmt.Errorf("err message: %w", appeal.ErrUnableToAddApprover), codes.InvalidArgument},
			{fmt.Errorf("err message: %w", appeal.ErrAppealNotFound), codes.NotFound},
			{fmt.Errorf("err message: %w", appeal.ErrApprovalNotFound), codes.NotFound},
			{errors.New("unexpected error"), codes.Internal},
		}

		for _, tc := range testCases {
			s.Run(fmt.Sprintf("should return %v error code if appeal service returns %q", tc.expectedStatusCode, tc.expectedError), func() {
				s.appealService.EXPECT().AddApprover(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, tc.expectedError).Once()

				req := &guardianv1beta1.AddApproverRequest{}
				res, err := s.grpcServer.AddApprover(context.Background(), req)

				s.Equal(tc.expectedStatusCode, status.Code(err))
				s.Nil(res)
				s.appealService.AssertExpectations(s.T())
			})
		}
	})

	s.Run("should return error if there's an error when parsing appeal", func() {
		expectedAppeal := &domain.Appeal{
			Creator: map[string]interface{}{
				"foo": make(chan int), // invalid json
			},
		}
		s.appealService.EXPECT().AddApprover(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(expectedAppeal, nil).Once()

		req := &guardianv1beta1.AddApproverRequest{}
		res, err := s.grpcServer.AddApprover(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestDeleteApprover() {
	s.Run("should return appeal details on success", func() {
		s.setup()
		timeNow := time.Now()

		appealID := uuid.New().String()
		approvalID := uuid.New().String()
		email := "user@example.com"
		expectedAppeal := &domain.Appeal{
			ID:            appealID,
			ResourceID:    uuid.New().String(),
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,
			Status:        domain.AppealStatusPending,
			AccountID:     "test-account-id",
			AccountType:   "test-account-type",
			CreatedBy:     "test-created-by",
			Role:          "test-role",
			Options: &domain.AppealOptions{
				Duration: "24h",
			},
			Resource: &domain.Resource{},
			Approvals: []*domain.Approval{
				{
					ID:        approvalID,
					Name:      "test-name",
					Status:    domain.ApprovalStatusPending,
					Approvers: []string{"approver1@example.com"},
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				},
			},
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		s.appealService.EXPECT().DeleteApprover(mock.AnythingOfType("*context.emptyCtx"), appealID, approvalID, email).Return(expectedAppeal, nil).Once()
		expectedResponse := &guardianv1beta1.DeleteApproverResponse{
			Appeal: &guardianv1beta1.Appeal{
				Id:            expectedAppeal.ID,
				ResourceId:    expectedAppeal.ResourceID,
				PolicyId:      expectedAppeal.PolicyID,
				PolicyVersion: uint32(expectedAppeal.PolicyVersion),
				Status:        expectedAppeal.Status,
				AccountId:     expectedAppeal.AccountID,
				Role:          expectedAppeal.Role,
				Options: &guardianv1beta1.AppealOptions{
					Duration: expectedAppeal.Options.Duration,
				},
				Resource:    &guardianv1beta1.Resource{},
				AccountType: expectedAppeal.AccountType,
				CreatedBy:   expectedAppeal.CreatedBy,
				Approvals: []*guardianv1beta1.Approval{
					{
						Id:        expectedAppeal.Approvals[0].ID,
						Name:      expectedAppeal.Approvals[0].Name,
						Status:    expectedAppeal.Approvals[0].Status,
						Approvers: expectedAppeal.Approvals[0].Approvers,
						CreatedAt: timestamppb.New(timeNow),
						UpdatedAt: timestamppb.New(timeNow),
					},
				},
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}

		req := &guardianv1beta1.DeleteApproverRequest{
			AppealId:   appealID,
			ApprovalId: approvalID,
			Email:      email,
		}
		res, err := s.grpcServer.DeleteApprover(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.appealService.AssertExpectations(s.T())
	})

	s.Run("should return error if appeal service returns error", func() {
		testCases := []struct {
			expectedError      error
			expectedStatusCode codes.Code
		}{
			{fmt.Errorf("err message: %w", appeal.ErrAppealIDEmptyParam), codes.InvalidArgument},
			{fmt.Errorf("err message: %w", appeal.ErrApprovalIDEmptyParam), codes.InvalidArgument},
			{fmt.Errorf("err message: %w", appeal.ErrApproverEmail), codes.InvalidArgument},
			{fmt.Errorf("err message: %w", appeal.ErrUnableToDeleteApprover), codes.InvalidArgument},
			{fmt.Errorf("err message: %w", appeal.ErrAppealNotFound), codes.NotFound},
			{fmt.Errorf("err message: %w", appeal.ErrApprovalNotFound), codes.NotFound},
			{errors.New("unexpected error"), codes.Internal},
		}

		for _, tc := range testCases {
			s.Run(fmt.Sprintf("should return %v error code if appeal service returns %q", tc.expectedStatusCode, tc.expectedError), func() {
				s.appealService.EXPECT().DeleteApprover(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, tc.expectedError).Once()

				req := &guardianv1beta1.DeleteApproverRequest{}
				res, err := s.grpcServer.DeleteApprover(context.Background(), req)

				s.Equal(tc.expectedStatusCode, status.Code(err))
				s.Nil(res)
				s.appealService.AssertExpectations(s.T())
			})
		}
	})

	s.Run("should return error if there's an error when parsing appeal", func() {
		expectedAppeal := &domain.Appeal{
			Creator: map[string]interface{}{
				"foo": make(chan int), // invalid json
			},
		}
		s.appealService.EXPECT().DeleteApprover(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(expectedAppeal, nil).Once()

		req := &guardianv1beta1.DeleteApproverRequest{}
		res, err := s.grpcServer.DeleteApprover(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.appealService.AssertExpectations(s.T())
	})
}
