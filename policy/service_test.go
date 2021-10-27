package policy_test

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/policy"
	"github.com/odpf/guardian/provider"
	"github.com/odpf/guardian/resource"
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
	s.service = policy.NewService(validator.New(), s.mockPolicyRepository, s.mockResourceService, s.mockProviderService)
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
				name: "step: empty conditions",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:       "step-1",
							Conditions: nil,
						},
					},
				},
			},
			{
				name: "step: without approvers/conditions",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name: "step-1",
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
							Name:      "a a",
							Approvers: "$resource.field",
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
							Name:      "step-1",
							Approvers: "$x",
						},
					},
				},
				expectedError: policy.ErrInvalidApprovers,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				actualError := s.service.Create(tc.policy)

				s.Error(actualError)
				if tc.expectedError != nil {
					s.Contains(actualError.Error(), tc.expectedError.Error())
				}
			})
		}
	})

	validPolicy := &domain.Policy{
		ID:      "id",
		Version: 1,
		Steps: []*domain.Step{
			{
				Name:      "test",
				Approvers: "user@email.com",
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
							Name:      "step-test",
							Approvers: "user@email.com",
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
					Name:      "test",
					Approvers: "user@email.com",
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
