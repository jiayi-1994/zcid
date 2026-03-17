package deployment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xjy/zcid/internal/environment"
	"github.com/xjy/zcid/pkg/argocd"
	"github.com/xjy/zcid/pkg/response"
)

// EnvironmentGetter gets environment by ID and project.
type EnvironmentGetter interface {
	Get(ctx context.Context, id, projectID string) (*environment.Environment, error)
}

type Service struct {
	repo       Repository
	envGetter  EnvironmentGetter
	argoClient argocd.ArgoClient
}

func NewService(repo Repository, envGetter EnvironmentGetter, argoClient argocd.ArgoClient) *Service {
	return &Service{
		repo:       repo,
		envGetter:  envGetter,
		argoClient: argoClient,
	}
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

	if err := s.argoClient.CreateOrUpdateApp(ctx, app); err != nil {
		errMsg := "创建/更新 ArgoCD 应用失败: " + err.Error()
		_ = s.repo.Update(ctx, d.ID, projectID, map[string]any{"status": string(StatusFailed), "error_message": errMsg})
		return nil, response.NewBizError(response.CodeDeployFailed, "创建/更新 ArgoCD 应用失败", err.Error())
	}

	if err := s.argoClient.SyncApp(ctx, appName); err != nil {
		errMsg := "触发 ArgoCD 同步失败: " + err.Error()
		_ = s.repo.Update(ctx, d.ID, projectID, map[string]any{"status": string(StatusFailed), "error_message": errMsg})
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
			_ = s.repo.Update(context.Background(), deployID, projectID, map[string]any{
				"status": string(StatusFailed), "error_message": "部署超时",
			})
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
				updates["finished_at"] = time.Now()
				_ = s.repo.Update(context.Background(), deployID, projectID, updates)
				return
			case "Degraded":
				updates["status"] = string(StatusDegraded)
				updates["finished_at"] = time.Now()
				_ = s.repo.Update(context.Background(), deployID, projectID, updates)
				return
			case "Progressing":
				updates["status"] = string(StatusSyncing)
				_ = s.repo.Update(context.Background(), deployID, projectID, updates)
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
	case "Progressing":
		updates["status"] = string(StatusSyncing)
	case "Suspended", "Missing", "Unknown":
		updates["status"] = string(StatusFailed)
		updates["error_message"] = "ArgoCD health: " + status.Health
	}
	if updateErr := s.repo.Update(ctx, deployID, projectID, updates); updateErr != nil {
		return d, response.NewBizError(response.CodeInternalServerError, "更新部署状态失败", updateErr.Error())
	}
	return s.repo.FindByID(ctx, deployID, projectID)
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
