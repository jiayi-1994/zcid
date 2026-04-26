package audit

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuditRepo struct {
	mu     sync.Mutex
	create func(ctx context.Context, log *AuditLog) error
	list   func(ctx context.Context, opts ListOpts) ([]*AuditLog, int64, error)
}

func (m *mockAuditRepo) Create(ctx context.Context, log *AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.create != nil {
		return m.create(ctx, log)
	}
	return nil
}

func (m *mockAuditRepo) List(ctx context.Context, opts ListOpts) ([]*AuditLog, int64, error) {
	if m.list != nil {
		return m.list(ctx, opts)
	}
	return nil, 0, nil
}

func TestService_LogAction(t *testing.T) {
	ctx := context.Background()
	var (
		mu      sync.Mutex
		created *AuditLog
	)
	repo := &mockAuditRepo{
		create: func(ctx context.Context, log *AuditLog) error {
			mu.Lock()
			defer mu.Unlock()
			created = log
			return nil
		},
	}
	svc := NewService(repo)
	svc.LogAction(ctx, "u1", "POST /test", "project", "p1", "success", "127.0.0.1", "")
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	c := created
	mu.Unlock()

	require.NotNil(t, c)
	assert.Equal(t, "u1", *c.UserID)
	assert.Equal(t, "POST /test", c.Action)
	assert.Equal(t, "project", c.ResourceType)
	assert.Equal(t, "p1", *c.ResourceID)
}

func TestService_LogAuthSecurityEvent(t *testing.T) {
	ctx := context.Background()
	var (
		mu      sync.Mutex
		created *AuditLog
	)
	repo := &mockAuditRepo{
		create: func(ctx context.Context, log *AuditLog) error {
			mu.Lock()
			defer mu.Unlock()
			created = log
			return nil
		},
	}
	svc := NewService(repo)
	svc.LogAuthSecurityEvent(ctx, AuthSecurityEvent{
		UserID:     "u1",
		Action:     "auth.login_failed",
		ResourceID: "u1",
		Result:     ResultFailure,
		IP:         "127.0.0.1",
		Detail: AuthSecurityDetail{
			PrincipalType: "user",
			Reason:        "invalid_credentials",
			Fields:        map[string]any{"username": "alice"},
		},
	})
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	c := created
	mu.Unlock()

	require.NotNil(t, c)
	assert.Equal(t, "u1", *c.UserID)
	assert.Equal(t, "auth.login_failed", c.Action)
	assert.Equal(t, ResourceTypeAuthSecurity, c.ResourceType)
	assert.Equal(t, ResultFailure, c.Result)
	assert.Equal(t, "u1", *c.ResourceID)
	assert.Equal(t, "127.0.0.1", *c.IP)
	require.NotNil(t, c.Detail)

	var detail AuthSecurityDetail
	require.NoError(t, json.Unmarshal([]byte(*c.Detail), &detail))
	assert.Equal(t, "auth.login_failed", detail.EventType)
	assert.Equal(t, "user", detail.PrincipalType)
	assert.Equal(t, "invalid_credentials", detail.Reason)
	assert.NotContains(t, *c.Detail, "password")
	assert.NotContains(t, *c.Detail, "token")
}

func TestService_LogAuthSecurityEventDefaults(t *testing.T) {
	ctx := context.Background()
	var created *AuditLog
	repo := &mockAuditRepo{
		create: func(ctx context.Context, log *AuditLog) error {
			created = log
			return nil
		},
	}
	svc := NewService(repo)
	svc.LogAuthSecurityEvent(ctx, AuthSecurityEvent{Detail: AuthSecurityDetail{EventType: "auth.logout"}})
	time.Sleep(100 * time.Millisecond)

	require.NotNil(t, created)
	assert.Equal(t, "auth.logout", created.Action)
	assert.Equal(t, ResourceTypeAuthSecurity, created.ResourceType)
	assert.Equal(t, ResultSuccess, created.Result)
}

func TestService_List(t *testing.T) {
	ctx := context.Background()
	repo := &mockAuditRepo{
		list: func(ctx context.Context, opts ListOpts) ([]*AuditLog, int64, error) {
			return []*AuditLog{{ID: "1", Action: "GET"}}, 1, nil
		},
	}
	svc := NewService(repo)
	list, total, err := svc.List(ctx, ListOpts{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(1), total)
}
