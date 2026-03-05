package git

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/internal/pipelinerun"
	"github.com/xjy/zcid/pkg/cache"
	"github.com/xjy/zcid/pkg/gitprovider"
	"github.com/xjy/zcid/pkg/response"
)

// PipelineMatchService finds pipelines that match an incoming webhook event.
type PipelineMatchService interface {
	FindMatchingPipelines(ctx context.Context, repoURL, branch, eventType string) ([]MatchedPipeline, error)
}

// MatchedPipeline represents a pipeline that matched a webhook event.
type MatchedPipeline struct {
	PipelineID string
	ProjectID  string
}

// RunTrigger triggers a pipeline run (injected to avoid circular dep).
type RunTrigger interface {
	TriggerRun(ctx context.Context, projectID, pipelineID, userID string, req pipelinerun.TriggerRunRequest) (*pipelinerun.PipelineRunResponse, error)
}

type WebhookHandler struct {
	service         *Service
	idempotentCache *cache.RedisCache
	pipelineMatcher PipelineMatchService
	runTrigger      RunTrigger
}

func NewWebhookHandler(service *Service, idempotentCache *cache.RedisCache, pipelineMatcher PipelineMatchService, runTrigger RunTrigger) *WebhookHandler {
	return &WebhookHandler{
		service:         service,
		idempotentCache: idempotentCache,
		pipelineMatcher: pipelineMatcher,
		runTrigger:      runTrigger,
	}
}

func (h *WebhookHandler) RegisterRoutes(router gin.IRoutes) {
	router.POST("/gitlab", h.HandleGitLab)
	router.POST("/github", h.HandleGitHub)
}

func (h *WebhookHandler) HandleGitLab(c *gin.Context) {
	deliveryID := c.GetHeader("X-Gitlab-Event-UUID")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "无法读取请求体", "")
		return
	}

	gitlabToken := c.GetHeader("X-Gitlab-Token")

	event, err := parseGitLabEvent(body, c.GetHeader("X-Gitlab-Event"))
	if err != nil {
		slog.Warn("GitLab Webhook payload 解析失败", slog.Any("error", err))
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "Payload 解析失败", err.Error())
		return
	}

	if _, verifyErr := h.service.VerifyGitLabWebhook(gitlabToken); verifyErr != nil {
		slog.Warn("GitLab Webhook 签名验证失败", slog.String("repo", event.RepoURL))
		response.Error(c, http.StatusUnauthorized, response.CodeGitWebhookSigInvalid, "Webhook 签名验证失败", "")
		return
	}

	h.processWebhookEvent(c, event, deliveryID)
}

func (h *WebhookHandler) HandleGitHub(c *gin.Context) {
	deliveryID := c.GetHeader("X-GitHub-Delivery")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "无法读取请求体", "")
		return
	}

	signature := c.GetHeader("X-Hub-Signature-256")
	eventType := c.GetHeader("X-GitHub-Event")

	event, err := parseGitHubEvent(body, eventType)
	if err != nil {
		slog.Warn("GitHub Webhook payload 解析失败", slog.Any("error", err))
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "Payload 解析失败", err.Error())
		return
	}

	if _, verifyErr := h.service.VerifyGitHubWebhook(signature, body); verifyErr != nil {
		slog.Warn("GitHub Webhook 签名验证失败", slog.String("repo", event.RepoURL))
		response.Error(c, http.StatusUnauthorized, response.CodeGitWebhookSigInvalid, "Webhook 签名验证失败", "")
		return
	}

	h.processWebhookEvent(c, event, deliveryID)
}

