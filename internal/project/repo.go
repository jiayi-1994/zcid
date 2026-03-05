package project

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrProjectNotFound  = errors.New("project not found")
	ErrProjectNameTaken = errors.New("project name already exists")
	ErrMemberExists     = errors.New("user is already a project member")
	ErrMemberNotFound   = errors.New("project member not found")
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, p *Project) error {
	if strings.TrimSpace(p.ID) == "" {
		p.ID = uuid.NewString()
	}

	err := r.db.WithContext(ctx).Create(p).Error
	if isUniqueConstraintError(err) {
		return ErrProjectNameTaken
	}
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	return nil
}

func (r *Repo) FindByID(ctx context.Context, id string) (*Project, error) {
	var p Project
	err := r.db.WithContext(ctx).
		Where("id = ? AND status != ?", id, ProjectStatusDeleted).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find project by id: %w", err)
	}
	return &p, nil
}

func (r *Repo) List(ctx context.Context, page, pageSize int) ([]*Project, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&Project{}).Where("status != ?", ProjectStatusDeleted)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}

	var projects []*Project
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&projects).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}

	return projects, total, nil
}

func (r *Repo) ListByIDs(ctx context.Context, ids []string, page, pageSize int) ([]*Project, int64, error) {
	if len(ids) == 0 {
		return []*Project{}, 0, nil
	}

	var total int64
	query := r.db.WithContext(ctx).Model(&Project{}).
		Where("id IN ? AND status != ?", ids, ProjectStatusDeleted)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count projects by ids: %w", err)
	}

	var projects []*Project
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&projects).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list projects by ids: %w", err)
	}

	return projects, total, nil
}

func (r *Repo) Update(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	res := r.db.WithContext(ctx).Model(&Project{}).
		Where("id = ? AND status != ?", id, ProjectStatusDeleted).
		Updates(updates)
	if isUniqueConstraintError(res.Error) {
		return ErrProjectNameTaken
	}
	if res.Error != nil {
		return fmt.Errorf("update project: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrProjectNotFound
	}

	return nil
}

func (r *Repo) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Model(&Project{}).
		Where("id = ? AND status != ?", id, ProjectStatusDeleted).
		Update("status", ProjectStatusDeleted)
	if res.Error != nil {
		return fmt.Errorf("soft delete project: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrProjectNotFound
	}

	return nil
}

func (r *Repo) AddMember(ctx context.Context, member *ProjectMember) error {
	if strings.TrimSpace(member.ID) == "" {
		member.ID = uuid.NewString()
	}

	err := r.db.WithContext(ctx).Create(member).Error
	if isUniqueConstraintError(err) {
		return ErrMemberExists
	}
	if err != nil {
		return fmt.Errorf("add project member: %w", err)
	}
	return nil
}

func (r *Repo) RemoveMembersByProject(ctx context.Context, projectID string) error {
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Delete(&ProjectMember{}).Error
	if err != nil {
		return fmt.Errorf("remove project members: %w", err)
	}
	return nil
}

func (r *Repo) GetUserProjectIDs(ctx context.Context, userID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).
		Model(&ProjectMember{}).
		Where("user_id = ?", userID).
		Pluck("project_id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("get user project ids: %w", err)
	}
	return ids, nil
}

func (r *Repo) IsProjectMember(ctx context.Context, projectID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check project membership: %w", err)
	}
	return count > 0, nil
}

func (r *Repo) GetMemberRole(ctx context.Context, projectID, userID string) (ProjectRole, error) {
	var member ProjectMember
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get member role: %w", err)
	}
	return member.Role, nil
}

func (r *Repo) ListMembers(ctx context.Context, projectID string) ([]MemberWithUsername, error) {
	var results []MemberWithUsername
	err := r.db.WithContext(ctx).
		Table("project_members pm").
		Select("pm.user_id as user_id, u.username as username, pm.role as role, pm.created_at as created_at").
		Joins("JOIN users u ON u.id = pm.user_id").
		Where("pm.project_id = ?", projectID).
		Order("pm.created_at ASC").
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("list project members: %w", err)
	}
	return results, nil
}

func (r *Repo) RemoveMember(ctx context.Context, projectID, userID string) error {
	res := r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&ProjectMember{})
	if res.Error != nil {
		return fmt.Errorf("remove project member: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (r *Repo) UpdateMemberRole(ctx context.Context, projectID, userID string, role ProjectRole) error {
	res := r.db.WithContext(ctx).
		Model(&ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Update("role", role)
	if res.Error != nil {
		return fmt.Errorf("update member role: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (r *Repo) SoftDeleteEnvironmentsByProject(ctx context.Context, projectID string) error {
	return r.db.WithContext(ctx).
		Table("environments").
		Where("project_id = ? AND status != ?", projectID, "deleted").
		Update("status", "deleted").Error
}

func (r *Repo) SoftDeleteServicesByProject(ctx context.Context, projectID string) error {
	return r.db.WithContext(ctx).
		Table("services").
		Where("project_id = ? AND status != ?", projectID, "deleted").
		Update("status", "deleted").Error
}

func (r *Repo) SoftDeleteVariablesByProject(ctx context.Context, projectID string) error {
	return r.db.WithContext(ctx).
		Table("variables").
		Where("project_id = ? AND status != ?", projectID, "deleted").
		Update("status", "deleted").Error
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
