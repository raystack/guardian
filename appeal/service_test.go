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
	mockRepository      *mocks.AppealRepository
	mockResourceService *mocks.ResourceService
	mockProviderService *mocks.ProviderService

	service *appeal.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(mocks.AppealRepository)
	s.mockResourceService = new(mocks.ResourceService)
	s.mockProviderService = new(mocks.ProviderService)

	s.service = appeal.NewService(s.mockRepository, s.mockResourceService, s.mockProviderService)
}

func (s *ServiceTestSuite) TestCreate() {
	s.Run("should return error if got error from resource service", func() {
		expectedError := errors.New("resource service error")
		s.mockResourceService.On("Find", mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Create("", []uint{})

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from provider service", func() {
		expectedResources := []*domain.Resource{}
		s.mockResourceService.On("Find", mock.Anything).Return(expectedResources, nil).Once()
		expectedError := errors.New("provider service error")
		s.mockProviderService.On("Find").Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Create("", []uint{})

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return error if got error from repository", func() {
		expectedResources := []*domain.Resource{}
		expectedProviders := []*domain.Provider{}
		s.mockResourceService.On("Find", mock.Anything).Return(expectedResources, nil).Once()
		s.mockProviderService.On("Find").Return(expectedProviders, nil).Once()
		expectedError := errors.New("repository error")
		s.mockRepository.On("BulkInsert", mock.Anything).Return(expectedError).Once()

		actualResult, actualError := s.service.Create("", []uint{})

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return appeals on success", func() {
		user := "test@email.com"
		resourceIDs := []uint{1, 2}
		expectedResources := []*domain.Resource{}
		for _, id := range resourceIDs {
			expectedResources = append(expectedResources, &domain.Resource{
				ID:           id,
				Type:         "resource_type_1",
				ProviderType: "provider_type",
				ProviderURN:  "provider1",
			})
		}
		expectedProviders := []*domain.Provider{
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
		expectedAppeals := []*domain.Appeal{}
		for _, r := range resourceIDs {
			expectedAppeals = append(expectedAppeals, &domain.Appeal{
				ResourceID:    r,
				PolicyID:      "policy_1",
				PolicyVersion: 1,
				Status:        domain.AppealStatusPending,
				Email:         user,
			})
		}
		expectedResult := []*domain.Appeal{}
		for i, a := range expectedAppeals {
			expectedAppeal := &domain.Appeal{}
			*expectedAppeal = *a
			expectedAppeal.ID = uint(i) + 1
			expectedResult = append(expectedResult, expectedAppeal)
		}
		expectedFilters := map[string]interface{}{"ids": resourceIDs}
		s.mockResourceService.On("Find", expectedFilters).Return(expectedResources, nil).Once()
		s.mockProviderService.On("Find").Return(expectedProviders, nil).Once()
		s.mockRepository.On("BulkInsert", expectedAppeals).Return(nil).Run(func(args mock.Arguments) {
			appeals := args.Get(0).([]*domain.Appeal)
			for i, a := range appeals {
				a.ID = expectedResult[i].ID
			}
		}).Once()

		actualResult, actualError := s.service.Create(user, resourceIDs)

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
