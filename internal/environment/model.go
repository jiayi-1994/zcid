package environment

import "time"

type EnvironmentStatus string

const (
	StatusActive  EnvironmentStatus = "active"
	StatusDeleted EnvironmentStatus = "deleted"
)

type Environment struct {
	ID          string            `gorm:"column:id"`
	ProjectID   string            `gorm:"column:project_id"`
	Name        string            `gorm:"column:name"`
	Namespace   string            `gorm:"column:namespace"`
	Description string            `gorm:"column:description"`
	Status      EnvironmentStatus `gorm:"column:status"`
	CreatedAt   time.Time         `gorm:"column:created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at"`
}

func (Environment) TableName() string {
	return "environments"
}
