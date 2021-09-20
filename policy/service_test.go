package policy_test

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/policy"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockPolicyRepository *mocks.PolicyRepository
	service              *policy.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockPolicyRepository = new(mocks.PolicyRepository)
	s.service = policy.NewService(validator.New(), s.mockPolicyRepository)
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
							Conditions: []*domain.Condition{},
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
			{
				name: "step: dependency doesn't exists",
				policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:      "step-1",
							Approvers: "$resource.field",
						},
						{
							Name:         "step-2",
							Approvers:    "$resource.field",
							Dependencies: []string{"step-x"},
						},
					},
				},
				expectedError: policy.ErrStepDependencyDoesNotExists,
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
