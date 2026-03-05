package deployment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("deployment not found")
)

type Repository interface {
	Create(ctx context.Context, d *Deployment) error
	FindByID(ctx context.Context, id, projectID string) (*Deployment, error)
	ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*Deployment, int64, error)
	ListByEnvironment(ctx context.Context, projectID, envID string, page, pageSize int) ([]*Deployment, int64, error)
	Update(ctx context.Context, id, projectID string, updates map[string]any) error
}

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, d *Deployment) error {
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *Repo) FindByID(ctx context.Context, id, projectID string) (*Deployment, error) {
	var d Deployment
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find deployment: %w", err)
	}
	return &d, nil
}

func (r *Repo) ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*Deployment, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&Deployment{}).Where("project_id = ?", projectID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count deployments: %w", err)
	}
	var list []*Deployment
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list deployments: %w", err)
	}
	return list, total, nil
}

func (r *Repo) ListByEnvironment(ctx context.Context, projectID, envID string, page, pageSize int) ([]*Deployment, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&Deployment{}).
		Where("project_id = ? AND environment_id = ?", projectID, envID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count deployments: %w", err)
	}
	var list []*Deployment
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list deployments: %w", err)
	}
	return list, total, nil
}

func (r *Repo) Update(ctx context.Context, id, projectID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&Deployment{}).
		Where("id = ? AND project_id = ?", id, projectID).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("update deployment: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
