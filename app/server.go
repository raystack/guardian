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
	"github.com/odpf/guardian/iam"
	"github.com/odpf/guardian/logger"
	"github.com/odpf/guardian/model"
	"github.com/odpf/guardian/notifier"
	"github.com/odpf/guardian/policy"
	"github.com/odpf/guardian/provider"
	"github.com/odpf/guardian/provider/bigquery"
	"github.com/odpf/guardian/provider/grafana"
	"github.com/odpf/guardian/provider/metabase"
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
		Level: c.Log.Level,
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

	iamClient, err := iam.NewClient(&c.IAM)
	if err != nil {
		return err
	}
	iamService := iam.NewService(iamClient)

	providers := []domain.ProviderInterface{
		bigquery.NewProvider(domain.ProviderTypeBigQuery, crypto),
		metabase.NewProvider(domain.ProviderTypeMetabase, crypto),
		grafana.NewProvider(domain.ProviderTypeGrafana, crypto),
	}

	notifier := notifier.NewSlackNotifier(c.SlackAccessToken)

	resourceService := resource.NewService(resourceRepository)
	policyService := policy.NewService(policyRepository)
	providerService := provider.NewService(
		providerRepository,
		resourceService,
		providers,
	)
	approvalService := approval.NewService(approvalRepository, policyService)
	appealService := appeal.NewService(
		appealRepository,
		approvalService,
		resourceService,
		providerService,
		policyService,
		iamService,
		notifier,
		logger,
	)

	providerHttpHandler := api.NewProviderHandler(providerService)
	policyHttpHandler := api.NewPolicyHandler(policyService)
	resourceHttpHandler := api.NewResourceHandler(resourceService)
	appealHttpHandler := api.NewAppealHandler(appealService)

	r := api.New(logger)

	// provider routes
	r.Methods(http.MethodGet).Path("/providers").HandlerFunc(providerHttpHandler.Find)
	r.Methods(http.MethodPost).Path("/providers").HandlerFunc(providerHttpHandler.Create)
	r.Methods(http.MethodPut).Path("/providers/{id}").HandlerFunc(providerHttpHandler.Update)

	// policy routes
	r.Methods(http.MethodGet).Path("/policies").HandlerFunc(policyHttpHandler.Find)
	r.Methods(http.MethodPost).Path("/policies").HandlerFunc(policyHttpHandler.Create)
	r.Methods(http.MethodPut).Path("/policies/{id}").HandlerFunc(policyHttpHandler.Update)

	// resource routes
	r.Methods(http.MethodGet).Path("/resources").HandlerFunc(resourceHttpHandler.Find)
	r.Methods(http.MethodPut).Path("/resources/{id}").HandlerFunc(resourceHttpHandler.Update)

	// appeal routes
	r.Methods(http.MethodPost).Path("/appeals").HandlerFunc(appealHttpHandler.Create)
	r.Methods(http.MethodGet).Path("/appeals").HandlerFunc(appealHttpHandler.Find)
	r.Methods(http.MethodGet).Path("/appeals/approvals").HandlerFunc(appealHttpHandler.GetPendingApprovals)
	r.Methods(http.MethodPost).Path("/appeals/{id}/approvals/{name}").HandlerFunc(appealHttpHandler.MakeAction)
	r.Methods(http.MethodPut).Path("/appeals/{id}/cancel").HandlerFunc(appealHttpHandler.Cancel)
	r.Methods(http.MethodPut).Path("/appeals/{id}/revoke").HandlerFunc(appealHttpHandler.Revoke)
	r.Methods(http.MethodGet).Path("/appeals/{id}").HandlerFunc(appealHttpHandler.GetByID)

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
	return store.New(&c.DB)
}
