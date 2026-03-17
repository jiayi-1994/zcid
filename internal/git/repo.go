package git

import (
	"errors"
	"time"

	"github.com/xjy/zcid/pkg/database"
	"gorm.io/gorm"
)

var (
	ErrNotFound     = errors.New("git connection not found")
	ErrNameDuplicate = errors.New("git connection name already exists")
)

type Repository interface {
	Create(conn *GitConnection) error
	GetByID(id string) (*GitConnection, error)
	List() ([]GitConnection, int64, error)
	ListByProviderType(providerType string) ([]GitConnection, error)
	GetByServerURL(serverURL string) (*GitConnection, error)
	Update(id string, updates map[string]interface{}) error
	SoftDelete(id string) error
}

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(conn *GitConnection) error {
	if err := r.db.Create(conn).Error; err != nil {
		if database.IsUniqueConstraintError(err) {
			return ErrNameDuplicate
		}
		return err
	}
	return nil
}

func (r *Repo) GetByID(id string) (*GitConnection, error) {
	var conn GitConnection
	err := r.db.Where("id = ? AND status != ?", id, StatusDeleted).First(&conn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &conn, nil
}

func (r *Repo) List() ([]GitConnection, int64, error) {
	var conns []GitConnection
	var total int64

	query := r.db.Where("status != ?", StatusDeleted)
	if err := query.Model(&GitConnection{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Find(&conns).Error; err != nil {
		return nil, 0, err
	}
	return conns, total, nil
}

func (r *Repo) ListByProviderType(providerType string) ([]GitConnection, error) {
	var conns []GitConnection
	err := r.db.Where("provider_type = ? AND status != ?", providerType, StatusDeleted).
		Order("created_at DESC").Find(&conns).Error
	if err != nil {
		return nil, err
	}
	return conns, nil
}

func (r *Repo) GetByServerURL(serverURL string) (*GitConnection, error) {
	var conn GitConnection
	err := r.db.Where("server_url = ? AND status != ?", serverURL, StatusDeleted).First(&conn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &conn, nil
}

func (r *Repo) Update(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	result := r.db.Model(&GitConnection{}).Where("id = ? AND status != ?", id, StatusDeleted).Updates(updates)
	if result.Error != nil {
		if database.IsUniqueConstraintError(result.Error) {
			return ErrNameDuplicate
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) SoftDelete(id string) error {
	result := r.db.Model(&GitConnection{}).Where("id = ? AND status != ?", id, StatusDeleted).
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
