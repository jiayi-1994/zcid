package variable

import (
	"errors"
	"testing"

	"github.com/xjy/zcid/pkg/crypto"
	"github.com/xjy/zcid/pkg/response"
)

type mockRepo struct {
	vars    map[string]*Variable
	nextErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{vars: make(map[string]*Variable)}
}

func (m *mockRepo) Create(v *Variable) error {
	if m.nextErr != nil {
		err := m.nextErr
		m.nextErr = nil
		return err
	}
	for _, existing := range m.vars {
		if existing.Status == StatusDeleted {
			continue
		}
		if existing.Key == v.Key && existing.Scope == v.Scope {
			if v.Scope == ScopeGlobal {
				return ErrKeyDuplicate
			}
			if v.Scope == ScopeProject && v.ProjectID != nil && existing.ProjectID != nil && *existing.ProjectID == *v.ProjectID {
				return ErrKeyDuplicate
			}
			if v.Scope == ScopePipeline && v.ProjectID != nil && existing.ProjectID != nil && *existing.ProjectID == *v.ProjectID &&
				v.PipelineID != nil && existing.PipelineID != nil && *existing.PipelineID == *v.PipelineID {
				return ErrKeyDuplicate
			}
		}
	}
	m.vars[v.ID] = v
	return nil
}

func (m *mockRepo) GetByID(id string) (*Variable, error) {
	if v, ok := m.vars[id]; ok && v.Status != StatusDeleted {
		return v, nil
	}
	return nil, ErrNotFound
}

