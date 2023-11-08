package server

import (
	"context"

	"github.com/goto/guardian/plugins/providers/dataplex"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/goto/guardian/core"
	"github.com/goto/guardian/core/activity"
	"github.com/goto/guardian/core/appeal"
	"github.com/goto/guardian/core/approval"
	"github.com/goto/guardian/core/grant"
	"github.com/goto/guardian/core/policy"
	"github.com/goto/guardian/core/provider"
	"github.com/goto/guardian/core/resource"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/internal/store/postgres"
	"github.com/goto/guardian/pkg/auth"
	"github.com/goto/guardian/pkg/log"
	"github.com/goto/guardian/plugins/identities"
	"github.com/goto/guardian/plugins/notifiers"
	"github.com/goto/guardian/plugins/providers/bigquery"
	"github.com/goto/guardian/plugins/providers/gcloudiam"
	"github.com/goto/guardian/plugins/providers/gcs"
	"github.com/goto/guardian/plugins/providers/grafana"
	"github.com/goto/guardian/plugins/providers/metabase"
	"github.com/goto/guardian/plugins/providers/noop"
	"github.com/goto/guardian/plugins/providers/shield"
	"github.com/goto/guardian/plugins/providers/tableau"
	"github.com/goto/salt/audit"
	audit_repos "github.com/goto/salt/audit/repositories"
	"google.golang.org/grpc/metadata"
)

type Services struct {
	ResourceService *resource.Service
	ActivityService *activity.Service
	ProviderService *provider.Service
	PolicyService   *policy.Service
	ApprovalService *approval.Service
	AppealService   *appeal.Service
	GrantService    *grant.Service
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

	sqldb, err := store.DB().DB()
	if err != nil {
		return nil, err
	}

	auditRepository := audit_repos.NewPostgresRepository(sqldb)
	auditRepository.Init(context.TODO())

	actorExtractor := getActorExtractor(deps.Config)

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
			md[domain.TraceIDKey] = traceID

			return md
		}),
		actorExtractor,
	)

	activityRepository := postgres.NewActivityRepository(store.DB())
	providerRepository := postgres.NewProviderRepository(store.DB())
	policyRepository := postgres.NewPolicyRepository(store.DB())
	resourceRepository := postgres.NewResourceRepository(store.DB())
	appealRepository := postgres.NewAppealRepository(store.DB())
	approvalRepository := postgres.NewApprovalRepository(store.DB())
	grantRepository := postgres.NewGrantRepository(store.DB())

	providerClients := []provider.Client{
		bigquery.NewProvider(domain.ProviderTypeBigQuery, deps.Crypto, deps.Logger),
		metabase.NewProvider(domain.ProviderTypeMetabase, deps.Crypto, deps.Logger),
		grafana.NewProvider(domain.ProviderTypeGrafana, deps.Crypto),
		tableau.NewProvider(domain.ProviderTypeTableau, deps.Crypto),
		gcloudiam.NewProvider(domain.ProviderTypeGCloudIAM, deps.Crypto, deps.Logger),
		noop.NewProvider(domain.ProviderTypeNoOp, deps.Logger),
		gcs.NewProvider(domain.ProviderTypeGCS, deps.Crypto),
		dataplex.NewProvider(domain.ProviderTypePolicyTag, deps.Crypto),
		shield.NewProvider(domain.ProviderTypeShield, deps.Logger),
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
	activityService := activity.NewService(activity.ServiceDeps{
		Repository:      activityRepository,
		ProviderService: providerService,
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
	grantService := grant.NewService(grant.ServiceDeps{
		Repository:      grantRepository,
		ProviderService: providerService,
		ResourceService: resourceService,
		Notifier:        deps.Notifier,
		Logger:          deps.Logger,
		Validator:       deps.Validator,
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
		GrantService:    grantService,
		IAMManager:      iamManager,
		Notifier:        deps.Notifier,
		Validator:       deps.Validator,
		Logger:          deps.Logger,
		AuditLogger:     auditLogger,
	})

	return &Services{
		resourceService,
		activityService,
		providerService,
		policyService,
		approvalService,
		appealService,
		grantService,
	}, nil
}

func getActorExtractor(config *Config) audit.AuditOption {
	var contextKey interface{}

	contextKey = authenticatedUserEmailContextKey{}
	if config.Auth.Provider == "oidc" {
		contextKey = auth.OIDCEmailContextKey{}
	}

	return audit.WithActorExtractor(func(ctx context.Context) (string, error) {
		if actor, ok := ctx.Value(contextKey).(string); ok {
			return actor, nil
		}
		return "", nil
	})
}
