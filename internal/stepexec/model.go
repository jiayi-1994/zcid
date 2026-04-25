package stepexec

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const (
	StatusPending     = "pending"
	StatusRunning     = "running"
	StatusSucceeded   = "succeeded"
	StatusFailed      = "failed"
	StatusCancelled   = "cancelled"
	StatusInterrupted = "interrupted"

	CommandArgsLimitKB    = 64 * 1024
	TektonResultsLimitKB  = 256 * 1024
	EnvPublicLimitKB      = 32 * 1024
	ParamsResolvedLimitKB = 32 * 1024
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
func RawArray() JSONRaw  { return JSONRaw([]byte("[]")) }

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
		*r = JSONRaw([]byte("{}"))
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
		*r = JSONRaw([]byte("{}"))
	}
	return nil
}

func (r JSONRaw) Bytes() []byte {
	return []byte(r)
}

type StepExecution struct {
	ID              string     `gorm:"column:id" json:"id"`
	PipelineRunID   string     `gorm:"column:pipeline_run_id" json:"pipelineRunId"`
	TaskRunName     string     `gorm:"column:task_run_name" json:"taskRunName"`
	StepName        string     `gorm:"column:step_name" json:"stepName"`
	StepIndex       int        `gorm:"column:step_index" json:"stepIndex"`
	Status          string     `gorm:"column:status" json:"status"`
	ImageRef        *string    `gorm:"column:image_ref" json:"imageRef,omitempty"`
	ImageDigest     *string    `gorm:"column:image_digest" json:"imageDigest,omitempty"`
	CommandArgs     JSONRaw    `gorm:"column:command_args;type:jsonb" json:"commandArgs"`
	EnvPublic       JSONRaw    `gorm:"column:env_public;type:jsonb" json:"envPublic"`
	SecretRefs      JSONRaw    `gorm:"column:secret_refs;type:jsonb" json:"secretRefs"`
	ParamsResolved  JSONRaw    `gorm:"column:params_resolved;type:jsonb" json:"paramsResolved"`
	WorkspaceMounts JSONRaw    `gorm:"column:workspace_mounts;type:jsonb" json:"workspaceMounts"`
	Resources       JSONRaw    `gorm:"column:resources;type:jsonb" json:"resources"`
	TektonResults   JSONRaw    `gorm:"column:tekton_results;type:jsonb" json:"tektonResults"`
	OutputDigests   JSONRaw    `gorm:"column:output_digests;type:jsonb" json:"outputDigests"`
	LogRef          JSONRaw    `gorm:"column:log_ref;type:jsonb" json:"-"`
	TraceID         *string    `gorm:"column:trace_id" json:"traceId,omitempty"`
	StartedAt       *time.Time `gorm:"column:started_at" json:"startedAt,omitempty"`
	FinishedAt      *time.Time `gorm:"column:finished_at" json:"finishedAt,omitempty"`
	DurationMS      *int64     `gorm:"column:duration_ms" json:"durationMs,omitempty"`
	ExitCode        *int       `gorm:"column:exit_code" json:"exitCode,omitempty"`
	CreatedAt       time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

func (StepExecution) TableName() string { return "step_executions" }

func (s *StepExecution) NormalizeJSON() {
	if len(s.CommandArgs) == 0 {
		s.CommandArgs = RawObject()
	}
	if len(s.EnvPublic) == 0 {
		s.EnvPublic = RawObject()
	}
	if len(s.SecretRefs) == 0 {
		s.SecretRefs = RawArray()
	}
	if len(s.ParamsResolved) == 0 {
		s.ParamsResolved = RawObject()
	}
	if len(s.WorkspaceMounts) == 0 {
		s.WorkspaceMounts = RawArray()
	}
	if len(s.Resources) == 0 {
		s.Resources = RawObject()
	}
	if len(s.TektonResults) == 0 {
		s.TektonResults = RawArray()
	}
	if len(s.OutputDigests) == 0 {
		s.OutputDigests = RawArray()
	}
	if len(s.LogRef) == 0 {
		s.LogRef = RawObject()
	}
}

func ApplySizeLimits(row *StepExecution) (truncated []string) {
	row.CommandArgs, truncated = truncateField(row.CommandArgs, CommandArgsLimitKB, "command_args", truncated)
	row.TektonResults, truncated = truncateField(row.TektonResults, TektonResultsLimitKB, "tekton_results", truncated)
	row.EnvPublic, truncated = truncateField(row.EnvPublic, EnvPublicLimitKB, "env_public", truncated)
	row.ParamsResolved, truncated = truncateField(row.ParamsResolved, ParamsResolvedLimitKB, "params_resolved", truncated)
	return truncated
}

func truncateField(raw JSONRaw, limit int, field string, truncated []string) (JSONRaw, []string) {
	if len(raw) <= limit {
		return raw, truncated
	}
	wrapped := map[string]any{
		"_truncated":     true,
		"field":          field,
		"original_bytes": len(raw),
	}
	return NewJSONRaw(wrapped), append(truncated, field)
}
