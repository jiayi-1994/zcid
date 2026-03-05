package pipelinerun

import (
	"context"
	"log/slog"

	"github.com/xjy/zcid/pkg/tekton"
)

// SecretInjector creates a temporary K8s secret for pipeline run secrets (FR13).
type SecretInjector interface {
	InjectSecrets(ctx context.Context, namespace, runID string, secrets map[string]string) (secretName string, err error)
}

// MockSecretInjector creates no real secret; logs and returns a placeholder name.
// TODO: Replace with real K8s secret creation when cluster is available.
type MockSecretInjector struct{}

func (m *MockSecretInjector) InjectSecrets(ctx context.Context, namespace, runID string, secrets map[string]string) (string, error) {
	_ = ctx
	_ = namespace
	_ = runID
	_ = secrets
	slog.Info("MOCK: Would create K8s secret for pipeline run", "runID", runID, "keys", len(secrets))
	return "zcid-run-" + runID, nil
}

// K8sClient abstracts Kubernetes/Tekton operations for pipeline execution.
// TODO: Replace with real K8s client when cluster is available.
type K8sClient interface {
	SubmitPipelineRun(ctx context.Context, namespace string, pr *tekton.PipelineRun) error
	DeletePipelineRun(ctx context.Context, namespace, name string) error
	GetPipelineRunStatus(ctx context.Context, namespace, name string) (string, error)
}

// MockK8sClient implements K8sClient with no-op stubs for local development.
type MockK8sClient struct{}

// SubmitPipelineRun logs and returns success without submitting to K8s.
func (m *MockK8sClient) SubmitPipelineRun(ctx context.Context, namespace string, pr *tekton.PipelineRun) error {
	// TODO: Replace with real K8s client when cluster is available
	name := ""
	if pr.ObjectMeta.Name != "" {
		name = pr.ObjectMeta.Name
	}
	slog.Info("MOCK: Would submit PipelineRun", "name", name, "namespace", namespace)
	return nil
}

// DeletePipelineRun logs and returns success without deleting from K8s.
func (m *MockK8sClient) DeletePipelineRun(ctx context.Context, namespace, name string) error {
	// TODO: Replace with real K8s client when cluster is available
	slog.Info("MOCK: Would delete PipelineRun", "name", name, "namespace", namespace)
	return nil
}

// GetPipelineRunStatus returns a mock status without querying K8s.
func (m *MockK8sClient) GetPipelineRunStatus(ctx context.Context, namespace, name string) (string, error) {
	// TODO: Replace with real K8s client when cluster is available
	slog.Info("MOCK: Would get PipelineRun status", "name", name, "namespace", namespace)
	return "Running", nil
}
