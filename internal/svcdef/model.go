package svcdef

import "time"

type ServiceStatus string

const (
	StatusActive  ServiceStatus = "active"
	StatusDeleted ServiceStatus = "deleted"
)

type ServiceDef struct {
	ID          string        `gorm:"column:id"`
	ProjectID   string        `gorm:"column:project_id"`
	Name        string        `gorm:"column:name"`
	Description string        `gorm:"column:description"`
	RepoURL     string        `gorm:"column:repo_url"`
	Status      ServiceStatus `gorm:"column:status"`
	CreatedAt   time.Time     `gorm:"column:created_at"`
	UpdatedAt   time.Time     `gorm:"column:updated_at"`
}

func (ServiceDef) TableName() string {
	return "services"
}
