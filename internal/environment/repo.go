package environment

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound       = errors.New("environment not found")
	ErrNameTaken      = errors.New("environment name already exists in this project")
	ErrNamespaceTaken = errors.New("namespace already in use")
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, e *Environment) error {
	if strings.TrimSpace(e.ID) == "" {
		e.ID = uuid.NewString()
	}

	err := r.db.WithContext(ctx).Create(e).Error
	if isUniqueConstraintError(err) {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "uk_environments_namespace") {
			return ErrNamespaceTaken
		}
		return ErrNameTaken
	}
	if err != nil {
		return fmt.Errorf("create environment: %w", err)
	}
	return nil
}

func (r *Repo) FindByID(ctx context.Context, id, projectID string) (*Environment, error) {
	var e Environment
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND status != ?", id, projectID, StatusDeleted).
		First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find environment: %w", err)
	}
	return &e, nil
}

func (r *Repo) ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*Environment, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&Environment{}).
		Where("project_id = ? AND status != ?", projectID, StatusDeleted)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count environments: %w", err)
	}

	var envs []*Environment
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&envs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list environments: %w", err)
	}
	return envs, total, nil
}

func (r *Repo) Update(ctx context.Context, id, projectID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	res := r.db.WithContext(ctx).Model(&Environment{}).
		Where("id = ? AND project_id = ? AND status != ?", id, projectID, StatusDeleted).
		Updates(updates)
	if isUniqueConstraintError(res.Error) {
		msg := strings.ToLower(res.Error.Error())
		if strings.Contains(msg, "uk_environments_namespace") {
			return ErrNamespaceTaken
		}
		return ErrNameTaken
	}
	if res.Error != nil {
		return fmt.Errorf("update environment: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) SoftDelete(ctx context.Context, id, projectID string) error {
	res := r.db.WithContext(ctx).Model(&Environment{}).
		Where("id = ? AND project_id = ? AND status != ?", id, projectID, StatusDeleted).
		Update("status", StatusDeleted)
	if res.Error != nil {
		return fmt.Errorf("soft delete environment: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
