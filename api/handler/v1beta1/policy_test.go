package v1beta1_test

import (
	"context"
	"errors"
	"time"

	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/core/policy"
	"github.com/odpf/guardian/domain"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *GrpcHandlersSuite) TestListPolicies() {
	s.Run("should return list of policies on success", func() {
		s.setup()

		expectedResponse := &guardianv1beta1.ListPoliciesResponse{
			Policies: []*guardianv1beta1.Policy{
				{
					Id: "test-policy",
				},
			},
		}
		dummyPolicies := []*domain.Policy{
			{ID: "test-policy"},
		}
		s.policyService.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx")).Return(dummyPolicies, nil).Once()

		req := &guardianv1beta1.ListPoliciesRequest{}
		res, err := s.grpcServer.ListPolicies(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.policyService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if policy service returns error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.policyService.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx")).Return(nil, expectedError).Once()

		req := &guardianv1beta1.ListPoliciesRequest{}
		res, err := s.grpcServer.ListPolicies(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.policyService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there's an error when parsing policy", func() {
		s.setup()

		dummyPolicies := []*domain.Policy{
			{
				ID: "test-policy",
				IAM: &domain.IAMConfig{
					Config: make(chan int), // invalid json
				},
			},
		}
		s.policyService.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx")).Return(dummyPolicies, nil).Once()

		req := &guardianv1beta1.ListPoliciesRequest{}
		res, err := s.grpcServer.ListPolicies(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.policyService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestGetPolicy() {
	s.Run("should return policy details on success", func() {
		s.setup()
		timeNow := time.Now()

		dummyPolicy := &domain.Policy{
			ID:          "test-policy",
			Version:     1,
			Description: "test-description",
			Steps: []*domain.Step{
				{
					Name:            "test-approval-step",
					Description:     "test-description",
					Strategy:        "auto",
					ApproveIf:       "true",
					RejectionReason: "test-rejection-message",
				},
			},
			Requirements: []*domain.Requirement{
				{
					On: &domain.RequirementTrigger{
						ProviderType: "test-provider-type",
					},
					Appeals: []*domain.AdditionalAppeal{
						{
							Resource: &domain.ResourceIdentifier{
								ID: "test-resource-id",
							},
							Role: "test-role",
							Policy: &domain.PolicyConfig{
								ID:      "test-policy",
								Version: 1,
							},
						},
					},
				},
			},
			IAM: &domain.IAMConfig{
				Provider: "slack",
				Config:   map[string]interface{}{"foo": "bar"},
				Schema:   map[string]string{"foo": "bar"},
			},
			AppealConfig: &domain.PolicyAppealConfig{
				DurationOptions: []domain.AppealDurationOption{
					{Name: "1 Day", Value: "24h"},
					{Name: "3 Days", Value: "72h"},
				},
			},
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		expectedIAMConfig, err := structpb.NewValue(dummyPolicy.IAM.Config)
		s.Require().NoError(err)
		expectedResponse := &guardianv1beta1.GetPolicyResponse{
			Policy: &guardianv1beta1.Policy{
				Id:          dummyPolicy.ID,
				Version:     uint32(dummyPolicy.Version),
				Description: dummyPolicy.Description,
				Steps: []*guardianv1beta1.Policy_ApprovalStep{
					{
						Name:            "test-approval-step",
						Description:     "test-description",
						Strategy:        "auto",
						ApproveIf:       "true",
						RejectionReason: "test-rejection-message",
					},
				},
				Requirements: []*guardianv1beta1.Policy_Requirement{
					{
						On: &guardianv1beta1.Policy_Requirement_RequirementTrigger{
							ProviderType: "test-provider-type",
						},
						Appeals: []*guardianv1beta1.Policy_Requirement_AdditionalAppeal{
							{
								Resource: &guardianv1beta1.Policy_Requirement_AdditionalAppeal_ResourceIdentifier{
									Id: "test-resource-id",
								},
								Role: "test-role",
								Policy: &guardianv1beta1.PolicyConfig{
									Id:      "test-policy",
									Version: 1,
								},
							},
						},
					},
				},
				Iam: &guardianv1beta1.Policy_IAM{
					Provider: "slack",
					Config:   expectedIAMConfig,
					Schema:   dummyPolicy.IAM.Schema,
				},
				Appeal: &guardianv1beta1.PolicyAppealConfig{
					DurationOptions: []*guardianv1beta1.PolicyAppealConfig_DurationOptions{
						{Name: "1 Day", Value: "24h"},
						{Name: "3 Days", Value: "72h"},
					},
				},
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}
		s.policyService.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), "test-policy", uint(1)).
			Return(dummyPolicy, nil).Once()

		req := &guardianv1beta1.GetPolicyRequest{
			Id:      "test-policy",
			Version: 1,
		}
		res, err := s.grpcServer.GetPolicy(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.policyService.AssertExpectations(s.T())
	})

	s.Run("should return not found error if policy not found", func() {
		s.setup()

		s.policyService.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("uint")).
			Return(nil, policy.ErrPolicyNotFound).Once()

		req := &guardianv1beta1.GetPolicyRequest{}
		res, err := s.grpcServer.GetPolicy(context.Background(), req)

		s.Equal(codes.NotFound, status.Code(err))
		s.Nil(res)
		s.policyService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if policy service returns error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.policyService.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("uint")).
			Return(nil, expectedError).Once()

		req := &guardianv1beta1.GetPolicyRequest{}
		res, err := s.grpcServer.GetPolicy(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.policyService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there's an error when parsing policy", func() {
		s.setup()

		dummyPolicy := &domain.Policy{

			ID: "test-policy",
			IAM: &domain.IAMConfig{
				Config: make(chan int), // invalid json
			},
		}
		s.policyService.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("uint")).
			Return(dummyPolicy, nil).Once()

		req := &guardianv1beta1.GetPolicyRequest{}
		res, err := s.grpcServer.GetPolicy(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.policyService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestCreatePolicy() {
	s.Run("should return policy details on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedPolicy := &domain.Policy{
			ID:          "test-policy",
			Description: "test-description",
			Steps: []*domain.Step{
				{
					Name:            "test-approval-step",
					Description:     "test-description",
					Strategy:        "auto",
					ApproveIf:       "true",
					RejectionReason: "test-rejection-message",
				},
			},
			Requirements: []*domain.Requirement{
				{
					On: &domain.RequirementTrigger{
						ProviderType: "test-provider-type",
					},
					Appeals: []*domain.AdditionalAppeal{
						{
							Resource: &domain.ResourceIdentifier{
								ID: "test-resource-id",
							},
							Role: "test-role",
							Policy: &domain.PolicyConfig{
								ID:      "test-policy",
								Version: 1,
							},
							Options: &domain.AppealOptions{
								Duration: "24h",
							},
						},
					},
				},
			},
			IAM: &domain.IAMConfig{
				Provider: "slack",
				Config:   map[string]interface{}{"foo": "bar"},
				Schema:   map[string]string{"foo": "bar"},
			},
			AppealConfig: &domain.PolicyAppealConfig{
				DurationOptions: []domain.AppealDurationOption{
					{Name: "1 Day", Value: "24h"},
					{Name: "3 Days", Value: "72h"},
				},
				AllowPermanentAccess:         true,
				AllowActiveAccessExtensionIn: "24h",
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
		expectedVersion := uint(1)
		expectedIAMConfig, err := structpb.NewValue(expectedPolicy.IAM.Config)
		s.Require().NoError(err)
		expectedResponse := &guardianv1beta1.CreatePolicyResponse{
			Policy: &guardianv1beta1.Policy{
				Id:          expectedPolicy.ID,
				Version:     uint32(expectedVersion),
				Description: expectedPolicy.Description,
				Steps: []*guardianv1beta1.Policy_ApprovalStep{
					{
						Name:            "test-approval-step",
						Description:     "test-description",
						Strategy:        "auto",
						ApproveIf:       "true",
						RejectionReason: "test-rejection-message",
					},
				},
				Requirements: []*guardianv1beta1.Policy_Requirement{
					{
						On: &guardianv1beta1.Policy_Requirement_RequirementTrigger{
							ProviderType: "test-provider-type",
						},
						Appeals: []*guardianv1beta1.Policy_Requirement_AdditionalAppeal{
							{
								Resource: &guardianv1beta1.Policy_Requirement_AdditionalAppeal_ResourceIdentifier{
									Id: "test-resource-id",
								},
								Role: "test-role",
								Policy: &guardianv1beta1.PolicyConfig{
									Id:      "test-policy",
									Version: 1,
								},
								Options: &guardianv1beta1.AppealOptions{
									Duration: "24h",
								},
							},
						},
					},
				},
				Iam: &guardianv1beta1.Policy_IAM{
					Provider: "slack",
					Config:   expectedIAMConfig,
					Schema:   expectedPolicy.IAM.Schema,
				},
				Appeal: &guardianv1beta1.PolicyAppealConfig{
					DurationOptions: []*guardianv1beta1.PolicyAppealConfig_DurationOptions{
						{Name: "1 Day", Value: "24h"},
						{Name: "3 Days", Value: "72h"},
					},
					AllowPermanentAccess:         true,
					AllowActiveAccessExtensionIn: "24h",
					Questions: []*guardianv1beta1.PolicyAppealConfig_Question{
						{
							Key:         "team",
							Question:    "What team are you in?",
							Required:    true,
							Description: "Please provide the name of the team you are in",
						},
					},
				},
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}
		s.policyService.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), expectedPolicy).
			Run(func(_a0 context.Context, _a1 *domain.Policy) {
				_a1.CreatedAt = timeNow
				_a1.UpdatedAt = timeNow
				_a1.Version = expectedVersion
			}).Return(nil).Once()

		req := &guardianv1beta1.CreatePolicyRequest{
			Policy: &guardianv1beta1.Policy{
				Id:          expectedPolicy.ID,
				Description: expectedPolicy.Description,
				Steps: []*guardianv1beta1.Policy_ApprovalStep{
					{
						Name:            "test-approval-step",
						Description:     "test-description",
						Strategy:        "auto",
						ApproveIf:       "true",
						RejectionReason: "test-rejection-message",
					},
				},
				Requirements: []*guardianv1beta1.Policy_Requirement{
					{
						On: &guardianv1beta1.Policy_Requirement_RequirementTrigger{
							ProviderType: "test-provider-type",
						},
						Appeals: []*guardianv1beta1.Policy_Requirement_AdditionalAppeal{
							{
								Resource: &guardianv1beta1.Policy_Requirement_AdditionalAppeal_ResourceIdentifier{
									Id: "test-resource-id",
								},
								Role: "test-role",
								Policy: &guardianv1beta1.PolicyConfig{
									Id:      "test-policy",
									Version: 1,
								},
								Options: &guardianv1beta1.AppealOptions{
									Duration: "24h",
								},
							},
						},
					},
				},
				Iam: &guardianv1beta1.Policy_IAM{
					Provider: "slack",
					Config:   expectedIAMConfig,
					Schema:   expectedPolicy.IAM.Schema,
				},
				Appeal: &guardianv1beta1.PolicyAppealConfig{
					DurationOptions: []*guardianv1beta1.PolicyAppealConfig_DurationOptions{
						{Name: "1 Day", Value: "24h"},
						{Name: "3 Days", Value: "72h"},
					},
					AllowPermanentAccess:         true,
					AllowActiveAccessExtensionIn: "24h",
					Questions: []*guardianv1beta1.PolicyAppealConfig_Question{
						{
							Key:         "team",
							Question:    "What team are you in?",
							Required:    true,
							Description: "Please provide the name of the team you are in",
						},
					},
				},
			},
		}
		res, err := s.grpcServer.CreatePolicy(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.policyService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if policy service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.policyService.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*domain.Policy")).Return(expectedError).Once()

		req := &guardianv1beta1.CreatePolicyRequest{}
		res, err := s.grpcServer.CreatePolicy(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.policyService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there's an error when parsing the policy", func() {
		s.setup()

		invalidPolicy := &domain.Policy{
			IAM: &domain.IAMConfig{
				Config: make(chan int), // invalid json
			},
		}
		s.policyService.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*domain.Policy")).Return(nil).
			Run(func(_a0 context.Context, _a1 *domain.Policy) {
				*_a1 = *invalidPolicy
			}).Once()

		req := &guardianv1beta1.CreatePolicyRequest{}
		res, err := s.grpcServer.CreatePolicy(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.policyService.AssertExpectations(s.T())
	})
}

func (s *GrpcHandlersSuite) TestUpdatePolicy() {
	s.Run("should return policy details on success", func() {
		s.setup()
		timeNow := time.Now()

		expectedPolicy := &domain.Policy{
			ID:          "test-policy",
			Description: "test-description",
			Steps: []*domain.Step{
				{
					Name:            "test-approval-step",
					Description:     "test-description",
					Strategy:        "auto",
					ApproveIf:       "true",
					RejectionReason: "test-rejection-message",
				},
			},
			Requirements: []*domain.Requirement{
				{
					On: &domain.RequirementTrigger{
						ProviderType: "test-provider-type",
					},
					Appeals: []*domain.AdditionalAppeal{
						{
							Resource: &domain.ResourceIdentifier{
								ID: "test-resource-id",
							},
							Role: "test-role",
							Policy: &domain.PolicyConfig{
								ID:      "test-policy",
								Version: 1,
							},
						},
					},
				},
			},
			IAM: &domain.IAMConfig{
				Provider: "slack",
				Config:   map[string]interface{}{"foo": "bar"},
				Schema:   map[string]string{"foo": "bar"},
			},
			AppealConfig: &domain.PolicyAppealConfig{
				DurationOptions: []domain.AppealDurationOption{
					{Name: "1 Day", Value: "24h"},
					{Name: "3 Days", Value: "72h"},
				},
			},
		}
		expectedVersion := uint(1)
		expectedIAMConfig, err := structpb.NewValue(expectedPolicy.IAM.Config)
		s.Require().NoError(err)
		expectedResponse := &guardianv1beta1.UpdatePolicyResponse{
			Policy: &guardianv1beta1.Policy{
				Id:          expectedPolicy.ID,
				Version:     uint32(expectedVersion),
				Description: expectedPolicy.Description,
				Steps: []*guardianv1beta1.Policy_ApprovalStep{
					{
						Name:            "test-approval-step",
						Description:     "test-description",
						Strategy:        "auto",
						ApproveIf:       "true",
						RejectionReason: "test-rejection-message",
					},
				},
				Requirements: []*guardianv1beta1.Policy_Requirement{
					{
						On: &guardianv1beta1.Policy_Requirement_RequirementTrigger{
							ProviderType: "test-provider-type",
						},
						Appeals: []*guardianv1beta1.Policy_Requirement_AdditionalAppeal{
							{
								Resource: &guardianv1beta1.Policy_Requirement_AdditionalAppeal_ResourceIdentifier{
									Id: "test-resource-id",
								},
								Role: "test-role",
								Policy: &guardianv1beta1.PolicyConfig{
									Id:      "test-policy",
									Version: 1,
								},
							},
						},
					},
				},
				Iam: &guardianv1beta1.Policy_IAM{
					Provider: "slack",
					Config:   expectedIAMConfig,
					Schema:   expectedPolicy.IAM.Schema,
				},
				Appeal: &guardianv1beta1.PolicyAppealConfig{
					DurationOptions: []*guardianv1beta1.PolicyAppealConfig_DurationOptions{
						{Name: "1 Day", Value: "24h"},
						{Name: "3 Days", Value: "72h"},
					},
				},
				CreatedAt: timestamppb.New(timeNow),
				UpdatedAt: timestamppb.New(timeNow),
			},
		}
		s.policyService.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), expectedPolicy).
			Run(func(_a0 context.Context, _a1 *domain.Policy) {
				_a1.CreatedAt = timeNow
				_a1.UpdatedAt = timeNow
				_a1.Version = expectedVersion
			}).Return(nil).Once()

		req := &guardianv1beta1.UpdatePolicyRequest{
			Id: expectedPolicy.ID,
			Policy: &guardianv1beta1.Policy{
				Description: expectedPolicy.Description,
				Steps: []*guardianv1beta1.Policy_ApprovalStep{
					{
						Name:            "test-approval-step",
						Description:     "test-description",
						Strategy:        "auto",
						ApproveIf:       "true",
						RejectionReason: "test-rejection-message",
					},
				},
				Requirements: []*guardianv1beta1.Policy_Requirement{
					{
						On: &guardianv1beta1.Policy_Requirement_RequirementTrigger{
							ProviderType: "test-provider-type",
						},
						Appeals: []*guardianv1beta1.Policy_Requirement_AdditionalAppeal{
							{
								Resource: &guardianv1beta1.Policy_Requirement_AdditionalAppeal_ResourceIdentifier{
									Id: "test-resource-id",
								},
								Role: "test-role",
								Policy: &guardianv1beta1.PolicyConfig{
									Id:      "test-policy",
									Version: 1,
								},
							},
						},
					},
				},
				Iam: &guardianv1beta1.Policy_IAM{
					Provider: "slack",
					Config:   expectedIAMConfig,
					Schema:   expectedPolicy.IAM.Schema,
				},
				Appeal: &guardianv1beta1.PolicyAppealConfig{
					DurationOptions: []*guardianv1beta1.PolicyAppealConfig_DurationOptions{
						{Name: "1 Day", Value: "24h"},
						{Name: "3 Days", Value: "72h"},
					},
				},
			},
		}
		res, err := s.grpcServer.UpdatePolicy(context.Background(), req)

		s.NoError(err)
		s.Equal(expectedResponse, res)
		s.policyService.AssertExpectations(s.T())
	})

	s.Run("should return not found error if policy not found", func() {
		s.setup()

		expectedError := policy.ErrPolicyNotFound
		s.policyService.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*domain.Policy")).Return(expectedError).Once()

		req := &guardianv1beta1.UpdatePolicyRequest{}
		res, err := s.grpcServer.UpdatePolicy(context.Background(), req)

		s.Equal(codes.NotFound, status.Code(err))
		s.Nil(res)
		s.policyService.AssertExpectations(s.T())
	})

	s.Run("should return invalid argument error if policy id is empty", func() {
		s.setup()

		expectedError := policy.ErrEmptyIDParam
		s.policyService.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*domain.Policy")).Return(expectedError).Once()

		req := &guardianv1beta1.UpdatePolicyRequest{}
		res, err := s.grpcServer.UpdatePolicy(context.Background(), req)

		s.Equal(codes.InvalidArgument, status.Code(err))
		s.Nil(res)
		s.policyService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if policy service returns an error", func() {
		s.setup()

		expectedError := errors.New("random error")
		s.policyService.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*domain.Policy")).Return(expectedError).Once()

		req := &guardianv1beta1.UpdatePolicyRequest{}
		res, err := s.grpcServer.UpdatePolicy(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.policyService.AssertExpectations(s.T())
	})

	s.Run("should return internal error if there's an error when parsing the policy", func() {
		s.setup()

		invalidPolicy := &domain.Policy{
			IAM: &domain.IAMConfig{
				Config: make(chan int), // invalid json
			},
		}
		s.policyService.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*domain.Policy")).Return(nil).
			Run(func(_a0 context.Context, _a1 *domain.Policy) {
				*_a1 = *invalidPolicy
			}).Once()

		req := &guardianv1beta1.UpdatePolicyRequest{}
		res, err := s.grpcServer.UpdatePolicy(context.Background(), req)

		s.Equal(codes.Internal, status.Code(err))
		s.Nil(res)
		s.policyService.AssertExpectations(s.T())
	})
}
