package notification

import "time"

type EventType string

type ChannelType string

const (
	EventBuildSuccess  EventType = "build_success"
	EventBuildFailed   EventType = "build_failed"
	EventDeploySuccess EventType = "deploy_success"
	EventDeployFailed  EventType = "deploy_failed"
)

const (
	ChannelWebhook ChannelType = "webhook"
	ChannelSlack   ChannelType = "slack"
)

type NotificationRule struct {
	ID           string      `gorm:"column:id"`
	ProjectID    string      `gorm:"column:project_id"`
	Name         string      `gorm:"column:name"`
	EventType    EventType   `gorm:"column:event_type"`
	ChannelType  ChannelType `gorm:"column:channel_type"`
	WebhookURL   string      `gorm:"column:webhook_url"`
	SlackToken   string      `gorm:"column:slack_token"`
	SlackChannel string      `gorm:"column:slack_channel"`
	Enabled      bool        `gorm:"column:enabled"`
	CreatedBy    string      `gorm:"column:created_by"`
	CreatedAt    time.Time   `gorm:"column:created_at"`
	UpdatedAt    time.Time   `gorm:"column:updated_at"`
}

func (NotificationRule) TableName() string {
	return "notification_rules"
}
