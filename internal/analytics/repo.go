package analytics

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Get(ctx context.Context, projectID string, since time.Time) (*Response, error)
}

type Repo struct{ db *gorm.DB }

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Get(ctx context.Context, projectID string, since time.Time) (*Response, error) {
	var summary Summary
	if err := r.db.WithContext(ctx).Raw(`
SELECT
  COUNT(*)::bigint AS total_runs,
  COALESCE(COUNT(*) FILTER (WHERE status = 'succeeded')::float / NULLIF(COUNT(*), 0), 0) AS success_rate,
  COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (finished_at - started_at)) * 1000) FILTER (WHERE started_at IS NOT NULL AND finished_at IS NOT NULL), 0)::bigint AS median_duration_ms,
  COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (finished_at - started_at)) * 1000) FILTER (WHERE started_at IS NOT NULL AND finished_at IS NOT NULL), 0)::bigint AS p95_duration_ms
FROM pipeline_runs
WHERE project_id = ? AND created_at >= ?`, projectID, since).Scan(&summary).Error; err != nil {
		return nil, fmt.Errorf("analytics summary: %w", err)
	}

	var daily []DailyStat
	if err := r.db.WithContext(ctx).Raw(`
SELECT
  TO_CHAR(DATE(created_at), 'YYYY-MM-DD') AS date,
  COUNT(*)::bigint AS total,
  COUNT(*) FILTER (WHERE status = 'succeeded')::bigint AS succeeded,
  COUNT(*) FILTER (WHERE status = 'failed')::bigint AS failed,
  COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (finished_at - started_at)) * 1000) FILTER (WHERE started_at IS NOT NULL AND finished_at IS NOT NULL), 0)::bigint AS median_duration_ms,
  COALESCE(COUNT(*) FILTER (WHERE status = 'succeeded')::float / NULLIF(COUNT(*), 0), 0) AS success_rate
FROM pipeline_runs
WHERE project_id = ? AND created_at >= ?
GROUP BY DATE(created_at)
ORDER BY DATE(created_at)`, projectID, since).Scan(&daily).Error; err != nil {
		return nil, fmt.Errorf("analytics daily stats: %w", err)
	}

	var steps []TopFailingStep
	if err := r.db.WithContext(ctx).Raw(`
SELECT
  se.step_name AS step_name,
  se.task_run_name AS task_run_name,
  COUNT(*) FILTER (WHERE se.status = 'failed')::bigint AS failure_count,
  COUNT(*)::bigint AS total_count,
  COALESCE(COUNT(*) FILTER (WHERE se.status = 'failed')::float / NULLIF(COUNT(*), 0), 0) AS failure_rate
FROM step_executions se
JOIN pipeline_runs pr ON pr.id = se.pipeline_run_id
WHERE pr.project_id = ? AND se.created_at >= ?
GROUP BY se.step_name, se.task_run_name
HAVING COUNT(*) FILTER (WHERE se.status = 'failed') > 0
ORDER BY failure_count DESC, total_count DESC
LIMIT 10`, projectID, since).Scan(&steps).Error; err != nil {
		return nil, fmt.Errorf("analytics top failing steps: %w", err)
	}

	var pipelines []TopPipeline
	if err := r.db.WithContext(ctx).Raw(`
SELECT
  p.id AS pipeline_id,
  p.name AS pipeline_name,
  COUNT(pr.id)::bigint AS run_count,
  COALESCE(COUNT(pr.id) FILTER (WHERE pr.status = 'succeeded')::float / NULLIF(COUNT(pr.id), 0), 0) AS success_rate
FROM pipelines p
JOIN pipeline_runs pr ON pr.pipeline_id = p.id
WHERE p.project_id = ? AND pr.created_at >= ? AND p.status != 'deleted'
GROUP BY p.id, p.name
ORDER BY run_count DESC, success_rate DESC
LIMIT 10`, projectID, since).Scan(&pipelines).Error; err != nil {
		return nil, fmt.Errorf("analytics top pipelines: %w", err)
	}

	return &Response{Summary: summary, DailyStats: daily, TopFailingSteps: steps, TopPipelines: pipelines}, nil
}
