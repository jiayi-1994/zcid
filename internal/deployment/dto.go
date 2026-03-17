package deployment

import "time"

// TriggerDeployRequest is the request body for triggering a deployment.
type TriggerDeployRequest struct {
	EnvironmentID string  `json:"environmentId" binding:"required"`
	Image         string  `json:"image" binding:"required"`
	PipelineRunID *string `json:"pipelineRunId,omitempty"`
}

// DeploymentResponse is the full deployment response.
type DeploymentResponse struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"projectId"`
	EnvironmentID string     `json:"environmentId"`
	PipelineRunID *string    `json:"pipelineRunId,omitempty"`
	Image         string     `json:"image"`
	Status        string     `json:"status"`
	ArgoAppName   *string    `json:"argoAppName,omitempty"`
	SyncStatus    *string    `json:"syncStatus,omitempty"`
	HealthStatus  *string    `json:"healthStatus,omitempty"`
	ErrorMessage  *string    `json:"errorMessage,omitempty"`
	DeployedBy    string     `json:"deployedBy"`
	StartedAt     *time.Time `json:"startedAt,omitempty"`
	FinishedAt    *time.Time `json:"finishedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// DeploymentSummary is a condensed deployment for list views.
type DeploymentSummary struct {
	ID            string    `json:"id"`
	EnvironmentID string    `json:"environmentId"`
	Image         string    `json:"image"`
	Status        string    `json:"status"`
	SyncStatus    *string   `json:"syncStatus,omitempty"`
	HealthStatus  *string   `json:"healthStatus,omitempty"`
	DeployedBy    string    `json:"deployedBy"`
	CreatedAt     time.Time `json:"createdAt"`
}

// DeploymentListResponse is the paginated list response.
type DeploymentListResponse struct {
	Items    []DeploymentSummary `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

func ToDeploymentResponse(d *Deployment) DeploymentResponse {
	resp := DeploymentResponse{
		ID:            d.ID,
		ProjectID:     d.ProjectID,
		EnvironmentID: d.EnvironmentID,
		Image:         d.Image,
		Status:        string(d.Status),
		DeployedBy:    d.DeployedBy,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
	if d.PipelineRunID != nil {
		resp.PipelineRunID = d.PipelineRunID
	}
	if d.ArgoAppName != nil {
		resp.ArgoAppName = d.ArgoAppName
	}
	if d.SyncStatus != nil {
		resp.SyncStatus = d.SyncStatus
	}
	if d.HealthStatus != nil {
		resp.HealthStatus = d.HealthStatus
	}
	if d.ErrorMessage != nil {
		resp.ErrorMessage = d.ErrorMessage
	}
	if d.StartedAt != nil {
		resp.StartedAt = d.StartedAt
	}
	if d.FinishedAt != nil {
		resp.FinishedAt = d.FinishedAt
	}
	return resp
}

func ToDeploymentSummary(d *Deployment) DeploymentSummary {
	sum := DeploymentSummary{
		ID:            d.ID,
		EnvironmentID: d.EnvironmentID,
		Image:         d.Image,
		Status:        string(d.Status),
		DeployedBy:    d.DeployedBy,
		CreatedAt:     d.CreatedAt,
	}
	if d.SyncStatus != nil {
		sum.SyncStatus = d.SyncStatus
	}
	if d.HealthStatus != nil {
		sum.HealthStatus = d.HealthStatus
	}
	return sum
}
