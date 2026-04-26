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

type BootstrapToken struct {
	ID        string     `gorm:"column:id"`
	TokenHash string     `gorm:"column:token_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

type AccessTokenType string

const (
	AccessTokenTypePersonal AccessTokenType = "personal"
	AccessTokenTypeProject  AccessTokenType = "project"
)

type AccessToken struct {
	ID          string          `gorm:"column:id"`
	TokenType   AccessTokenType `gorm:"column:token_type"`
	Name        string          `gorm:"column:name"`
	TokenPrefix string          `gorm:"column:token_prefix"`
	TokenHash   string          `gorm:"column:token_hash"`
	Scopes      string          `gorm:"column:scopes"`
	UserID      *string         `gorm:"column:user_id"`
	ProjectID   *string         `gorm:"column:project_id"`
	CreatedBy   string          `gorm:"column:created_by"`
	ExpiresAt   time.Time       `gorm:"column:expires_at"`
	LastUsedAt  *time.Time      `gorm:"column:last_used_at"`
	RevokedAt   *time.Time      `gorm:"column:revoked_at"`
	RevokedBy   *string         `gorm:"column:revoked_by"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
}

func (BootstrapToken) TableName() string {
	return "bootstrap_tokens"
}

func (AccessToken) TableName() string {
	return "access_tokens"
}

func (User) TableName() string {
	return "users"
}
