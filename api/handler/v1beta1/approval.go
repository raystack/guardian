package v1beta1

import (
	"context"

	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) ListUserApprovals(ctx context.Context, req *guardianv1beta1.ListUserApprovalsRequest) (*guardianv1beta1.ListUserApprovalsResponse, error) {
	user, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	approvals, err := s.listApprovals(ctx, &domain.ListApprovalsFilter{
		AccountID: req.GetAccountId(),
		CreatedBy: user,
		Statuses:  req.GetStatuses(),
		OrderBy:   req.GetOrderBy(),
	})
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.ListUserApprovalsResponse{
		Approvals: approvals,
	}, nil
}

func (s *GRPCServer) ListApprovals(ctx context.Context, req *guardianv1beta1.ListApprovalsRequest) (*guardianv1beta1.ListApprovalsResponse, error) {
	approvals, err := s.listApprovals(ctx, &domain.ListApprovalsFilter{
		AccountID: req.GetAccountId(),
		CreatedBy: req.GetCreatedBy(),
		Statuses:  req.GetStatuses(),
		OrderBy:   req.GetOrderBy(),
	})
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.ListApprovalsResponse{
		Approvals: approvals,
	}, nil
}

func (s *GRPCServer) UpdateApproval(ctx context.Context, req *guardianv1beta1.UpdateApprovalRequest) (*guardianv1beta1.UpdateApprovalResponse, error) {
	actor, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	id := req.GetId()
	a, err := s.appealService.MakeAction(ctx, domain.ApprovalAction{
		AppealID:     id,
		ApprovalName: req.GetApprovalName(),
		Actor:        actor,
		Action:       req.GetAction().GetAction(),
		Reason:       req.GetAction().GetReason(),
	})
	if err != nil {
		switch err {
		case appeal.ErrAppealStatusCanceled,
			appeal.ErrAppealStatusApproved,
			appeal.ErrAppealStatusRejected,
			appeal.ErrAppealStatusTerminated,
			appeal.ErrAppealStatusUnrecognized,
			appeal.ErrApprovalDependencyIsPending,
			appeal.ErrApprovalStatusUnrecognized,
			appeal.ErrApprovalStatusApproved,
			appeal.ErrApprovalStatusRejected,
			appeal.ErrApprovalStatusSkipped,
			appeal.ErrActionInvalidValue:
			return nil, status.Errorf(codes.InvalidArgument, "unable to process the request: %v", err)
		case appeal.ErrActionForbidden:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		case appeal.ErrApprovalNameNotFound:
			return nil, status.Errorf(codes.NotFound, "appeal not found: %v", id)
		default:
			return nil, status.Errorf(codes.Internal, "failed to update approval: %v", err)
		}
	}

	appealProto, err := s.adapter.ToAppealProto(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse appeal: %v", err)
	}

	return &guardianv1beta1.UpdateApprovalResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) listApprovals(ctx context.Context, filters *domain.ListApprovalsFilter) ([]*guardianv1beta1.Approval, error) {
	approvals, err := s.approvalService.ListApprovals(ctx, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get approval list: %s", err)
	}

	approvalProtos := []*guardianv1beta1.Approval{}
	for _, a := range approvals {
		approvalProto, err := s.adapter.ToApprovalProto(a)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse approval: %v: %s", a.ID, err)
		}
		approvalProtos = append(approvalProtos, approvalProto)
	}

	return approvalProtos, nil
}
