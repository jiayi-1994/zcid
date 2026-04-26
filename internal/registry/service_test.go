package registry

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/xjy/zcid/pkg/crypto"
	"github.com/xjy/zcid/pkg/response"
)

type mockRepo struct {
	regs    map[string]*Registry
	nextErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{regs: make(map[string]*Registry)}
}

func (m *mockRepo) Create(r *Registry) error {
	if m.nextErr != nil {
		err := m.nextErr
		m.nextErr = nil
		return err
	}
	for _, existing := range m.regs {
		if existing.Status == StatusDeleted {
			continue
		}
		if existing.Name == r.Name {
			return ErrNameDuplicate
		}
	}
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	m.regs[r.ID] = r
	return nil
}

func (m *mockRepo) GetByID(id string) (*Registry, error) {
	if r, ok := m.regs[id]; ok && r.Status != StatusDeleted {
		return r, nil
	}
	return nil, ErrNotFound
}

func (m *mockRepo) List() ([]Registry, int64, error) {
	var result []Registry
	for _, r := range m.regs {
		if r.Status != StatusDeleted {
			result = append(result, *r)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepo) Update(id string, updates map[string]interface{}) error {
	r, ok := m.regs[id]
	if !ok || r.Status == StatusDeleted {
		return ErrNotFound
	}
	if m.nextErr != nil {
		err := m.nextErr
		m.nextErr = nil
		return err
	}
	if name, ok := updates["name"].(string); ok {
		for _, existing := range m.regs {
			if existing.ID != id && existing.Status != StatusDeleted && existing.Name == name {
				return ErrNameDuplicate
			}
		}
		r.Name = name
	}
	return nil
}

func (m *mockRepo) SoftDelete(id string) error {
	r, ok := m.regs[id]
	if !ok || r.Status == StatusDeleted {
		return ErrNotFound
	}
	r.Status = StatusDeleted
	return nil
}

func (m *mockRepo) GetDefault() (*Registry, error) {
	for _, r := range m.regs {
		if r.Status != StatusDeleted && r.IsDefault {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) SetDefault(id string) error {
	r, ok := m.regs[id]
	if !ok || r.Status == StatusDeleted {
		return ErrNotFound
	}
	for _, reg := range m.regs {
		reg.IsDefault = false
	}
	r.IsDefault = true
	return nil
}

func testCrypto() *crypto.AESCrypto {
	c, _ := crypto.NewAESCrypto([]byte("01234567890123456789012345678901"))
	return c
}

func TestCreateRegistry(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	reg, err := svc.Create(CreateRegistryRequest{
		Name:      "harbor-prod",
		Type:      "harbor",
		URL:       "https://harbor.example.com",
		Username:  "admin",
		Password:  "secret",
		IsDefault: true,
	}, "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg.Name != "harbor-prod" || reg.URL != "https://harbor.example.com" {
		t.Fatalf("unexpected registry: %+v", reg)
	}
	if !reg.IsDefault {
		t.Fatal("expected is_default true")
	}
	if reg.PasswordEncrypted == "secret" {
		t.Fatal("password should be encrypted")
	}
}

func TestTestConnection(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	res, err := svc.TestConnection(TestConnectionRequest{
		URL:      server.URL,
		Username: "admin",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatal("expected success")
	}
}

func TestSetDefault(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	r1, _ := svc.Create(CreateRegistryRequest{
		Name: "reg1", Type: "harbor", URL: "https://a.com",
	}, "user1")
	r2, _ := svc.Create(CreateRegistryRequest{
		Name: "reg2", Type: "harbor", URL: "https://b.com",
	}, "user1")

	def, err := svc.GetDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def != nil {
		t.Fatal("expected no default initially")
	}

	_, err = svc.Update(r2.ID, UpdateRegistryRequest{IsDefault: boolPtr(true)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def, _ = svc.GetDefault()
	if def == nil || def.ID != r2.ID {
		t.Fatalf("expected default to be reg2, got %v", def)
	}
	_ = r1
}

func TestDeleteRegistry(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	reg, _ := svc.Create(CreateRegistryRequest{
		Name: "reg1", Type: "harbor", URL: "https://a.com",
	}, "user1")

	if err := svc.Delete(reg.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.Get(reg.ID)
	if err == nil {
		t.Fatal("expected not found after delete")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeRegistryNotFound {
		t.Fatalf("expected CodeRegistryNotFound, got %v", err)
	}
}

func TestCreateRegistry_NameDup(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, testCrypto())

	_, _ = svc.Create(CreateRegistryRequest{
		Name: "harbor", Type: "harbor", URL: "https://a.com",
	}, "user1")

	_, err := svc.Create(CreateRegistryRequest{
		Name: "harbor", Type: "harbor", URL: "https://b.com",
	}, "user1")
	if err == nil {
		t.Fatal("expected duplicate name error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeRegistryNameDup {
		t.Fatalf("expected CodeRegistryNameDup, got %v", err)
	}
}

func boolPtr(b bool) *bool {
	return &b
}
