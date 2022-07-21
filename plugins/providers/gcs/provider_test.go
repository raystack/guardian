package gcs_test

import (
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/gcs"
	"github.com/stretchr/testify/assert"
)

func TestGetType(t *testing.T) {
	t.Run("should return the typeName of the provider", func(t *testing.T) {
		expectedTypeName := "test-typeName"
		crypto := new(mocks.Crypto)
		p := gcs.NewProvider(expectedTypeName, crypto)

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestCreateConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		p := initProvider()
		providerURN := "test-URN"
		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
		}

		actualError := p.CreateConfig(pc)
		assert.Nil(t, actualError)
	})
}

func TestGrantAccess(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		p := initProvider()
		providerURN := "test-URN"
		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: "valid-Credentials",
			Resources:   []*domain.ResourceConfig{{}},
		}
		a := &domain.Appeal{}

		actualError := p.GrantAccess(pc, a)
		assert.Nil(t, actualError)
	})
}

func TestRevokeAccess(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		p := initProvider()
		providerURN := "test-URN"
		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: "valid-Credentials",
			Resources:   []*domain.ResourceConfig{{}},
		}
		a := &domain.Appeal{}

		actualError := p.RevokeAccess(pc, a)
		assert.Nil(t, actualError)
	})
}

func TestGetRoles(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		p := initProvider()
		providerURN := "test-URN"
		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: "valid-Credentials",
			Resources:   []*domain.ResourceConfig{{}},
		}
		expectedRoles := []*domain.Role(nil)

		actualRoles, _ := p.GetRoles(pc, gcs.ResourceTypeBucket)

		//	assert.Nil(t, actualError)
		assert.Equal(t, expectedRoles, actualRoles)
	})
}

func TestGetAccountType(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		p := initProvider()
		expectedAccountTypes := []string{"user", "serviceAccount", "group", "domain"}

		actualAccountypes := p.GetAccountTypes()

		assert.Equal(t, expectedAccountTypes, actualAccountypes)
	})
}

func initProvider() *gcs.Provider {
	crypto := new(mocks.Crypto)
	return gcs.NewProvider("gcs", crypto)
}
