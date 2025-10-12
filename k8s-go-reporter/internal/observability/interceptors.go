package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// Metric instruments for request tracking
	requestCounter   metric.Int64Counter
	requestDuration  metric.Float64Histogram
	requestsInFlight metric.Int64UpDownCounter
)

// InitMetrics initializes OpenTelemetry metric instruments
// This should be called once during application startup after OTel provider is set
func InitMetrics() error {
	meter := otel.Meter("k8s-go-reporter")

	var err error

	// Counter for total requests
	requestCounter, err = meter.Int64Counter(
		"grpc.server.requests.total",
		metric.WithDescription("Total number of gRPC requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	// Histogram for request duration
	requestDuration, err = meter.Float64Histogram(
		"grpc.server.request.duration",
		metric.WithDescription("Duration of gRPC requests in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	// UpDownCounter for in-flight requests
	requestsInFlight, err = meter.Int64UpDownCounter(
		"grpc.server.requests.in_flight",
		metric.WithDescription("Number of in-flight gRPC requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	return nil
}

// NewOTelStatsHandler creates a gRPC stats handler for distributed tracing
// This uses the official otelgrpc instrumentation for automatic trace context propagation
func NewOTelStatsHandler() grpc.ServerOption {
	return grpc.StatsHandler(otelgrpc.NewServerHandler())
}

// NewMetricsInterceptor creates a gRPC unary server interceptor for metrics collection
// This tracks request count, duration, and in-flight requests
func NewMetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Increment in-flight requests
		if requestsInFlight != nil {
			requestsInFlight.Add(ctx, 1)
			defer requestsInFlight.Add(ctx, -1)
		}

		// Execute the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code from error
		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			} else {
				statusCode = codes.Unknown
			}
		}

		// Create attributes
		attrs := []attribute.KeyValue{
			attribute.String("grpc.method", info.FullMethod),
			attribute.String("grpc.status", statusCode.String()),
		}

		// Record metrics
		if requestCounter != nil {
			requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
		}

		if requestDuration != nil {
			requestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))
		}

		return resp, err
	}
}
