package pipelinerun

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjy/zcid/internal/pipeline"
	"github.com/xjy/zcid/pkg/response"
	"github.com/xjy/zcid/pkg/tekton"
)

type mockRepo struct {
	create           func(ctx context.Context, r *PipelineRun) error
	getByID          func(ctx context.Context, id, projectID string) (*PipelineRun, error)
	getByIDPipeline  func(ctx context.Context, id, projectID, pipelineID string) (*PipelineRun, error)
	getNextRunNumber func(ctx context.Context, pipelineID string) (int, error)
	listByPipeline   func(ctx context.Context, pipelineID, projectID string, page, pageSize int) ([]*PipelineRun, int64, error)
	listRunning      func(ctx context.Context, pipelineID string) ([]*PipelineRun, error)
	update           func(ctx context.Context, id, projectID string, updates map[string]interface{}) error
	updateStatus     func(ctx context.Context, id, projectID string, status RunStatus, errorMsg *string) error
	countRunning     func(ctx context.Context, pipelineID string) (int64, error)
	updateArtifacts  func(ctx context.Context, id, projectID string, artifacts ArtifactSlice) error
}

func (m *mockRepo) Create(ctx context.Context, r *PipelineRun) error {
	if m.create != nil {
		return m.create(ctx, r)
	}
	return nil
}

func (m *mockRepo) GetByIDAndProject(ctx context.Context, id, projectID string) (*PipelineRun, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id, projectID)
	}
	return nil, ErrNotFound
}

func (m *mockRepo) GetByIDProjectPipeline(ctx context.Context, id, projectID, pipelineID string) (*PipelineRun, error) {
	if m.getByIDPipeline != nil {
		return m.getByIDPipeline(ctx, id, projectID, pipelineID)
	}
	return nil, ErrNotFound
}

func (m *mockRepo) GetNextRunNumber(ctx context.Context, pipelineID string) (int, error) {
	if m.getNextRunNumber != nil {
		return m.getNextRunNumber(ctx, pipelineID)
	}
	return 1, nil
}

