package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/odpf/guardian/api"
	"github.com/odpf/guardian/appeal"
	"github.com/odpf/guardian/approval"
	"github.com/odpf/guardian/crypto"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/identitymanager"
	"github.com/odpf/guardian/logger"
	"github.com/odpf/guardian/model"
	"github.com/odpf/guardian/notifier"
	"github.com/odpf/guardian/policy"
	"github.com/odpf/guardian/provider"
	"github.com/odpf/guardian/provider/bigquery"
	"github.com/odpf/guardian/resource"
	"github.com/odpf/guardian/scheduler"
	"github.com/odpf/guardian/store"
	"gorm.io/gorm"
)

// RunServer runs the application server
func RunServer(c *Config) error {
	db, err := getDB(c)
	if err != nil {
		return err
	}

	logger, err := logger.New(&logger.Config{
		Level: c.LogLevel,
	})
	if err != nil {
		return err
	}

	crypto := crypto.NewAES(c.EncryptionSecretKeyKey)

	providerRepository := provider.NewRepository(db)
	policyRepository := policy.NewRepository(db)
	resourceRepository := resource.NewRepository(db)
	appealRepository := appeal.NewRepository(db)
	approvalRepository := approval.NewRepository(db)

	identityManagerClient := identitymanager.NewClient(
		&identitymanager.ClientConfig{
			URL:        c.IdentityManagerURL,
			HttpClient: &http.Client{},
		},
	)
	identityManagerService := identitymanager.NewService(identityManagerClient)

	providers := []domain.ProviderInterface{
		bigquery.NewProvider(domain.ProviderTypeBigQuery, crypto),
	}

	notifier := notifier.NewSlackNotifier(c.SlackAccessToken)

	resourceService := resource.NewService(resourceRepository)
	providerService := provider.NewService(
		providerRepository,
		resourceService,
		providers,
	)
	policyService := policy.NewService(policyRepository)
	approvalService := approval.NewService(approvalRepository)
	appealService := appeal.NewService(
		appealRepository,
		approvalService,
		resourceService,
		providerService,
		policyService,
		identityManagerService,
		notifier,
		logger,
	)

	r := api.New(logger)
	provider.SetupHandler(r, providerService)
	policy.SetupHandler(r, policyService)
	resource.SetupHandler(r, resourceService)
	appeal.SetupHandler(r, appealService)

	providerJobHandler := provider.NewJobHandler(providerService)
	appealJobHandler := appeal.NewJobHandler(appealService)

	tasks := []*scheduler.Task{
		{
			CronTab: "0 */2 * * *",
			Func:    providerJobHandler.GetResources,
		},
		{
			CronTab: "0/20 * * * *",
			Func:    appealJobHandler.RevokeExpiredAccess,
		},
	}
	s, err := scheduler.New(tasks)
	if err != nil {
		return err
	}
	s.Run()

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
		&model.Appeal{},
		&model.Approval{},
		&model.Approver{},
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
