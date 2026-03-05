package notification

import "time"

type EventType string

const (
	EventBuildSuccess  EventType = "build_success"
	EventBuildFailed   EventType = "build_failed"
	EventDeploySuccess EventType = "deploy_success"
	EventDeployFailed  EventType = "deploy_failed"
)

type NotificationRule struct {
	ID         string    `gorm:"column:id"`
	ProjectID  string    `gorm:"column:project_id"`
	Name       string    `gorm:"column:name"`
	EventType  EventType `gorm:"column:event_type"`
	WebhookURL string    `gorm:"column:webhook_url"`
	Enabled    bool      `gorm:"column:enabled"`
	CreatedBy  string    `gorm:"column:created_by"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (NotificationRule) TableName() string {
	return "notification_rules"
}
