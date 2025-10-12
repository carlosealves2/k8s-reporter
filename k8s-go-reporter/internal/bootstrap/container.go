package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/carlosealves2/k8s-go-reporter/internal/config"
	grpcServer "github.com/carlosealves2/k8s-go-reporter/internal/grpc"
	"github.com/carlosealves2/k8s-go-reporter/internal/logger"
	"github.com/carlosealves2/k8s-go-reporter/internal/observability"
	"github.com/carlosealves2/k8s-go-reporter/internal/services"
)

// Container holds all application dependencies
// This follows the Dependency Inversion Principle (DIP) and makes the application testable
type Container struct {
	Config       config.Config
	K8sClient    kubernetes.Interface
	Service      services.Service
	Server       grpcServer.ServerInterface
	Logger       logger.Logger
	OTelProvider *observability.OTelProvider
}

// NewConfig creates application configuration
// Exported for testing purposes
func NewConfig() config.Config {
	return config.NewBuilder().
		WithEnv().
		Validate().
		Build()
}

// NewK8sClient creates a Kubernetes client with automatic environment detection
// It tries in-cluster config first (production), then falls back to kubeconfig (local development)
// This follows the Open/Closed Principle (OCP) - extensible without modification
// Returns kubernetes.Interface for testability (DIP)
// Exported for testing purposes
func NewK8sClient() (kubernetes.Interface, error) {
	// Try in-cluster configuration first (when running inside Kubernetes)
	k8sConfig, err := rest.InClusterConfig()
	if err == nil {
		clientSet, err := kubernetes.NewForConfig(k8sConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes client from in-cluster config: %w", err)
		}
		return clientSet, nil
	}

	// Fallback to kubeconfig for local development
	kubeconfig := getKubeconfigPath()
	k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes config (tried in-cluster and kubeconfig): %w", err)
	}

	clientSet, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client from kubeconfig: %w", err)
	}

	return clientSet, nil
}

// getKubeconfigPath returns the path to kubeconfig file
// It checks KUBECONFIG env var first, then falls back to default location
func getKubeconfigPath() string {
	// Check KUBECONFIG environment variable
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// Fallback to default location: ~/.kube/config
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".kube", "config")
}

// NewService creates the application service
// Exported for testing purposes
func NewService(k8sClient kubernetes.Interface) services.Service {
	return services.NewK8sService(k8sClient)
}

// NewLogger creates the application logger
// Exported for testing purposes
func NewLogger() logger.Logger {
	return logger.NewStdLogger()
}

// NewOTelProvider creates the OpenTelemetry provider based on configuration
// Returns nil if OpenTelemetry is disabled
// Exported for testing purposes
func NewOTelProvider(ctx context.Context, cfg config.Config) (*observability.OTelProvider, error) {
	return observability.NewOTelProvider(ctx, cfg.OTel())
}

// NewGRPCServer creates the gRPC server with all dependencies
// Accepts optional gRPC server options for OpenTelemetry interceptors
// Exported for testing purposes
func NewGRPCServer(cfg config.Config, service services.Service, opts ...grpc.ServerOption) grpcServer.ServerInterface {
	handler := grpcServer.NewHandler(service)
	return grpcServer.NewServer(cfg.Server().Port(), handler, opts...)
}

// NewContainer creates a new dependency injection container with all dependencies initialized
// This is the main factory function that wires everything together
func NewContainer() (*Container, error) {
	ctx := context.Background()

	// Create logger
	log := NewLogger()

	// Create configuration
	cfg := NewConfig()

	// Initialize OpenTelemetry if enabled
	otelProvider, err := NewOTelProvider(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenTelemetry provider: %w", err)
	}

	// Initialize metrics if OTel is enabled
	if otelProvider != nil {
		if err := observability.InitMetrics(); err != nil {
			return nil, fmt.Errorf("failed to initialize OpenTelemetry metrics: %w", err)
		}
		log.Infof("OpenTelemetry initialized: traces=%v metrics=%v endpoint=%s",
			cfg.OTel().TracesEnabled(), cfg.OTel().MetricsEnabled(), cfg.OTel().Endpoint())
	}

	// Create Kubernetes client
	k8sClient, err := NewK8sClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Create service
	service := NewService(k8sClient)

	// Prepare gRPC server options with OpenTelemetry instrumentation if enabled
	var grpcOpts []grpc.ServerOption
	if otelProvider != nil {
		// Add tracing via stats handler
		if cfg.OTel().TracesEnabled() {
			grpcOpts = append(grpcOpts, observability.NewOTelStatsHandler())
		}
		// Add custom metrics interceptor
		if cfg.OTel().MetricsEnabled() {
			grpcOpts = append(grpcOpts, grpc.UnaryInterceptor(observability.NewMetricsInterceptor()))
		}
	}

	// Create gRPC server with interceptors
	server := NewGRPCServer(cfg, service, grpcOpts...)

	return &Container{
		Config:       cfg,
		K8sClient:    k8sClient,
		Service:      service,
		Server:       server,
		Logger:       log,
		OTelProvider: otelProvider,
	}, nil
}
