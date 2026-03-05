package audit

import "time"

type AuditLog struct {
	ID           string    `gorm:"column:id"`
	UserID       *string   `gorm:"column:user_id"`
	Action       string    `gorm:"column:action"`
	ResourceType string    `gorm:"column:resource_type"`
	ResourceID   *string   `gorm:"column:resource_id"`
	Result       string    `gorm:"column:result"`
	IP           *string   `gorm:"column:ip"`
	Detail       *string   `gorm:"column:detail"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
