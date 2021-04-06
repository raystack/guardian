package appeal_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository             *mocks.AppealRepository
	mockApprovalService        *mocks.ApprovalService
	mockResourceService        *mocks.ResourceService
	mockProviderService        *mocks.ProviderService
	mockPolicyService          *mocks.PolicyService
	mockIdentityManagerService *mocks.IdentityManagerService

	service *appeal.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(mocks.AppealRepository)
	s.mockApprovalService = new(mocks.ApprovalService)
	s.mockResourceService = new(mocks.ResourceService)
	s.mockProviderService = new(mocks.ProviderService)
	s.mockPolicyService = new(mocks.PolicyService)
	s.mockIdentityManagerService = new(mocks.IdentityManagerService)

	s.service = appeal.NewService(
		s.mockRepository,
		s.mockApprovalService,
		s.mockResourceService,
		s.mockProviderService,
		s.mockPolicyService,
		s.mockIdentityManagerService,
	)
}

func (s *ServiceTestSuite) TestGetByID() {
	s.Run("should return error if id is empty/0", func() {
		expectedError := appeal.ErrAppealIDEmptyParam

		actualResult, actualError := s.service.GetByID(0)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got any from repository", func() {
		expectedError := errors.New("repository error")
		s.mockRepository.On("GetByID", mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.GetByID(1)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return record on success", func() {
		expectedID := uint(1)
		expectedResult := &domain.Appeal{
			ID: uint(expectedID),
		}
		s.mockRepository.On("GetByID", expectedID).Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.GetByID(expectedID)

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
	})
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return error if got any from repository", func() {
		expectedError := errors.New("unexpected repository error")
		s.mockRepository.On("Find", mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find(map[string]interface{}{})

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return records on success", func() {
		expectedFilters := map[string]interface{}{
			"user": "user@email.com",
		}
		expectedResult := []*domain.Appeal{
			{
				ID:         1,
				ResourceID: 1,
				User:       "user@email.com",
				Role:       "viewer",
			},
		}
		s.mockRepository.On("Find", expectedFilters).Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.Find(expectedFilters)

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
	})
}

func (s *ServiceTestSuite) TestCreate() {
	s.Run("should return error if got error from resource service", func() {
		expectedError := errors.New("resource service error")
		s.mockResourceService.On("Find", mock.Anything).Return(nil, expectedError).Once()

		actualError := s.service.Create([]*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from provider service", func() {
		expectedResources := []*domain.Resource{}
		s.mockResourceService.On("Find", mock.Anything).Return(expectedResources, nil).Once()
		expectedError := errors.New("provider service error")
		s.mockProviderService.On("Find").Return(nil, expectedError).Once()

		actualError := s.service.Create([]*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from policy service", func() {
		expectedResources := []*domain.Resource{}
		expectedProviders := []*domain.Provider{}
		s.mockResourceService.On("Find", mock.Anything).Return(expectedResources, nil).Once()
		s.mockProviderService.On("Find").Return(expectedProviders, nil).Once()
		expectedError := errors.New("policy service error")
		s.mockPolicyService.On("Find").Return(nil, expectedError).Once()

		actualError := s.service.Create([]*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from repository", func() {
		expectedResources := []*domain.Resource{}
		expectedProviders := []*domain.Provider{}
		expectedPolicies := []*domain.Policy{}
		s.mockResourceService.On("Find", mock.Anything).Return(expectedResources, nil).Once()
		s.mockProviderService.On("Find").Return(expectedProviders, nil).Once()
		s.mockPolicyService.On("Find").Return(expectedPolicies, nil).Once()
		expectedError := errors.New("repository error")
		s.mockRepository.On("BulkInsert", mock.Anything).Return(expectedError).Once()

		actualError := s.service.Create([]*domain.Appeal{})

		s.EqualError(actualError, expectedError.Error())
	})

	user := "test@email.com"
	resourceIDs := []uint{1, 2}
	resources := []*domain.Resource{}
	for _, id := range resourceIDs {
		resources = append(resources, &domain.Resource{
			ID:           id,
			Type:         "resource_type_1",
			ProviderType: "provider_type",
			ProviderURN:  "provider1",
		})
	}
	providers := []*domain.Provider{
		{
			ID:   1,
			Type: "provider_type",
			URN:  "provider1",
			Config: &domain.ProviderConfig{
				Resources: []*domain.ResourceConfig{
					{
						Type: "resource_type_1",
						Policy: &domain.PolicyConfig{
							ID:      "policy_1",
							Version: 1,
						},
					},
				},
			},
		},
	}
	policies := []*domain.Policy{
		{
			ID:      "policy_1",
			Version: 1,
			Steps: []*domain.Step{
				{
					Name: "step_1",
				},
				{
					Name: "step_2",
				},
			},
		},
	}
	expectedAppealsInsertionParam := []*domain.Appeal{}
	for _, r := range resourceIDs {
		expectedAppealsInsertionParam = append(expectedAppealsInsertionParam, &domain.Appeal{
			ResourceID:    r,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
			Status:        domain.AppealStatusPending,
			User:          user,
			Approvals: []*domain.Approval{
				{
					Name:          "step_1",
					Status:        domain.ApprovalStatusPending,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
				},
				{
					Name:          "step_2",
					Status:        domain.ApprovalStatusPending,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
				},
			},
		})
	}
	expectedResult := []*domain.Appeal{
		{
			ID:            1,
			ResourceID:    1,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
			Status:        domain.AppealStatusPending,
			User:          user,
			Approvals: []*domain.Approval{
				{
					ID:            1,
					Name:          "step_1",
					Status:        domain.ApprovalStatusPending,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
				},
				{
					ID:            2,
					Name:          "step_2",
					Status:        domain.ApprovalStatusPending,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
				},
			},
		},
		{
			ID:            2,
			ResourceID:    2,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
			Status:        domain.AppealStatusPending,
			User:          user,
			Approvals: []*domain.Approval{
				{
					ID:            1,
					Name:          "step_1",
					Status:        domain.ApprovalStatusPending,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
				},
				{
					ID:            2,
					Name:          "step_2",
					Status:        domain.ApprovalStatusPending,
					PolicyID:      "policy_1",
					PolicyVersion: 1,
				},
			},
		},
	}

	s.Run("should return appeals on success", func() {
		expectedFilters := map[string]interface{}{"ids": resourceIDs}
		s.mockResourceService.On("Find", expectedFilters).Return(resources, nil).Once()
		s.mockProviderService.On("Find").Return(providers, nil).Once()
		s.mockPolicyService.On("Find").Return(policies, nil).Once()
		s.mockRepository.
			On("BulkInsert", expectedAppealsInsertionParam).
			Return(nil).
			Run(func(args mock.Arguments) {
				appeals := args.Get(0).([]*domain.Appeal)
				for i, a := range appeals {
					a.ID = expectedResult[i].ID
					for j, approval := range a.Approvals {
						approval.ID = expectedResult[i].Approvals[j].ID
					}
				}
			}).
			Once()

		appeals := []*domain.Appeal{
			{
				User:       user,
				ResourceID: 1,
			},
			{
				User:       user,
				ResourceID: 2,
			},
		}
		actualError := s.service.Create(appeals)

		s.Equal(expectedResult, appeals)
		s.Nil(actualError)
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
