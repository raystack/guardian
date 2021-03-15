package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/odpf/guardian/api"
	"github.com/odpf/guardian/crypto"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/model"
	"github.com/odpf/guardian/policy"
	"github.com/odpf/guardian/provider"
	"github.com/odpf/guardian/provider/bigquery"
	"github.com/odpf/guardian/store"
	"gorm.io/gorm"
)

// RunServer runs the application server
func RunServer(c *Config) error {
	db, err := getDB(c)
	if err != nil {
		return err
	}

	crypto := crypto.NewAES(c.EncryptionSecretKeyKey)

	providerRepository := provider.NewRepository(db)
	policyRepository := policy.NewRepository(db)

	providers := []domain.ProviderInterface{
		bigquery.NewProvider(domain.ProviderTypeBigQuery, crypto),
	}

	providerService := provider.NewService(providerRepository, providers)
	policyService := policy.NewService(policyRepository)

	r := api.New()
	provider.SetupHandler(r, providerService)
	policy.SetupHandler(r, policyService)

	log.Printf("running server on port %d\n", c.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", c.Port), r)
}

// Migrate runs the schema migration scripts
func Migrate(c *Config) error {
	db, err := getDB(c)
	if err != nil {
		return err
	}

	models := []interface{}{
		&model.Provider{},
		&model.Policy{},
		&model.Resource{},
	}
	return store.Migrate(db, models...)
}

func getDB(c *Config) (*gorm.DB, error) {
	return store.New(&store.Config{
		Host:     c.DBHost,
		User:     c.DBUser,
		Password: c.DBPassword,
		Name:     c.DBName,
		Port:     c.DBPort,
		SslMode:  c.DBSslMode,
	})
}
