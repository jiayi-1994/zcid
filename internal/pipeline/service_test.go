package pipeline

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/xjy/zcid/pkg/response"
)

type mockRepo struct {
	pipelines map[string]*Pipeline
	nextErr   error
}

func newMockRepo() *mockRepo {
	return &mockRepo{pipelines: make(map[string]*Pipeline)}
}

func (m *mockRepo) Create(ctx context.Context, p *Pipeline) error {
	if m.nextErr != nil {
		err := m.nextErr
		m.nextErr = nil
		return err
	}
	for _, existing := range m.pipelines {
		if existing.ProjectID == p.ProjectID && existing.Name == p.Name && existing.Status != StatusDeleted {
			return ErrNameDuplicate
		}
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.pipelines[p.ID] = p
	return nil
}

func (m *mockRepo) GetByIDAndProject(ctx context.Context, id, projectID string) (*Pipeline, error) {
	if m.nextErr != nil {
		err := m.nextErr
		m.nextErr = nil
		return nil, err
	}
	p, ok := m.pipelines[id]
	if !ok || p.Status == StatusDeleted || p.ProjectID != projectID {
		return nil, ErrNotFound
	}
	return p, nil
}

func (m *mockRepo) List(ctx context.Context, projectID string, page, pageSize int) ([]*Pipeline, int64, error) {
	if m.nextErr != nil {
		err := m.nextErr
		m.nextErr = nil
		return nil, 0, err
	}
	var result []*Pipeline
	for _, p := range m.pipelines {
		if p.ProjectID == projectID && p.Status != StatusDeleted {
			result = append(result, p)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	total := int64(len(result))
	offset := (page - 1) * pageSize
	if offset >= len(result) {
		return []*Pipeline{}, total, nil
	}
	end := offset + pageSize
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (m *mockRepo) ListByTriggerType(ctx context.Context, triggerType TriggerType) ([]*Pipeline, error) {
	if m.nextErr != nil {
		err := m.nextErr
		m.nextErr = nil
		return nil, err
	}
	var result []*Pipeline
	for _, p := range m.pipelines {
		if p.TriggerType == triggerType && p.Status != StatusDeleted {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockRepo) Update(ctx context.Context, id, projectID string, updates map[string]any) error {
	if m.nextErr != nil {
		err := m.nextErr
		m.nextErr = nil
		return err
	}
	p, ok := m.pipelines[id]
	if !ok || p.Status == StatusDeleted || p.ProjectID != projectID {
		return ErrNotFound
	}
	if name, ok := updates["name"]; ok {
		for _, ep := range m.pipelines {
			if ep.ProjectID == projectID && ep.Name == name.(string) && ep.ID != id && ep.Status != StatusDeleted {
				return ErrNameDuplicate
			}
		}
		p.Name = name.(string)
	}
	if desc, ok := updates["description"]; ok {
		p.Description = desc.(string)
	}
	if status, ok := updates["status"]; ok {
		p.Status = PipelineStatus(status.(string))
	}
	if config, ok := updates["config"]; ok {
		p.Config = config.(PipelineConfig)
	}
	if tt, ok := updates["trigger_type"]; ok {
		p.TriggerType = TriggerType(tt.(string))
	}
	if cp, ok := updates["concurrency_policy"]; ok {
		p.ConcurrencyPolicy = ConcurrencyPolicy(cp.(string))
	}
	p.UpdatedAt = time.Now()
	return nil
}

func (m *mockRepo) SoftDelete(ctx context.Context, id, projectID string) error {
	if m.nextErr != nil {
		err := m.nextErr
		m.nextErr = nil
		return err
	}
	p, ok := m.pipelines[id]
	if !ok || p.Status == StatusDeleted || p.ProjectID != projectID {
		return ErrNotFound
	}
	p.Status = StatusDeleted
	return nil
}

func (m *mockRepo) ExistsByNameAndProject(ctx context.Context, projectID, name string, excludeID string) (bool, error) {
	for _, p := range m.pipelines {
		if p.ProjectID == projectID && p.Name == name && p.Status != StatusDeleted {
			if excludeID != "" && p.ID == excludeID {
				continue
			}
			return true, nil
		}
	}
	return false, nil
}

func TestCreatePipeline_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{
		Name: "build-go",
		Config: PipelineConfig{
			Stages: []StageConfig{
				{ID: "s1", Name: "Build", Steps: []StepConfig{{ID: "st1", Name: "compile", Type: "shell"}}},
			},
		},
	}

	p, err := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "build-go" {
		t.Errorf("expected name 'build-go', got '%s'", p.Name)
	}
	if p.Config.SchemaVersion != "1.0" {
		t.Errorf("expected schemaVersion '1.0', got '%s'", p.Config.SchemaVersion)
	}
	if p.Status != StatusDraft {
		t.Errorf("expected status 'draft', got '%s'", p.Status)
	}
	if p.ProjectID != "proj-1" {
		t.Errorf("expected projectId 'proj-1', got '%s'", p.ProjectID)
	}
	if p.TriggerType != TriggerManual {
		t.Errorf("expected triggerType 'manual', got '%s'", p.TriggerType)
	}
}

func TestCreatePipeline_NameDuplicate(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{Name: "build-go"}
	_, err := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err = svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
	bizErr, ok := err.(*response.BizError)
	if !ok {
		t.Fatalf("expected BizError, got %T", err)
	}
	if bizErr.Code != response.CodePipelineNameDup {
		t.Errorf("expected code %d, got %d", response.CodePipelineNameDup, bizErr.Code)
	}
}

func TestCreatePipeline_InvalidTriggerType(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{Name: "test", TriggerType: "invalid"}
	_, err := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")
	if err == nil {
		t.Fatal("expected error for invalid trigger type")
	}
	bizErr, ok := err.(*response.BizError)
	if !ok {
		t.Fatalf("expected BizError, got %T", err)
	}
	if bizErr.Code != response.CodeValidation {
		t.Errorf("expected code %d, got %d", response.CodeValidation, bizErr.Code)
	}
}

func TestGetPipeline_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.GetPipeline(context.Background(), "nonexistent", "proj-1")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	bizErr, ok := err.(*response.BizError)
	if !ok {
		t.Fatalf("expected BizError, got %T", err)
	}
	if bizErr.Code != response.CodePipelineNotFound {
		t.Errorf("expected code %d, got %d", response.CodePipelineNotFound, bizErr.Code)
	}
}

func TestGetPipeline_CrossProjectDenied(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{Name: "secret-pipeline"}
	p, _ := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")

	_, err := svc.GetPipeline(context.Background(), p.ID, "proj-2")
	if err == nil {
		t.Fatal("expected error for cross-project access")
	}
	bizErr, ok := err.(*response.BizError)
	if !ok {
		t.Fatalf("expected BizError, got %T", err)
	}
	if bizErr.Code != response.CodePipelineNotFound {
		t.Errorf("expected code %d, got %d", response.CodePipelineNotFound, bizErr.Code)
	}
}

func TestCopyPipeline_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{
		Name: "original",
		Config: PipelineConfig{
			SchemaVersion: "1.0",
			Stages: []StageConfig{
				{ID: "s1", Name: "Build", Steps: []StepConfig{{ID: "st1", Name: "test", Type: "shell"}}},
			},
		},
	}
	original, err := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	copied, err := svc.CopyPipeline(context.Background(), original.ID, "proj-1", "user-2")
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	if copied.Name != "original-copy" {
		t.Errorf("expected name 'original-copy', got '%s'", copied.Name)
	}
	if copied.ID == original.ID {
		t.Error("copied pipeline should have a different ID")
	}
	if copied.Status != StatusDraft {
		t.Errorf("expected status 'draft', got '%s'", copied.Status)
	}
	if copied.CreatedBy != "user-2" {
		t.Errorf("expected createdBy 'user-2', got '%s'", copied.CreatedBy)
	}
	if len(copied.Config.Stages) != 1 {
		t.Errorf("expected 1 stage, got %d", len(copied.Config.Stages))
	}
}

func TestCopyPipeline_DuplicateName(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{Name: "build"}
	original, _ := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")

	copyReq := CreatePipelineRequest{Name: "build-copy"}
	_, _ = svc.CreatePipeline(context.Background(), "proj-1", copyReq, "user-1")

	copied, err := svc.CopyPipeline(context.Background(), original.ID, "proj-1", "user-1")
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}
	if copied.Name != "build-copy-2" {
		t.Errorf("expected name 'build-copy-2', got '%s'", copied.Name)
	}
}

func TestCopyPipeline_CrossProjectDenied(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{Name: "secret"}
	original, _ := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")

	_, err := svc.CopyPipeline(context.Background(), original.ID, "proj-2", "user-2")
	if err == nil {
		t.Fatal("expected error for cross-project copy")
	}
}

func TestUpdatePipeline_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{Name: "old-name"}
	p, _ := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")

	newName := "new-name"
	updated, err := svc.UpdatePipeline(context.Background(), p.ID, "proj-1", UpdatePipelineRequest{Name: &newName})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "new-name" {
		t.Errorf("expected name 'new-name', got '%s'", updated.Name)
	}
}

func TestUpdatePipeline_InvalidStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{Name: "test"}
	p, _ := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")

	badStatus := "invalid"
	_, err := svc.UpdatePipeline(context.Background(), p.ID, "proj-1", UpdatePipelineRequest{Status: &badStatus})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	bizErr, ok := err.(*response.BizError)
	if !ok {
		t.Fatalf("expected BizError, got %T", err)
	}
	if bizErr.Code != response.CodeValidation {
		t.Errorf("expected code %d, got %d", response.CodeValidation, bizErr.Code)
	}
}

