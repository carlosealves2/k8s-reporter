# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build tools (protoc and make)
RUN apk add --no-cache protobuf-dev protobuf make

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install protoc Go plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy Makefile and proto files
COPY Makefile ./
COPY proto/ ./proto/

# Generate protobuf files using Makefile
RUN make proto

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the application using Makefile variables
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o k8s-go-reporter ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/k8s-go-reporter .

# Expose port
EXPOSE 50051

# Run the application
CMD ["./k8s-go-reporter"]
