package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xjy/zcid/pkg/crypto"
	"github.com/xjy/zcid/pkg/response"
)

type IdempotencyChecker interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

type Service struct {
	repo      Repository
	idemCache IdempotencyChecker
	crypto    *crypto.AESCrypto
	slack     SlackSender
	baseURL   string
}

func NewService(repo Repository, idemCache IdempotencyChecker) *Service {
	return &Service{repo: repo, idemCache: idemCache, slack: NewSlackSender()}
}

func (s *Service) SetCrypto(aesCrypto *crypto.AESCrypto) { s.crypto = aesCrypto }
func (s *Service) SetSlackBaseURL(baseURL string) {
	s.baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
}
func (s *Service) SetSlackSender(sender SlackSender) { s.slack = sender }

var validEventTypes = map[EventType]bool{
	EventBuildSuccess: true, EventBuildFailed: true,
	EventDeploySuccess: true, EventDeployFailed: true,
}

func (s *Service) Create(ctx context.Context, projectID, userID string, req CreateRuleRequest) (*NotificationRule, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.WebhookURL = strings.TrimSpace(req.WebhookURL)
	req.SlackToken = strings.TrimSpace(req.SlackToken)
	req.SlackChannel = strings.TrimSpace(req.SlackChannel)
	req.ChannelType = normalizeChannelType(req.ChannelType)
	if req.Name == "" {
		return nil, response.NewBizError(response.CodeValidation, "name is required", "")
	}
	if !validEventTypes[req.EventType] {
		return nil, response.NewBizError(response.CodeValidation, "eventType must be build_success, build_failed, deploy_success, or deploy_failed", "")
	}
	if err := s.validateChannel(req.ChannelType, req.WebhookURL, req.SlackToken, req.SlackChannel, true); err != nil {
		return nil, err
	}
	slackToken := ""
	if req.ChannelType == ChannelSlack {
		encrypted, err := s.encryptSlackToken(req.SlackToken)
		if err != nil {
			return nil, err
		}
		slackToken = encrypted
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	rule := &NotificationRule{
		ID:           uuid.NewString(),
		ProjectID:    projectID,
		Name:         req.Name,
		EventType:    req.EventType,
		ChannelType:  req.ChannelType,
		WebhookURL:   req.WebhookURL,
		SlackToken:   slackToken,
		SlackChannel: req.SlackChannel,
		Enabled:      enabled,
		CreatedBy:    userID,
	}
	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "failed to create notification rule", err.Error())
	}
	return rule, nil
}

func (s *Service) Get(ctx context.Context, projectID, ruleID string) (*NotificationRule, error) {
	rule, err := s.repo.FindByID(ctx, ruleID, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeNotifRuleNotFound, "notification rule not found", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "failed to get notification rule", err.Error())
	}
	return rule, nil
}

func (s *Service) List(ctx context.Context, projectID string, page, pageSize int) ([]*NotificationRule, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListByProject(ctx, projectID, page, pageSize)
}

func (s *Service) Update(ctx context.Context, projectID, ruleID string, req UpdateRuleRequest) (*NotificationRule, error) {
	existing, err := s.repo.FindByID(ctx, ruleID, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeNotifRuleNotFound, "notification rule not found", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "failed to update notification rule", err.Error())
	}
	updates := make(map[string]any)
	nextChannel := normalizeChannelType(existing.ChannelType)
	webhookURL := existing.WebhookURL
	slackToken := existing.SlackToken
	slackChannel := existing.SlackChannel
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.EventType != nil {
		if !validEventTypes[*req.EventType] {
			return nil, response.NewBizError(response.CodeValidation, "invalid eventType", "")
		}
		updates["event_type"] = string(*req.EventType)
	}
	if req.ChannelType != nil {
		nextChannel = normalizeChannelType(*req.ChannelType)
		updates["channel_type"] = string(nextChannel)
	}
	if req.WebhookURL != nil {
		webhookURL = strings.TrimSpace(*req.WebhookURL)
		updates["webhook_url"] = webhookURL
	}
	if req.SlackToken != nil && strings.TrimSpace(*req.SlackToken) != "" {
		encrypted, encErr := s.encryptSlackToken(strings.TrimSpace(*req.SlackToken))
		if encErr != nil {
			return nil, encErr
		}
		slackToken = encrypted
		updates["slack_token"] = encrypted
	}
	if req.SlackChannel != nil {
		slackChannel = strings.TrimSpace(*req.SlackChannel)
		updates["slack_channel"] = slackChannel
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if err := s.validateChannel(nextChannel, webhookURL, slackToken, slackChannel, false); err != nil {
		return nil, err
	}
	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.repo.Update(ctx, ruleID, projectID, updates); err != nil {
			return nil, response.NewBizError(response.CodeInternalServerError, "failed to update notification rule", err.Error())
		}
	}
	return s.repo.FindByID(ctx, ruleID, projectID)
}

func (s *Service) Delete(ctx context.Context, projectID, ruleID string) error {
	if err := s.repo.Delete(ctx, ruleID, projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodeNotifRuleNotFound, "notification rule not found", "")
		}
		return response.NewBizError(response.CodeInternalServerError, "failed to delete notification rule", err.Error())
	}
	return nil
}

