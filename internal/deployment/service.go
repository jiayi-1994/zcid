package deployment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xjy/zcid/internal/environment"
	"github.com/xjy/zcid/internal/notification"
	"github.com/xjy/zcid/internal/signal"
	"github.com/xjy/zcid/pkg/argocd"
	"github.com/xjy/zcid/pkg/response"
)

// EnvironmentGetter gets environment by ID and project.
type EnvironmentGetter interface {
	Get(ctx context.Context, id, projectID string) (*environment.Environment, error)
}

type NotificationDispatcher interface {
	SendWebhook(ctx context.Context, projectID string, event notification.EventType, payload map[string]any) error
}

type Service struct {
	repo          Repository
	envGetter     EnvironmentGetter
	argoClient    argocd.ArgoClient
	signals       *signal.Service
	notifications NotificationDispatcher
}

func NewService(repo Repository, envGetter EnvironmentGetter, argoClient argocd.ArgoClient) *Service {
	return &Service{
		repo:       repo,
		envGetter:  envGetter,
		argoClient: argoClient,
	}
}

func (s *Service) SetSignalService(signals *signal.Service) {
	s.signals = signals
}

func (s *Service) SetNotificationService(service *notification.Service) {
	s.notifications = service
}

func (s *Service) SetNotificationDispatcher(dispatcher NotificationDispatcher) {
	s.notifications = dispatcher
}

func (s *Service) TriggerDeploy(ctx context.Context, projectID, userID string, req TriggerDeployRequest) (*Deployment, error) {
	req.EnvironmentID = strings.TrimSpace(req.EnvironmentID)
	req.Image = strings.TrimSpace(req.Image)
	if req.EnvironmentID == "" || req.Image == "" {
		return nil, response.NewBizError(response.CodeValidation, "environmentId 和 image 必填", "")
	}

	env, err := s.envGetter.Get(ctx, req.EnvironmentID, projectID)
	if err != nil {
		if errors.Is(err, environment.ErrNotFound) {
			return nil, response.NewBizError(response.CodeNotFound, "环境不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "查询环境失败", err.Error())
	}

	appName := fmt.Sprintf("zcid-%s-%s", projectID, env.Name)
	if len(appName) > 200 {
		appName = appName[:200]
	}

	app := &argocd.ArgoApp{
		Name:           appName,
		Project:        "default",
		RepoURL:        "https://github.com/example/zcid-manifests",
		Path:           "overlays/" + env.Name,
		TargetRevision: "HEAD",
		Namespace:      env.Namespace,
		Image:          req.Image,
	}

	now := time.Now()
	pendingSync := "Pending"
	d := &Deployment{
		ID:            uuid.NewString(),
		ProjectID:     projectID,
		EnvironmentID: env.ID,
		Image:         req.Image,
		Status:        StatusPending,
		ArgoAppName:   &appName,
		SyncStatus:    &pendingSync,
		DeployedBy:    userID,
		StartedAt:     &now,
	}
	if req.PipelineRunID != nil {
		d.PipelineRunID = req.PipelineRunID
	}

	if err := s.repo.Create(ctx, d); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "保存部署记录失败", err.Error())
	}
	s.recordDeploymentSignals(ctx, d, "deployment.created", "Deployment queued")

	if err := s.argoClient.CreateOrUpdateApp(ctx, app); err != nil {
		errMsg := "创建/更新 ArgoCD 应用失败: " + err.Error()
		_ = s.repo.Update(ctx, d.ID, projectID, map[string]any{"status": string(StatusFailed), "error_message": errMsg})
		d.Status = StatusFailed
		d.ErrorMessage = &errMsg
		s.recordDeploymentSignals(ctx, d, "deployment.argocd_error", errMsg)
		s.notifyDeployment(ctx, d, notification.EventDeployFailed)
		return nil, response.NewBizError(response.CodeDeployFailed, "创建/更新 ArgoCD 应用失败", err.Error())
	}

	if err := s.argoClient.SyncApp(ctx, appName); err != nil {
		errMsg := "触发 ArgoCD 同步失败: " + err.Error()
		_ = s.repo.Update(ctx, d.ID, projectID, map[string]any{"status": string(StatusFailed), "error_message": errMsg})
		d.Status = StatusFailed
		d.ErrorMessage = &errMsg
		s.recordDeploymentSignals(ctx, d, "deployment.sync_error", errMsg)
		s.notifyDeployment(ctx, d, notification.EventDeployFailed)
		return nil, response.NewBizError(response.CodeDeploySyncFailed, "触发 ArgoCD 同步失败", err.Error())
	}

	syncStatus := "Syncing"
	healthStatus := "Progressing"
	_ = s.repo.Update(ctx, d.ID, projectID, map[string]any{
		"status":        string(StatusSyncing),
		"sync_status":   syncStatus,
		"health_status": healthStatus,
	})
	d.Status = StatusSyncing
	d.SyncStatus = &syncStatus
	d.HealthStatus = &healthStatus
	s.recordDeploymentSignals(ctx, d, "deployment.syncing", "Deployment sync started")

	go s.syncDeployStatus(d.ID, projectID)

	return d, nil
}

