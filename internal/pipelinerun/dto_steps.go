package pipelinerun

import (
	"encoding/json"
	"time"

	"github.com/xjy/zcid/internal/stepexec"
)

type StepExecutionResponse struct {
	ID              string          `json:"id"`
	PipelineRunID   string          `json:"pipelineRunId"`
	TaskRunName     string          `json:"taskRunName"`
	StepName        string          `json:"stepName"`
	StepIndex       int             `json:"stepIndex"`
	Status          string          `json:"status"`
	ImageRef        *string         `json:"imageRef,omitempty"`
	ImageDigest     *string         `json:"imageDigest,omitempty"`
	SecretRefs      json.RawMessage `json:"secretRefs"`
	WorkspaceMounts json.RawMessage `json:"workspaceMounts"`
	Resources       json.RawMessage `json:"resources"`
	OutputDigests   json.RawMessage `json:"outputDigests"`
	TraceID         *string         `json:"traceId,omitempty"`
	StartedAt       *time.Time      `json:"startedAt,omitempty"`
	FinishedAt      *time.Time      `json:"finishedAt,omitempty"`
	DurationMS      *int64          `json:"durationMs,omitempty"`
	ExitCode        *int            `json:"exitCode,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

type StepExecutionListResponse struct {
	Items []StepExecutionResponse `json:"items"`
}

func toStepExecutionResponse(row stepexec.StepExecution) StepExecutionResponse {
	return StepExecutionResponse{
		ID: row.ID, PipelineRunID: row.PipelineRunID, TaskRunName: row.TaskRunName, StepName: row.StepName,
		StepIndex: row.StepIndex, Status: row.Status, ImageRef: row.ImageRef, ImageDigest: row.ImageDigest,
		SecretRefs: raw(row.SecretRefs, `[]`), WorkspaceMounts: raw(row.WorkspaceMounts, `[]`), Resources: raw(row.Resources, `{}`),
		OutputDigests: raw(row.OutputDigests, `[]`), TraceID: row.TraceID,
		StartedAt: row.StartedAt, FinishedAt: row.FinishedAt, DurationMS: row.DurationMS, ExitCode: row.ExitCode,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func raw(v stepexec.JSONRaw, fallback string) json.RawMessage {
	if len(v) == 0 || !json.Valid([]byte(v)) {
		return json.RawMessage(fallback)
	}
	return json.RawMessage(v)
}
