package tracing

import (
	"context"
	"crypto/x509"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type Config struct {
	Enabled     bool              `mapstructure:"enabled" default:"false"`
	ServiceName string            `mapstructure:"service_name" default:"guardian"`
	Labels      map[string]string `mapstructure:"labels"`
	Exporter    string            `mapstructure:"exporter" default:"stdout"`
	OTLP        struct {
		Headers  map[string]string `mapstructure:"headers"`
		Endpoint string            `mapstructure:"endpoint" default:"otlp.nr-data.net:443"`
	} `mapstructure:"otlp"`
}

// InitTracer initializes the opentelemetry tracer
// to be used by grpc and http clients and servers.
func InitTracer(cfg Config) (func(), error) {
	if !cfg.Enabled {
		return func() {}, nil
	}

	exporter, err := getExporter(cfg)
	if err != nil {
		return nil, err
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName)),
		),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	shutdown := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			fmt.Printf("failed to shutdown tracer: %v", err)
		}
	}
	return shutdown, err
}

func getExporter(cfg Config) (sdktrace.SpanExporter, error) {
	switch cfg.Exporter {
	case "otlp":
		ctx := context.Background()
		pool, err := x509.SystemCertPool()
		if err != nil {
			panic(err)
		}
		creds := credentials.NewClientTLSFromCert(pool, "")

		return otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithHeaders(cfg.OTLP.Headers),
			otlptracegrpc.WithEndpoint(cfg.OTLP.Endpoint),
			otlptracegrpc.WithTLSCredentials(creds),
		)
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	}
	return nil, nil
}