func TestUpdatePipeline_NameConflict(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, _ = svc.CreatePipeline(context.Background(), "proj-1", CreatePipelineRequest{Name: "existing"}, "user-1")
	p2, _ := svc.CreatePipeline(context.Background(), "proj-1", CreatePipelineRequest{Name: "other"}, "user-1")

	conflictName := "existing"
	_, err := svc.UpdatePipeline(context.Background(), p2.ID, "proj-1", UpdatePipelineRequest{Name: &conflictName})
	if err == nil {
		t.Fatal("expected error for name conflict")
	}
	bizErr, ok := err.(*response.BizError)
	if !ok {
		t.Fatalf("expected BizError, got %T", err)
	}
	if bizErr.Code != response.CodePipelineNameDup {
		t.Errorf("expected code %d, got %d", response.CodePipelineNameDup, bizErr.Code)
	}
}

func TestDeletePipeline_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	p, _ := svc.CreatePipeline(context.Background(), "proj-1", CreatePipelineRequest{Name: "to-delete"}, "user-1")

	err := svc.DeletePipeline(context.Background(), p.ID, "proj-1")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = svc.GetPipeline(context.Background(), p.ID, "proj-1")
	if err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestDeletePipeline_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	err := svc.DeletePipeline(context.Background(), "nonexistent", "proj-1")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	bizErr, ok := err.(*response.BizError)
	if !ok {
		t.Fatalf("expected BizError, got %T", err)
	}
	if bizErr.Code != response.CodePipelineNotFound {
		t.Errorf("expected code %d, got %d", response.CodePipelineNotFound, bizErr.Code)
	}
}

