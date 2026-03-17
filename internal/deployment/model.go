package deployment

import "time"

type DeployStatus string

const (
	StatusPending    DeployStatus = "pending"
	StatusSyncing    DeployStatus = "syncing"
	StatusHealthy    DeployStatus = "healthy"
	StatusDegraded   DeployStatus = "degraded"
	StatusFailed     DeployStatus = "failed"
	StatusRolledBack DeployStatus = "rolled_back"
)

type Deployment struct {
	ID            string       `gorm:"column:id"`
	ProjectID     string       `gorm:"column:project_id"`
	EnvironmentID string       `gorm:"column:environment_id"`
	PipelineRunID *string      `gorm:"column:pipeline_run_id"`
	Image         string       `gorm:"column:image"`
	Status        DeployStatus `gorm:"column:status"`
	ArgoAppName   *string      `gorm:"column:argo_app_name"`
	SyncStatus    *string      `gorm:"column:sync_status"`
	HealthStatus  *string      `gorm:"column:health_status"`
	ErrorMessage  *string      `gorm:"column:error_message"`
	DeployedBy    string       `gorm:"column:deployed_by"`
	StartedAt     *time.Time   `gorm:"column:started_at"`
	FinishedAt    *time.Time   `gorm:"column:finished_at"`
	CreatedAt     time.Time    `gorm:"column:created_at"`
	UpdatedAt     time.Time    `gorm:"column:updated_at"`
}

func (Deployment) TableName() string {
	return "deployments"
}
