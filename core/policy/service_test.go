package policy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/goto/guardian/core/policy"
	policymocks "github.com/goto/guardian/core/policy/mocks"
	"github.com/goto/guardian/core/provider"
	"github.com/goto/guardian/core/resource"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/mocks"
	"github.com/goto/guardian/plugins/identities"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockPolicyRepository *policymocks.Repository
	mockResourceService  *policymocks.ResourceService
	mockProviderService  *policymocks.ProviderService
	mockAuditLogger      *policymocks.AuditLogger
	mockCrypto           *mocks.Crypto
	service              *policy.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockPolicyRepository = new(policymocks.Repository)
	s.mockResourceService = new(policymocks.ResourceService)
	s.mockProviderService = new(policymocks.ProviderService)
	s.mockAuditLogger = new(policymocks.AuditLogger)

	s.mockCrypto = new(mocks.Crypto)
	v := validator.New()
	iamManager := identities.NewManager(s.mockCrypto, v)

	s.service = policy.NewService(policy.ServiceDeps{
		Repository:      s.mockPolicyRepository,
		ResourceService: s.mockResourceService,
		ProviderService: s.mockProviderService,
		IAMManager:      iamManager,
		AuditLogger:     s.mockAuditLogger,
		Validator:       validator.New(),
	})
}