func TestListTemplates(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	templates := svc.ListTemplates()
	if len(templates) != 5 {
		t.Errorf("expected 5 templates, got %d", len(templates))
	}

	ids := make(map[string]bool)
	for _, tmpl := range templates {
		ids[tmpl.ID] = true
	}
	expected := []string{"go-microservice", "java-maven", "java-jar-traditional", "frontend-node", "generic-docker"}
	for _, id := range expected {
		if !ids[id] {
			t.Errorf("missing template: %s", id)
		}
	}
}

func TestGetTemplate_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	tmpl, err := svc.GetTemplate("go-microservice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.Name != "Go 微服务" {
		t.Errorf("expected name 'Go 微服务', got '%s'", tmpl.Name)
	}
	if len(tmpl.Params) == 0 {
		t.Error("expected params to be non-empty")
	}
	if len(tmpl.Config.Stages) != 3 {
		t.Errorf("expected 3 stages, got %d", len(tmpl.Config.Stages))
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.GetTemplate("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
}

func TestCreateFromTemplate_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{
		Name:       "my-go-service",
		TemplateID: "go-microservice",
		TemplateParams: map[string]string{
			"repoUrl":   "https://github.com/example/repo",
			"branch":    "develop",
			"imageName": "registry.example.com/my-service",
			"goVersion": "1.24",
		},
	}

	p, err := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "my-go-service" {
		t.Errorf("expected name 'my-go-service', got '%s'", p.Name)
	}
	if len(p.Config.Stages) != 3 {
		t.Fatalf("expected 3 stages from template, got %d", len(p.Config.Stages))
	}
	if p.Config.Stages[1].Steps[0].Image != "golang:1.24" {
		t.Errorf("expected image 'golang:1.24', got '%s'", p.Config.Stages[1].Steps[0].Image)
	}
}

