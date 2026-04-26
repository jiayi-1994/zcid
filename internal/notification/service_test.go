package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjy/zcid/pkg/crypto"
	"github.com/xjy/zcid/pkg/response"
)

type mockNotificationRepo struct {
	create             func(ctx context.Context, r *NotificationRule) error
	findByID           func(ctx context.Context, id, projectID string) (*NotificationRule, error)
	listByProject      func(ctx context.Context, projectID string, page, pageSize int) ([]*NotificationRule, int64, error)
	listByProjectEvent func(ctx context.Context, projectID string, event EventType) ([]*NotificationRule, error)
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

func TestService_CreateSlackRuleEncryptsToken(t *testing.T) {
	ctx := context.Background()
	aesCrypto, err := crypto.NewAESCrypto([]byte("01234567890123456789012345678901"))
	require.NoError(t, err)
	var created *NotificationRule
	repo := &mockNotificationRepo{
		create: func(ctx context.Context, r *NotificationRule) error {
			created = r
			return nil
		},
	}
	svc := NewService(repo, nil)
	svc.SetCrypto(aesCrypto)

	got, err := svc.Create(ctx, "p1", "u1", CreateRuleRequest{
		Name:         "slack",
		EventType:    EventBuildFailed,
		ChannelType:  ChannelSlack,
		SlackToken:   "xoxb-secret",
		SlackChannel: "#builds",
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, got.ID, created.ID)
	assert.NotEqual(t, "xoxb-secret", created.SlackToken)
	decrypted, err := aesCrypto.Decrypt(created.SlackToken)
	require.NoError(t, err)
	assert.Equal(t, "xoxb-secret", decrypted)
	assert.Equal(t, "#builds", created.SlackChannel)
}

func TestService_SendWebhookDispatchesSlackRule(t *testing.T) {
	ctx := context.Background()
	aesCrypto, err := crypto.NewAESCrypto([]byte("01234567890123456789012345678901"))
	require.NoError(t, err)
	encryptedToken, err := aesCrypto.Encrypt("xoxb-secret")
	require.NoError(t, err)
	repo := &mockNotificationRepo{
		listByProjectEvent: func(ctx context.Context, projectID string, event EventType) ([]*NotificationRule, error) {
			assert.Equal(t, "p1", projectID)
			assert.Equal(t, EventBuildSuccess, event)
			return []*NotificationRule{{
				ID:           "r1",
				ProjectID:    "p1",
				Name:         "slack",
				EventType:    EventBuildSuccess,
				ChannelType:  ChannelSlack,
				SlackToken:   encryptedToken,
				SlackChannel: "#builds",
				Enabled:      true,
			}}, nil
		},
	}
	sender := &captureSlackSender{}
	svc := NewService(repo, nil)
	svc.SetCrypto(aesCrypto)
	svc.SetSlackSender(sender)
	svc.SetSlackBaseURL("https://zcid.example")

	err = svc.SendWebhook(ctx, "p1", EventBuildSuccess, map[string]any{
		"pipelineId":   "pipe1",
		"pipelineName": "Build",
		"runId":        "run1",
		"status":       "succeeded",
		"branch":       "main",
		"commitSha":    "abcdef123456",
		"triggeredBy":  "u1",
	})
	require.NoError(t, err)
	require.Len(t, sender.calls, 1)
	assert.Equal(t, "xoxb-secret", sender.calls[0].token)
	assert.Equal(t, "#builds", sender.calls[0].channel)
	assert.Equal(t, "Build", sender.calls[0].event.PipelineName)
	assert.Equal(t, "https://zcid.example", sender.calls[0].event.BaseURL)
}

func TestService_TestDispatchesSlackRule(t *testing.T) {
	ctx := context.Background()
	aesCrypto, err := crypto.NewAESCrypto([]byte("01234567890123456789012345678901"))
	require.NoError(t, err)
	encryptedToken, err := aesCrypto.Encrypt("xoxb-test")
	require.NoError(t, err)
	repo := &mockNotificationRepo{
		findByID: func(ctx context.Context, id, projectID string) (*NotificationRule, error) {
			return &NotificationRule{
				ID:           id,
				ProjectID:    projectID,
				Name:         "slack",
				EventType:    EventDeployFailed,
				ChannelType:  ChannelSlack,
				SlackToken:   encryptedToken,
				SlackChannel: "#deploys",
			}, nil
		},
	}
	sender := &captureSlackSender{}
	svc := NewService(repo, nil)
	svc.SetCrypto(aesCrypto)
	svc.SetSlackSender(sender)

	require.NoError(t, svc.Test(ctx, "p1", "r1"))
	require.Len(t, sender.calls, 1)
	assert.Equal(t, "xoxb-test", sender.calls[0].token)
	assert.Equal(t, "#deploys", sender.calls[0].channel)
	assert.Equal(t, "sample-run", sender.calls[0].event.RunID)
	assert.Equal(t, string(EventDeployFailed), sender.calls[0].event.Status)
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

type captureSlackSender struct {
	calls []slackCall
}

type slackCall struct {
	token   string
	channel string
	event   BuildEvent
}

func (s *captureSlackSender) SendBuildNotification(ctx context.Context, botToken, channel string, event BuildEvent) error {
	s.calls = append(s.calls, slackCall{token: botToken, channel: channel, event: event})
	return nil
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
