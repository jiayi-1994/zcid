package signal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("health signal not found")

type Repository interface {
	Create(ctx context.Context, row *HealthSignal) error
	ListLatestByTarget(ctx context.Context, projectID string, targetType TargetType, targetID string, limit int) ([]HealthSignal, error)
}

type Repo struct{ db *gorm.DB }

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Create(ctx context.Context, row *HealthSignal) error {
	if row == nil {
		return fmt.Errorf("create health signal: nil row")
	}
	if row.ID == "" {
		row.ID = uuid.NewString()
	}
	row.Normalize(time.Now())
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return fmt.Errorf("create health signal: %w", err)
	}
	return nil
}

func (r *Repo) ListLatestByTarget(ctx context.Context, projectID string, targetType TargetType, targetID string, limit int) ([]HealthSignal, error) {
	if limit <= 0 {
		limit = 20
	}
	var rows []HealthSignal
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND target_type = ? AND target_id = ?", projectID, targetType, targetID).
		Order("observed_at DESC, created_at DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list health signals: %w", err)
	}
	return rows, nil
}
