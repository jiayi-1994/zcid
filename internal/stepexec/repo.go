package stepexec

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNotFound = errors.New("step execution not found")

type Repository interface {
	Upsert(ctx context.Context, row *StepExecution) error
	ListByPipelineRun(ctx context.Context, runID string) ([]StepExecution, error)
	DeleteExpired(ctx context.Context, cutoff time.Time, batchSize int, maxBatches int) (int, bool, error)
	FinalizeRun(ctx context.Context, runID, terminalStatus string) error
}

type Repo struct{ db *gorm.DB }

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Upsert(ctx context.Context, row *StepExecution) error {
	if row == nil {
		return fmt.Errorf("upsert step execution: nil row")
	}
	if row.ID == "" {
		row.ID = uuid.NewString()
	}
	if row.Status == "" {
		row.Status = StatusPending
	}
	row.NormalizeJSON()
	ApplySizeLimits(row)
	now := time.Now()
	if row.CreatedAt.IsZero() {
		row.CreatedAt = now
	}
	row.UpdatedAt = now

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "pipeline_run_id"}, {Name: "task_run_name"}, {Name: "step_name"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"step_index", "status", "image_ref", "image_digest", "command_args", "env_public",
			"secret_refs", "params_resolved", "workspace_mounts", "resources", "tekton_results",
			"output_digests", "log_ref", "trace_id", "started_at", "finished_at", "duration_ms",
			"exit_code", "updated_at",
		}),
		Where: clause.Where{Exprs: []clause.Expression{clause.Expr{
			SQL:  "excluded.status NOT IN ? OR step_executions.status IN ?",
			Vars: []interface{}{[]string{StatusRunning, StatusPending}, []string{StatusRunning, StatusPending}},
		}}},
	}).Create(row).Error
	if err != nil {
		return fmt.Errorf("upsert step execution: %w", err)
	}
	return nil
}

func (r *Repo) ListByPipelineRun(ctx context.Context, runID string) ([]StepExecution, error) {
	var rows []StepExecution
	err := r.db.WithContext(ctx).
		Where("pipeline_run_id = ?", runID).
		Order("task_run_name ASC, step_index ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list step executions: %w", err)
	}
	return rows, nil
}

func (r *Repo) DeleteExpired(ctx context.Context, cutoff time.Time, batchSize int, maxBatches int) (int, bool, error) {
	if batchSize <= 0 {
		batchSize = 1000
	}
	if maxBatches <= 0 {
		maxBatches = 100
	}
	orphanCutoff := cutoff.Add(-30 * 24 * time.Hour)
	total := 0
	for i := 0; i < maxBatches; i++ {
		res := r.db.WithContext(ctx).Exec(`
DELETE FROM step_executions
WHERE id IN (
  SELECT id FROM step_executions
  WHERE finished_at < ? OR (finished_at IS NULL AND started_at < ?)
  ORDER BY COALESCE(finished_at, started_at) ASC
  LIMIT ?
)`, cutoff, orphanCutoff, batchSize)
		if res.Error != nil {
			return total, false, fmt.Errorf("delete expired step executions: %w", res.Error)
		}
		deleted := int(res.RowsAffected)
		total += deleted
		if deleted < batchSize {
			return total, false, nil
		}
	}
	return total, true, nil
}

func (r *Repo) FinalizeRun(ctx context.Context, runID, terminalStatus string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&StepExecution{}).
		Where("pipeline_run_id = ? AND status IN ?", runID, []string{StatusPending, StatusRunning}).
		Updates(map[string]interface{}{
			"status":      StatusInterrupted,
			"finished_at": now,
			"updated_at":  now,
		})
	if res.Error != nil {
		return fmt.Errorf("finalize step executions for run %s (%s): %w", runID, terminalStatus, res.Error)
	}
	return nil
}
