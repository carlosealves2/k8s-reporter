package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcServer "github.com/carlosealves2/k8s-go-reporter/internal/grpc"
	"github.com/carlosealves2/k8s-go-reporter/internal/logger"
	"github.com/carlosealves2/k8s-go-reporter/internal/observability"
)

// Application manages the application lifecycle
// This follows the Single Responsibility Principle (SRP) - only handles start/stop
type Application struct {
	server       grpcServer.ServerInterface
	logger       logger.Logger
	otelProvider *observability.OTelProvider
}

// Run starts the application with the provided dependency container
// This follows the Dependency Inversion Principle (DIP) - accepts dependencies instead of creating them
func Run(container *Container, version, commit string) error {
	container.Logger.Infof("Initializing application with port: %s, version %s, commit %s",
		container.Config.Server().Port(), version, commit)

	// Setup OpenTelemetry shutdown if enabled
	if container.OTelProvider != nil {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			container.Logger.Info("Shutting down OpenTelemetry provider...")
			if err := container.OTelProvider.Shutdown(shutdownCtx); err != nil {
				container.Logger.Errorf("Failed to shutdown OpenTelemetry provider: %v", err)
			} else {
				container.Logger.Info("OpenTelemetry provider stopped successfully")
			}
		}()
	}

	// Create application instance with injected dependencies
	application := &Application{
		server:       container.Server,
		logger:       container.Logger,
		otelProvider: container.OTelProvider,
	}

	// Start the application
	return application.start()
}

// start starts the gRPC server with graceful shutdown
func (a *Application) start() error {
	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		a.logger.Info("Starting gRPC server...")
		if err := a.server.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		a.logger.Info("Shutdown signal received, stopping server...")
		a.server.Stop()
		a.logger.Info("Server stopped gracefully")
		return nil
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}
