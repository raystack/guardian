package grafana_test

import (
	"testing"

	"github.com/odpf/guardian/provider/grafana"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("should return error if config is invalid", func(t *testing.T) {
		invalidConfig := &grafana.ClientConfig{}

		actualClient, actualError := grafana.NewClient(invalidConfig)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return error if config.Host is not a valid url", func(t *testing.T) {
		invalidHostConfig := &grafana.ClientConfig{
			Host:   "invalid-url",
			ApiKey: "test-api-key",
		}

		actualClient, actualError := grafana.NewClient(invalidHostConfig)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return client and nil error on success", func(t *testing.T) {
		// TO DO
	})
}
