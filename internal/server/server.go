package server

import (
	"context"
	"fmt"
	"github.com/odpf/guardian/pkg/auth"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-playground/validator/v10"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/guardian/jobs"
	"github.com/odpf/guardian/pkg/crypto"
	"github.com/odpf/guardian/pkg/scheduler"
	"github.com/odpf/guardian/pkg/tracing"
	"github.com/odpf/guardian/plugins/notifiers"
	audit_repos "github.com/odpf/salt/audit/repositories"
	"github.com/odpf/salt/log"
	"github.com/odpf/salt/mux"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	ConfigFileName = "config"
)

const (
	GRPCMaxClientSendSize = 32 << 20
	defaultGracePeriod    = 5 * time.Second
)

// RunServer runs the application server
func RunServer(config *Config) error {
	logger := log.NewLogrus(log.LogrusWithLevel(config.LogLevel))
	crypto := crypto.NewAES(config.EncryptionSecretKeyKey)
	validator := validator.New()
	notifier, err := notifiers.NewClient(&config.Notifier)
	if err != nil {
		return err
	}

	shutdown, err := tracing.InitTracer(config.Telemetry)
	if err != nil {
		return err
	}
	defer shutdown()

	services, err := InitServices(ServiceDeps{
		Config:    config,
		Logger:    logger,
		Validator: validator,
		Notifier:  notifier,
		Crypto:    crypto,
	})
	if err != nil {
		return fmt.Errorf("initializing services: %w", err)
	}

	jobHandler := jobs.NewHandler(
		logger,
		services.GrantService,
		services.ProviderService,
		notifier,
	)

	// init scheduler
	// TODO: allow timeout configuration for job handler context
	jobsMap := map[JobType]func(context.Context) error{
		FetchResources:            jobHandler.FetchResources,
		ExpiringGrantNotification: jobHandler.GrantExpirationReminder,
		RevokeExpiredGrants:       jobHandler.RevokeExpiredGrants,
	}

	enabledJobs := fetchJobsToRun(config)
	tasks := make([]*scheduler.Task, 0)
	for _, job := range enabledJobs {
		fn := jobsMap[job.JobType]
		task := scheduler.Task{
			CronTab: job.Interval,
			Func:    func() error { return fn(context.Background()) },
		}
		tasks = append(tasks, &task)
	}

	s, err := scheduler.New(tasks)
	if err != nil {
		return err
	}
	s.Run()

	// init grpc server
	logrusEntry := logrus.NewEntry(logrus.New()) // TODO: get logrus instance from `logger` var

	authInterceptor, err := getAuthInterceptor(config)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_logrus.StreamServerInterceptor(logrusEntry),
			otelgrpc.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(
				grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
					logger.Error(string(debug.Stack()))
					return status.Errorf(codes.Internal, "Internal error, please check log")
				}),
			),
			grpc_logrus.UnaryServerInterceptor(logrusEntry),
			authInterceptor,
			otelgrpc.UnaryServerInterceptor(),
		)),
	)
	protoAdapter := handlerv1beta1.NewAdapter()
	guardianv1beta1.RegisterGuardianServiceServer(grpcServer, handlerv1beta1.NewGRPCServer(
		services.ResourceService,
		services.ActivityService,
		services.ProviderService,
		services.PolicyService,
		services.AppealService,
		services.ApprovalService,
		services.GrantService,
		protoAdapter,
		config.Auth.Default.HeaderKey,
	))

	// init http proxy
	timeoutGrpcDialCtx, grpcDialCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer grpcDialCancel()

	headerMatcher := makeHeaderMatcher(config)
	gwmux := runtime.NewServeMux(
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
		runtime.WithIncomingHeaderMatcher(headerMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	address := fmt.Sprintf(":%d", config.Port)
	grpcConn, err := grpc.DialContext(
		timeoutGrpcDialCtx,
		address,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(GRPCMaxClientSendSize),
			grpc.MaxCallSendMsgSize(GRPCMaxClientSendSize),
		),
	)
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

	logger.Info(fmt.Sprintf("server running on %s", address))

	return mux.Serve(runtimeCtx, address,
		mux.WithHTTP(baseMux),
		mux.WithGRPC(grpcServer),
		mux.WithGracePeriod(defaultGracePeriod),
	)
}

