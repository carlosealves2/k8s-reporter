package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/carlosealves2/k8s-go-reporter/internal/config"
)

// OTelProvider manages OpenTelemetry trace and metric providers
// This follows the Single Responsibility Principle - only handles OTel lifecycle
type OTelProvider struct {
	traceProvider  *trace.TracerProvider
	metricProvider *metric.MeterProvider
	shutdownFuncs  []func(context.Context) error
}

// NewOTelProvider initializes OpenTelemetry providers based on configuration
// Returns nil if OpenTelemetry is disabled
// This follows the Dependency Inversion Principle - depends on config interface
func NewOTelProvider(ctx context.Context, cfg config.OTelConfig) (*OTelProvider, error) {
	// If OTel is disabled, return nil (no-op)
	if !cfg.Enabled() {
		return nil, nil
	}

	provider := &OTelProvider{
		shutdownFuncs: make([]func(context.Context) error, 0),
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName()),
			semconv.ServiceVersion(cfg.ServiceVersion()),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Setup trace provider if enabled
	if cfg.TracesEnabled() {
		traceProvider, shutdown, err := setupTraceProvider(ctx, cfg, res)
		if err != nil {
			return nil, fmt.Errorf("failed to setup trace provider: %w", err)
		}
		provider.traceProvider = traceProvider
		provider.shutdownFuncs = append(provider.shutdownFuncs, shutdown)

		// Set global trace provider
		otel.SetTracerProvider(traceProvider)
	}

	// Setup metric provider if enabled
	if cfg.MetricsEnabled() {
		metricProvider, shutdown, err := setupMetricProvider(ctx, cfg, res)
		if err != nil {
			return nil, fmt.Errorf("failed to setup metric provider: %w", err)
		}
		provider.metricProvider = metricProvider
		provider.shutdownFuncs = append(provider.shutdownFuncs, shutdown)

		// Set global meter provider
		otel.SetMeterProvider(metricProvider)
	}

	return provider, nil
}

// Shutdown gracefully shuts down all OpenTelemetry providers
// This ensures all telemetry data is flushed before application exit
func (p *OTelProvider) Shutdown(ctx context.Context) error {
	if p == nil {
		return nil
	}

	var errors []error
	for _, shutdown := range p.shutdownFuncs {
		if err := shutdown(ctx); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to shutdown OTel providers: %v", errors)
	}

	return nil
}

// setupTraceProvider creates and configures the OTLP trace exporter and provider
func setupTraceProvider(ctx context.Context, cfg config.OTelConfig, res *resource.Resource) (*trace.TracerProvider, func(context.Context) error, error) {
	// Configure gRPC connection options
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint()),
	}

	// Add insecure option if specified
	if cfg.Insecure() {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	// Create OTLP trace exporter
	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Create trace provider with batch span processor
	traceProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(5*time.Second),
			trace.WithMaxExportBatchSize(512),
		),
	)

	shutdown := func(ctx context.Context) error {
		return traceProvider.Shutdown(ctx)
	}

	return traceProvider, shutdown, nil
}

// setupMetricProvider creates and configures the OTLP metric exporter and provider
func setupMetricProvider(ctx context.Context, cfg config.OTelConfig, res *resource.Resource) (*metric.MeterProvider, func(context.Context) error, error) {
	// Configure gRPC connection options
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint()),
	}

	// Add insecure option if specified
	if cfg.Insecure() {
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	// Create OTLP metric exporter
	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	// Create metric provider with periodic reader
	metricProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter,
			metric.WithInterval(10*time.Second),
		)),
	)

	shutdown := func(ctx context.Context) error {
		return metricProvider.Shutdown(ctx)
	}

	return metricProvider, shutdown, nil
}

// GetDialOptions returns gRPC dial options for OTLP exporters based on configuration
// This helper is used by both trace and metric exporters
func getDialOptions(cfg config.OTelConfig) []grpc.DialOption {
	opts := []grpc.DialOption{}

	if cfg.Insecure() {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	return opts
}
