package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, opts ListOpts) ([]*AuditLog, int64, error)
}

type ListOpts struct {
	UserID     *string
	Action     *string
	ResourceType *string
	ResourceID   *string
	StartTime  *time.Time
	EndTime    *time.Time
	Page       int
	PageSize   int
}

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, log *AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.NewString()
	}
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *Repo) List(ctx context.Context, opts ListOpts) ([]*AuditLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&AuditLog{})
	if opts.UserID != nil && *opts.UserID != "" {
		query = query.Where("user_id = ?", *opts.UserID)
	}
	if opts.Action != nil && *opts.Action != "" {
		query = query.Where("action = ?", *opts.Action)
	}
	if opts.ResourceType != nil && *opts.ResourceType != "" {
		query = query.Where("resource_type = ?", *opts.ResourceType)
	}
	if opts.ResourceID != nil && *opts.ResourceID != "" {
		query = query.Where("resource_id = ?", *opts.ResourceID)
	}
	if opts.StartTime != nil {
		query = query.Where("created_at >= ?", *opts.StartTime)
	}
	if opts.EndTime != nil {
		query = query.Where("created_at <= ?", *opts.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var list []*AuditLog
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	return list, total, nil
}
