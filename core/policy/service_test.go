package policy_test

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/core/policy"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/core/resource"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/identities"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockPolicyRepository *mocks.PolicyRepository
	mockResourceService  *mocks.ResourceService
	mockProviderService  *mocks.ProviderService
	service              *policy.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockPolicyRepository = new(mocks.PolicyRepository)
	s.mockResourceService = new(mocks.ResourceService)
	s.mockProviderService = new(mocks.ProviderService)

	mockCrypto := new(mocks.Crypto)
	v := validator.New()
	iamManager := identities.NewManager(mockCrypto, v)

	s.service = policy.NewService(
		validator.New(),
		s.mockPolicyRepository,
		s.mockResourceService,
		s.mockProviderService,
		iamManager,
	)
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
					ID: "a	a",
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
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				actualError := s.service.Create(tc.policy)

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

			actualError := s.service.Create(policy)

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
	}

	s.Run("should return error if got error from the policy repository", func() {
		expectedError := errors.New("error from repository")
		s.mockPolicyRepository.On("Create", mock.Anything).Return(expectedError).Once()

		actualError := s.service.Create(validPolicy)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should set initial version to 1", func() {
		p := &domain.Policy{
			ID:    "test",
			Steps: validPolicy.Steps,
		}

		expectedVersion := uint(1)
		s.mockPolicyRepository.On("Create", p).Return(nil).Once()

		actualError := s.service.Create(p)

		s.Nil(actualError)
		s.Equal(expectedVersion, p.Version)
		s.mockPolicyRepository.AssertExpectations(s.T())
	})

	s.Run("should pass the model from the param", func() {
		s.mockPolicyRepository.On("Create", validPolicy).Return(nil).Once()

		actualError := s.service.Create(validPolicy)

		s.Nil(actualError)
		s.mockPolicyRepository.AssertExpectations(s.T())
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
									ID: 1,
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
									ID: 1,
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
									ID: 1,
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
							On("Get", &domain.ResourceIdentifier{}).
							Return(tc.expectedResource, tc.expectedResourceServiceGetError).
							Once()
						if tc.expectedResource != nil {
							s.mockProviderService.
								On("GetOne", tc.expectedResource.ProviderType, tc.expectedResource.ProviderURN).
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
									On("ValidateAppeal", expectedAppeal, tc.expectedProvider).
									Return(tc.expectedProviderServiceValidateAppealError).
									Once()
							}
						}
					}
				}

				actualError := s.service.Create(policy)

				s.Error(actualError)
			})
		}
	})

	s.Run("valid requirements", func() {
		expectedResource := &domain.Resource{
			ID:           1,
			ProviderType: "provider-type-test",
			ProviderURN:  "provider-urn-test",
		}
		expectedProvider := &domain.Provider{}
		additionalAppeals := []*domain.AdditionalAppeal{
			{
				Resource: &domain.ResourceIdentifier{
					ID: 1,
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
				policy := &domain.Policy{
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
							On("Get", aa.Resource).
							Return(expectedResource, nil).
							Once()
						s.mockProviderService.
							On("GetOne", expectedResource.ProviderType, expectedResource.ProviderURN).
							Return(expectedProvider, nil).
							Once()
						expectedAppeal := &domain.Appeal{
							ResourceID: expectedResource.ID,
							Resource:   expectedResource,
							Role:       aa.Role,
							Options:    aa.Options,
						}
						s.mockProviderService.
							On("ValidateAppeal", expectedAppeal, expectedProvider).
							Return(nil).
							Once()
					}
				}
				s.mockPolicyRepository.On("Create", policy).Return(nil).Once()

				actualError := s.service.Create(policy)
				s.Nil(actualError)
			})
		}
	})
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockPolicyRepository.On("Find").Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find()

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedResult := []*domain.Policy{}
		s.mockPolicyRepository.On("Find").Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.Find()

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
		s.mockPolicyRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestGetOne() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockPolicyRepository.On("GetOne", mock.Anything, mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.GetOne("", 0)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedResult := &domain.Policy{}
		s.mockPolicyRepository.On("GetOne", mock.Anything, mock.Anything).Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.GetOne("", 0)

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
		s.mockPolicyRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestUpdate() {
	s.Run("should return error if policy id doesn't exists", func() {
		p := &domain.Policy{}
		expectedError := policy.ErrEmptyIDParam

		actualError := s.service.Update(p)

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return increment policy version", func() {
		p := &domain.Policy{
			ID: "id",
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
		s.mockPolicyRepository.On("GetOne", p.ID, p.Version).Return(expectedLatestPolicy, nil).Once()
		s.mockPolicyRepository.On("Create", p).Return(nil)

		s.service.Update(p)

		s.mockPolicyRepository.AssertExpectations(s.T())
		s.Equal(expectedNewVersion, p.Version)
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
