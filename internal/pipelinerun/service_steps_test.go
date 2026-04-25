package pipelinerun

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjy/zcid/internal/stepexec"
	"github.com/xjy/zcid/pkg/tekton"
)

type mockStepRepo struct {
	rows []stepexec.StepExecution
	err  error
}

func (m *mockStepRepo) Upsert(ctx context.Context, row *stepexec.StepExecution) error { return nil }
func (m *mockStepRepo) ListByPipelineRun(ctx context.Context, runID string) ([]stepexec.StepExecution, error) {
	return m.rows, m.err
}
func (m *mockStepRepo) DeleteExpired(ctx context.Context, cutoff time.Time, batchSize int, maxBatches int) (int, bool, error) {
	return 0, false, nil
}
func (m *mockStepRepo) FinalizeRun(ctx context.Context, runID, terminalStatus string) error {
	return nil
}

func TestGetStepExecutions_ProjectScopedAndJSONRaw(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{getByIDPipeline: func(_ context.Context, id, projectID, pipelineID string) (*PipelineRun, error) {
		if id == "run-1" && projectID == "proj-1" && pipelineID == "pipe-1" {
			return &PipelineRun{ID: "run-1", ProjectID: "proj-1", PipelineID: "pipe-1"}, nil
		}
		return nil, ErrNotFound
	}}
	steps := &mockStepRepo{rows: []stepexec.StepExecution{{
		ID: "step-1", PipelineRunID: "run-1", TaskRunName: "task-a", StepName: "build", StepIndex: 0,
		Status: stepexec.StatusRunning, CommandArgs: stepexec.NewJSONRaw(map[string]any{"command": []string{"make"}}),
		EnvPublic: stepexec.NewJSONRaw(map[string]any{"PUBLIC": "true"}), SecretRefs: stepexec.NewJSONRaw([]map[string]any{{"name": "token", "source": "secretKeyRef"}}),
	}}}
	svc := NewService(repo, &mockPipelineGetter{}, nil, tekton.NewTranslator(), &MockK8sClient{}, &MockSecretInjector{}, steps)

	resp, err := svc.GetStepExecutions(ctx, "proj-1", "pipe-1", "run-1")
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "task-a", resp.Items[0].TaskRunName)
	assert.NotContains(t, string(resp.Items[0].SecretRefs), "secret-value")

	_, err = svc.GetStepExecutions(ctx, "proj-1", "wrong-pipe", "run-1")
	assert.Error(t, err)
}
