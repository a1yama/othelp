package othelp

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config holds the configuration for OTel initialization.
type Config struct {
	// ServiceName is the name of the service (required).
	ServiceName string

	// Exporter specifies which exporter to use: "otlp" (default) or "stdout".
	Exporter string

	// Endpoint is the collector endpoint for OTLP exporter (default: "localhost:4317").
	Endpoint string

	// Insecure disables TLS for the OTLP exporter (default: false).
	Insecure bool
}

// ShutdownFunc gracefully shuts down the tracer provider.
type ShutdownFunc func(ctx context.Context) error

// Init initializes OpenTelemetry with sensible defaults.
// Returns a shutdown function that should be deferred.
//
//	shutdown, err := othelp.Init(ctx, othelp.Config{
//	    ServiceName: "myapp",
//	    Exporter:    "otlp",
//	    Endpoint:    "localhost:4317",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer shutdown(ctx)
func Init(ctx context.Context, cfg Config) (ShutdownFunc, error) {
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("othelp: ServiceName is required")
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("othelp: failed to create resource: %w", err)
	}

	exporter, err := createExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("othelp: failed to create exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}, nil
}

func createExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	switch cfg.Exporter {
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "otlp", "":
		endpoint := cfg.Endpoint
		if endpoint == "" {
			endpoint = "localhost:4317"
		}
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(endpoint),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		return otlptracegrpc.New(ctx, opts...)
	default:
		return nil, fmt.Errorf("unsupported exporter: %q (use \"otlp\" or \"stdout\")", cfg.Exporter)
	}
}
