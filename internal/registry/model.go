package registry

import "time"

// RegistryType represents the image registry provider type
type RegistryType string

const (
	RegistryTypeHarbor   RegistryType = "harbor"
	RegistryTypeDockerHub RegistryType = "dockerhub"
	RegistryTypeGHCR      RegistryType = "ghcr"
)

// RegistryStatus represents the registry status
type RegistryStatus string

const (
	StatusActive   RegistryStatus = "active"
	StatusDisabled RegistryStatus = "disabled"
	StatusDeleted  RegistryStatus = "deleted"
)

// Registry is the GORM model for image registries
type Registry struct {
	ID               string         `gorm:"column:id;primaryKey"`
	Name             string         `gorm:"column:name;not null"`
	Type             RegistryType   `gorm:"column:type;not null"`
	URL              string         `gorm:"column:url;not null"`
	Username         string         `gorm:"column:username"`
	PasswordEncrypted string        `gorm:"column:password_encrypted"`
	IsDefault        bool           `gorm:"column:is_default;not null"`
	Status           RegistryStatus `gorm:"column:status;not null"`
	CreatedBy        string         `gorm:"column:created_by;not null"`
	CreatedAt        time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;not null"`
}

// TableName returns the table name for Registry
func (Registry) TableName() string {
	return "registries"
}
