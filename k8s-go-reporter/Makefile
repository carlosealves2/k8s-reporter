# Makefile for k8s-go-reporter

# Variables
APP_NAME := k8s-go-reporter
BINARY_NAME := k8s-go-reporter
CMD_PATH := ./cmd/server
PROTO_DIR := proto/k8sreporter/v1
PROTO_FILE := $(PROTO_DIR)/k8s_reporter.proto
GEN_DIR := internal/gen

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt

# Build flags
LDFLAGS := -w -s
BUILD_FLAGS := CGO_ENABLED=0

# Docker
DOCKER := docker
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := latest

# Protoc
PROTOC := protoc
PROTOC_GEN_GO := --go_out=. --go_opt=module=k8s-go-reporter
PROTOC_GEN_GRPC := --go-grpc_out=. --go-grpc_opt=module=k8s-go-reporter

.PHONY: all build build-linux clean proto test fmt vet docker-build docker-push tilt-up tilt-down deps help

# Default target
all: clean proto build

##@ General

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

proto: ## Generate protobuf files
	@echo "Generating protobuf files..."
	@$(PROTOC) $(PROTOC_GEN_GO) $(PROTOC_GEN_GRPC) $(PROTO_FILE)
	@echo "Protobuf files generated successfully"

build: proto ## Build the application (local architecture)
	@echo "Building $(APP_NAME)..."
	@$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BINARY_NAME)"

build-linux: proto ## Build the application for Linux (amd64)
	@echo "Building $(APP_NAME) for Linux..."
	@$(BUILD_FLAGS) GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	@echo "Build complete: $(BINARY_NAME)-linux-amd64"

run: build ## Build and run the application
	@echo "Running $(APP_NAME)..."
	@./$(BINARY_NAME)

fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GOFMT) -w -s .
	@echo "Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	@$(GOCMD) vet ./...
	@echo "Vet complete"

lint: fmt vet ## Run formatters and linters
	@echo "Linting complete"

test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

##@ Dependencies

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@echo "Dependencies downloaded"

deps-tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	@$(GOMOD) tidy
	@echo "Dependencies tidied"

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	@$(GOMOD) verify
	@echo "Dependencies verified"

deps-upgrade: ## Upgrade dependencies
	@echo "Upgrading dependencies..."
	@$(GOGET) -u ./...
	@$(GOMOD) tidy
	@echo "Dependencies upgraded"

##@ Docker

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -f Dockerfile .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-build-dev: ## Build Docker dev image
	@echo "Building Docker dev image..."
	@$(DOCKER) build -t $(DOCKER_IMAGE):dev -f Dockerfile.dev .
	@echo "Docker dev image built: $(DOCKER_IMAGE):dev"

docker-push: ## Push Docker image
	@echo "Pushing Docker image..."
	@$(DOCKER) push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Docker image pushed"

##@ Tilt

tilt-up: ## Start Tilt development environment
	@echo "Starting Tilt..."
	@tilt up

tilt-down: ## Stop Tilt development environment
	@echo "Stopping Tilt..."
	@tilt down --delete-namespaces

tilt-logs: ## View Tilt logs
	@tilt logs

##@ Protobuf Tools

install-proto-tools: ## Install protoc plugins
	@echo "Installing protoc plugins..."
	@$(GOGET) google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@$(GOGET) google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Protoc plugins installed"

##@ Cleanup

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-linux-amd64
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

clean-proto: ## Clean generated protobuf files
	@echo "Cleaning generated protobuf files..."
	@rm -rf $(GEN_DIR)
	@echo "Generated protobuf files cleaned"

clean-all: clean clean-proto ## Clean everything
	@echo "All cleaned"

##@ Kubernetes

k8s-apply: ## Apply Kubernetes manifests
	@echo "Applying Kubernetes manifests..."
	@kubectl apply -f k8s/namespace.yaml
	@kubectl apply -f k8s/rbac.yaml
	@kubectl apply -f k8s/deployment.yaml
	@kubectl apply -f k8s/service.yaml
	@echo "Manifests applied"

k8s-delete: ## Delete Kubernetes resources
	@echo "Deleting Kubernetes resources..."
	@kubectl delete -f k8s/service.yaml --ignore-not-found=true
	@kubectl delete -f k8s/deployment.yaml --ignore-not-found=true
	@kubectl delete -f k8s/rbac.yaml --ignore-not-found=true
	@kubectl delete -f k8s/namespace.yaml --ignore-not-found=true
	@echo "Resources deleted"

k8s-logs: ## View application logs
	@kubectl logs -l app=$(APP_NAME) -n $(APP_NAME) -f

k8s-restart: ## Restart deployment
	@kubectl rollout restart deployment/$(APP_NAME) -n $(APP_NAME)

k8s-status: ## Check deployment status
	@kubectl get pods -l app=$(APP_NAME) -n $(APP_NAME)
	@kubectl get svc -l app=$(APP_NAME) -n $(APP_NAME)

##@ Info

version: ## Show version information
	@echo "App: $(APP_NAME)"
	@echo "Go version: $$($(GOCMD) version)"
	@echo "Protoc version: $$($(PROTOC) --version)"

check-tools: ## Check if required tools are installed
	@echo "Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "Go is not installed"; exit 1; }
	@command -v protoc >/dev/null 2>&1 || { echo "Protoc is not installed"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "Docker is not installed"; exit 1; }
	@command -v kubectl >/dev/null 2>&1 || { echo "Kubectl is not installed"; exit 1; }
	@command -v tilt >/dev/null 2>&1 || { echo "Tilt is not installed"; exit 1; }
	@echo "All required tools are installed ✓"