func (m *mockRepo) ListByPipeline(ctx context.Context, pipelineID, projectID string, page, pageSize int) ([]*PipelineRun, int64, error) {
	if m.listByPipeline != nil {
		return m.listByPipeline(ctx, pipelineID, projectID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockRepo) Update(ctx context.Context, id, projectID string, updates map[string]interface{}) error {
	if m.update != nil {
		return m.update(ctx, id, projectID, updates)
	}
	return nil
}

func (m *mockRepo) UpdateStatus(ctx context.Context, id, projectID string, status RunStatus, errorMsg *string) error {
	if m.updateStatus != nil {
		return m.updateStatus(ctx, id, projectID, status, errorMsg)
	}
	return nil
}

func (m *mockRepo) ListRunning(ctx context.Context, pipelineID string) ([]*PipelineRun, error) {
	if m.listRunning != nil {
		return m.listRunning(ctx, pipelineID)
	}
	return nil, nil
}

func (m *mockRepo) CountRunning(ctx context.Context, pipelineID string) (int64, error) {
	if m.countRunning != nil {
		return m.countRunning(ctx, pipelineID)
	}
	return 0, nil
}

func (m *mockRepo) UpdateArtifacts(ctx context.Context, id, projectID string, artifacts ArtifactSlice) error {
	if m.updateArtifacts != nil {
		return m.updateArtifacts(ctx, id, projectID, artifacts)
	}
	return nil
}

type mockPipelineGetter struct {
	get func(ctx context.Context, id, projectID string) (*pipeline.Pipeline, error)
}

func (m *mockPipelineGetter) GetByIDAndProject(ctx context.Context, id, projectID string) (*pipeline.Pipeline, error) {
	if m.get != nil {
		return m.get(ctx, id, projectID)
	}
	return nil, pipeline.ErrNotFound
}

type mockK8sClient struct {
	submit func(ctx context.Context, namespace string, pr *tekton.PipelineRun) error
	delete func(ctx context.Context, namespace, name string) error
	status func(ctx context.Context, namespace, name string) (string, error)
}

func (m *mockK8sClient) SubmitPipelineRun(ctx context.Context, namespace string, pr *tekton.PipelineRun) error {
	if m.submit != nil {
		return m.submit(ctx, namespace, pr)
	}
	return nil
}

func (m *mockK8sClient) DeletePipelineRun(ctx context.Context, namespace, name string) error {
	if m.delete != nil {
		return m.delete(ctx, namespace, name)
	}
	return nil
}

func (m *mockK8sClient) GetPipelineRunStatus(ctx context.Context, namespace, name string) (string, error) {
	if m.status != nil {
		return m.status(ctx, namespace, name)
	}
	return "Running", nil
}

func TestTriggerRun_Success(t *testing.T) {
	ctx := context.Background()

	var createdRun *PipelineRun
	repo := &mockRepo{
		getNextRunNumber: func(_ context.Context, _ string) (int, error) { return 1, nil },
		countRunning:     func(_ context.Context, _ string) (int64, error) { return 0, nil },
		create: func(_ context.Context, r *PipelineRun) error {
			createdRun = r
			r.ID = "run-123"
			return nil
		},
		update: func(_ context.Context, _, _ string, _ map[string]interface{}) error { return nil },
	}

	pipe := &pipeline.Pipeline{
		ID:                "pipeline-abc",
		ProjectID:         "proj-1",
		Config:            pipeline.PipelineConfig{Stages: []pipeline.StageConfig{{ID: "s1", Name: "build", Steps: []pipeline.StepConfig{{ID: "s1", Name: "build", Type: "build", Image: "alpine"}}}}},
		TriggerType:       pipeline.TriggerManual,
		ConcurrencyPolicy: pipeline.ConcurrencyQueue,
	}

	pipeGetter := &mockPipelineGetter{
		get: func(_ context.Context, _, _ string) (*pipeline.Pipeline, error) { return pipe, nil },
	}

	svc := NewService(repo, pipeGetter, nil, tekton.NewTranslator(), &MockK8sClient{}, &MockSecretInjector{})
	resp, err := svc.TriggerRun(ctx, "proj-1", "pipeline-abc", "user-1", TriggerRunRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "run-123", resp.ID)
	assert.Equal(t, 1, resp.RunNumber)
	assert.Equal(t, "queued", resp.Status)
	assert.NotNil(t, createdRun)
}

func TestTriggerRun_ConcurrencyReject(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		countRunning: func(_ context.Context, _ string) (int64, error) { return 1, nil },
	}

	pipe := &pipeline.Pipeline{
		ID:                "pipeline-abc",
		ProjectID:         "proj-1",
		Config:            pipeline.PipelineConfig{Stages: []pipeline.StageConfig{{ID: "s1", Name: "build", Steps: []pipeline.StepConfig{{Image: "alpine"}}}}},
		ConcurrencyPolicy: pipeline.ConcurrencyReject,
	}

	pipeGetter := &mockPipelineGetter{
		get: func(_ context.Context, _, _ string) (*pipeline.Pipeline, error) { return pipe, nil },
	}

	svc := NewService(repo, pipeGetter, nil, tekton.NewTranslator(), &MockK8sClient{}, &MockSecretInjector{})
	resp, err := svc.TriggerRun(ctx, "proj-1", "pipeline-abc", "user-1", TriggerRunRequest{})
	assert.Error(t, err)
	assert.Nil(t, resp)
	var bizErr *response.BizError
	if errors.As(err, &bizErr) {
		assert.Equal(t, response.CodeRunConcurrency, bizErr.Code)
	}
}

func TestTriggerRun_ConcurrencyCancelOld(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		getNextRunNumber: func(_ context.Context, _ string) (int, error) { return 2, nil },
		countRunning:     func(_ context.Context, _ string) (int64, error) { return 1, nil },
		create: func(_ context.Context, r *PipelineRun) error {
			r.ID = "run-456"
			return nil
		},
		update: func(_ context.Context, _, _ string, _ map[string]interface{}) error { return nil },
	}

	pipe := &pipeline.Pipeline{
		ID:                "pipeline-abc",
		ProjectID:         "proj-1",
		Config:            pipeline.PipelineConfig{Stages: []pipeline.StageConfig{{ID: "s1", Name: "build", Steps: []pipeline.StepConfig{{Image: "alpine"}}}}},
		ConcurrencyPolicy: pipeline.ConcurrencyCancelOld,
	}

	pipeGetter := &mockPipelineGetter{
		get: func(_ context.Context, _, _ string) (*pipeline.Pipeline, error) { return pipe, nil },
	}

	svc := NewService(repo, pipeGetter, nil, tekton.NewTranslator(), &MockK8sClient{}, &MockSecretInjector{})
	resp, err := svc.TriggerRun(ctx, "proj-1", "pipeline-abc", "user-1", TriggerRunRequest{})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "run-456", resp.ID)
}

