package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/xjy/zcid/pkg/database"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("pipeline not found")
	ErrNameDuplicate = errors.New("pipeline name already exists in this project")
)

type Repository interface {
	Create(ctx context.Context, p *Pipeline) error
	GetByIDAndProject(ctx context.Context, id, projectID string) (*Pipeline, error)
	List(ctx context.Context, projectID string, page, pageSize int) ([]*Pipeline, int64, error)
	ListByTriggerType(ctx context.Context, triggerType TriggerType) ([]*Pipeline, error)
	Update(ctx context.Context, id, projectID string, updates map[string]any) error
	SoftDelete(ctx context.Context, id, projectID string) error
	ExistsByNameAndProject(ctx context.Context, projectID, name string, excludeID string) (bool, error)
}

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, p *Pipeline) error {
	if strings.TrimSpace(p.ID) == "" {
		p.ID = uuid.NewString()
	}

	err := r.db.WithContext(ctx).Create(p).Error
	if database.IsUniqueConstraintError(err) {
		return ErrNameDuplicate
	}
	if err != nil {
		return fmt.Errorf("create pipeline: %w", err)
	}
	return nil
}

func (r *Repo) GetByIDAndProject(ctx context.Context, id, projectID string) (*Pipeline, error) {
	var p Pipeline
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND status != ?", id, projectID, StatusDeleted).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get pipeline by id and project: %w", err)
	}
	return &p, nil
}

func (r *Repo) List(ctx context.Context, projectID string, page, pageSize int) ([]*Pipeline, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&Pipeline{}).
		Where("project_id = ? AND status != ?", projectID, StatusDeleted)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count pipelines: %w", err)
	}

	var pipelines []*Pipeline
	offset := (page - 1) * pageSize
	err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&pipelines).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list pipelines: %w", err)
	}

	return pipelines, total, nil
}

func (r *Repo) ListByTriggerType(ctx context.Context, triggerType TriggerType) ([]*Pipeline, error) {
	var pipelines []*Pipeline
	err := r.db.WithContext(ctx).Model(&Pipeline{}).
		Where("trigger_type = ? AND status != ?", triggerType, StatusDeleted).
		Find(&pipelines).Error
	if err != nil {
		return nil, fmt.Errorf("list pipelines by trigger type: %w", err)
	}
	return pipelines, nil
}

func (r *Repo) Update(ctx context.Context, id, projectID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	res := r.db.WithContext(ctx).Model(&Pipeline{}).
		Where("id = ? AND project_id = ? AND status != ?", id, projectID, StatusDeleted).
		Updates(updates)
	if database.IsUniqueConstraintError(res.Error) {
		return ErrNameDuplicate
	}
	if res.Error != nil {
		return fmt.Errorf("update pipeline: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) SoftDelete(ctx context.Context, id, projectID string) error {
	res := r.db.WithContext(ctx).Model(&Pipeline{}).
		Where("id = ? AND project_id = ? AND status != ?", id, projectID, StatusDeleted).
		Update("status", StatusDeleted)
	if res.Error != nil {
		return fmt.Errorf("soft delete pipeline: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) ExistsByNameAndProject(ctx context.Context, projectID, name string, excludeID string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&Pipeline{}).
		Where("project_id = ? AND name = ? AND status != ?", projectID, name, StatusDeleted)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("check pipeline name: %w", err)
	}
	return count > 0, nil
}
