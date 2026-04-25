package pipelinerun

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("pipeline run not found")
)

type Repository interface {
	Create(ctx context.Context, r *PipelineRun) error
	GetByIDAndProject(ctx context.Context, id, projectID string) (*PipelineRun, error)
	GetByIDProjectPipeline(ctx context.Context, id, projectID, pipelineID string) (*PipelineRun, error)
	GetNextRunNumber(ctx context.Context, pipelineID string) (int, error)
	ListByPipeline(ctx context.Context, pipelineID, projectID string, page, pageSize int) ([]*PipelineRun, int64, error)
	ListRunning(ctx context.Context, pipelineID string) ([]*PipelineRun, error)
	Update(ctx context.Context, id, projectID string, updates map[string]interface{}) error
	UpdateStatus(ctx context.Context, id, projectID string, status RunStatus, errorMsg *string) error
	CountRunning(ctx context.Context, pipelineID string) (int64, error)
	UpdateArtifacts(ctx context.Context, id, projectID string, artifacts ArtifactSlice) error
}

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, run *PipelineRun) error {
	if run.ID == "" {
		run.ID = uuid.New().String()
	}
	err := r.db.WithContext(ctx).Create(run).Error
	if err != nil {
		return fmt.Errorf("create pipeline run: %w", err)
	}
	return nil
}

func (r *Repo) GetByIDAndProject(ctx context.Context, id, projectID string) (*PipelineRun, error) {
	var run PipelineRun
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get pipeline run: %w", err)
	}
	return &run, nil
}

func (r *Repo) GetByIDProjectPipeline(ctx context.Context, id, projectID, pipelineID string) (*PipelineRun, error) {
	var run PipelineRun
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND pipeline_id = ?", id, projectID, pipelineID).
		First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get pipeline run by pipeline: %w", err)
	}
	return &run, nil
}

func (r *Repo) GetNextRunNumber(ctx context.Context, pipelineID string) (int, error) {
	var maxNum int
	err := r.db.WithContext(ctx).Model(&PipelineRun{}).
		Select("COALESCE(MAX(run_number), 0)").
		Where("pipeline_id = ?", pipelineID).
		Scan(&maxNum).Error
	if err != nil {
		return 0, fmt.Errorf("get next run number: %w", err)
	}
	return maxNum + 1, nil
}

func (r *Repo) ListByPipeline(ctx context.Context, pipelineID, projectID string, page, pageSize int) ([]*PipelineRun, int64, error) {
	base := r.db.WithContext(ctx).Model(&PipelineRun{}).
		Where("pipeline_id = ? AND project_id = ?", pipelineID, projectID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count pipeline runs: %w", err)
	}

	var runs []*PipelineRun
	offset := (page - 1) * pageSize
	err := base.Order("run_number DESC").Offset(offset).Limit(pageSize).Find(&runs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list pipeline runs: %w", err)
	}
	return runs, total, nil
}

func (r *Repo) Update(ctx context.Context, id, projectID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&PipelineRun{}).
		Where("id = ? AND project_id = ?", id, projectID).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("update pipeline run: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) UpdateStatus(ctx context.Context, id, projectID string, status RunStatus, errorMsg *string) error {
	updates := map[string]interface{}{"status": string(status)}
	if errorMsg != nil {
		updates["error_message"] = *errorMsg
	}
	return r.Update(ctx, id, projectID, updates)
}

func (r *Repo) ListRunning(ctx context.Context, pipelineID string) ([]*PipelineRun, error) {
	var runs []*PipelineRun
	err := r.db.WithContext(ctx).
		Where("pipeline_id = ? AND status IN ?", pipelineID, []RunStatus{StatusPending, StatusQueued, StatusRunning}).
		Find(&runs).Error
	if err != nil {
		return nil, fmt.Errorf("list running: %w", err)
	}
	return runs, nil
}

func (r *Repo) CountRunning(ctx context.Context, pipelineID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&PipelineRun{}).
		Where("pipeline_id = ? AND status IN ?", pipelineID, []RunStatus{StatusPending, StatusQueued, StatusRunning}).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count running: %w", err)
	}
	return count, nil
}

func (r *Repo) UpdateArtifacts(ctx context.Context, id, projectID string, artifacts ArtifactSlice) error {
	return r.Update(ctx, id, projectID, map[string]interface{}{"artifacts": artifacts})
}
