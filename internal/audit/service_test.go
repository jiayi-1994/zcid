package audit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuditRepo struct {
	create func(ctx context.Context, log *AuditLog) error
	list   func(ctx context.Context, opts ListOpts) ([]*AuditLog, int64, error)
}

func (m *mockAuditRepo) Create(ctx context.Context, log *AuditLog) error {
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
	var created *AuditLog
	repo := &mockAuditRepo{
		create: func(ctx context.Context, log *AuditLog) error {
			created = log
			return nil
		},
	}
	svc := NewService(repo)
	svc.LogAction(ctx, "u1", "POST /test", "project", "p1", "success", "127.0.0.1", "")
	time.Sleep(50 * time.Millisecond)
	require.NotNil(t, created)
	assert.Equal(t, "u1", *created.UserID)
	assert.Equal(t, "POST /test", created.Action)
	assert.Equal(t, "project", created.ResourceType)
	assert.Equal(t, "p1", *created.ResourceID)
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
