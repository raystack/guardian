package providers_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/providers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockProviderRepository *mocks.ProviderRepository
	service                *providers.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockProviderRepository = new(mocks.ProviderRepository)
	s.service = providers.NewService(s.mockProviderRepository)
}

func (s *ServiceTestSuite) TestCreate() {
	config := "config string"
	provider := &domain.Provider{
		Config: config,
	}

	s.Run("should return error if got error from the provider repository", func() {
		expectedError := errors.New("error from repository")
		s.mockProviderRepository.On("Create", mock.Anything).Return(expectedError).Once()

		actualError := s.service.Create(&domain.Provider{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should pass the model from the param", func() {
		s.mockProviderRepository.On("Create", provider).Return(nil).Once()

		actualError := s.service.Create(provider)

		s.Nil(actualError)
		s.mockProviderRepository.AssertExpectations(s.T())
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
