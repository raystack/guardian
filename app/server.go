package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/odpf/guardian/api/handler/v1"
	pb "github.com/odpf/guardian/api/proto/guardian"
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
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
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

	providerJobHandler := provider.NewJobHandler(providerService)
	appealJobHandler := appeal.NewJobHandler(appealService)

	// init scheduler
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

	// init grpc server
	grpcServer := grpc.NewServer()
	protoAdapter := v1.NewAdapter()
	pb.RegisterGuardianServiceServer(grpcServer, v1.NewGuardianServiceServer(
		resourceService,
		providerService,
		policyService,
		appealService,
		protoAdapter,
	))

	// init http proxy
	timeoutGrpcDialCtx, grpcDialCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer grpcDialCancel()

	gwmux := runtime.NewServeMux(
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
	)
	address := fmt.Sprintf(":%d", c.Port)
	grpcConn, err := grpc.DialContext(timeoutGrpcDialCtx, address, grpc.WithInsecure())
	if err != nil {
		return err
	}

	runtimeCtx, runtimeCancel := context.WithCancel(context.Background())
	defer runtimeCancel()

	if err := pb.RegisterGuardianServiceHandler(runtimeCtx, gwmux, grpcConn); err != nil {
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

	log.Println("server running on port:", c.Port)
	if err := server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			return err
		}
	}

	return nil
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
