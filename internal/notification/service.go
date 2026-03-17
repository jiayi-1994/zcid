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
	"github.com/xjy/zcid/pkg/response"
)

type IdempotencyChecker interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

type Service struct {
	repo      Repository
	idemCache IdempotencyChecker
}

func NewService(repo Repository, idemCache IdempotencyChecker) *Service {
	return &Service{repo: repo, idemCache: idemCache}
}

var validEventTypes = map[EventType]bool{
	EventBuildSuccess: true, EventBuildFailed: true,
	EventDeploySuccess: true, EventDeployFailed: true,
}

func (s *Service) Create(ctx context.Context, projectID, userID string, req CreateRuleRequest) (*NotificationRule, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.WebhookURL = strings.TrimSpace(req.WebhookURL)
	if req.Name == "" || req.WebhookURL == "" {
		return nil, response.NewBizError(response.CodeValidation, "name and webhookUrl are required", "")
	}
	if !validEventTypes[req.EventType] {
		return nil, response.NewBizError(response.CodeValidation, "eventType must be build_success, build_failed, deploy_success, or deploy_failed", "")
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	rule := &NotificationRule{
		ID:         uuid.NewString(),
		ProjectID:  projectID,
		Name:       req.Name,
		EventType:  req.EventType,
		WebhookURL: req.WebhookURL,
		Enabled:    enabled,
		CreatedBy:  userID,
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
	_, err := s.repo.FindByID(ctx, ruleID, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeNotifRuleNotFound, "notification rule not found", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "failed to update notification rule", err.Error())
	}
	updates := make(map[string]any)
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.EventType != nil {
		updates["event_type"] = string(*req.EventType)
	}
	if req.WebhookURL != nil {
		updates["webhook_url"] = strings.TrimSpace(*req.WebhookURL)
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
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
	body, _ := json.Marshal(payload)
	idemKey := buildIdempotencyKey(projectID, event, payload)
	for _, rule := range rules {
		if s.idemCache != nil {
			if _, err := s.idemCache.Get(ctx, idemKey+":"+rule.ID); err == nil {
				continue
			}
		}
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
		if s.idemCache != nil {
			_ = s.idemCache.Set(ctx, idemKey+":"+rule.ID, "1", 5*time.Minute)
		}
	}
	return nil
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
