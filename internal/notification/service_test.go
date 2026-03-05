package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjy/zcid/pkg/response"
)

type mockNotificationRepo struct {
	create             func(ctx context.Context, r *NotificationRule) error
	findByID           func(ctx context.Context, id, projectID string) (*NotificationRule, error)
	listByProject      func(ctx context.Context, projectID string, page, pageSize int) ([]*NotificationRule, int64, error)
	listByProjectEvent  func(ctx context.Context, projectID string, event EventType) ([]*NotificationRule, error)
	update             func(ctx context.Context, id, projectID string, updates map[string]any) error
	delete             func(ctx context.Context, id, projectID string) error
}

func (m *mockNotificationRepo) Create(ctx context.Context, r *NotificationRule) error {
	if m.create != nil {
		return m.create(ctx, r)
	}
	return nil
}

func (m *mockNotificationRepo) FindByID(ctx context.Context, id, projectID string) (*NotificationRule, error) {
	if m.findByID != nil {
		return m.findByID(ctx, id, projectID)
	}
	return nil, ErrNotFound
}

func (m *mockNotificationRepo) ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*NotificationRule, int64, error) {
	if m.listByProject != nil {
		return m.listByProject(ctx, projectID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockNotificationRepo) ListByProjectAndEvent(ctx context.Context, projectID string, event EventType) ([]*NotificationRule, error) {
	if m.listByProjectEvent != nil {
		return m.listByProjectEvent(ctx, projectID, event)
	}
	return nil, nil
}

func (m *mockNotificationRepo) Update(ctx context.Context, id, projectID string, updates map[string]any) error {
	if m.update != nil {
		return m.update(ctx, id, projectID, updates)
	}
	return nil
}

func (m *mockNotificationRepo) Delete(ctx context.Context, id, projectID string) error {
	if m.delete != nil {
		return m.delete(ctx, id, projectID)
	}
	return nil
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	rule := &NotificationRule{ID: "r1", ProjectID: "p1", Name: "test", WebhookURL: "https://example.com/webhook", EventType: EventBuildSuccess, Enabled: true}
	repo := &mockNotificationRepo{
		create: func(ctx context.Context, r *NotificationRule) error {
			*r = *rule
			return nil
		},
	}
	svc := NewService(repo, nil)
	enabled := true
	got, err := svc.Create(ctx, "p1", "u1", CreateRuleRequest{
		Name:       "test",
		EventType:  EventBuildSuccess,
		WebhookURL: "https://example.com/webhook",
		Enabled:    &enabled,
	})
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "test", got.Name)
}

func TestService_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockNotificationRepo{
		findByID: func(ctx context.Context, id, projectID string) (*NotificationRule, error) {
			return nil, ErrNotFound
		},
	}
	svc := NewService(repo, nil)
	_, err := svc.Get(ctx, "p1", "r1")
	require.Error(t, err)
	var bizErr *response.BizError
	require.True(t, errors.As(err, &bizErr))
	assert.Equal(t, response.CodeNotifRuleNotFound, bizErr.Code)
}

func TestService_Delete_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockNotificationRepo{
		delete: func(ctx context.Context, id, projectID string) error {
			return ErrNotFound
		},
	}
	svc := NewService(repo, nil)
	err := svc.Delete(ctx, "p1", "r1")
	require.Error(t, err)
	var bizErr *response.BizError
	require.True(t, errors.As(err, &bizErr))
	assert.Equal(t, response.CodeNotifRuleNotFound, bizErr.Code)
}
