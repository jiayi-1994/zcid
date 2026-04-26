package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type BuildEvent struct {
	ProjectID    string
	ProjectName  string
	PipelineID   string
	PipelineName string
	RunID        string
	Status       string
	Branch       string
	CommitSHA    string
	Duration     string
	TriggeredBy  string
	BaseURL      string
}

type SlackSender interface {
	SendBuildNotification(ctx context.Context, botToken, channel string, event BuildEvent) error
}

type SlackHTTPSender struct{ httpClient *http.Client }

func NewSlackSender() *SlackHTTPSender {
	return &SlackHTTPSender{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (s *SlackHTTPSender) SendBuildNotification(ctx context.Context, botToken, channel string, event BuildEvent) error {
	payload := buildSlackPayload(channel, event)
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal slack payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://slack.com/api/chat.postMessage", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create slack request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+botToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack send: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode slack response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("slack API error: %s", result.Error)
	}
	return nil
}

func buildSlackPayload(channel string, event BuildEvent) map[string]any {
	status := strings.TrimSpace(event.Status)
	if status == "" {
		status = "unknown"
	}
	text := fmt.Sprintf("[zcid] %s: %s — %s", fallback(event.PipelineName, "Pipeline"), status, fallback(event.ProjectName, event.ProjectID))
	commit := event.CommitSHA
	if len(commit) > 8 {
		commit = commit[:8]
	}
	viewURL := strings.TrimRight(event.BaseURL, "/")
	if viewURL != "" && event.ProjectID != "" && event.PipelineID != "" && event.RunID != "" {
		viewURL = fmt.Sprintf("%s/projects/%s/pipelines/%s/runs/%s", viewURL, event.ProjectID, event.PipelineID, event.RunID)
	} else {
		viewURL = ""
	}
	fields := []map[string]any{
		mrkdwnField("Project", fallback(event.ProjectName, event.ProjectID)),
		mrkdwnField("Pipeline", event.PipelineName),
		mrkdwnField("Branch", fallback(event.Branch, "-")),
		mrkdwnField("Commit", fallback(commit, "-")),
		mrkdwnField("Duration", fallback(event.Duration, "-")),
		mrkdwnField("Triggered by", fallback(event.TriggeredBy, "-")),
	}
	blocks := []map[string]any{
		{"type": "header", "text": map[string]any{"type": "plain_text", "text": fmt.Sprintf("[zcid] Pipeline Run — %s %s", statusEmoji(status), status), "emoji": true}},
		{"type": "section", "fields": fields},
	}
	if viewURL != "" {
		blocks = append(blocks, map[string]any{"type": "actions", "elements": []map[string]any{{"type": "button", "text": map[string]any{"type": "plain_text", "text": "View Run", "emoji": true}, "url": viewURL}}})
	}
	return map[string]any{"channel": channel, "text": text, "attachments": []map[string]any{{"color": slackColor(status), "blocks": blocks}}}
}

func mrkdwnField(label string, value string) map[string]any {
	return map[string]any{"type": "mrkdwn", "text": fmt.Sprintf("*%s*\n%s", label, fallback(value, "-"))}
}

func fallback(value string, alt string) string {
	if strings.TrimSpace(value) == "" {
		return alt
	}
	return value
}

func statusEmoji(status string) string {
	switch strings.ToLower(status) {
	case "succeeded", "success", "build_success", "deploy_success":
		return "✅"
	case "failed", "failure", "build_failed", "deploy_failed":
		return "❌"
	case "cancelled", "canceled":
		return "⚠️"
	default:
		return "ℹ️"
	}
}

func slackColor(status string) string {
	switch strings.ToLower(status) {
	case "succeeded", "success", "build_success", "deploy_success":
		return "#22c55e"
	case "failed", "failure", "build_failed", "deploy_failed":
		return "#ef4444"
	case "cancelled", "canceled":
		return "#f59e0b"
	default:
		return "#3b82f6"
	}
}
