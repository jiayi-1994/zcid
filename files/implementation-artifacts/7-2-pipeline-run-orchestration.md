# Story 7.2: Pipeline Run Orchestration

**Status:** done

## Summary
Implemented pipeline run lifecycle: trigger, list, get, cancel with mock K8s client.

## Deliverables
- `migrations/000013_create_pipeline_runs.up.sql` - pipeline_runs table with status, git info, params, artifacts
- `internal/pipelinerun/model.go` - PipelineRun GORM model, Artifact, RunStatus
- `internal/pipelinerun/dto.go` - TriggerRunRequest, PipelineRunResponse, PipelineRunListResponse
- `internal/pipelinerun/repo.go` - CRUD, GetNextRunNumber, ListByPipeline, UpdateStatus, CountRunning
- `internal/pipelinerun/executor.go` - K8sClient interface, MockK8sClient with TODO for real K8s
- `internal/pipelinerun/service.go` - TriggerRun, CancelRun, GetRun, ListRuns with concurrency policy
- `internal/pipelinerun/handler.go` - POST/GET /runs, GET /:runId, POST /:runId/cancel
- `internal/pipelinerun/service_test.go` - TriggerRun success/reject/cancelOld, CancelRun, GetRun scope, artifacts tests

## Notes
- Routes under /api/v1/projects/:id/pipelines/:pipelineId/runs
- Concurrency: queue (allow), reject (error if running), cancel_old (proceed; full cancel logic TODO)
- Variables resolved via variableService.ResolveVariables(projectID)
