package auth

import "time"

const (
	AccessTokenTTL  = 30 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

type UserStatus string

type SystemRole string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"

	SystemRoleAdmin        SystemRole = "admin"
	SystemRoleProjectAdmin SystemRole = "project_admin"
	SystemRoleMember       SystemRole = "member"
)

type User struct {
	ID           string     `gorm:"column:id"`
	Username     string     `gorm:"column:username"`
	PasswordHash string     `gorm:"column:password_hash"`
	Status       UserStatus `gorm:"column:status"`
	Role         SystemRole `gorm:"column:role"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}
