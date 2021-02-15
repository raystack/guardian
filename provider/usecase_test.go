package provider_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/provider"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UsecaseTestSuite struct {
	suite.Suite
	mockProviderRepository *mocks.ProviderRepository
	usecase                *provider.Usecase
}

func (s *UsecaseTestSuite) SetupTest() {
	s.mockProviderRepository = new(mocks.ProviderRepository)
	s.usecase = provider.NewUsecase(s.mockProviderRepository)
}

func (s *UsecaseTestSuite) TestCreate() {
	config := "config string"
	provider := &domain.Provider{
		Config: config,
	}

	s.Run("should return error if got error from the provider repository", func() {
		expectedError := errors.New("error from repository")
		s.mockProviderRepository.On("Create", mock.Anything).Return(expectedError).Once()

		actualError := s.usecase.Create(&domain.Provider{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should pass the model from the param", func() {
		s.mockProviderRepository.On("Create", provider).Return(nil).Once()

		actualError := s.usecase.Create(provider)

		s.Nil(actualError)
		s.mockProviderRepository.AssertExpectations(s.T())
	})
}

func TestUsecase(t *testing.T) {
	suite.Run(t, new(UsecaseTestSuite))
}
