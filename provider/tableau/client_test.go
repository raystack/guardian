package tableau_test

import (
	"testing"

	"github.com/odpf/guardian/provider/tableau"
	"github.com/stretchr/testify/assert"
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
		//TODO
	})

	t.Run("should return client and nil error on success", func(t *testing.T) {
		// TODO: test http request execution
	})
}