// Migrate runs the schema migration scripts
func Migrate(c *Config) error {
	store, err := getStore(c)
	if err != nil {
		return err
	}

	sqldb, _ := store.DB().DB()

	auditRepository := audit_repos.NewPostgresRepository(sqldb)
	if err := auditRepository.Init(context.Background()); err != nil {
		return fmt.Errorf("initializing audit repository: %w", err)
	}

	return store.Migrate()
}

func getStore(c *Config) (*postgres.Store, error) {
	store, err := postgres.NewStore(&c.DB)
	if c.Telemetry.Enabled {
		if err := store.DB().Use(otelgorm.NewPlugin()); err != nil {
			return store, err
		}
	}
	return store, err
}

func makeHeaderMatcher(c *Config) func(key string) (string, bool) {
	return func(key string) (string, bool) {
		switch strings.ToLower(key) {
		case
			strings.ToLower(c.Auth.Default.HeaderKey),
			strings.ToLower(c.AuditLogTraceIDHeaderKey):
			return key, true
		default:
			return runtime.DefaultHeaderMatcher(key)
		}
	}
}

func fetchJobsToRun(config *Config) []*JobConfig {
	jobsToRun := make([]*JobConfig, 0)

	if config.Jobs.FetchResources.Enabled {
		job := config.Jobs.FetchResources
		job.JobType = FetchResources
		jobsToRun = append(jobsToRun, &job)
	}

	if config.Jobs.ExpiringAccessNotification.Enabled || config.Jobs.ExpiringGrantNotification.Enabled {
		job := config.Jobs.ExpiringAccessNotification
		job.JobType = ExpiringGrantNotification
		jobsToRun = append(jobsToRun, &job)
	}

	if config.Jobs.RevokeExpiredAccess.Enabled || config.Jobs.RevokeExpiredGrants.Enabled {
		job := config.Jobs.RevokeExpiredAccess
		job.JobType = RevokeExpiredGrants
		jobsToRun = append(jobsToRun, &job)
	}

	jobScheduleMapping := fetchDefaultJobScheduleMapping()
	for _, jobConfig := range jobsToRun {
		schedule, ok := jobScheduleMapping[jobConfig.JobType]
		if ok && jobConfig.Interval == "" {
			jobConfig.Interval = schedule
		}
	}

	return jobsToRun
}

func fetchDefaultJobScheduleMapping() map[JobType]string {
	return map[JobType]string{
		FetchResources:            "0 */2 * * *",
		RevokeExpiredGrants:       "*/20 * * * *",
		ExpiringGrantNotification: "0 9 * * *",
	}
}

func getAuthInterceptor(config *Config) (grpc.UnaryServerInterceptor, error) {
	// default fallback to user email on header
	authInterceptor := withAuthenticatedUserEmail(config.Auth.Default.HeaderKey)

	if config.Auth.Provider == "oidc" {
		idtokenValidator, err := idtoken.NewValidator(context.Background())
		if err != nil {
			return nil, err
		}

		params := &auth.OidcValidatorParams{
			Audience:          config.Auth.Oidc.Audience,
			ValidEmailDomains: config.Auth.Oidc.EligibleEmailDomains,
			HeaderKey:         config.Auth.Default.HeaderKey,
			ContextKey:        AuthenticatedUserEmailContextKey{},
		}

		bearerTokenValidator := auth.NewOidcValidator(idtokenValidator, params)
		authInterceptor = bearerTokenValidator.WithOidcValidator()
	}

	return authInterceptor, nil
}
