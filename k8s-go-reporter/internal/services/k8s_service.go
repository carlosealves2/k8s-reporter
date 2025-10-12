package services

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// K8sService implements the Service interface using Kubernetes client-go
// This follows the Single Responsibility Principle (SOLID)
type K8sService struct {
	clientSet kubernetes.Interface
}

// NewK8sService creates a new Kubernetes service implementation
// Accepts kubernetes.Interface for testability (DIP)
func NewK8sService(clientSet kubernetes.Interface) Service {
	return &K8sService{
		clientSet: clientSet,
	}
}

// ListNamespaces returns all namespace names in the cluster
func (s *K8sService) ListNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := s.clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		// Map Kubernetes API errors to domain errors
		if errors.IsNotFound(err) {
			return nil, NewNotFoundError("namespaces not found", err)
		}
		if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			return nil, NewUnauthorizedError("access denied to list namespaces", err)
		}
		return nil, NewInternalError("failed to list namespaces", err)
	}

	names := make([]string, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		names = append(names, ns.Name)
	}

	return names, nil
}

// ListPods returns pod information for a specific namespace
// Assumes namespace has already been validated by the transport layer (handler)
func (s *K8sService) ListPods(ctx context.Context, namespace string) ([]Pod, error) {
	pods, err := s.clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		// Map Kubernetes API errors to domain errors
		if errors.IsNotFound(err) {
			return nil, NewNotFoundError(fmt.Sprintf("namespace '%s' not found", namespace), err)
		}
		if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			return nil, NewUnauthorizedError(fmt.Sprintf("access denied to list pods in namespace '%s'", namespace), err)
		}
		return nil, NewInternalError(fmt.Sprintf("failed to list pods in namespace '%s'", namespace), err)
	}

	result := make([]Pod, 0, len(pods.Items))
	for _, p := range pods.Items {
		result = append(result, Pod{
			Name:      p.Name,
			Namespace: p.Namespace,
			Status:    string(p.Status.Phase),
			NodeName:  p.Spec.NodeName,
		})
	}

	return result, nil
}