func TestCreateFromTemplate_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{
		Name:       "test",
		TemplateID: "nonexistent-template",
	}

	_, err := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
}

func TestCreateFromTemplate_MissingRequiredParams(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{
		Name:           "test",
		TemplateID:     "go-microservice",
		TemplateParams: map[string]string{},
	}

	_, err := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")
	if err == nil {
		t.Fatal("expected error for missing required params")
	}
	bizErr, ok := err.(*response.BizError)
	if !ok {
		t.Fatalf("expected BizError, got %T", err)
	}
	if bizErr.Code != response.CodeValidation {
		t.Errorf("expected code %d, got %d", response.CodeValidation, bizErr.Code)
	}
}

func TestCreateFromTemplate_NilParams(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{
		Name:       "test",
		TemplateID: "go-microservice",
	}

	_, err := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")
	if err == nil {
		t.Fatal("expected error for nil template params (missing required)")
	}
}

func TestCreateFromTemplate_DefaultValues(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{
		Name:       "test-defaults",
		TemplateID: "go-microservice",
		TemplateParams: map[string]string{
			"repoUrl":   "https://github.com/example/repo",
			"imageName": "registry.example.com/my-service",
		},
	}

	p, err := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Config.Stages[1].Steps[0].Image != "golang:1.24" {
		t.Errorf("expected default goVersion '1.24' applied, got image '%s'", p.Config.Stages[1].Steps[0].Image)
	}
}

func TestCreateFromTemplate_IllegalParamChars(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	req := CreatePipelineRequest{
		Name:       "test-injection",
		TemplateID: "go-microservice",
		TemplateParams: map[string]string{
			"repoUrl":   "https://github.com/example/repo",
			"branch":    "main",
			"imageName": "evil\"injected",
			"goVersion": "1.24",
		},
	}

	_, err := svc.CreatePipeline(context.Background(), "proj-1", req, "user-1")
	if err == nil {
		t.Fatal("expected error for illegal characters in params")
	}
}

func TestTemplateListOrder(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	templates := svc.ListTemplates()
	if len(templates) < 2 {
		t.Fatal("expected at least 2 templates")
	}
	for i := 1; i < len(templates); i++ {
		if templates[i].ID < templates[i-1].ID {
			t.Errorf("templates not sorted: %s before %s", templates[i-1].ID, templates[i].ID)
		}
	}
}

func TestListPipelines_Pagination(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	for i := 0; i < 5; i++ {
		_, _ = svc.CreatePipeline(context.Background(), "proj-1", CreatePipelineRequest{
			Name: "pipeline-" + string(rune('A'+i)),
		}, "user-1")
		time.Sleep(time.Millisecond)
	}

	pipelines, total, err := svc.ListPipelines(context.Background(), "proj-1", 1, 3)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(pipelines) != 3 {
		t.Errorf("expected 3 items, got %d", len(pipelines))
	}
}
