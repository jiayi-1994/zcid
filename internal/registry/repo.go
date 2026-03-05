package registry

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("registry not found")
	ErrNameDuplicate = errors.New("registry name already exists")
)

// Repository defines the registry data access interface
type Repository interface {
	Create(r *Registry) error
	GetByID(id string) (*Registry, error)
	List() ([]Registry, int64, error)
	Update(id string, updates map[string]interface{}) error
	SoftDelete(id string) error
	GetDefault() (*Registry, error)
	SetDefault(id string) error
}

// Repo implements Repository with GORM
type Repo struct {
	db *gorm.DB
}

// NewRepo creates a new Repo
func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

// Create inserts a new registry
func (r *Repo) Create(reg *Registry) error {
	if reg.ID == "" {
		reg.ID = uuid.New().String()
	}
	if err := r.db.Create(reg).Error; err != nil {
		if isUniqueConstraintError(err) {
			return ErrNameDuplicate
		}
		return err
	}
	return nil
}

// GetByID retrieves a registry by ID, excluding deleted
func (r *Repo) GetByID(id string) (*Registry, error) {
	var reg Registry
	err := r.db.Where("id = ? AND status != ?", id, StatusDeleted).First(&reg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &reg, nil
}

// List returns all non-deleted registries with total count
func (r *Repo) List() ([]Registry, int64, error) {
	var regs []Registry
	var total int64

	query := r.db.Where("status != ?", StatusDeleted)
	if err := query.Model(&Registry{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("is_default DESC, created_at DESC").Find(&regs).Error; err != nil {
		return nil, 0, err
	}
	return regs, total, nil
}

// Update updates a registry by ID
func (r *Repo) Update(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	result := r.db.Model(&Registry{}).Where("id = ? AND status != ?", id, StatusDeleted).Updates(updates)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return ErrNameDuplicate
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// SoftDelete marks a registry as deleted
func (r *Repo) SoftDelete(id string) error {
	result := r.db.Model(&Registry{}).Where("id = ? AND status != ?", id, StatusDeleted).
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

// GetDefault returns the default registry, or nil if none
func (r *Repo) GetDefault() (*Registry, error) {
	var reg Registry
	err := r.db.Where("is_default = ? AND status != ?", true, StatusDeleted).First(&reg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &reg, nil
}

// SetDefault sets the given registry as default and unsets others
func (r *Repo) SetDefault(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Unset all defaults
		if err := tx.Model(&Registry{}).Where("status != ?", StatusDeleted).Update("is_default", false).Error; err != nil {
			return err
		}
		// Set the new default
		result := tx.Model(&Registry{}).Where("id = ? AND status != ?", id, StatusDeleted).Update("is_default", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func isUniqueConstraintError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}
