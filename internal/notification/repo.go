package notification

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("notification rule not found")

type Repository interface {
	Create(ctx context.Context, r *NotificationRule) error
	FindByID(ctx context.Context, id, projectID string) (*NotificationRule, error)
	ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*NotificationRule, int64, error)
	ListByProjectAndEvent(ctx context.Context, projectID string, event EventType) ([]*NotificationRule, error)
	Update(ctx context.Context, id, projectID string, updates map[string]any) error
	Delete(ctx context.Context, id, projectID string) error
}

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, rule *NotificationRule) error {
	if rule.ID == "" {
		rule.ID = uuid.NewString()
	}
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *Repo) FindByID(ctx context.Context, id, projectID string) (*NotificationRule, error) {
	var rule NotificationRule
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find notification rule: %w", err)
	}
	return &rule, nil
}

func (r *Repo) ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*NotificationRule, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&NotificationRule{}).Where("project_id = ?", projectID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count notification rules: %w", err)
	}
	var list []*NotificationRule
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list notification rules: %w", err)
	}
	return list, total, nil
}

func (r *Repo) ListByProjectAndEvent(ctx context.Context, projectID string, event EventType) ([]*NotificationRule, error) {
	var list []*NotificationRule
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND event_type = ? AND enabled = ?", projectID, event, true).
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list notification rules by event: %w", err)
	}
	return list, nil
}

func (r *Repo) Update(ctx context.Context, id, projectID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&NotificationRule{}).
		Where("id = ? AND project_id = ?", id, projectID).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("update notification rule: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, id, projectID string) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		Delete(&NotificationRule{})
	if res.Error != nil {
		return fmt.Errorf("delete notification rule: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
