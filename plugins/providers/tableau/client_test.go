package tableau_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/tableau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewClient(t *testing.T) {
	t.Run("should return error if config is invalid", func(t *testing.T) {
		invalidConfig := &tableau.ClientConfig{}

		actualClient, actualError := tableau.NewClient(invalidConfig)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return error if config.Host is not a valid url", func(t *testing.T) {
		invalidHostConfig := &tableau.ClientConfig{
			Username:   "test-username",
			Password:   "test-password",
			ContentURL: "test-content-url",
			Host:       "invalid-url",
		}

		actualClient, actualError := tableau.NewClient(invalidHostConfig)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return error if got error retrieving the session token", func(t *testing.T) {
		mockHttpClient := new(mocks.HTTPClient)
		config := &tableau.ClientConfig{
			Username:   "test-username",
			Password:   "test-password",
			Host:       "http://localhost",
			ContentURL: "test-content-url",
			HTTPClient: mockHttpClient,
		}

		expectedError := errors.New("request error")
		mockHttpClient.On("Do", mock.Anything).Return(nil, expectedError).Once()
		actualClient, actualError := tableau.NewClient(config)

		mockHttpClient.AssertExpectations(t)
		assert.Nil(t, actualClient)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return client and nil error on success", func(t *testing.T) {
		// TODO: test http request execution
	})
}
