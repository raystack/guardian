package v1beta1

import (
	"context"
	"errors"

	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/core/provideractivity"
	"github.com/odpf/guardian/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) GetActivity(ctx context.Context, req *guardianv1beta1.GetActivityRequest) (*guardianv1beta1.GetActivityResponse, error) {
	activity, err := s.providerActivityService.GetOne(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, provideractivity.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "activity not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get activity: %v", err)
	}

	activityProto, err := s.adapter.ToProviderActivity(activity)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse proto: %v", err)
	}

	return &guardianv1beta1.GetActivityResponse{
		Activity: activityProto,
	}, nil
}

func (s *GRPCServer) ListActivities(ctx context.Context, req *guardianv1beta1.ListActivitiesRequest) (*guardianv1beta1.ListActivitiesResponse, error) {
	filter := domain.ListProviderActivitiesFilter{
		ProviderIDs: req.GetProviderIds(),
		AccountIDs:  req.GetAccountIds(),
		ResourceIDs: req.GetResourceIds(),
		Types:       req.GetTypes(),
	}
	if req.GetTimestampGte() != nil {
		t := req.GetTimestampGte().AsTime()
		filter.TimestampGte = &t
	}
	if req.GetTimestampLte() != nil {
		t := req.GetTimestampLte().AsTime()
		filter.TimestampLte = &t
	}

	activities, err := s.providerActivityService.Find(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list activities: %v", err)
	}

	activityProtos := []*guardianv1beta1.ProviderActivity{}
	for _, a := range activities {
		activityProto, err := s.adapter.ToProviderActivity(a)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse proto: %v", err)
		}
		activityProtos = append(activityProtos, activityProto)
	}

	return &guardianv1beta1.ListActivitiesResponse{
		Activities: activityProtos,
	}, nil
}

func (s *GRPCServer) ImportActivities(ctx context.Context, req *guardianv1beta1.ImportActivitiesRequest) (*guardianv1beta1.ImportActivitiesResponse, error) {
	filter := domain.ImportActivitiesFilter{
		ProviderID:  req.GetProviderId(),
		ResourceIDs: req.GetResourceIds(),
		AccountIDs:  req.GetAccountIds(),
	}
	if req.GetTimestampGte() != nil {
		t := req.GetTimestampGte().AsTime()
		filter.TimestampGte = &t
	}
	if req.GetTimestampLte() != nil {
		t := req.GetTimestampLte().AsTime()
		filter.TimestampLte = &t
	}

	activities, err := s.providerActivityService.Import(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to import activities: %v", err)
	}

	activityProtos := []*guardianv1beta1.ProviderActivity{}
	for _, a := range activities {
		activity, err := s.adapter.ToProviderActivity(a)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse proto: %v", err)
		}

		activityProtos = append(activityProtos, activity)
	}

	return &guardianv1beta1.ImportActivitiesResponse{
		Activities: activityProtos,
	}, nil
}