func (s *Service) syncDeployStatus(deployID, projectID string) {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()
	timeout := time.After(10 * time.Minute)

	for {
		select {
		case <-timeout:
			d, _ := s.repo.FindByID(context.Background(), deployID, projectID)
			now := time.Now()
			_ = s.repo.Update(context.Background(), deployID, projectID, map[string]any{
				"status": string(StatusFailed), "error_message": "部署超时", "finished_at": now, "updated_at": now,
			})
			if d != nil {
				d.Status = StatusFailed
				d.ErrorMessage = ptrString("部署超时")
				d.FinishedAt = &now
				s.recordDeploymentSignals(context.Background(), d, "deployment.timeout", "Deployment timed out")
				s.notifyDeployment(context.Background(), d, notification.EventDeployFailed)
			}
			return
		case <-ticker.C:
			d, err := s.repo.FindByID(context.Background(), deployID, projectID)
			if err != nil || d.ArgoAppName == nil {
				return
			}
			status, err := s.argoClient.GetAppStatus(context.Background(), *d.ArgoAppName)
			if err != nil {
				continue
			}
			updates := map[string]any{
				"sync_status":   status.Sync,
				"health_status": status.Health,
				"updated_at":    time.Now(),
			}
			switch status.Health {
			case "Healthy":
				updates["status"] = string(StatusHealthy)
				now := time.Now()
				updates["finished_at"] = now
				_ = s.repo.Update(context.Background(), deployID, projectID, updates)
				d.Status = StatusHealthy
				d.SyncStatus = &status.Sync
				d.HealthStatus = &status.Health
				d.FinishedAt = &now
				s.recordDeploymentSignals(context.Background(), d, "deployment.healthy", "Deployment is healthy")
				s.notifyDeployment(context.Background(), d, notification.EventDeploySuccess)
				return
			case "Degraded":
				updates["status"] = string(StatusDegraded)
				now := time.Now()
				updates["finished_at"] = now
				_ = s.repo.Update(context.Background(), deployID, projectID, updates)
				d.Status = StatusDegraded
				d.SyncStatus = &status.Sync
				d.HealthStatus = &status.Health
				d.FinishedAt = &now
				s.recordDeploymentSignals(context.Background(), d, "deployment.degraded", "Deployment is degraded")
				s.notifyDeployment(context.Background(), d, notification.EventDeployFailed)
				return
			case "Progressing":
				updates["status"] = string(StatusSyncing)
				_ = s.repo.Update(context.Background(), deployID, projectID, updates)
				d.Status = StatusSyncing
				d.SyncStatus = &status.Sync
				d.HealthStatus = &status.Health
				s.recordDeploymentSignals(context.Background(), d, "deployment.progressing", "Deployment is still progressing")
			}
		}
	}
}

