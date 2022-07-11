package server

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/odpf/guardian/core"
	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/core/approval"
	"github.com/odpf/guardian/core/policy"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/core/resource"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/guardian/plugins/identities"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/guardian/plugins/providers/bigquery"
	"github.com/odpf/guardian/plugins/providers/gcloudiam"
	"github.com/odpf/guardian/plugins/providers/grafana"
	"github.com/odpf/guardian/plugins/providers/metabase"
	"github.com/odpf/guardian/plugins/providers/noop"
	"github.com/odpf/guardian/plugins/providers/tableau"
	"github.com/odpf/salt/audit"
	audit_repos "github.com/odpf/salt/audit/repositories"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/metadata"
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

	auditRepository := audit_repos.NewPostgresRepository(store.DB())
	auditLogger := audit.New(
		audit.WithRepository(auditRepository),
		audit.WithMetadataExtractor(func(ctx context.Context) map[string]interface{} {
			md := map[string]interface{}{
				"app_name":    "guardian",
				"app_version": core.Version,
			}

			// trace id
			var traceID string
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				if rawTraceID := md.Get(deps.Config.AuditLogTraceIDHeaderKey); len(rawTraceID) > 0 {
					traceID = rawTraceID[0]
				}
			}
			if traceID == "" {
				traceID = uuid.New().String()
			}
			md["trace_id"] = traceID

			return md
		}),
		audit.WithActorExtractor(func(ctx context.Context) (string, error) {
			if actor, ok := ctx.Value(authenticatedUserEmailContextKey{}).(string); ok {
				return actor, nil
			}
			return "", nil
		}),
	)

	providerRepository := postgres.NewProviderRepository(store.DB())
	policyRepository := postgres.NewPolicyRepository(store.DB())
	resourceRepository := postgres.NewResourceRepository(store.DB())
	appealRepository := postgres.NewAppealRepository(store.DB())
	approvalRepository := postgres.NewApprovalRepository(store.DB())

	providerClients := []provider.Client{
		bigquery.NewProvider(domain.ProviderTypeBigQuery, deps.Crypto),
		metabase.NewProvider(domain.ProviderTypeMetabase, deps.Crypto, deps.Logger),
		grafana.NewProvider(domain.ProviderTypeGrafana, deps.Crypto),
		tableau.NewProvider(domain.ProviderTypeTableau, deps.Crypto),
		gcloudiam.NewProvider(domain.ProviderTypeGCloudIAM, deps.Crypto),
		noop.NewProvider(domain.ProviderTypeNoOp, deps.Logger),
	}

	iamManager := identities.NewManager(deps.Crypto, deps.Validator)

	resourceService := resource.NewService(resource.ServiceDeps{
		Repository:  resourceRepository,
		Logger:      deps.Logger,
		AuditLogger: auditLogger,
	})
	providerService := provider.NewService(provider.ServiceDeps{
		Repository:      providerRepository,
		ResourceService: resourceService,
		Clients:         providerClients,
		Validator:       deps.Validator,
		Logger:          deps.Logger,
		AuditLogger:     auditLogger,
	})
	policyService := policy.NewService(policy.ServiceDeps{
		Repository:      policyRepository,
		ResourceService: resourceService,
		ProviderService: providerService,
		IAMManager:      iamManager,
		Validator:       deps.Validator,
		Logger:          deps.Logger,
		AuditLogger:     auditLogger,
	})
	approvalService := approval.NewService(approval.ServiceDeps{
		Repository:    approvalRepository,
		PolicyService: policyService,
	})
	appealService := appeal.NewService(appeal.ServiceDeps{
		Repository:      appealRepository,
		ResourceService: resourceService,
		ApprovalService: approvalService,
		ProviderService: providerService,
		PolicyService:   policyService,
		IAMManager:      iamManager,
		Notifier:        deps.Notifier,
		Validator:       deps.Validator,
		Logger:          deps.Logger,
		AuditLogger:     auditLogger,
	})

	return &Services{
		resourceService,
		providerService,
		policyService,
		approvalService,
		appealService,
	}, nil
}
