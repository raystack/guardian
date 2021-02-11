package usecases_test

import (
	"errors"

	"github.com/odpf/guardian/models"
	"github.com/stretchr/testify/mock"
)

func (s *UseCaseTestSuite) TestProvider() {
	s.Run("create", func() {
		config := "config string"
		providerModel := &models.Provider{
			Config: config,
		}

		s.Run("should return error if got error from the provider repository", func() {
			expectedError := errors.New("error from repository")
			s.mockProviderRepository.On("Create", mock.Anything).Return(expectedError).Once()

			actualError := s.useCases.Provider.Create(&models.Provider{})

			s.EqualError(actualError, expectedError.Error())
		})

		s.Run("should pass the model from the param", func() {
			s.mockProviderRepository.On("Create", providerModel).Return(nil).Once()

			actualError := s.useCases.Provider.Create(providerModel)

			s.Nil(actualError)
			s.mockProviderRepository.AssertExpectations(s.T())
		})
	})
}
