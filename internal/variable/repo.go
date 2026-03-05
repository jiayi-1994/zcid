package variable

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrNotFound     = errors.New("variable not found")
	ErrKeyDuplicate = errors.New("variable key already exists in this scope")
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(v *Variable) error {
	if err := r.db.Create(v).Error; err != nil {
		if isUniqueConstraintError(err) {
			return ErrKeyDuplicate
		}
		return err
	}
	return nil
}

func (r *Repo) GetByID(id string) (*Variable, error) {
	var v Variable
	err := r.db.Where("id = ? AND status != ?", id, StatusDeleted).First(&v).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *Repo) ListByProject(projectID string) ([]Variable, int64, error) {
	var vars []Variable
	var total int64

	query := r.db.Where("project_id = ? AND scope = ? AND status != ?", projectID, ScopeProject, StatusDeleted)
	if err := query.Model(&Variable{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("key ASC").Find(&vars).Error; err != nil {
		return nil, 0, err
	}
	return vars, total, nil
}

func (r *Repo) ListGlobal() ([]Variable, int64, error) {
	var vars []Variable
	var total int64

	query := r.db.Where("scope = ? AND status != ?", ScopeGlobal, StatusDeleted)
	if err := query.Model(&Variable{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("key ASC").Find(&vars).Error; err != nil {
		return nil, 0, err
	}
	return vars, total, nil
}

func (r *Repo) Update(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	result := r.db.Model(&Variable{}).Where("id = ? AND status != ?", id, StatusDeleted).Updates(updates)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return ErrKeyDuplicate
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) SoftDelete(id string) error {
	result := r.db.Model(&Variable{}).Where("id = ? AND status != ?", id, StatusDeleted).
		Updates(map[string]interface{}{
			"status":     StatusDeleted,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) ListGlobalAndProject(projectID string) ([]Variable, error) {
	var vars []Variable
	err := r.db.Where(
		"((scope = ? AND status != ?) OR (scope = ? AND project_id = ? AND status != ?))",
		ScopeGlobal, StatusDeleted, ScopeProject, projectID, StatusDeleted,
	).Order("scope ASC, key ASC").Find(&vars).Error
	return vars, err
}

func (r *Repo) ListByPipelineScope(projectID, pipelineID string) ([]Variable, error) {
	var vars []Variable
	err := r.db.Where(
		"scope = ? AND project_id = ? AND pipeline_id = ? AND status != ?",
		ScopePipeline, projectID, pipelineID, StatusDeleted,
	).Order("key ASC").Find(&vars).Error
	return vars, err
}

func isUniqueConstraintError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}
