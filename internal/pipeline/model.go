package pipeline

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type PipelineStatus string

const (
	StatusDraft   PipelineStatus = "draft"
	StatusActive  PipelineStatus = "active"
	StatusDisabled PipelineStatus = "disabled"
	StatusDeleted PipelineStatus = "deleted"
)

type TriggerType string

const (
	TriggerManual    TriggerType = "manual"
	TriggerWebhook   TriggerType = "webhook"
	TriggerScheduled TriggerType = "scheduled"
)

type ConcurrencyPolicy string

const (
	ConcurrencyQueue     ConcurrencyPolicy = "queue"
	ConcurrencyCancelOld ConcurrencyPolicy = "cancel_old"
	ConcurrencyReject    ConcurrencyPolicy = "reject"
)

type Pipeline struct {
	ID                string            `gorm:"column:id"`
	ProjectID         string            `gorm:"column:project_id"`
	Name              string            `gorm:"column:name"`
	Description       string            `gorm:"column:description"`
	Status            PipelineStatus    `gorm:"column:status"`
	Config            PipelineConfig    `gorm:"column:config;type:jsonb"`
	TriggerType       TriggerType       `gorm:"column:trigger_type"`
	ConcurrencyPolicy ConcurrencyPolicy `gorm:"column:concurrency_policy"`
	CreatedBy         string            `gorm:"column:created_by"`
	CreatedAt         time.Time         `gorm:"column:created_at"`
	UpdatedAt         time.Time         `gorm:"column:updated_at"`
}

func (Pipeline) TableName() string {
	return "pipelines"
}

type PipelineConfig struct {
	SchemaVersion string            `json:"schemaVersion"`
	Stages        []StageConfig     `json:"stages"`
	Params        []ParamConfig     `json:"params,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type StageConfig struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Steps []StepConfig `json:"steps"`
}

type StepConfig struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Image   string            `json:"image,omitempty"`
	Command []string          `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Config  map[string]any    `json:"config,omitempty"`
}

type ParamConfig struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Description  string `json:"description,omitempty"`
	Required     bool   `json:"required"`
}

func (c PipelineConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *PipelineConfig) Scan(value interface{}) error {
	if value == nil {
		*c = PipelineConfig{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	default:
		return fmt.Errorf("PipelineConfig.Scan: unsupported type %T", value)
	}
}
