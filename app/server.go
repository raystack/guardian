package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/core/approval"
	"github.com/odpf/guardian/core/policy"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/core/resource"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/crypto"
	"github.com/odpf/guardian/internal/scheduler"
	"github.com/odpf/guardian/plugins/identities"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/guardian/plugins/providers"
	"github.com/odpf/guardian/plugins/providers/bigquery"
	"github.com/odpf/guardian/plugins/providers/gcloudiam"
	"github.com/odpf/guardian/plugins/providers/grafana"
	"github.com/odpf/guardian/plugins/providers/metabase"
	"github.com/odpf/guardian/plugins/providers/tableau"
	"github.com/odpf/guardian/store/postgres"
	"github.com/odpf/salt/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	ConfigFileName = "config"
)

type Jobs struct {
	FetchResourcesInterval             string `mapstructure:"fetch_resources_interval" default:"0 */2 * * *"`
	RevokeExpiredAccessInterval        string `mapstructure:"revoke_expired_access_interval" default:"*/20 * * * *"`
	ExpiringAccessNotificationInterval string `mapstructure:"expiring_access_notification_interval" default:"0 9 * * *"`
}

// RunServer runs the application server
func RunServer(c *Config) error {
	store, err := getStore(c)
	if err != nil {
		return err
	}

	logger := log.NewLogrus(log.LogrusWithLevel(c.LogLevel))
	crypto := crypto.NewAES(c.EncryptionSecretKeyKey)
	v := validator.New()

	db := store.DB()
	providerRepository := postgres.NewProviderRepository(db)
	policyRepository := postgres.NewPolicyRepository(db)
	resourceRepository := postgres.NewResourceRepository(db)
	appealRepository := postgres.NewAppealRepository(db)
	approvalRepository := postgres.NewApprovalRepository(db)

	providerClients := []providers.Client{
		bigquery.NewProvider(domain.ProviderTypeBigQuery, crypto),
		metabase.NewProvider(domain.ProviderTypeMetabase, crypto),
		grafana.NewProvider(domain.ProviderTypeGrafana, crypto),
		tableau.NewProvider(domain.ProviderTypeTableau, crypto),
		gcloudiam.NewProvider(domain.ProviderTypeGCloudIAM, crypto),
	}

	notifier, err := notifiers.NewClient(&c.Notifier)
	if err != nil {
		return err
	}

	iamManager := identities.NewManager(crypto, v)

	resourceService := resource.NewService(resourceRepository)
	providerService := provider.NewService(
		logger,
		v,
		providerRepository,
		resourceService,
		providerClients,
	)
	policyService := policy.NewService(
		v,
		policyRepository,
		resourceService,
		providerService,
		iamManager,
	)
	approvalService := approval.NewService(approvalRepository, policyService)
	appealService := appeal.NewService(
		appealRepository,
		approvalService,
		resourceService,
		providerService,
		policyService,
		iamManager,
		notifier,
		logger,
	)

	providerJobHandler := provider.NewJobHandler(providerService)
	appealJobHandler := appeal.NewJobHandler(logger, appealService, notifier)

	// init scheduler
	tasks := []*scheduler.Task{
		{
			CronTab: c.Jobs.FetchResourcesInterval,
			Func:    providerJobHandler.GetResources,
		},
		{
			CronTab: c.Jobs.RevokeExpiredAccessInterval,
			Func:    appealJobHandler.RevokeExpiredAccess,
		},
		{
			CronTab: c.Jobs.ExpiringAccessNotificationInterval,
			Func:    appealJobHandler.NotifyAboutToExpireAccess,
		},
	}
	s, err := scheduler.New(tasks)
	if err != nil {
		return err
	}
	s.Run()

	// init grpc server
	grpcServer := grpc.NewServer()
	protoAdapter := handlerv1beta1.NewAdapter()
	guardianv1beta1.RegisterGuardianServiceServer(grpcServer, handlerv1beta1.NewGRPCServer(
		resourceService,
		providerService,
		policyService,
		appealService,
		approvalService,
		protoAdapter,
		c.AuthenticatedUserHeaderKey,
	))

	// init http proxy
	timeoutGrpcDialCtx, grpcDialCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer grpcDialCancel()

	headerMatcher := makeHeaderMatcher(c)
	gwmux := runtime.NewServeMux(
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
		runtime.WithIncomingHeaderMatcher(headerMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
		}),
	)
	address := fmt.Sprintf(":%d", c.Port)
	grpcConn, err := grpc.DialContext(timeoutGrpcDialCtx, address, grpc.WithInsecure())
	if err != nil {
		return err
	}

	runtimeCtx, runtimeCancel := context.WithCancel(context.Background())
	defer runtimeCancel()

	if err := guardianv1beta1.RegisterGuardianServiceHandler(runtimeCtx, gwmux, grpcConn); err != nil {
		return err
	}

	baseMux := http.NewServeMux()
	baseMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})
	baseMux.Handle("/api/", http.StripPrefix("/api", gwmux))

	server := &http.Server{
		Handler:      grpcHandlerFunc(grpcServer, baseMux),
		Addr:         address,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Info(fmt.Sprintf("server running on port: %d", c.Port))
	if err := server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			return err
		}
	}

	return nil
}

// Migrate runs the schema migration scripts
func Migrate(c *Config) error {
	store, err := getStore(c)
	if err != nil {
		return err
	}
	return store.Migrate()
}

func getStore(c *Config) (*postgres.Store, error) {
	return postgres.NewStore(&c.DB)
}

// grpcHandlerFunc routes http1 calls to baseMux and http2 with grpc header to grpcServer.
// Using a single port for proxying both http1 & 2 protocols will degrade http performance
// but for our usecase the convenience per performance tradeoff is better suited
// if in future, this does become a bottleneck(which I highly doubt), we can break the service
// into two ports, default port for grpc and default+1 for grpc-gateway proxy.
// We can also use something like a connection multiplexer
// https://github.com/soheilhy/cmux to achieve the same.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

func makeHeaderMatcher(c *Config) func(key string) (string, bool) {
	return func(key string) (string, bool) {
		switch strings.ToLower(key) {
		case strings.ToLower(c.AuthenticatedUserHeaderKey):
			return key, true
		default:
			return runtime.DefaultHeaderMatcher(key)
		}
	}
}
