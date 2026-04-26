package signal

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type TargetType string

const (
	TargetService     TargetType = "service"
	TargetEnvironment TargetType = "environment"
	TargetPipeline    TargetType = "pipeline"
	TargetDeployment  TargetType = "deployment"
	TargetIntegration TargetType = "integration"
)

type Status string

const (
	StatusHealthy  Status = "healthy"
	StatusWarning  Status = "warning"
	StatusDegraded Status = "degraded"
	StatusUnknown  Status = "unknown"
	StatusStale    Status = "stale"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

type JSONRaw json.RawMessage

func NewJSONRaw(v any) JSONRaw {
	if v == nil {
		return JSONRaw([]byte("{}"))
	}
	b, err := json.Marshal(v)
	if err != nil {
		return JSONRaw([]byte("{}"))
	}
	return JSONRaw(b)
}

func RawObject() JSONRaw { return JSONRaw([]byte("{}")) }

func (r JSONRaw) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}
	if !json.Valid(r) {
		return nil, fmt.Errorf("JSONRaw.MarshalJSON: invalid JSON")
	}
	return []byte(r), nil
}

func (r JSONRaw) Value() (driver.Value, error) {
	if len(r) == 0 {
		return []byte("{}"), nil
	}
	if !json.Valid(r) {
		return nil, fmt.Errorf("JSONRaw.Value: invalid JSON")
	}
	return []byte(r), nil
}

func (r *JSONRaw) Scan(value interface{}) error {
	if value == nil {
		*r = RawObject()
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*r = append((*r)[0:0], v...)
	case string:
		*r = append((*r)[0:0], []byte(v)...)
	default:
		return fmt.Errorf("JSONRaw.Scan: unsupported type %T", value)
	}
	if len(*r) == 0 {
		*r = RawObject()
	}
	return nil
}

type HealthSignal struct {
	ID            string     `gorm:"column:id" json:"id"`
	ProjectID     string     `gorm:"column:project_id" json:"projectId"`
	TargetType    TargetType `gorm:"column:target_type" json:"targetType"`
	TargetID      string     `gorm:"column:target_id" json:"targetId"`
	Source        string     `gorm:"column:source" json:"source"`
	Status        Status     `gorm:"column:status" json:"status"`
	Severity      Severity   `gorm:"column:severity" json:"severity"`
	Reason        string     `gorm:"column:reason" json:"reason"`
	Message       string     `gorm:"column:message" json:"message"`
	ObservedValue JSONRaw    `gorm:"column:observed_value;type:jsonb" json:"observedValue"`
	ObservedAt    time.Time  `gorm:"column:observed_at" json:"observedAt"`
	StaleAfter    *time.Time `gorm:"column:stale_after" json:"staleAfter,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

func (HealthSignal) TableName() string { return "health_signals" }

func (s *HealthSignal) Normalize(now time.Time) {
	if len(s.ObservedValue) == 0 {
		s.ObservedValue = RawObject()
	}
	if s.ObservedAt.IsZero() {
		s.ObservedAt = now
	}
	if s.Severity == "" {
		s.Severity = SeverityInfo
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
}

func (s HealthSignal) EffectiveStatus(now time.Time) Status {
	if s.StaleAfter != nil && !s.StaleAfter.After(now) {
		return StatusStale
	}
	return s.Status
}