func (m *mockRepo) ListByProject(projectID string) ([]Variable, int64, error) {
	var result []Variable
	for _, v := range m.vars {
		if v.ProjectID != nil && *v.ProjectID == projectID && v.Scope == ScopeProject && v.Status != StatusDeleted {
			result = append(result, *v)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepo) ListGlobal() ([]Variable, int64, error) {
	var result []Variable
	for _, v := range m.vars {
		if v.Scope == ScopeGlobal && v.Status != StatusDeleted {
			result = append(result, *v)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepo) Update(id string, updates map[string]interface{}) error {
	v, ok := m.vars[id]
	if !ok || v.Status == StatusDeleted {
		return ErrNotFound
	}
	if val, ok := updates["value"]; ok {
		v.Value = val.(string)
	}
	if val, ok := updates["description"]; ok {
		v.Description = val.(string)
	}
	return nil
}

func (m *mockRepo) SoftDelete(id string) error {
	v, ok := m.vars[id]
	if !ok || v.Status == StatusDeleted {
		return ErrNotFound
	}
	v.Status = StatusDeleted
	return nil
}

func (m *mockRepo) ListGlobalAndProject(projectID string) ([]Variable, error) {
	var result []Variable
	for _, v := range m.vars {
		if v.Status == StatusDeleted {
			continue
		}
		if v.Scope == ScopeGlobal || (v.Scope == ScopeProject && v.ProjectID != nil && *v.ProjectID == projectID) {
			result = append(result, *v)
		}
	}
	return result, nil
}

func (m *mockRepo) ListByPipelineScope(projectID, pipelineID string) ([]Variable, error) {
	var result []Variable
	for _, v := range m.vars {
		if v.Status == StatusDeleted || v.Scope != ScopePipeline {
			continue
		}
		if v.ProjectID != nil && *v.ProjectID == projectID && v.PipelineID != nil && *v.PipelineID == pipelineID {
			result = append(result, *v)
		}
	}
	return result, nil
}

func testCrypto() *crypto.AESCrypto {
	c, _ := crypto.NewAESCrypto([]byte("01234567890123456789012345678901"))
	return c
}

func TestCreateVariable_Plain(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	v, err := svc.CreateVariable(ScopeGlobal, nil, nil, CreateVariableRequest{
		Key: "DB_HOST", Value: "localhost",
	}, "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Key != "DB_HOST" || v.Value != "localhost" {
		t.Fatalf("unexpected variable: %+v", v)
	}
	if v.VarType != TypePlain {
		t.Fatalf("expected plain type, got %s", v.VarType)
	}
}

func TestCreateVariable_Secret(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	v, err := svc.CreateVariable(ScopeGlobal, nil, nil, CreateVariableRequest{
		Key: "DB_PASSWORD", Value: "secret123", VarType: "secret",
	}, "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.VarType != TypeSecret {
		t.Fatalf("expected secret type, got %s", v.VarType)
	}
	if v.Value == "secret123" {
		t.Fatal("value should be encrypted, not plaintext")
	}
}

func TestCreateVariable_Duplicate(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	_, _ = svc.CreateVariable(ScopeGlobal, nil, nil, CreateVariableRequest{
		Key: "KEY1", Value: "val1",
	}, "user1")

	_, err := svc.CreateVariable(ScopeGlobal, nil, nil, CreateVariableRequest{
		Key: "KEY1", Value: "val2",
	}, "user1")
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeVarDuplicate {
		t.Fatalf("expected CodeVarDuplicate, got %v", err)
	}
}

func TestCreateVariable_Secret_NoCrypto(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	_, err := svc.CreateVariable(ScopeGlobal, nil, nil, CreateVariableRequest{
		Key: "SECRET", Value: "val", VarType: "secret",
	}, "user1")
	if err == nil {
		t.Fatal("expected error when crypto is nil")
	}
}

func TestGetVariable(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	created, _ := svc.CreateVariable(ScopeGlobal, nil, nil, CreateVariableRequest{
		Key: "K1", Value: "V1",
	}, "user1")

	v, err := svc.GetVariable(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Key != "K1" {
		t.Fatalf("expected key K1, got %s", v.Key)
	}
}

func TestGetVariable_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	_, err := svc.GetVariable("nonexistent")
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestDeleteVariable(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	created, _ := svc.CreateVariable(ScopeGlobal, nil, nil, CreateVariableRequest{
		Key: "K1", Value: "V1",
	}, "user1")

	if err := svc.DeleteVariable(created.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.GetVariable(created.ID)
	if err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestUpdateVariable(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	created, _ := svc.CreateVariable(ScopeGlobal, nil, nil, CreateVariableRequest{
		Key: "K1", Value: "V1",
	}, "user1")

	newVal := "V2"
	err := svc.UpdateVariable(created.ID, UpdateVariableRequest{Value: &newVal}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, _ := svc.GetVariable(created.ID)
	if v.Value != "V2" {
		t.Fatalf("expected V2, got %s", v.Value)
	}
}

func TestGetMergedVariables(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	pid := "proj1"
	_, _ = svc.CreateVariable(ScopeGlobal, nil, nil, CreateVariableRequest{
		Key: "DB_HOST", Value: "global-db",
	}, "user1")
	_, _ = svc.CreateVariable(ScopeGlobal, nil, nil, CreateVariableRequest{
		Key: "REDIS_HOST", Value: "global-redis",
	}, "user1")
	_, _ = svc.CreateVariable(ScopeProject, &pid, nil, CreateVariableRequest{
		Key: "DB_HOST", Value: "project-db",
	}, "user1")

	merged, err := svc.GetMergedVariables(pid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := make(map[string]string)
	for _, v := range merged {
		m[v.Key] = v.Value
	}

	if m["DB_HOST"] != "project-db" {
		t.Fatalf("expected project-db, got %s", m["DB_HOST"])
	}
	if m["REDIS_HOST"] != "global-redis" {
		t.Fatalf("expected global-redis, got %s", m["REDIS_HOST"])
	}
}

func TestResolveVariables_DecryptsSecrets(t *testing.T) {
	repo := newMockRepo()
	c := testCrypto()
	svc := NewService(repo, c)

	_, _ = svc.CreateVariable(ScopeGlobal, nil, nil, CreateVariableRequest{
		Key: "API_KEY", Value: "my-secret-key", VarType: "secret",
	}, "user1")

	resolved, err := svc.ResolveVariables("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, v := range resolved {
		if v.Key == "API_KEY" {
			if v.Value != "my-secret-key" {
				t.Fatalf("expected decrypted value, got %s", v.Value)
			}
			return
		}
	}
	t.Fatal("API_KEY not found in resolved variables")
}

func TestFilterForRole(t *testing.T) {
	vars := []Variable{
		{Key: "K1", VarType: TypePlain},
		{Key: "K2", VarType: TypeSecret},
		{Key: "K3", VarType: TypePlain},
	}

	adminFiltered := FilterForRole(vars, "admin")
	if len(adminFiltered) != 3 {
		t.Fatalf("admin should see all 3 vars, got %d", len(adminFiltered))
	}

	memberFiltered := FilterForRole(vars, "member")
	if len(memberFiltered) != 2 {
		t.Fatalf("member should see 2 plain vars, got %d", len(memberFiltered))
	}
	for _, v := range memberFiltered {
		if v.VarType == TypeSecret {
			t.Fatal("member should not see secret vars")
		}
	}
}
