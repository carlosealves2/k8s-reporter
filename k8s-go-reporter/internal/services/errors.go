package services

import "fmt"

// ErrorType defines the type of error that occurred
type ErrorType int

const (
	// ErrorTypeNotFound indicates a resource was not found
	ErrorTypeNotFound ErrorType = iota
	// ErrorTypeUnauthorized indicates lack of authorization
	ErrorTypeUnauthorized
	// ErrorTypeInvalidInput indicates invalid input parameters
	ErrorTypeInvalidInput
	// ErrorTypeInternal indicates an internal server error
	ErrorTypeInternal
)

// K8sError represents a domain-specific error with type information
// This allows proper error handling and mapping to transport layer errors (e.g., gRPC codes)
type K8sError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// Error implements the error interface
func (e *K8sError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap implements the errors.Unwrap interface for error chaining
func (e *K8sError) Unwrap() error {
	return e.Cause
}

// NewNotFoundError creates a new NotFound error
func NewNotFoundError(msg string, cause error) *K8sError {
	return &K8sError{
		Type:    ErrorTypeNotFound,
		Message: msg,
		Cause:   cause,
	}
}

// NewUnauthorizedError creates a new Unauthorized error
func NewUnauthorizedError(msg string, cause error) *K8sError {
	return &K8sError{
		Type:    ErrorTypeUnauthorized,
		Message: msg,
		Cause:   cause,
	}
}

// NewInvalidInputError creates a new InvalidInput error
func NewInvalidInputError(msg string, cause error) *K8sError {
	return &K8sError{
		Type:    ErrorTypeInvalidInput,
		Message: msg,
		Cause:   cause,
	}
}

// NewInternalError creates a new Internal error
func NewInternalError(msg string, cause error) *K8sError {
	return &K8sError{
		Type:    ErrorTypeInternal,
		Message: msg,
		Cause:   cause,
	}
}
