# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

K8s Go Reporter is a gRPC service for querying Kubernetes resources (namespaces, pods). It runs in-cluster with RBAC permissions and follows Clean Architecture and SOLID principles.

## Key Commands

### Development
```bash
make all              # Full build: proto generation + binary compilation
make proto            # Generate protobuf files only
make build            # Build binary (local architecture)
make run              # Build and run locally
make test             # Run tests with race detection and coverage
make test-coverage    # Generate HTML coverage report
make lint             # Format and vet code
```

### Protobuf Workflow
```bash
make install-proto-tools  # Install protoc-gen-go and protoc-gen-go-grpc
make proto                # Generate Go code from proto files
make clean-proto          # Remove generated files
```

### Hot Reload Development (Tilt)
```bash
make tilt-up    # Start Tilt with hot reload (~2-3s rebuild on code changes)
make tilt-down  # Stop Tilt and clean up resources
```

**Important**: Tilt manifests reference `../k8s/` (parent directory), not `k8s/` in project root.

### Kubernetes
```bash
make k8s-apply    # Deploy to cluster
make k8s-logs     # Stream logs
make k8s-restart  # Rollout restart
make k8s-delete   # Remove all resources
```

## Architecture

### Dependency Flow (Clean Architecture)
```
cmd/server (main.go)
  ↓
internal/bootstrap (Container - DI)
  ↓
├── internal/config (Builder pattern)
├── internal/services (Business logic - Service interface)
│   └── k8s_service.go (Kubernetes implementation)
├── internal/grpc (Transport layer)
│   ├── server.go (Lifecycle management)
│   └── handler.go (gRPC handlers)
└── internal/observability (OpenTelemetry)
```

### Key Interfaces

**services.Service** (`internal/services/service.go:22-25`)
- Composed of `NamespaceReader` and `PodReader` interfaces (ISP)
- All business logic implementations must satisfy this interface
- The gRPC layer depends on this abstraction (DIP)

**grpcServer.ServerInterface** (`internal/grpc/interfaces.go`)
- Defines `Start()` and `Stop()` methods for server lifecycle
- Allows test mocking of the entire gRPC server

**kubernetes.Interface**
- Used throughout for K8s client dependency injection
- Enables testing without real cluster (DIP)

### Bootstrap Container Pattern

`internal/bootstrap/container.go` is the DI container that wires all dependencies:
1. **NewContainer()**: Main factory (line 116) - creates all dependencies in order
2. **NewConfig()**: Builds config from environment variables
3. **NewK8sClient()**: Auto-detects in-cluster vs kubeconfig
4. **NewOTelProvider()**: Conditional initialization based on config
5. **NewGRPCServer()**: Injects service + optional OpenTelemetry interceptors

## Configuration

All config is environment-based using the Builder pattern (`internal/config/builder.go`):

### Required
- `PORT`: gRPC port (default: "50051")

### Optional (OpenTelemetry)
- `OTEL_ENABLED`: Enable observability (default: false)
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP collector endpoint (required if enabled)
- `OTEL_EXPORTER_OTLP_INSECURE`: Use insecure connection (default: false)
- `OTEL_TRACES_ENABLED`: Enable distributed tracing (default: true)
- `OTEL_METRICS_ENABLED`: Enable custom metrics (default: true)
- `OTEL_SERVICE_NAME`: Service name (default: "k8s-go-reporter")
- `OTEL_SERVICE_VERSION`: Service version (default: "unknown")

**Validation**: Config is validated via `builder.Validate()` which fails fast with `log.Fatalf` on invalid values.

## Adding New gRPC Methods

1. **Define in proto** (`proto/k8sreporter/v1/k8s_reporter.proto`)
   ```protobuf
   service K8sReporterService {
     rpc NewMethod(NewRequest) returns (NewResponse);
   }
   ```

2. **Regenerate code**: `make proto`

3. **Add to Service interface** (`internal/services/service.go`)
   ```go
   type Service interface {
       NamespaceReader
       PodReader
       NewMethod(ctx context.Context, ...) (Result, error)
   }
   ```

4. **Implement in k8s_service.go** (`internal/services/k8s_service.go`)
   ```go
   func (s *K8sService) NewMethod(ctx context.Context, ...) (Result, error) {
       // Business logic using s.clientSet
   }
   ```

5. **Add handler** (`internal/grpc/handler.go`)
   ```go
   func (h *Handler) NewMethod(ctx context.Context, req *pb.NewRequest) (*pb.NewResponse, error) {
       // 1. Validate input (transport concern)
       // 2. Delegate to h.service
       // 3. Map result using ToProto helper
   }
   ```

## OpenTelemetry Integration

When `OTEL_ENABLED=true`:
- **Traces**: Automatic via `otelgrpc.NewServerHandler()` stats handler (line 155 in container.go)
- **Metrics**: Custom interceptor (`observability.NewMetricsInterceptor()`) that records:
  - `grpc.server.requests.total` (Counter)
  - `grpc.server.request.duration` (Histogram)
  - `grpc.server.requests.in_flight` (UpDownCounter)

All metrics include `grpc.method` and `grpc.status` attributes.

## Kubernetes Client Behavior

`NewK8sClient()` in `bootstrap/container.go:46-70`:
1. First attempts `rest.InClusterConfig()` (production)
2. Falls back to `clientcmd.BuildConfigFromFlags()` with kubeconfig (local dev)
3. Kubeconfig path: `$KUBECONFIG` env var → `~/.kube/config` default

## Testing

Run single test:
```bash
go test -v -run TestName ./path/to/package
```

Test with specific Go file:
```bash
go test -v ./internal/services/k8s_service_test.go
```

Coverage for package:
```bash
go test -coverprofile=coverage.out ./internal/services
go tool cover -html=coverage.out
```

## Code Organization Notes

- **Generated code**: `internal/gen/` is gitignored, regenerate via `make proto`
- **No tests yet**: The codebase currently has no test files
- **Error handling**: Domain errors in `internal/services/errors.go` are mapped to gRPC codes in `internal/grpc/errors.go`
- **Validation**: Input validation happens at transport layer (`grpc/validation.go`), not in business logic
- **Mappers**: `internal/grpc/mapper.go` converts between domain models (`services.Pod`) and protobuf messages

## Tiltfile Workflow

The Tilt setup (`Tiltfile` in project root) orchestrates:
1. **proto-gen**: Generates protobuf files
2. **k8s-go-reporter-compile**: Builds Linux binary to `.tiltbuild/`
3. **docker_build_with_restart**: Hot-swaps binary in container without full rebuild
4. **k8s_resource**: Deploys to cluster with port-forward on `:50051`

Live update syncs `.tiltbuild/k8s-go-reporter` → `/app/k8s-go-reporter` in container.

## Testing gRPC Services

Using grpcurl:
```bash
# List all services
grpcurl -plaintext localhost:50051 list

# Call ListNamespaces
grpcurl -plaintext localhost:50051 k8sreporter.v1.K8SReporterService/ListNamespaces

# Call ListPods with JSON payload
grpcurl -plaintext -d '{"namespace":"default"}' localhost:50051 k8sreporter.v1.K8SReporterService/ListPods
```

## RBAC Requirements

The service needs:
- **ServiceAccount**: `k8s-go-reporter`
- **ClusterRole**: `get`, `list`, `watch` on `namespaces` and `pods`
- **ClusterRoleBinding**: Binds SA to ClusterRole

Manifests are in parent directory: `../k8s/` (relative to project root).