func (s *ServiceTestSuite) TestCreate() {
	s.Run("should return error if policy is invalid", func() {
		validSteps := []*domain.Step{
			{
				Name: "step-1",
			},
		}

		testCases := []struct {
			name          string
			policy        *domain.Policy
			expectedError error
		}{
			{
				name: "id contains space(s)",
				policy: &domain.Policy{
					ID:      "a a",
					Version: 1,
					Steps:   validSteps,
				},
				expectedError: policy.ErrIDContainsWhitespaces,
			},
			{
				name: "id contains tab(s)",
				policy: &domain.Policy{
					ID:      "a	a",
					Version: 1,
					Steps:   validSteps,
				},
				expectedError: policy.ErrIDContainsWhitespaces,
			},
			{
				name: "nil steps",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
				},
			},
			{
				name: "empty steps",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps:   []*domain.Step{},
				},
			},
			{
				name: "step: empty name",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{},
					},
				},
			},
			{
				name: "step: with empty strategy",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:     "step-1",
							Strategy: "",
						},
					},
				},
			},
			{
				name: "step: empty ApproveIf",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:      "step-1",
							Strategy:  "auto",
							ApproveIf: "",
						},
					},
				},
			},
			{
				name: "step: empty approvers",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:      "step-1",
							Strategy:  "manual",
							Approvers: []string{},
						},
					},
				},
			},
			{
				name: "step: step with strategy:auto doesn't contain ApproveIf",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:     "step-1",
							Strategy: "auto",
							Approvers: []string{
								"$resource.field",
							},
						},
					},
				},
			},
			{
				name: "step: step with strategy:manual doesn't contain Approvers",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:      "step-1",
							Strategy:  "manual",
							ApproveIf: "true",
						},
					},
				},
			},
			{
				name: "step: invalid strategy",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:     "step-1",
							Strategy: "invalid-strategy",
							Approvers: []string{
								"$resource.field",
							},
						},
					},
				},
			},
			{
				name: "step: name contains whitespaces",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:     "a a",
							Strategy: "manual",
							Approvers: []string{
								"$appeal.field",
							},
						},
					},
				},
				expectedError: policy.ErrStepNameContainsWhitespaces,
			},
			{
				name: "step: invalid approvers key",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:     "step-1",
							Strategy: "manual",
							Approvers: []string{
								"$x",
							},
						},
					},
				},
			},
			{
				name: "step: invalid AllowActiveAccessExtensionIn",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:     "step-1",
							Strategy: "manual",
							Approvers: []string{
								"approver@email.com",
							},
						},
					},
					AppealConfig: &domain.PolicyAppealConfig{
						AllowActiveAccessExtensionIn: "invalid-duration",
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
				},
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				actualError := s.service.Create(context.Background(), tc.policy)

				s.Error(actualError)
				if tc.expectedError != nil {
					s.ErrorIs(actualError, tc.expectedError)
				}
			})
		}

		s.Run("iam: invalid iam provider", func() {
			policy := &domain.Policy{
				ID:      "test-id",
				Version: 1,
				Steps: []*domain.Step{
					{
						Name:     "step-1",
						Strategy: "manual",
						Approvers: []string{
							"approver@email.com",
						},
					},
				},
				IAM: &domain.IAMConfig{
					Provider: "invalid-provider",
				},
			}
			expectedError := identities.ErrUnknownProviderType

			actualError := s.service.Create(context.Background(), policy)

			s.ErrorIs(actualError, expectedError)
		})
	})

	validPolicy := &domain.Policy{
		ID:      "id",
		Version: 1,
		Steps: []*domain.Step{
			{
				Name:     "test",
				Strategy: "manual",
				Approvers: []string{
					"user@email.com",
				},
			},
		},
		AppealConfig: &domain.PolicyAppealConfig{
			DurationOptions: []domain.AppealDurationOption{
				{
					Name:  "1 day",
					Value: "24h",
				},
				{
					Name:  "2 days",
					Value: "48h",
				},
			},
			AllowPermanentAccess:         false,
			AllowActiveAccessExtensionIn: "24h",
		},
		IAM: &domain.IAMConfig{
			Provider: "http",
			Config: map[string]interface{}{
				"url": "http://test-localhost:8080",
				"auth": map[string]interface{}{
					"type":     "basic",
					"username": "test-user",
					"password": "test-password",
				},
			},
		},
	}

	s.Run("should return error if got error from the policy repository", func() {
		expectedError := errors.New("error from repository")
		s.mockPolicyRepository.EXPECT().Create(mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.Anything).Return(expectedError).Once()
		s.mockCrypto.EXPECT().Encrypt("test-password").Return("test-password", nil).Once()
		s.mockCrypto.EXPECT().Decrypt("test-password").Return("test-password", nil).Once()

		actualError := s.service.Create(context.Background(), validPolicy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should set initial version to 1", func() {
		p := &domain.Policy{
			ID:    "test",
			Steps: validPolicy.Steps,
		}

		expectedVersion := uint(1)
		s.mockPolicyRepository.EXPECT().Create(mock.MatchedBy(func(ctx context.Context) bool { return true }), p).Return(nil).Once()
		s.mockCrypto.EXPECT().Encrypt("test-password").Return("test-password", nil).Once()
		s.mockCrypto.EXPECT().Decrypt("test-password").Return("test-password", nil).Once()
		s.mockAuditLogger.EXPECT().Log(mock.Anything, policy.AuditKeyPolicyCreate, mock.Anything).Return(nil).Once()

		actualError := s.service.Create(context.Background(), p)

		s.Nil(actualError)
		s.Equal(expectedVersion, p.Version)
		s.mockPolicyRepository.AssertExpectations(s.T())
		s.mockAuditLogger.AssertExpectations(s.T())
	})

	s.Run("should pass the model from the param", func() {
		s.mockPolicyRepository.EXPECT().Create(mock.MatchedBy(func(ctx context.Context) bool { return true }), validPolicy).Return(nil).Once()
		s.mockAuditLogger.EXPECT().Log(mock.Anything, policy.AuditKeyPolicyCreate, mock.Anything).Return(nil).Once()
		s.mockCrypto.EXPECT().Encrypt("test-password").Return("test-password", nil).Once()
		s.mockCrypto.EXPECT().Decrypt("test-password").Return("test-password", nil).Once()

		actualError := s.service.Create(context.Background(), validPolicy)

		s.Nil(actualError)
		s.mockPolicyRepository.AssertExpectations(s.T())
		s.mockAuditLogger.AssertExpectations(s.T())
	})

	s.Run("with dryRun true", func() {
		s.Run("with valid policy should not call repository", func() {
			ctx := policy.WithDryRun(context.Background())

			actualError := s.service.Create(ctx, validPolicy)

			s.Nil(actualError)
			s.mockPolicyRepository.AssertNotCalled(s.T(), "Create")
			s.mockAuditLogger.AssertNotCalled(s.T(), "Log")
		})
	})
}

