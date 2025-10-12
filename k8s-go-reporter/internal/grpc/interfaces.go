package grpc

// ServerInterface defines the contract for gRPC server lifecycle management
// This allows for dependency injection and improves testability (DIP)
type ServerInterface interface {
	// Start starts the gRPC server and blocks until an error occurs
	Start() error
	// Stop gracefully stops the gRPC server
	Stop()
}
