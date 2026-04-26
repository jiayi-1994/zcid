package svcdef

import (
	"context"
	"strings"
	"time"

	"github.com/xjy/zcid/internal/signal"
	"github.com/xjy/zcid/pkg/response"
)

type VitalsSignalTarget struct {
	TargetType signal.TargetType
	TargetID   string
}

type VitalsPipeline struct {
	ID        string
	Name      string
	Status    string
	RepoURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type VitalsRun struct {
	ID           string
	PipelineID   string
	PipelineName string
	RunNumber    int
	Status       string
	StartedAt    *time.Time
	FinishedAt   *time.Time
	ErrorMessage *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type VitalsDeployment struct {
	ID              string
	EnvironmentID   string
	EnvironmentName string
	PipelineRunID   *string
	Image           string
	Status          string
	SyncStatus      *string
	HealthStatus    *string
	ErrorMessage    *string
	StartedAt       *time.Time
	FinishedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type VitalsStepWarning struct {
	StepName     string
	TaskRunName  string
	Status       string
	DurationMS   *int64
	ExitCode     *int
	StartedAt    *time.Time
	FinishedAt   *time.Time
	RunID        string
	PipelineID   string
	PipelineName string
	RunNumber    int
	CreatedAt    time.Time
}

type VitalsSignal struct {
	ID            string         `json:"id"`
	TargetType    string         `json:"targetType"`
	TargetID      string         `json:"targetId"`
	Source        string         `json:"source"`
	Status        string         `json:"status"`
	RawStatus     string         `json:"rawStatus"`
	Severity      string         `json:"severity"`
	Reason        string         `json:"reason"`
	Message       string         `json:"message"`
	ObservedValue signal.JSONRaw `json:"observedValue"`
	ObservedAt    time.Time      `json:"observedAt"`
	StaleAfter    *time.Time     `json:"staleAfter,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
}

type ServiceVitalsResponse struct {
	Service           ServiceResponse             `json:"service"`
	Summary           ServiceVitalsSummary        `json:"summary"`
	LinkedPipelines   []VitalsPipelineResponse    `json:"linkedPipelines"`
	RecentRuns        []VitalsRunResponse         `json:"recentRuns"`
	LatestDeployments []VitalsDeploymentResponse  `json:"latestDeployments"`
	ActiveSignals     []VitalsSignal              `json:"activeSignals"`
	Warnings          []VitalsStepWarningResponse `json:"warnings"`
	EmptyStates       []string                    `json:"emptyStates"`
	RefreshedAt       string                      `json:"refreshedAt"`
}

type ServiceVitalsSummary struct {
	Status             string `json:"status"`
	Reason             string `json:"reason"`
	LastSignalAt       string `json:"lastSignalAt,omitempty"`
	HasDeliveryData    bool   `json:"hasDeliveryData"`
	HasDeploymentData  bool   `json:"hasDeploymentData"`
	ActiveWarningCount int    `json:"activeWarningCount"`
}

type VitalsPipelineResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	RepoURL   string `json:"repoUrl"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type VitalsRunResponse struct {
	ID           string `json:"id"`
	PipelineID   string `json:"pipelineId"`
	PipelineName string `json:"pipelineName"`
	RunNumber    int    `json:"runNumber"`
	Status       string `json:"status"`
	StartedAt    string `json:"startedAt,omitempty"`
	FinishedAt   string `json:"finishedAt,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

type VitalsDeploymentResponse struct {
	ID              string `json:"id"`
	EnvironmentID   string `json:"environmentId"`
	EnvironmentName string `json:"environmentName"`
	PipelineRunID   string `json:"pipelineRunId,omitempty"`
	Image           string `json:"image"`
	Status          string `json:"status"`
	SyncStatus      string `json:"syncStatus,omitempty"`
	HealthStatus    string `json:"healthStatus,omitempty"`
	ErrorMessage    string `json:"errorMessage,omitempty"`
	StartedAt       string `json:"startedAt,omitempty"`
	FinishedAt      string `json:"finishedAt,omitempty"`
	CreatedAt       string `json:"createdAt"`
}

type VitalsStepWarningResponse struct {
	StepName     string `json:"stepName"`
	TaskRunName  string `json:"taskRunName"`
	Status       string `json:"status"`
	DurationMS   *int64 `json:"durationMs,omitempty"`
	ExitCode     *int   `json:"exitCode,omitempty"`
	RunID        string `json:"runId"`
	PipelineID   string `json:"pipelineId"`
	PipelineName string `json:"pipelineName"`
	RunNumber    int    `json:"runNumber"`
	RunPath      string `json:"runPath"`
	CreatedAt    string `json:"createdAt"`
}

func (s *Service) GetVitals(ctx context.Context, projectID, serviceID string) (*ServiceVitalsResponse, error) {
	svc, err := s.Get(ctx, serviceID, projectID)
	if err != nil {
		return nil, err
	}

	pipelines, err := s.repo.ListLinkedPipelines(ctx, svc)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "查询服务关联流水线失败", err.Error())
	}
	pipelineIDs := make([]string, 0, len(pipelines))
	for _, p := range pipelines {
		pipelineIDs = append(pipelineIDs, p.ID)
	}
	runs, err := s.repo.ListRecentRuns(ctx, projectID, pipelineIDs, 10)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "查询服务运行记录失败", err.Error())
	}
	runIDs := make([]string, 0, len(runs))
	for _, run := range runs {
		runIDs = append(runIDs, run.ID)
	}
	deployments, err := s.repo.ListLatestDeployments(ctx, projectID, []string(svc.EnvironmentIDs), 10)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "查询服务部署记录失败", err.Error())
	}
	stepWarnings, err := s.repo.ListFailedSteps(ctx, projectID, runIDs, 10)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "查询服务步骤告警失败", err.Error())
	}

	targets := signalTargetsForVitals(svc, pipelines, deployments)
	signals, err := s.repo.ListLatestSignals(ctx, projectID, targets, 5)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "查询服务健康信号失败", err.Error())
	}

	now := time.Now()
	resp := &ServiceVitalsResponse{
		Service:           ToServiceResponse(svc),
		LinkedPipelines:   toVitalsPipelineResponses(pipelines),
		RecentRuns:        toVitalsRunResponses(runs),
		LatestDeployments: toVitalsDeploymentResponses(deployments),
		ActiveSignals:     signals,
		Warnings:          toVitalsStepWarningResponses(projectID, stepWarnings),
		RefreshedAt:       formatTime(now),
	}
	resp.EmptyStates = serviceVitalsEmptyStates(resp)
	resp.Summary = summarizeVitals(resp)
	return resp, nil
}

func signalTargetsForVitals(svc *ServiceDef, pipelines []VitalsPipeline, deployments []VitalsDeployment) []VitalsSignalTarget {
	targets := []VitalsSignalTarget{{TargetType: signal.TargetService, TargetID: svc.ID}}
	for _, integration := range []string{"database", "redis", "k8s", "registry"} {
		targets = append(targets, VitalsSignalTarget{TargetType: signal.TargetIntegration, TargetID: integration})
	}
	for _, id := range svc.EnvironmentIDs {
		if strings.TrimSpace(id) != "" {
			targets = append(targets, VitalsSignalTarget{TargetType: signal.TargetEnvironment, TargetID: id})
		}
	}
	for _, pipeline := range pipelines {
		targets = append(targets, VitalsSignalTarget{TargetType: signal.TargetPipeline, TargetID: pipeline.ID})
	}
	for _, deployment := range deployments {
		targets = append(targets, VitalsSignalTarget{TargetType: signal.TargetDeployment, TargetID: deployment.ID})
	}
	return targets
}

func summarizeVitals(resp *ServiceVitalsResponse) ServiceVitalsSummary {
	status := "unknown"
	reason := "No service health signals or delivery evidence yet"
	lastSignalAt := ""
	var lastSignalAtTime time.Time
	warnings := len(resp.Warnings)

	if len(resp.ActiveSignals) > 0 {
		for _, sig := range resp.ActiveSignals {
			if sig.ObservedAt.After(lastSignalAtTime) {
				lastSignalAtTime = sig.ObservedAt
				lastSignalAt = formatTime(sig.ObservedAt)
			}
			status = worseStatus(status, sig.Status)
			if sig.Message != "" {
				reason = sig.Message
			} else if sig.Reason != "" {
				reason = sig.Reason
			}
			if sig.Status == string(signal.StatusWarning) || sig.Status == string(signal.StatusDegraded) || sig.Status == string(signal.StatusStale) {
				warnings++
			}
		}
	}
	for _, d := range resp.LatestDeployments {
		status = worseStatus(status, deploymentToHealth(d.Status))
		if d.ErrorMessage != "" {
			reason = d.ErrorMessage
		}
	}
	for _, run := range resp.RecentRuns {
		switch run.Status {
		case "failed", "cancelled":
			status = worseStatus(status, "warning")
			if run.ErrorMessage != "" {
				reason = run.ErrorMessage
			}
		case "running", "queued", "pending":
			status = worseStatus(status, "warning")
		case "succeeded":
			if status == "unknown" {
				status = "healthy"
				reason = "Recent delivery evidence is healthy"
			}
		}
	}
	if len(resp.Warnings) > 0 {
		status = worseStatus(status, "warning")
		reason = "Recent pipeline steps need attention"
	}
	if len(resp.ActiveSignals) == 0 && len(resp.RecentRuns) == 0 && len(resp.LatestDeployments) == 0 {
		status = "unknown"
	}
	return ServiceVitalsSummary{
		Status:             status,
		Reason:             reason,
		LastSignalAt:       lastSignalAt,
		HasDeliveryData:    len(resp.RecentRuns) > 0,
		HasDeploymentData:  len(resp.LatestDeployments) > 0,
		ActiveWarningCount: warnings,
	}
}

func worseStatus(current, candidate string) string {
	rank := map[string]int{"healthy": 0, "unknown": 1, "stale": 2, "warning": 3, "degraded": 4}
	if rank[candidate] > rank[current] {
		return candidate
	}
	return current
}

func deploymentToHealth(status string) string {
	switch status {
	case "healthy":
		return "healthy"
	case "degraded", "failed":
		return "degraded"
	case "pending", "syncing":
		return "warning"
	default:
		return "unknown"
	}
}

func serviceVitalsEmptyStates(resp *ServiceVitalsResponse) []string {
	var out []string
	if len(resp.LinkedPipelines) == 0 {
		out = append(out, "No linked pipelines. Add pipeline IDs or matching repo metadata to this service.")
	}
	if len(resp.LatestDeployments) == 0 {
		out = append(out, "No linked deployments. Add environment IDs to this service or deploy to a linked environment.")
	}
	if len(resp.ActiveSignals) == 0 {
		out = append(out, "No active health signals yet. Signals will appear after deployments, checks, or pipeline evidence emit them.")
	}
	return out
}

func toVitalsPipelineResponses(rows []VitalsPipeline) []VitalsPipelineResponse {
	out := make([]VitalsPipelineResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, VitalsPipelineResponse{
			ID: row.ID, Name: row.Name, Status: row.Status, RepoURL: row.RepoURL,
			CreatedAt: formatTime(row.CreatedAt), UpdatedAt: formatTime(row.UpdatedAt),
		})
	}
	return out
}

func toVitalsRunResponses(rows []VitalsRun) []VitalsRunResponse {
	out := make([]VitalsRunResponse, 0, len(rows))
	for _, row := range rows {
		item := VitalsRunResponse{
			ID: row.ID, PipelineID: row.PipelineID, PipelineName: row.PipelineName,
			RunNumber: row.RunNumber, Status: row.Status, CreatedAt: formatTime(row.CreatedAt),
		}
		if row.StartedAt != nil {
			item.StartedAt = formatTime(*row.StartedAt)
		}
		if row.FinishedAt != nil {
			item.FinishedAt = formatTime(*row.FinishedAt)
		}
		if row.ErrorMessage != nil {
			item.ErrorMessage = *row.ErrorMessage
		}
		out = append(out, item)
	}
	return out
}

func toVitalsDeploymentResponses(rows []VitalsDeployment) []VitalsDeploymentResponse {
	out := make([]VitalsDeploymentResponse, 0, len(rows))
	for _, row := range rows {
		item := VitalsDeploymentResponse{
			ID: row.ID, EnvironmentID: row.EnvironmentID, EnvironmentName: row.EnvironmentName,
			Image: row.Image, Status: row.Status, CreatedAt: formatTime(row.CreatedAt),
		}
		if row.PipelineRunID != nil {
			item.PipelineRunID = *row.PipelineRunID
		}
		if row.SyncStatus != nil {
			item.SyncStatus = *row.SyncStatus
		}
		if row.HealthStatus != nil {
			item.HealthStatus = *row.HealthStatus
		}
		if row.ErrorMessage != nil {
			item.ErrorMessage = *row.ErrorMessage
		}
		if row.StartedAt != nil {
			item.StartedAt = formatTime(*row.StartedAt)
		}
		if row.FinishedAt != nil {
			item.FinishedAt = formatTime(*row.FinishedAt)
		}
		out = append(out, item)
	}
	return out
}

func toVitalsStepWarningResponses(projectID string, rows []VitalsStepWarning) []VitalsStepWarningResponse {
	out := make([]VitalsStepWarningResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, VitalsStepWarningResponse{
			StepName: row.StepName, TaskRunName: row.TaskRunName, Status: row.Status,
			DurationMS: row.DurationMS, ExitCode: row.ExitCode, RunID: row.RunID,
			PipelineID: row.PipelineID, PipelineName: row.PipelineName, RunNumber: row.RunNumber,
			RunPath:   "/projects/" + projectID + "/pipelines/" + row.PipelineID + "/runs/" + row.RunID,
			CreatedAt: formatTime(row.CreatedAt),
		})
	}
	return out
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