func (h *WebhookHandler) processWebhookEvent(c *gin.Context, event *gitprovider.WebhookEvent, deliveryID string) {
	idempotencyKey := gitprovider.IdempotencyKeyWithDelivery(event, deliveryID)

	if h.idempotentCache != nil {
		ctx := c.Request.Context()
		if _, err := h.idempotentCache.Get(ctx, idempotencyKey); err == nil {
			slog.Info("Webhook 事件幂等去重", slog.String("key", idempotencyKey))
			response.Success(c, gin.H{"status": "duplicate", "message": "事件已处理"})
			return
		}
	}

	slog.Info("Webhook 事件接收成功",
		slog.String("provider", string(event.Provider)),
		slog.String("repo", event.RepoName),
		slog.String("branch", event.Branch),
		slog.String("commit", event.CommitSHA),
		slog.String("event", event.EventType),
	)

	ctx := c.Request.Context()
	if h.pipelineMatcher != nil && h.runTrigger != nil {
		matched, err := h.pipelineMatcher.FindMatchingPipelines(ctx, event.RepoURL, event.Branch, event.EventType)
		if err != nil {
			slog.Warn("流水线匹配失败", slog.Any("error", err))
		} else {
			for _, m := range matched {
				req := pipelinerun.TriggerRunRequest{
					Params:    map[string]string{"GIT_BRANCH": event.Branch, "GIT_COMMIT": event.CommitSHA},
					GitBranch: event.Branch,
					GitCommit: event.CommitSHA,
				}
				if _, runErr := h.runTrigger.TriggerRun(ctx, m.ProjectID, m.PipelineID, "webhook", req); runErr != nil {
					slog.Warn("Webhook 触发流水线失败",
						slog.String("pipeline", m.PipelineID),
						slog.String("project", m.ProjectID),
						slog.Any("error", runErr))
				} else {
					slog.Info("Webhook 已触发流水线", slog.String("pipeline", m.PipelineID), slog.String("project", m.ProjectID))
				}
			}
		}
	}

	if h.idempotentCache != nil {
		_ = h.idempotentCache.Set(c.Request.Context(), idempotencyKey, "1", 0)
	}

	response.Success(c, gin.H{
		"status":  "received",
		"event":   event.EventType,
		"repo":    event.RepoName,
		"branch":  event.Branch,
		"commit":  event.CommitSHA,
	})
}

func parseGitLabEvent(body []byte, eventHeader string) (*gitprovider.WebhookEvent, error) {
	var payload struct {
		Ref        string `json:"ref"`
		After      string `json:"after"`
		Repository struct {
			Name     string `json:"name"`
			URL      string `json:"url"`
			Homepage string `json:"homepage"`
		} `json:"repository"`
		Commits []struct {
			Message string `json:"message"`
			Author  struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		} `json:"commits"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	if payload.Repository.Name == "" || payload.After == "" {
		return nil, fmt.Errorf("missing required fields: repository.name or after(commit SHA)")
	}

	branch := payload.Ref
	if len(branch) > 11 && branch[:11] == "refs/heads/" {
		branch = branch[11:]
	}

	commitMsg := ""
	authorName := ""
	authorEmail := ""
	if len(payload.Commits) > 0 {
		commitMsg = payload.Commits[len(payload.Commits)-1].Message
		authorName = payload.Commits[len(payload.Commits)-1].Author.Name
		authorEmail = payload.Commits[len(payload.Commits)-1].Author.Email
	}

	return &gitprovider.WebhookEvent{
		Provider:    gitprovider.ProviderGitLab,
		EventType:   normalizeEventType(eventHeader),
		RepoURL:     payload.Repository.Homepage,
		RepoName:    payload.Repository.Name,
		Branch:      branch,
		CommitSHA:   payload.After,
		CommitMsg:   commitMsg,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
	}, nil
}

func parseGitHubEvent(body []byte, eventHeader string) (*gitprovider.WebhookEvent, error) {
	var payload struct {
		Ref        string `json:"ref"`
		After      string `json:"after"`
		Repository struct {
			FullName string `json:"full_name"`
			Name     string `json:"name"`
			HTMLURL  string `json:"html_url"`
			CloneURL string `json:"clone_url"`
		} `json:"repository"`
		HeadCommit struct {
			Message string `json:"message"`
			Author  struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		} `json:"head_commit"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	if payload.Repository.FullName == "" || payload.After == "" {
		return nil, fmt.Errorf("missing required fields: repository.full_name or after(commit SHA)")
	}

	branch := payload.Ref
	if len(branch) > 11 && branch[:11] == "refs/heads/" {
		branch = branch[11:]
	}

	return &gitprovider.WebhookEvent{
		Provider:    gitprovider.ProviderGitHub,
		EventType:   normalizeEventType(eventHeader),
		RepoURL:     payload.Repository.HTMLURL,
		RepoName:    payload.Repository.FullName,
		Branch:      branch,
		CommitSHA:   payload.After,
		CommitMsg:   payload.HeadCommit.Message,
		AuthorName:  payload.HeadCommit.Author.Name,
		AuthorEmail: payload.HeadCommit.Author.Email,
	}, nil
}

func normalizeEventType(header string) string {
	switch header {
	case "Push Hook", "push":
		return "push"
	case "Merge Request Hook", "pull_request":
		return "merge_request"
	case "Tag Push Hook", "create":
		return "tag_push"
	default:
		return header
	}
}
