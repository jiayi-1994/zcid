package project

import "time"

type ProjectStatus string

const (
	ProjectStatusActive  ProjectStatus = "active"
	ProjectStatusDeleted ProjectStatus = "deleted"
)

type ProjectRole string

const (
	RoleProjectAdmin ProjectRole = "project_admin"
	RoleMember       ProjectRole = "member"
)

type Project struct {
	ID          string        `gorm:"column:id"`
	Name        string        `gorm:"column:name"`
	Description string        `gorm:"column:description"`
	OwnerID     string        `gorm:"column:owner_id"`
	Status      ProjectStatus `gorm:"column:status"`
	CreatedAt   time.Time     `gorm:"column:created_at"`
	UpdatedAt   time.Time     `gorm:"column:updated_at"`
}

func (Project) TableName() string {
	return "projects"
}

type ProjectMember struct {
	ID        string      `gorm:"column:id"`
	ProjectID string      `gorm:"column:project_id"`
	UserID    string      `gorm:"column:user_id"`
	Role      ProjectRole `gorm:"column:role"`
	CreatedAt time.Time   `gorm:"column:created_at"`
}

func (ProjectMember) TableName() string {
	return "project_members"
}
