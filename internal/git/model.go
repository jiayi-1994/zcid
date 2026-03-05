package git

import "time"

type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusTokenExpired ConnectionStatus = "token_expired"
	StatusDeleted      ConnectionStatus = "deleted"
)

type TokenType string

const (
	TokenPAT   TokenType = "pat"
	TokenOAuth TokenType = "oauth"
)

type GitConnection struct {
	ID            string           `gorm:"column:id"`
	Name          string           `gorm:"column:name"`
	ProviderType  string           `gorm:"column:provider_type"`
	ServerURL     string           `gorm:"column:server_url"`
	AccessToken   string           `gorm:"column:access_token"`
	RefreshToken  *string          `gorm:"column:refresh_token"`
	TokenType     TokenType        `gorm:"column:token_type"`
	WebhookSecret string           `gorm:"column:webhook_secret"`
	Status        ConnectionStatus `gorm:"column:status"`
	Description   string           `gorm:"column:description"`
	CreatedBy     string           `gorm:"column:created_by"`
	CreatedAt     time.Time        `gorm:"column:created_at"`
	UpdatedAt     time.Time        `gorm:"column:updated_at"`

	PlainTokenMask string `gorm:"-" json:"-"`
}

func (GitConnection) TableName() string {
	return "git_connections"
}
