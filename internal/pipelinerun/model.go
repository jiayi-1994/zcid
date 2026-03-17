package pipelinerun

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type RunStatus string

const (
	StatusPending   RunStatus = "pending"
	StatusQueued    RunStatus = "queued"
	StatusRunning   RunStatus = "running"
	StatusSucceeded RunStatus = "succeeded"
	StatusFailed    RunStatus = "failed"
	StatusCancelled RunStatus = "cancelled"
)

type Artifact struct {
	Type string `json:"type"` // image, file
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size,omitempty"`
}

// JSONMap stores map[string]string for JSONB
type JSONMap map[string]string

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = map[string]string{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return fmt.Errorf("JSONMap.Scan: unsupported type %T", value)
	}
}

// JSONBytes stores raw JSON for JSONB
type JSONBytes []byte

func (b JSONBytes) Value() (driver.Value, error) {
	if b == nil {
		return []byte("{}"), nil
	}
	return []byte(b), nil
}

func (b *JSONBytes) Scan(value interface{}) error {
	if value == nil {
		*b = []byte("{}")
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*b = append((*b)[0:0], v...)
		return nil
	case string:
		*b = []byte(v)
		return nil
	default:
		return fmt.Errorf("JSONBytes.Scan: unsupported type %T", value)
	}
}

// ArtifactSlice for JSONB storage
type ArtifactSlice []Artifact

func (s ArtifactSlice) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]Artifact{})
	}
	return json.Marshal(s)
}

func (s *ArtifactSlice) Scan(value interface{}) error {
	if value == nil {
		*s = []Artifact{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("ArtifactSlice.Scan: unsupported type %T", value)
	}
}

type PipelineRun struct {
	ID             string        `gorm:"column:id"`
	PipelineID     string        `gorm:"column:pipeline_id"`
	ProjectID      string        `gorm:"column:project_id"`
	RunNumber      int           `gorm:"column:run_number"`
	Status         RunStatus     `gorm:"column:status"`
	TriggerType    string        `gorm:"column:trigger_type"`
	TriggeredBy    *string       `gorm:"column:triggered_by"`
	GitBranch      *string       `gorm:"column:git_branch"`
	GitCommit      *string       `gorm:"column:git_commit"`
	GitAuthor      *string       `gorm:"column:git_author"`
	GitMessage     *string       `gorm:"column:git_message"`
	ConfigSnapshot JSONBytes     `gorm:"column:config_snapshot;type:jsonb"`
	Params         JSONMap       `gorm:"column:params;type:jsonb"`
	TektonName     *string       `gorm:"column:tekton_name"`
	Namespace      *string       `gorm:"column:namespace"`
	StartedAt      *time.Time    `gorm:"column:started_at"`
	FinishedAt     *time.Time    `gorm:"column:finished_at"`
	ErrorMessage   *string       `gorm:"column:error_message"`
	Artifacts      ArtifactSlice `gorm:"column:artifacts;type:jsonb"`
	CreatedAt      time.Time     `gorm:"column:created_at"`
	UpdatedAt      time.Time     `gorm:"column:updated_at"`
}

func (PipelineRun) TableName() string {
	return "pipeline_runs"
}
