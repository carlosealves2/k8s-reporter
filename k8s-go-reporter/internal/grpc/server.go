package grpc

import (
	"fmt"
	"net"

	pb "github.com/carlosealves2/k8s-go-reporter/internal/gen/k8sreporter/v1"
	"google.golang.org/grpc"
)

// Server handles the gRPC server lifecycle
// This follows the Single Responsibility Principle (SOLID)
type Server struct {
	grpcServer *grpc.Server
	port       string
	handler    pb.K8SReporterServiceServer
}

// NewServer creates a new gRPC server with dependency injection
// Accepts handler and optional grpc.ServerOption for extensibility (OCP)
func NewServer(port string, handler pb.K8SReporterServiceServer, opts ...grpc.ServerOption) *Server {
	return &Server{
		grpcServer: grpc.NewServer(opts...),
		port:       port,
		handler:    handler,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	// Register the handler
	pb.RegisterK8SReporterServiceServer(s.grpcServer, s.handler)

	return s.grpcServer.Serve(listener)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
