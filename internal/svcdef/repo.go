package svcdef

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xjy/zcid/internal/pipeline"
	"github.com/xjy/zcid/internal/signal"
	"github.com/xjy/zcid/pkg/database"
	"gorm.io/gorm"
)

var (
	ErrNotFound  = errors.New("service not found")
	ErrNameTaken = errors.New("service name already exists in this project")
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, s *ServiceDef) error {
	if strings.TrimSpace(s.ID) == "" {
		s.ID = uuid.NewString()
	}

	err := r.db.WithContext(ctx).Create(s).Error
	if database.IsUniqueConstraintError(err) {
		return ErrNameTaken
	}
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	return nil
}

func (r *Repo) FindByID(ctx context.Context, id, projectID string) (*ServiceDef, error) {
	var s ServiceDef
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND status != ?", id, projectID, StatusDeleted).
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find service: %w", err)
	}
	return &s, nil
}

func (r *Repo) ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*ServiceDef, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&ServiceDef{}).
		Where("project_id = ? AND status != ?", projectID, StatusDeleted)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count services: %w", err)
	}

	var services []*ServiceDef
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&services).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list services: %w", err)
	}
	return services, total, nil
}

func (r *Repo) Update(ctx context.Context, id, projectID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	res := r.db.WithContext(ctx).Model(&ServiceDef{}).
		Where("id = ? AND project_id = ? AND status != ?", id, projectID, StatusDeleted).
		Updates(updates)
	if database.IsUniqueConstraintError(res.Error) {
		return ErrNameTaken
	}
	if res.Error != nil {
		return fmt.Errorf("update service: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) SoftDelete(ctx context.Context, id, projectID string) error {
	res := r.db.WithContext(ctx).Model(&ServiceDef{}).
		Where("id = ? AND project_id = ? AND status != ?", id, projectID, StatusDeleted).
		Update("status", StatusDeleted)
	if res.Error != nil {
		return fmt.Errorf("soft delete service: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) ListLinkedPipelines(ctx context.Context, svc *ServiceDef) ([]VitalsPipeline, error) {
	if svc == nil {
		return []VitalsPipeline{}, nil
	}

	var rows []pipeline.Pipeline
	query := r.db.WithContext(ctx).
		Where("project_id = ? AND status != ?", svc.ProjectID, pipeline.StatusDeleted).
		Order("created_at DESC")
	if len(svc.PipelineIDs) > 0 {
		query = query.Where("id IN ?", []string(svc.PipelineIDs))
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list linked pipelines: %w", err)
	}

	out := make([]VitalsPipeline, 0, len(rows))
	for _, row := range rows {
		repoURL := pipelineRepoURL(row.Config.Metadata)
		if len(svc.PipelineIDs) == 0 && strings.TrimSpace(svc.RepoURL) != "" && !sameRepoURL(repoURL, svc.RepoURL) {
			continue
		}
		out = append(out, VitalsPipeline{
			ID:        row.ID,
			Name:      row.Name,
			Status:    string(row.Status),
			RepoURL:   repoURL,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func pipelineRepoURL(metadata map[string]string) string {
	for _, key := range []string{"repoUrl", "repo_url", "repository", "gitUrl", "git_url"} {
		if value := strings.TrimSpace(metadata[key]); value != "" {
			return value
		}
	}
	return ""
}

func sameRepoURL(a, b string) bool {
	normalize := func(v string) string {
		return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(v)), ".git")
	}
	return normalize(a) != "" && normalize(a) == normalize(b)
}

func (r *Repo) ListRecentRuns(ctx context.Context, projectID string, pipelineIDs []string, limit int) ([]VitalsRun, error) {
	if len(pipelineIDs) == 0 {
		return []VitalsRun{}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	var rows []VitalsRun
	err := r.db.WithContext(ctx).Raw(`
		SELECT pr.id, pr.pipeline_id, p.name AS pipeline_name, pr.run_number, pr.status,
		       pr.started_at, pr.finished_at, pr.error_message, pr.created_at, pr.updated_at
		FROM pipeline_runs pr
		JOIN pipelines p ON p.id = pr.pipeline_id
		WHERE pr.project_id = ? AND pr.pipeline_id IN ?
		ORDER BY pr.created_at DESC
		LIMIT ?`, projectID, pipelineIDs, limit).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list recent service runs: %w", err)
	}
	return rows, nil
}

func (r *Repo) ListLatestDeployments(ctx context.Context, projectID string, environmentIDs []string, limit int) ([]VitalsDeployment, error) {
	if len(environmentIDs) == 0 {
		return []VitalsDeployment{}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	var rows []VitalsDeployment
	err := r.db.WithContext(ctx).Raw(`
		SELECT d.id, d.environment_id, e.name AS environment_name, d.pipeline_run_id,
		       d.image, d.status, d.sync_status, d.health_status, d.error_message,
		       d.started_at, d.finished_at, d.created_at, d.updated_at
		FROM deployments d
		JOIN environments e ON e.id = d.environment_id
		WHERE d.project_id = ? AND d.environment_id IN ?
		ORDER BY d.created_at DESC
		LIMIT ?`, projectID, environmentIDs, limit).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list latest service deployments: %w", err)
	}
	return rows, nil
}

func (r *Repo) ListFailedSteps(ctx context.Context, projectID string, runIDs []string, limit int) ([]VitalsStepWarning, error) {
	if len(runIDs) == 0 {
		return []VitalsStepWarning{}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	var rows []VitalsStepWarning
	err := r.db.WithContext(ctx).Raw(`
		SELECT se.step_name, se.task_run_name, se.status, se.duration_ms, se.exit_code,
		       se.started_at, se.finished_at, pr.id AS run_id, pr.pipeline_id,
		       p.name AS pipeline_name, pr.run_number, se.created_at
		FROM step_executions se
		JOIN pipeline_runs pr ON pr.id = se.pipeline_run_id
		JOIN pipelines p ON p.id = pr.pipeline_id
		WHERE pr.project_id = ? AND pr.id IN ? AND se.status IN ('failed', 'cancelled', 'interrupted')
		ORDER BY se.created_at DESC
		LIMIT ?`, projectID, runIDs, limit).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list failed service steps: %w", err)
	}
	return rows, nil
}

func (r *Repo) ListLatestSignals(ctx context.Context, projectID string, targets []VitalsSignalTarget, limit int) ([]VitalsSignal, error) {
	if len(targets) == 0 {
		return []VitalsSignal{}, nil
	}
	if limit <= 0 {
		limit = 20
	}
	out := make([]VitalsSignal, 0, len(targets))
	for _, target := range targets {
		var rows []signal.HealthSignal
		err := r.db.WithContext(ctx).
			Where("project_id = ? AND target_type = ? AND target_id = ?", projectID, target.TargetType, target.TargetID).
			Order("observed_at DESC, created_at DESC").
			Limit(limit).
			Find(&rows).Error
		if err != nil {
			return nil, fmt.Errorf("list service signals: %w", err)
		}
		for _, row := range rows {
			out = append(out, VitalsSignal{
				ID:            row.ID,
				TargetType:    string(row.TargetType),
				TargetID:      row.TargetID,
				Source:        row.Source,
				Status:        string(row.EffectiveStatus(time.Now())),
				RawStatus:     string(row.Status),
				Severity:      string(row.Severity),
				Reason:        row.Reason,
				Message:       row.Message,
				ObservedValue: row.ObservedValue,
				ObservedAt:    row.ObservedAt,
				StaleAfter:    row.StaleAfter,
				CreatedAt:     row.CreatedAt,
			})
		}
	}
	return out, nil
}
