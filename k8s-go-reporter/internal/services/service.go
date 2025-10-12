package services

import "context"

// NamespaceReader defines operations for reading namespace information
// This follows the Interface Segregation Principle (ISP) - clients depend only on what they need
type NamespaceReader interface {
	// ListNamespaces returns all namespace names in the cluster
	ListNamespaces(ctx context.Context) ([]string, error)
}

// PodReader defines operations for reading pod information
// This follows the Interface Segregation Principle (ISP) - clients depend only on what they need
type PodReader interface {
	// ListPods returns pod information for a specific namespace
	ListPods(ctx context.Context, namespace string) ([]Pod, error)
}

// Service defines the complete business logic operations for Kubernetes reporting
// It composes smaller, focused interfaces following ISP
// This design prevents future violations when adding new resource types (Deployments, ConfigMaps, etc.)
type Service interface {
	NamespaceReader
	PodReader
}

// Pod represents basic pod information
type Pod struct {
	Name      string
	Namespace string
	Status    string
	NodeName  string
}
