package pipelinerun

import "time"

// TriggerRunRequest is the request body for triggering a pipeline run
type TriggerRunRequest struct {
	Params    map[string]string `json:"params,omitempty"`
	GitBranch string            `json:"gitBranch,omitempty"`
	GitCommit string            `json:"gitCommit,omitempty"`
}

// PipelineRunResponse is the full run response
type PipelineRunResponse struct {
	ID           string            `json:"id"`
	PipelineID   string            `json:"pipelineId"`
	ProjectID    string            `json:"projectId"`
	RunNumber    int               `json:"runNumber"`
	Status       string            `json:"status"`
	TriggerType  string            `json:"triggerType"`
	TriggeredBy  *string           `json:"triggeredBy,omitempty"`
	GitBranch    *string           `json:"gitBranch,omitempty"`
	GitCommit    *string           `json:"gitCommit,omitempty"`
	GitAuthor    *string           `json:"gitAuthor,omitempty"`
	GitMessage   *string           `json:"gitMessage,omitempty"`
	Params       map[string]string `json:"params,omitempty"`
	TektonName   *string           `json:"tektonName,omitempty"`
	Namespace    *string           `json:"namespace,omitempty"`
	StartedAt    *time.Time        `json:"startedAt,omitempty"`
	FinishedAt   *time.Time        `json:"finishedAt,omitempty"`
	ErrorMessage *string           `json:"errorMessage,omitempty"`
	Artifacts    []Artifact        `json:"artifacts,omitempty"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

// PipelineRunSummary is a short summary for list responses
type PipelineRunSummary struct {
	ID          string     `json:"id"`
	PipelineID  string     `json:"pipelineId"`
	RunNumber   int        `json:"runNumber"`
	Status      string     `json:"status"`
	TriggerType string     `json:"triggerType"`
	TriggeredBy *string    `json:"triggeredBy,omitempty"`
	GitBranch   *string    `json:"gitBranch,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	FinishedAt  *time.Time `json:"finishedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// PipelineRunListResponse is the paginated list response
type PipelineRunListResponse struct {
	Items    []PipelineRunSummary `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}