func TestCancelRun_Success(t *testing.T) {
	ctx := context.Background()

	tektonName := "run-abc-1"
	namespace := "zcid-run"
	repo := &mockRepo{
		getByID: func(_ context.Context, _, _ string) (*PipelineRun, error) {
			return &PipelineRun{ID: "run-1", ProjectID: "proj-1", Status: StatusRunning, TektonName: &tektonName, Namespace: &namespace}, nil
		},
		update: func(_ context.Context, _, _ string, _ map[string]interface{}) error { return nil },
	}

	k8s := &mockK8sClient{
		delete: func(_ context.Context, _, _ string) error { return nil },
	}

	svc := NewService(repo, &mockPipelineGetter{}, nil, tekton.NewTranslator(), k8s, &MockSecretInjector{})
	err := svc.CancelRun(ctx, "proj-1", "run-1")
	require.NoError(t, err)
}

func TestCancelRun_NotRunning(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		getByID: func(_ context.Context, _, _ string) (*PipelineRun, error) {
			return &PipelineRun{ID: "run-1", ProjectID: "proj-1", Status: StatusSucceeded}, nil
		},
	}

	svc := NewService(repo, &mockPipelineGetter{}, nil, tekton.NewTranslator(), &MockK8sClient{}, &MockSecretInjector{})
	err := svc.CancelRun(ctx, "proj-1", "run-1")
	assert.Error(t, err)
	var bizErr *response.BizError
	if errors.As(err, &bizErr) {
		assert.Equal(t, response.CodeRunAlreadyDone, bizErr.Code)
	}
}

func TestGetRun_ProjectScope(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		getByID: func(_ context.Context, id, projectID string) (*PipelineRun, error) {
			if id == "run-1" && projectID == "proj-1" {
				return &PipelineRun{ID: "run-1", ProjectID: "proj-1", PipelineID: "pipe-1", RunNumber: 1, Status: StatusSucceeded}, nil
			}
			return nil, ErrNotFound
		},
	}

	svc := NewService(repo, &mockPipelineGetter{}, nil, tekton.NewTranslator(), &MockK8sClient{}, &MockSecretInjector{})
	run, err := svc.GetRun(ctx, "proj-1", "run-1")
	require.NoError(t, err)
	assert.Equal(t, "run-1", run.ID)
	assert.Equal(t, "succeeded", run.Status)

	_, err = svc.GetRun(ctx, "proj-2", "run-1")
	assert.Error(t, err)
}

func TestUpdateArtifacts_Success(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		getByID: func(_ context.Context, _, _ string) (*PipelineRun, error) {
			return &PipelineRun{ID: "run-1", ProjectID: "proj-1"}, nil
		},
		updateArtifacts: func(_ context.Context, _, _ string, artifacts ArtifactSlice) error {
			assert.Len(t, artifacts, 1)
			assert.Equal(t, "image", artifacts[0].Type)
			assert.Equal(t, "myapp:v1", artifacts[0].Name)
			return nil
		},
	}

	svc := NewService(repo, &mockPipelineGetter{}, nil, tekton.NewTranslator(), &MockK8sClient{}, &MockSecretInjector{})
	err := svc.UpdateArtifacts(ctx, "proj-1", "run-1", []Artifact{{Type: "image", Name: "myapp:v1", URL: "harbor.example.com/myapp:v1"}})
	require.NoError(t, err)
}

func TestGetArtifacts_Success(t *testing.T) {
	ctx := context.Background()

	artifacts := ArtifactSlice{{Type: "file", Name: "dist.zip", URL: "s3://bucket/dist.zip", Size: 1024}}
	repo := &mockRepo{
		getByID: func(_ context.Context, _, _ string) (*PipelineRun, error) {
			return &PipelineRun{ID: "run-1", ProjectID: "proj-1", Artifacts: artifacts}, nil
		},
	}

	svc := NewService(repo, &mockPipelineGetter{}, nil, tekton.NewTranslator(), &MockK8sClient{}, &MockSecretInjector{})
	result, err := svc.GetArtifacts(ctx, "proj-1", "run-1")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "file", result[0].Type)
	assert.Equal(t, "dist.zip", result[0].Name)
}
