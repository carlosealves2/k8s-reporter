package grpc

import (
	"regexp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Kubernetes namespace naming rules:
// - Must be a valid DNS label (RFC 1123)
// - Lowercase alphanumeric characters or '-'
// - Must start and end with an alphanumeric character
// - Maximum length 63 characters
var namespaceRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

const maxNamespaceLength = 63

// ValidateNamespace validates a Kubernetes namespace according to DNS label rules
// Returns gRPC status error if validation fails
// This follows the Single Responsibility Principle - dedicated validation logic
func ValidateNamespace(namespace string) error {
	// Check if namespace is empty
	if namespace == "" {
		return status.Error(codes.InvalidArgument, "namespace cannot be empty")
	}

	// Check length
	if len(namespace) > maxNamespaceLength {
		return status.Errorf(codes.InvalidArgument,
			"namespace '%s' exceeds maximum length of %d characters",
			namespace, maxNamespaceLength)
	}

	// Check DNS label format (RFC 1123)
	if !namespaceRegex.MatchString(namespace) {
		return status.Errorf(codes.InvalidArgument,
			"namespace '%s' is invalid: must consist of lowercase alphanumeric characters or '-', "+
				"and must start and end with an alphanumeric character",
			namespace)
	}

	return nil
}

// ValidateNamespaceOrDefault validates a namespace and returns default if empty
// This is useful for optional namespace parameters
func ValidateNamespaceOrDefault(namespace, defaultNamespace string) (string, error) {
	if namespace == "" {
		namespace = defaultNamespace
	}

	if err := ValidateNamespace(namespace); err != nil {
		return "", err
	}

	return namespace, nil
}
