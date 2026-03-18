package pipelinerun

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/xjy/zcid/pkg/tekton"
)

type SecretInjector interface {
	InjectSecrets(ctx context.Context, namespace, runID string, secrets map[string]string) (secretName string, err error)
}

type MockSecretInjector struct{}

func (m *MockSecretInjector) InjectSecrets(ctx context.Context, namespace, runID string, secrets map[string]string) (string, error) {
	slog.Info("MOCK: Would create K8s secret for pipeline run", "runID", runID, "keys", len(secrets))
	return "zcid-run-" + runID, nil
}

type K8sClient interface {
	SubmitPipelineRun(ctx context.Context, namespace string, pr *tekton.PipelineRun) error
	DeletePipelineRun(ctx context.Context, namespace, name string) error
	GetPipelineRunStatus(ctx context.Context, namespace, name string) (string, error)
}

type mockRun struct {
	status    string
	startedAt time.Time
}

type MockK8sClient struct {
	mu   sync.Mutex
	runs map[string]*mockRun
}

func (m *MockK8sClient) init() {
	if m.runs == nil {
		m.runs = make(map[string]*mockRun)
	}
}

func (m *MockK8sClient) SubmitPipelineRun(ctx context.Context, namespace string, pr *tekton.PipelineRun) error {
	m.mu.Lock()
	m.init()
	name := pr.Metadata.Name
	m.runs[name] = &mockRun{status: "Pending", startedAt: time.Now()}
	m.mu.Unlock()

	slog.Info("MOCK: Submitted PipelineRun", "name", name, "namespace", namespace)

	go m.simulateLifecycle(name)
	return nil
}

func (m *MockK8sClient) simulateLifecycle(name string) {
	time.Sleep(2 * time.Second)
	m.mu.Lock()
	if r, ok := m.runs[name]; ok {
		r.status = "Running"
	}
	m.mu.Unlock()
	slog.Info("MOCK: PipelineRun status → Running", "name", name)

	duration := 5 + rand.Intn(10)
	time.Sleep(time.Duration(duration) * time.Second)

	finalStatus := "Succeeded"
	if rand.Intn(10) < 2 {
		finalStatus = "Failed"
	}

	m.mu.Lock()
	if r, ok := m.runs[name]; ok {
		r.status = finalStatus
	}
	m.mu.Unlock()
	slog.Info("MOCK: PipelineRun status → "+finalStatus, "name", name)
}

func (m *MockK8sClient) DeletePipelineRun(ctx context.Context, namespace, name string) error {
	m.mu.Lock()
	m.init()
	delete(m.runs, name)
	m.mu.Unlock()
	slog.Info("MOCK: Deleted PipelineRun", "name", name, "namespace", namespace)
	return nil
}

func (m *MockK8sClient) GetPipelineRunStatus(ctx context.Context, namespace, name string) (string, error) {
	m.mu.Lock()
	m.init()
	r, ok := m.runs[name]
	m.mu.Unlock()
	if !ok {
		return "Unknown", nil
	}
	return r.status, nil
}
