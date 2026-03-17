package svcdef

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
