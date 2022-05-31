package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/core/approval"
	"github.com/odpf/guardian/core/policy"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/core/resource"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/guardian/plugins/identities"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/guardian/plugins/providers"
	"github.com/odpf/guardian/plugins/providers/bigquery"
	"github.com/odpf/guardian/plugins/providers/gcloudiam"
	"github.com/odpf/guardian/plugins/providers/grafana"
	"github.com/odpf/guardian/plugins/providers/metabase"
	"github.com/odpf/guardian/plugins/providers/tableau"
	"github.com/odpf/salt/log"
)

type Services struct {
	ResourceService *resource.Service
	ProviderService *provider.Service
	PolicyService   *policy.Service
	ApprovalService *approval.Service
	AppealService   *appeal.Service
}

type ServiceDeps struct {
	Config *Config
	// TODO: make items below as options
	Logger    log.Logger
	Validator *validator.Validate
	Notifier  notifiers.Client
	Crypto    domain.Crypto
}

func InitServices(deps ServiceDeps) (*Services, error) {
	store, err := getStore(deps.Config)
	if err != nil {
		return nil, err
	}

	providerRepository := postgres.NewProviderRepository(store.DB())
	policyRepository := postgres.NewPolicyRepository(store.DB())
	resourceRepository := postgres.NewResourceRepository(store.DB())
	appealRepository := postgres.NewAppealRepository(store.DB())
	approvalRepository := postgres.NewApprovalRepository(store.DB())

	providerClients := []providers.Client{
		bigquery.NewProvider(domain.ProviderTypeBigQuery, deps.Crypto),
		metabase.NewProvider(domain.ProviderTypeMetabase, deps.Crypto, deps.Logger),
		grafana.NewProvider(domain.ProviderTypeGrafana, deps.Crypto),
		tableau.NewProvider(domain.ProviderTypeTableau, deps.Crypto),
		gcloudiam.NewProvider(domain.ProviderTypeGCloudIAM, deps.Crypto),
	}

	iamManager := identities.NewManager(deps.Crypto, deps.Validator)

	resourceService := resource.NewService(resourceRepository)
	providerService := provider.NewService(
		deps.Logger,
		deps.Validator,
		providerRepository,
		resourceService,
		providerClients,
	)
	policyService := policy.NewService(
		deps.Validator,
		policyRepository,
		resourceService,
		providerService,
		iamManager,
	)
	approvalService := approval.NewService(
		approvalRepository,
		policyService,
	)
	appealService := appeal.NewService(
		appealRepository,
		approvalService,
		resourceService,
		providerService,
		policyService,
		iamManager,
		deps.Notifier,
		deps.Logger,
	)

	return &Services{
		resourceService,
		providerService,
		policyService,
		approvalService,
		appealService,
	}, nil
}
