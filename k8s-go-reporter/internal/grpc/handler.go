package grpc

import (
	"context"

	pb "github.com/carlosealves2/k8s-go-reporter/internal/gen/k8sreporter/v1"
	"github.com/carlosealves2/k8s-go-reporter/internal/services"
)

// Handler implements the gRPC service interface
// It delegates to the services.Service for business logic (Dependency Inversion Principle)
type Handler struct {
	pb.UnimplementedK8SReporterServiceServer
	service services.Service
}

// NewHandler creates a new gRPC handler
func NewHandler(service services.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ListNamespaces handles the ListNamespaces RPC
func (h *Handler) ListNamespaces(ctx context.Context, req *pb.ListNamespacesRequest) (*pb.ListNamespacesResponse, error) {
	namespaces, err := h.service.ListNamespaces(ctx)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	return &pb.ListNamespacesResponse{
		Namespaces: namespaces,
	}, nil
}

// ListPods handles the ListPods RPC
func (h *Handler) ListPods(ctx context.Context, req *pb.ListPodsRequest) (*pb.ListPodsResponse, error) {
	// Validate input (transport layer responsibility)
	if err := ValidateNamespace(req.GetNamespace()); err != nil {
		return nil, err
	}

	// Delegate to service
	pods, err := h.service.ListPods(ctx, req.GetNamespace())
	if err != nil {
		return nil, ToGRPCError(err)
	}

	// Convert using mapper (SRP - separation of concerns)
	return &pb.ListPodsResponse{
		Pods: PodsToProto(pods),
	}, nil
}