func (s *Service) GetDeployStatus(ctx context.Context, projectID, deployID string) (*Deployment, error) {
	d, err := s.repo.FindByID(ctx, deployID, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeDeployNotFound, "部署不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "查询部署失败", err.Error())
	}
	if d.ArgoAppName == nil || *d.ArgoAppName == "" {
		return d, nil
	}
	previousStatus := d.Status
	status, err := s.argoClient.GetAppStatus(ctx, *d.ArgoAppName)
	if err != nil {
		return d, response.NewBizError(response.CodeDeploySyncFailed, "获取 ArgoCD 状态失败", err.Error())
	}
	updates := map[string]any{
		"sync_status":   status.Sync,
		"health_status": status.Health,
	}
	switch status.Health {
	case "Healthy":
		updates["status"] = string(StatusHealthy)
		now := time.Now()
		updates["finished_at"] = now
	case "Degraded":
		updates["status"] = string(StatusDegraded)
		now := time.Now()
		updates["finished_at"] = now
	case "Progressing":
		updates["status"] = string(StatusSyncing)
	case "Suspended", "Missing", "Unknown":
		updates["status"] = string(StatusFailed)
		updates["error_message"] = "ArgoCD health: " + status.Health
		now := time.Now()
		updates["finished_at"] = now
	}
	if updateErr := s.repo.Update(ctx, deployID, projectID, updates); updateErr != nil {
		return d, response.NewBizError(response.CodeInternalServerError, "更新部署状态失败", updateErr.Error())
	}
	updated, err := s.repo.FindByID(ctx, deployID, projectID)
	if err == nil {
		s.recordDeploymentSignals(ctx, updated, "deployment.refresh", "Deployment status refreshed")
		if shouldNotifyDeploymentTransition(previousStatus, updated.Status) {
			s.notifyDeployment(ctx, updated, deploymentEventForStatus(updated.Status))
		}
	}
	return updated, err
}

func (s *Service) GetDeployment(ctx context.Context, projectID, deployID string) (*Deployment, error) {
	return s.GetDeployStatus(ctx, projectID, deployID)
}

func (s *Service) ListDeployments(ctx context.Context, projectID string, page, pageSize int) ([]*Deployment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListByProject(ctx, projectID, page, pageSize)
}

// RefreshStatus fetches fresh status from ArgoCD and updates the deployment record.
func (s *Service) RefreshStatus(ctx context.Context, projectID, deployID string) (*Deployment, error) {
	return s.GetDeployStatus(ctx, projectID, deployID)
}

// ResyncDeploy triggers ArgoCD re-sync for the deployment.
func (s *Service) ResyncDeploy(ctx context.Context, projectID, deployID string) (*Deployment, error) {
	d, err := s.repo.FindByID(ctx, deployID, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeDeployNotFound, "部署不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "查询部署失败", err.Error())
	}
	if d.ArgoAppName == nil || *d.ArgoAppName == "" {
		return nil, response.NewBizError(response.CodeDeploySyncFailed, "无 ArgoCD 应用名，无法同步", "")
	}
	if err := s.argoClient.SyncApp(ctx, *d.ArgoAppName); err != nil {
		return nil, response.NewBizError(response.CodeDeploySyncFailed, "触发 ArgoCD 同步失败", err.Error())
	}
	syncStatus := "Syncing"
	_ = s.repo.Update(ctx, deployID, projectID, map[string]any{"sync_status": syncStatus})
	return s.repo.FindByID(ctx, deployID, projectID)
}

// RollbackDeploy rolls back to the previous successful deployment's image.
func (s *Service) RollbackDeploy(ctx context.Context, projectID, deployID, userID string) (*Deployment, error) {
	d, err := s.repo.FindByID(ctx, deployID, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeDeployNotFound, "部署不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "查询部署失败", err.Error())
	}
	prevList, _, err := s.repo.ListByEnvironment(ctx, projectID, d.EnvironmentID, 1, 50)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "查询部署历史失败", err.Error())
	}
	var prevImage string
	for _, p := range prevList {
		if p.ID != d.ID && (p.Status == StatusHealthy || p.Status == StatusDegraded) && p.Image != "" {
			prevImage = p.Image
			break
		}
	}
	if prevImage == "" {
		return nil, response.NewBizError(response.CodeDeployRollbackErr, "无可用于回滚的先前部署", "")
	}
	return s.TriggerDeploy(ctx, projectID, userID, TriggerDeployRequest{
		EnvironmentID: d.EnvironmentID,
		Image:         prevImage,
	})
}

