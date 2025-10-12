package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/carlosealves2/k8s-go-reporter/internal/services"
)

// ToGRPCError converts domain errors to gRPC status errors
// This follows the Single Responsibility Principle - dedicated error mapping logic
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	// Try to unwrap to K8sError
	var k8sErr *services.K8sError
	if errors.As(err, &k8sErr) {
		// Map domain error type to gRPC code
		var code codes.Code
		switch k8sErr.Type {
		case services.ErrorTypeNotFound:
			code = codes.NotFound
		case services.ErrorTypeUnauthorized:
			code = codes.PermissionDenied
		case services.ErrorTypeInvalidInput:
			code = codes.InvalidArgument
		case services.ErrorTypeInternal:
			code = codes.Internal
		default:
			code = codes.Unknown
		}
		return status.Error(code, err.Error())
	}

	// If not a K8sError, return as Unknown
	return status.Error(codes.Unknown, err.Error())
}
