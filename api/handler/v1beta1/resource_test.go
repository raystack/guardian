package v1beta1_test

import (
	"context"
	"errors"
	"time"

	guardianv1beta1 "github.com/raystack/guardian/api/proto/raystack/guardian/v1beta1"
	"github.com/raystack/guardian/core/resource"
	"github.com/raystack/guardian/domain"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *GrpcHandlersSuite) TestListResources() {
	s.Run("should return list of resources on success", func() {
		s.setup()
		timeNow := time.Now()

		ctx := context.Background()

		req := &guardianv1beta1.ListResourcesRequest{
			IsDeleted:    true,
			Type:         "test-type",
			Urn:          "test-urn",
			ProviderType: "test-provider-type",
			ProviderUrn:  "test-provider-urn",
			Name:         "test-name",
			Details: []string{
				"key1:value1",
				"key2.key3:value2",
			},
		}

		expectedDetails := map[string]interface{}{
			"key1": "value1",
			"key2": map[string]interface{}{
				"key3": "value2",
			},
		}
		expectedDetailsProto, err := structpb.NewStruct(expectedDetails)
		s.Require().NoError(err)
		expectedResponse := &guardianv1beta1.ListResourcesResponse{
			Resources: []*guardianv1beta1.Resource{
				{
					Id:           "123",
					IsDeleted:    true,
					Type:         "test-type",
					Urn:          "test-urn",
					ProviderType: "test-provider-type",
					ProviderUrn:  "test-provider-urn",
					Name:         "test-name",
					CreatedAt:    timestamppb.New(timeNow),
					UpdatedAt:    timestamppb.New(timeNow),
					Details:      expectedDetailsProto,
				},
			},
		}
		expectedFilters := domain.ListResourcesFilter{
			IsDeleted:    true,
			ResourceType: "test-type",
			ResourceURN:  "test-urn",
			ProviderType: "test-provider-type",
			ProviderURN:  "test-provider-urn",
			Name:         "test-name",
			Details: map[string]string{
				"key1":      "value1",
				"key2.key3": "value2",
			},
		}
		dummyResources := []*domain.Resource{
			{
				ID:           "123",
				IsDeleted:    true,
				Type:         "test-type",
				URN:          "test-urn",
				ProviderType: "test-provider-type",
				ProviderURN:  "test-provider-urn",
				Name:         "test-name",
				CreatedAt:    timeNow,
				UpdatedAt:    timeNow,
				Details:      expectedDetails,
			},
		}
		s.resourceService.EXPECT().Find(ctx, expectedFilters).Return(dummyResources, nil)

		res, err := s.grpcServer.ListResources(ctx, req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.resourceService.AssertExpectations(s.T())
	})

	s.Run("should return error if resource service returns error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.resourceService.EXPECT().Find(mock.AnythingOfType("context.backgroundCtx"), mock.Anything).Return(nil, expectedError).Once()

		req := &guardianv1beta1.ListResourcesRequest{}
		res, err := s.grpcServer.ListResources(context.Background(), req)

		s.Nil(res)
		s.Equal(codes.Internal, status.Code(err))
		s.resourceService.AssertExpectations(s.T())
	})

	s.Run("should return error if there is an error when parsing the resources", func() {
		s.setup()

		invalidResources := []*domain.Resource{
			{
				Details: map[string]interface{}{
					"key": make(chan int), // invalid json
				},
			},
		}
		s.resourceService.EXPECT().Find(mock.AnythingOfType("context.backgroundCtx"), mock.Anything).Return(invalidResources, nil).Once()

		req := &guardianv1beta1.ListResourcesRequest{}
		res, err := s.grpcServer.ListResources(context.Background(), req)

		s.Nil(res)
		s.Equal(codes.Internal, status.Code(err))
		s.resourceService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestGetResource() {
	s.Run("should return resource details on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedID := "test-id"
		expectedResource := &domain.Resource{
			ID:        expectedID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		s.resourceService.EXPECT().GetOne(mock.AnythingOfType("context.backgroundCtx"), expectedID).Return(expectedResource, nil).Once()
		expectedResponse := &guardianv1beta1.GetResourceResponse{
			Resource: &guardianv1beta1.Resource{
				Id:        expectedID,
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}

		req := &guardianv1beta1.GetResourceRequest{Id: expectedID}
		res, err := s.grpcServer.GetResource(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.resourceService.AssertExpectations(s.T())
	})

	s.Run("should return not found error if resource not found", func() {
		s.setup()

		s.resourceService.EXPECT().GetOne(mock.AnythingOfType("context.backgroundCtx"), mock.Anything).Return(nil, resource.ErrRecordNotFound)

		req := &guardianv1beta1.GetResourceRequest{Id: "unknown-id"}
		res, err := s.grpcServer.GetResource(context.Background(), req)

		s.Equal(codes.NotFound, status.Code(err))
		s.Nil(res)
		s.resourceService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if resource service returns error", func() {
		s.setup()

		expectedError := errors.New("randome error")
		s.resourceService.EXPECT().GetOne(mock.AnythingOfType("context.backgroundCtx"), mock.Anything).Return(nil, expectedError)

		req := &guardianv1beta1.GetResourceRequest{Id: "unknown-id"}
		res, err := s.grpcServer.GetResource(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.resourceService.AssertExpectations(s.T())
	})

	s.Run("should return error if there is an error when parsing the resource", func() {
		s.setup()
		timeNow := time.Now()

		expectedID := "test-id"
		expectedResource := &domain.Resource{
			ID:        expectedID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
			Details: map[string]interface{}{
				"key": make(chan int), // invalid json
			},
		}
		s.resourceService.EXPECT().GetOne(mock.AnythingOfType("context.backgroundCtx"), expectedID).Return(expectedResource, nil).Once()

		req := &guardianv1beta1.GetResourceRequest{Id: expectedID}
		res, err := s.grpcServer.GetResource(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.resourceService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestUpdateResource() {
	s.Run("should return resource details on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedID := "test-id"
		expectedResource := &domain.Resource{
			ID:   expectedID,
			Name: "new-name",
		}
		s.resourceService.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), expectedResource).Return(nil).
			Run(func(_a0 context.Context, _a1 *domain.Resource) {
				_a1.CreatedAt = timeNow
				_a1.UpdatedAt = timeNow
			}).Once()
		expectedResponse := &guardianv1beta1.UpdateResourceResponse{
			Resource: &guardianv1beta1.Resource{
				Id:        expectedID,
				Name:      expectedResource.Name,
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}

		req := &guardianv1beta1.UpdateResourceRequest{
			Id: expectedID,
			Resource: &guardianv1beta1.Resource{
				Name: "new-name",
			},
		}
		res, err := s.grpcServer.UpdateResource(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.resourceService.AssertExpectations(s.T())
	})

	s.Run("should return not found error if resource not found", func() {
		s.setup()

		s.resourceService.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*domain.Resource")).Return(resource.ErrRecordNotFound)

		req := &guardianv1beta1.UpdateResourceRequest{Id: "unknown-id"}
		res, err := s.grpcServer.UpdateResource(context.Background(), req)

		s.Equal(codes.NotFound, status.Code(err))
		s.Nil(res)
		s.resourceService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if resource service returns error", func() {
		s.setup()

		expectedError := errors.New("randome error")
		s.resourceService.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*domain.Resource")).Return(expectedError)

		req := &guardianv1beta1.UpdateResourceRequest{Id: "unknown-id"}
		res, err := s.grpcServer.UpdateResource(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.resourceService.AssertExpectations(s.T())
	})

	s.Run("should return error if there is an error when parsing the resource", func() {
		s.setup()

		s.resourceService.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*domain.Resource")).Return(nil).
			Run(func(_a0 context.Context, _a1 *domain.Resource) {
				_a1.Details = map[string]interface{}{
					"key": make(chan int), // invalid json
				}
			}).Once()

		req := &guardianv1beta1.UpdateResourceRequest{Id: "test-id"}
		res, err := s.grpcServer.UpdateResource(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.resourceService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestDeleteResource() {
	s.Run("should return no error on success", func() {
		s.setup()

		expectedID := "test-id"
		s.resourceService.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), expectedID).Return(nil)

		req := &guardianv1beta1.DeleteResourceRequest{Id: expectedID}
		res, err := s.grpcServer.DeleteResource(context.Background(), req)

		s.NoError(err)
		s.Equal(&guardianv1beta1.DeleteResourceResponse{}, res)
		s.resourceService.AssertExpectations(s.T())
	})

	s.Run("should return not found error if resource not found", func() {
		s.setup()

		s.resourceService.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).Return(resource.ErrRecordNotFound)

		req := &guardianv1beta1.DeleteResourceRequest{Id: "unknown-id"}
		res, err := s.grpcServer.DeleteResource(context.Background(), req)

		s.Equal(codes.NotFound, status.Code(err))
		s.Nil(res)
		s.resourceService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if resource service returns error", func() {
		s.setup()

		expectedError := errors.New("randome error")
		s.resourceService.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).Return(expectedError)

		req := &guardianv1beta1.DeleteResourceRequest{Id: "unknown-id"}
		res, err := s.grpcServer.DeleteResource(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.resourceService.AssertExpectations(s.T())
	})
}