func (s *Service) notifyDeployment(ctx context.Context, d *Deployment, event notification.EventType) {
	if s.notifications == nil || d == nil {
		return
	}
	payload := map[string]any{
		"deploymentId":  d.ID,
		"environmentId": d.EnvironmentID,
		"pipelineRunId": stringValue(d.PipelineRunID),
		"pipelineName":  "Deploy to " + d.EnvironmentID,
		"runId":         d.ID,
		"status":        string(d.Status),
		"syncStatus":    stringValue(d.SyncStatus),
		"healthStatus":  stringValue(d.HealthStatus),
		"image":         d.Image,
		"triggeredBy":   d.DeployedBy,
		"errorMessage":  stringValue(d.ErrorMessage),
	}
	if d.StartedAt != nil && d.FinishedAt != nil && d.FinishedAt.After(*d.StartedAt) {
		payload["duration"] = d.FinishedAt.Sub(*d.StartedAt).Round(time.Second).String()
	}
	if err := s.notifications.SendWebhook(ctx, d.ProjectID, event, payload); err != nil {
		slog.Warn("failed to send deployment notification", slog.Any("error", err), slog.String("deploymentID", d.ID), slog.String("event", string(event)))
	}
}

func shouldNotifyDeploymentTransition(previous, current DeployStatus) bool {
	if previous == current {
		return false
	}
	return current == StatusHealthy || current == StatusDegraded || current == StatusFailed
}

func deploymentEventForStatus(status DeployStatus) notification.EventType {
	if status == StatusHealthy {
		return notification.EventDeploySuccess
	}
	return notification.EventDeployFailed
}

func (s *Service) recordDeploymentSignals(ctx context.Context, d *Deployment, reason, message string) {
	if s.signals == nil || d == nil {
		return
	}
	status, severity := deploymentSignalStatus(d.Status)
	staleAfter := time.Now().Add(30 * time.Minute)
	value := map[string]any{
		"deploymentStatus": string(d.Status),
		"syncStatus":       stringValue(d.SyncStatus),
		"healthStatus":     stringValue(d.HealthStatus),
		"image":            d.Image,
		"errorMessage":     stringValue(d.ErrorMessage),
	}
	targets := []struct {
		targetType signal.TargetType
		targetID   string
	}{
		{signal.TargetDeployment, d.ID},
		{signal.TargetEnvironment, d.EnvironmentID},
	}
	for _, target := range targets {
		if _, err := s.signals.Record(ctx, signal.RecordInput{
			ProjectID:     d.ProjectID,
			TargetType:    target.targetType,
			TargetID:      target.targetID,
			Source:        "deployment",
			Status:        status,
			Severity:      severity,
			Reason:        reason,
			Message:       message,
			ObservedValue: value,
			StaleAfter:    &staleAfter,
		}); err != nil {
			slog.Warn("failed to record deployment health signal", slog.Any("error", err), slog.String("deploymentID", d.ID), slog.String("targetID", target.targetID))
		}
	}
}

func deploymentSignalStatus(status DeployStatus) (signal.Status, signal.Severity) {
	switch status {
	case StatusHealthy:
		return signal.StatusHealthy, signal.SeverityInfo
	case StatusFailed, StatusDegraded:
		return signal.StatusDegraded, signal.SeverityCritical
	case StatusPending, StatusSyncing:
		return signal.StatusWarning, signal.SeverityWarning
	default:
		return signal.StatusUnknown, signal.SeverityInfo
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func ptrString(value string) *string {
	return &value
}

// GetDeployHistory returns paginated deployment history for an environment.
func (s *Service) GetDeployHistory(ctx context.Context, projectID, envID string, page, pageSize int) ([]*Deployment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListByEnvironment(ctx, projectID, envID, page, pageSize)
}