func (s *ServiceTestSuite) TestPolicyRequirements() {
	s.Run("validations", func() {
		testCases := []struct {
			name         string
			requirements []*domain.Requirement

			expectedResource                *domain.Resource
			expectedResourceServiceGetError error

			expectedProvider                   *domain.Provider
			expectedProviderServiceGetOneError error

			expectedProviderServiceValidateAppealError error
		}{
			{
				name: "target resource doesn't exist",
				requirements: []*domain.Requirement{
					{
						Appeals: []*domain.AdditionalAppeal{
							{
								Resource: &domain.ResourceIdentifier{
									ID: "1",
								},
							},
						},
					},
				},
				expectedResource:                nil,
				expectedResourceServiceGetError: resource.ErrRecordNotFound,
			},
			{
				name: "provider not found/deleted",
				requirements: []*domain.Requirement{
					{
						Appeals: []*domain.AdditionalAppeal{
							{
								Resource: &domain.ResourceIdentifier{
									ID: "1",
								},
							},
						},
					},
				},
				expectedResource: &domain.Resource{
					ProviderType: "test-provider-type",
					ProviderURN:  "test-provider-urn",
				},
				expectedProvider:                   nil,
				expectedProviderServiceGetOneError: provider.ErrRecordNotFound,
			},
			{
				name: "provider invalidates appeal",
				requirements: []*domain.Requirement{
					{
						Appeals: []*domain.AdditionalAppeal{
							{
								Resource: &domain.ResourceIdentifier{
									ID: "1",
								},
							},
						},
					},
				},
				expectedResource: &domain.Resource{
					ProviderType: "test-provider-type",
					ProviderURN:  "test-provider-urn",
				},
				expectedProvider: &domain.Provider{},
				expectedProviderServiceValidateAppealError: errors.New("test invalid appeal"),
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				policy := &domain.Policy{
					ID:      "policy-tes",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name: "step-test",
							Approvers: []string{
								"user@email.com",
							},
						},
					},
					Requirements: tc.requirements,
				}

				for _, r := range tc.requirements {
					for _, aa := range r.Appeals {
						s.mockResourceService.
							EXPECT().Get(mock.Anything, &domain.ResourceIdentifier{}).
							Return(tc.expectedResource, tc.expectedResourceServiceGetError).
							Once()
						if tc.expectedResource != nil {
							s.mockProviderService.
								EXPECT().GetOne(mock.Anything, tc.expectedResource.ProviderType, tc.expectedResource.ProviderURN).
								Return(tc.expectedProvider, tc.expectedProviderServiceGetOneError).
								Once()
							if tc.expectedProviderServiceGetOneError == nil {
								expectedAppeal := &domain.Appeal{
									ResourceID: tc.expectedResource.ID,
									Resource:   tc.expectedResource,
									Role:       aa.Role,
									Options:    aa.Options,
								}
								s.mockProviderService.
									EXPECT().ValidateAppeal(mock.Anything, expectedAppeal, tc.expectedProvider, mock.Anything).
									Return(tc.expectedProviderServiceValidateAppealError).
									Once()
							}
						}
					}
				}

				actualError := s.service.Create(context.Background(), policy)

				s.Error(actualError)
			})
		}
	})

	s.Run("valid requirements", func() {
		resourceID := uuid.New().String()
		expectedResource := &domain.Resource{
			ID:           resourceID,
			ProviderType: "provider-type-test",
			ProviderURN:  "provider-urn-test",
		}
		expectedProvider := &domain.Provider{}
		additionalAppeals := []*domain.AdditionalAppeal{
			{
				Resource: &domain.ResourceIdentifier{
					ID: resourceID,
				},
				Role: "viewer",
			},
		}

		testCases := []struct {
			name         string
			requirements []*domain.Requirement
		}{
			{
				name: "requirement condition on ProviderType only",
				requirements: []*domain.Requirement{
					{
						On: &domain.RequirementTrigger{
							ProviderType: "test-bigquery",
						},
						Appeals: additionalAppeals,
					},
				},
			},
			{
				name: "requirement condition on Role only",
				requirements: []*domain.Requirement{
					{
						On: &domain.RequirementTrigger{
							Role: "test-viewer",
						},
						Appeals: additionalAppeals,
					},
				},
			},
			{
				name: "appeal identifier using provider type+urn and resource type+urn",
				requirements: []*domain.Requirement{
					{
						On: &domain.RequirementTrigger{
							Role: "test-viewer",
						},
						Appeals: []*domain.AdditionalAppeal{
							{
								Resource: &domain.ResourceIdentifier{
									ProviderType: "test-provider-type",
									ProviderURN:  "test-provider-urn",
									Type:         "test-type",
									URN:          "test-urn",
								},
								Role: "viewer",
							},
						},
					},
				},
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				p := &domain.Policy{
					ID:      "policy-test",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:     "step-test",
							Strategy: "manual",
							Approvers: []string{
								"user@email.com",
							},
						},
					},
					Requirements: tc.requirements,
				}

				for _, r := range tc.requirements {
					for _, aa := range r.Appeals {
						s.mockResourceService.
							EXPECT().Get(mock.Anything, aa.Resource).
							Return(expectedResource, nil).
							Once()
						s.mockProviderService.
							EXPECT().GetOne(mock.Anything, expectedResource.ProviderType, expectedResource.ProviderURN).
							Return(expectedProvider, nil).
							Once()
						expectedAppeal := &domain.Appeal{
							ResourceID: expectedResource.ID,
							Resource:   expectedResource,
							Role:       aa.Role,
							Options:    aa.Options,
						}
						expectedAppeal.SetDefaults()
						s.mockProviderService.
							EXPECT().ValidateAppeal(mock.Anything, expectedAppeal, expectedProvider, mock.Anything).
							Return(nil).
							Once()
					}
				}
				s.mockPolicyRepository.EXPECT().Create(mock.MatchedBy(func(ctx context.Context) bool { return true }), p).Return(nil).Once()
				s.mockAuditLogger.EXPECT().Log(mock.Anything, policy.AuditKeyPolicyCreate, mock.Anything).Return(nil).Once()

				actualError := s.service.Create(context.Background(), p)
				s.Nil(actualError)
			})
		}
	})
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockPolicyRepository.EXPECT().Find(mock.MatchedBy(func(ctx context.Context) bool { return true })).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find(context.Background())

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedResult := []*domain.Policy{
			{
				IAM: &domain.IAMConfig{
					Provider: "http",
					Config: map[string]interface{}{
						"url": "http://test-localhost:8080",
						"auth": map[string]interface{}{
							"type":     "basic",
							"username": "test-user",
							"password": "test-password",
						},
					},
				},
			},
		}
		s.mockPolicyRepository.EXPECT().Find(mock.MatchedBy(func(ctx context.Context) bool { return true })).Return(expectedResult, nil).Once()
		s.mockCrypto.EXPECT().Decrypt("test-password").Return("test-password", nil).Once()

		actualResult, actualError := s.service.Find(context.Background())

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
		s.mockPolicyRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestGetOne() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockPolicyRepository.EXPECT().GetOne(mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.Anything, mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.GetOne(context.Background(), "", 0)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedResult := &domain.Policy{
			IAM: &domain.IAMConfig{
				Provider: "http",
				Config: map[string]interface{}{
					"url": "http://test-localhost:8080",
					"auth": map[string]interface{}{
						"type":     "basic",
						"username": "test-user",
						"password": "test-password",
					},
				},
			},
		}
		s.mockPolicyRepository.EXPECT().GetOne(mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.Anything, mock.Anything).Return(expectedResult, nil).Once()
		s.mockCrypto.EXPECT().Decrypt("test-password").Return("test-password", nil).Once()

		actualResult, actualError := s.service.GetOne(context.Background(), "", 0)

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
		s.mockPolicyRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestUpdate() {
	s.Run("should return error if policy id doesn't exists", func() {
		p := &domain.Policy{}
		expectedError := policy.ErrEmptyIDParam

		actualError := s.service.Update(context.Background(), p)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return increment policy version", func() {
		p := &domain.Policy{
			ID:      "id",
			Version: 5,
			Steps: []*domain.Step{
				{
					Name:     "test",
					Strategy: "manual",
					Approvers: []string{
						"user@email.com",
					},
				},
			},
			IAM: &domain.IAMConfig{
				Provider: "http",
				Config: map[string]interface{}{
					"url": "http://test-localhost:8080",
					"auth": map[string]interface{}{
						"type":     "basic",
						"username": "test-user",
						"password": "test-password",
					},
				},
			},
		}

		expectedLatestPolicy := &domain.Policy{
			ID:      p.ID,
			Version: 5,
		}
		expectedNewVersion := uint(6)
		s.mockPolicyRepository.EXPECT().GetOne(mock.MatchedBy(func(ctx context.Context) bool { return true }), p.ID, uint(0)).Return(expectedLatestPolicy, nil).Once()
		s.mockPolicyRepository.EXPECT().Create(mock.MatchedBy(func(ctx context.Context) bool { return true }), p).Return(nil)
		s.mockCrypto.EXPECT().Encrypt("test-password").Return("test-password", nil).Once()
		s.mockCrypto.EXPECT().Decrypt("test-password").Return("test-password", nil).Once()
		s.mockAuditLogger.EXPECT().Log(mock.Anything, policy.AuditKeyPolicyUpdate, mock.Anything).Return(nil).Once()

		s.service.Update(context.Background(), p)

		s.mockPolicyRepository.AssertExpectations(s.T())
		s.mockAuditLogger.AssertExpectations(s.T())
		s.Equal(expectedNewVersion, p.Version)
	})

	s.Run("with dryRun true", func() {
		s.Run("with valid policy should not call repository", func() {
			p := &domain.Policy{
				ID:      "id",
				Version: 5,
				Steps: []*domain.Step{
					{
						Name:     "test",
						Strategy: "manual",
						Approvers: []string{
							"user@email.com",
						},
					},
				},
			}

			expectedLatestPolicy := &domain.Policy{
				ID:      p.ID,
				Version: 5,
			}
			expectedNewVersion := uint(6)

			ctx := policy.WithDryRun(context.Background())

			s.mockPolicyRepository.EXPECT().GetOne(ctx, p.ID, uint(0)).Return(expectedLatestPolicy, nil).Once()

			actualError := s.service.Update(ctx, p)

			s.Nil(actualError)

			s.mockPolicyRepository.AssertExpectations(s.T())
			s.mockPolicyRepository.AssertNotCalled(s.T(), "Create")
			s.mockAuditLogger.AssertNotCalled(s.T(), "Log")
			s.Equal(expectedNewVersion, p.Version)
		})
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