func (s *Service) SendWebhook(ctx context.Context, projectID string, event EventType, payload map[string]any) error {
	rules, err := s.repo.ListByProjectAndEvent(ctx, projectID, event)
	if err != nil {
		return response.NewBizError(response.CodeInternalServerError, "failed to list notification rules", err.Error())
	}
	payload["eventType"] = string(event)
	payload["projectId"] = projectID
	idemKey := buildIdempotencyKey(projectID, event, payload)
	for _, rule := range rules {
		if s.idemCache != nil {
			if _, err := s.idemCache.Get(ctx, idemKey+":"+rule.ID); err == nil {
				continue
			}
		}
		switch normalizeChannelType(rule.ChannelType) {
		case ChannelSlack:
			if err := s.sendSlack(ctx, rule, payload); err != nil {
				return err
			}
		default:
			body, _ := json.Marshal(payload)
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, rule.WebhookURL, bytes.NewReader(body))
			if err != nil {
				return response.NewBizError(response.CodeNotifSendFailed, "failed to create webhook request", err.Error())
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Idempotency-Key", idemKey)
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return response.NewBizError(response.CodeNotifSendFailed, "failed to send webhook", err.Error())
			}
			resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return response.NewBizError(response.CodeNotifSendFailed, "webhook returned non-2xx status", "")
			}
		}
		if s.idemCache != nil {
			_ = s.idemCache.Set(ctx, idemKey+":"+rule.ID, "1", 5*time.Minute)
		}
	}
	return nil
}

func (s *Service) Test(ctx context.Context, projectID, ruleID string) error {
	rule, err := s.repo.FindByID(ctx, ruleID, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodeNotifRuleNotFound, "notification rule not found", "")
		}
		return response.NewBizError(response.CodeInternalServerError, "failed to get notification rule", err.Error())
	}
	payload := map[string]any{
		"eventType":    string(rule.EventType),
		"projectId":    projectID,
		"projectName":  "zcid project",
		"pipelineName": "Sample pipeline",
		"runId":        "sample-run",
		"status":       string(rule.EventType),
		"branch":       "main",
		"commitSha":    "12345678",
		"duration":     "1m 20s",
		"triggeredBy":  "notification-test",
	}
	if normalizeChannelType(rule.ChannelType) == ChannelSlack {
		return s.sendSlack(ctx, rule, payload)
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rule.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return response.NewBizError(response.CodeNotifSendFailed, "failed to create webhook request", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return response.NewBizError(response.CodeNotifSendFailed, "failed to send webhook", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return response.NewBizError(response.CodeNotifSendFailed, "webhook returned non-2xx status", "")
	}
	return nil
}

func (s *Service) sendSlack(ctx context.Context, rule *NotificationRule, payload map[string]any) error {
	if s.crypto == nil {
		return response.NewBizError(response.CodeDecryptFailed, "encryption service not configured", "")
	}
	if s.slack == nil {
		return response.NewBizError(response.CodeNotifSendFailed, "slack sender not configured", "")
	}
	token, err := s.crypto.Decrypt(rule.SlackToken)
	if err != nil {
		return response.NewBizError(response.CodeDecryptFailed, "failed to decrypt slack token", err.Error())
	}
	event := BuildEvent{
		ProjectID:    stringValue(payload, "projectId"),
		ProjectName:  stringValue(payload, "projectName"),
		PipelineID:   stringValue(payload, "pipelineId"),
		PipelineName: stringValue(payload, "pipelineName"),
		RunID:        stringValue(payload, "runId"),
		Status:       stringValue(payload, "status"),
		Branch:       stringValue(payload, "branch"),
		CommitSHA:    firstNonEmpty(stringValue(payload, "commitSha"), stringValue(payload, "gitCommit")),
		Duration:     stringValue(payload, "duration"),
		TriggeredBy:  stringValue(payload, "triggeredBy"),
		BaseURL:      s.baseURL,
	}
	if event.ProjectID == "" {
		event.ProjectID = rule.ProjectID
	}
	if event.Status == "" {
		event.Status = string(rule.EventType)
	}
	if err := s.slack.SendBuildNotification(ctx, token, rule.SlackChannel, event); err != nil {
		return response.NewBizError(response.CodeNotifSendFailed, "failed to send slack message", err.Error())
	}
	return nil
}

func (s *Service) encryptSlackToken(token string) (string, error) {
	if s.crypto == nil {
		return "", response.NewBizError(response.CodeDecryptFailed, "encryption service not configured", "encryption key not set")
	}
	encrypted, err := s.crypto.Encrypt(token)
	if err != nil {
		return "", response.NewBizError(response.CodeDecryptFailed, "failed to encrypt slack token", err.Error())
	}
	return encrypted, nil
}

func (s *Service) validateChannel(channel ChannelType, webhookURL, slackToken, slackChannel string, creating bool) error {
	switch normalizeChannelType(channel) {
	case ChannelWebhook:
		if strings.TrimSpace(webhookURL) == "" {
			return response.NewBizError(response.CodeValidation, "webhookUrl is required for webhook notifications", "")
		}
	case ChannelSlack:
		if creating && strings.TrimSpace(slackToken) == "" {
			return response.NewBizError(response.CodeValidation, "slackToken is required for slack notifications", "")
		}
		if strings.TrimSpace(slackToken) == "" || strings.TrimSpace(slackChannel) == "" {
			return response.NewBizError(response.CodeValidation, "slack token and channel are required for slack notifications", "")
		}
	default:
		return response.NewBizError(response.CodeValidation, "invalid channelType", "")
	}
	return nil
}

func normalizeChannelType(channel ChannelType) ChannelType {
	if channel == "" {
		return ChannelWebhook
	}
	return channel
}

func stringValue(payload map[string]any, key string) string {
	value, _ := payload[key].(string)
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func buildIdempotencyKey(projectID string, event EventType, payload map[string]any) string {
	if id, ok := payload["id"].(string); ok && id != "" {
		return projectID + ":" + string(event) + ":" + id
	}
	if runID, ok := payload["runId"].(string); ok && runID != "" {
		return projectID + ":" + string(event) + ":" + runID
	}
	return projectID + ":" + string(event) + ":" + uuid.NewString()
}
