package k8s

import "context"

// SecretData represents a key-value pair for K8s Secret injection.
type SecretData struct {
	Key   string
	Value string
}

// SecretInjector defines the interface for managing temporary K8s Secrets.
// Implementation will be provided in Epic 7 (Pipeline Execution).
type SecretInjector interface {
	CreateSecret(ctx context.Context, namespace, name string, data []SecretData) error
	DeleteSecret(ctx context.Context, namespace, name string) error
}
