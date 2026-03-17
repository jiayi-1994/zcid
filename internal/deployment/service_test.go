package deployment

import (
	"context"
	"errors"
	"testing"

	"github.com/xjy/zcid/internal/environment"
	"github.com/xjy/zcid/pkg/argocd"
	"github.com/xjy/zcid/pkg/response"
)

type mockRepo struct {
	createErr     error
	findByIDErr   error
	findByIDDep   *Deployment
	listByProject []*Deployment
	listTotal     int64
	listErr       error
	updateErr     error
}

func (m *mockRepo) Create(ctx context.Context, d *Deployment) error {
	if m.createErr != nil {
		return m.createErr
	}
	return nil
}

func (m *mockRepo) FindByID(ctx context.Context, id, projectID string) (*Deployment, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	if m.findByIDDep != nil {
		return m.findByIDDep, nil
	}
	return nil, ErrNotFound
}

func (m *mockRepo) ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*Deployment, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listByProject, m.listTotal, nil
}

func (m *mockRepo) ListByEnvironment(ctx context.Context, projectID, envID string, page, pageSize int) ([]*Deployment, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listByProject, m.listTotal, nil
}

func (m *mockRepo) Update(ctx context.Context, id, projectID string, updates map[string]any) error {
	return m.updateErr
}

type mockEnvGetter struct {
	env *environment.Environment
	err error
}

func (m *mockEnvGetter) Get(ctx context.Context, id, projectID string) (*environment.Environment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.env, nil
}

func TestTriggerDeploy(t *testing.T) {
	env := &environment.Environment{ID: "e1", ProjectID: "p1", Name: "dev", Namespace: "zcid-dev"}
	argo := &argocd.MockArgoClient{}
	envGetter := &mockEnvGetter{env: env}
	repo := &mockRepo{}

	svc := NewService(repo, envGetter, argo)
	d, err := svc.TriggerDeploy(context.Background(), "p1", "u1", TriggerDeployRequest{
		EnvironmentID: "e1",
		Image:         "nginx:latest",
	})
	if err != nil {
		t.Fatalf("TriggerDeploy: %v", err)
	}
	if d == nil {
		t.Fatal("expected deployment")
	}
	if d.ProjectID != "p1" || d.EnvironmentID != "e1" || d.Image != "nginx:latest" {
		t.Errorf("unexpected deployment: %+v", d)
	}
}

func TestGetDeployStatus(t *testing.T) {
	appName := "zcid-p1-dev"
	dep := &Deployment{ID: "d1", ProjectID: "p1", ArgoAppName: &appName}
	repo := &mockRepo{findByIDDep: dep}
	envGetter := &mockEnvGetter{}
	argo := &argocd.MockArgoClient{}

	svc := NewService(repo, envGetter, argo)
	d, err := svc.GetDeployStatus(context.Background(), "p1", "d1")
	if err != nil {
		t.Fatalf("GetDeployStatus: %v", err)
	}
	if d == nil {
		t.Fatal("expected deployment")
	}
}

func TestGetDeployStatus_NotFound(t *testing.T) {
	repo := &mockRepo{findByIDErr: ErrNotFound}
	svc := NewService(repo, &mockEnvGetter{}, &argocd.MockArgoClient{})
	_, err := svc.GetDeployStatus(context.Background(), "p1", "d1")
	if err == nil {
		t.Fatal("expected error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeDeployNotFound {
		t.Errorf("expected CodeDeployNotFound, got %v", err)
	}
}

func TestListDeployments(t *testing.T) {
	repo := &mockRepo{
		listByProject: []*Deployment{{ID: "d1", ProjectID: "p1", Image: "nginx:1"}},
		listTotal:     1,
	}
	svc := NewService(repo, &mockEnvGetter{}, &argocd.MockArgoClient{})
	list, total, err := svc.ListDeployments(context.Background(), "p1", 1, 20)
	if err != nil {
		t.Fatalf("ListDeployments: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Errorf("expected 1 item, got %d/%d", len(list), total)
	}
}
