# Story 8.5: Run History & Frontend Pages

**Status:** done

## Summary
Implemented PipelineRun API service, run list page, run detail page, and integrated into app routes.

## Deliverables
- `web/src/services/pipelineRun.ts` - PipelineRun, PipelineRunSummary, PipelineRunList, Artifact, LogEntry; fetchPipelineRuns, fetchPipelineRun, triggerPipelineRun, cancelPipelineRun, fetchRunArtifacts, fetchArchivedLogs
- `web/src/pages/projects/pipelines/PipelineRunListPage.tsx` - Table (run#, status, trigger, triggered by, branch, duration, time); View details, Cancel; Trigger run
- `web/src/pages/projects/pipelines/PipelineRunDetailPage.tsx` - Run details (status, trigger, git, params); log placeholder; artifacts list; Cancel if running
- `web/src/App.tsx` - Routes: `/pipelines/:pipelineId/runs`, `/pipelines/:pipelineId/runs/:runId`
- `web/src/pages/projects/pipelines/PipelineListPage.tsx` - "运行历史" link per pipeline
- `web/src/components/pipeline/RunPipelineModal.tsx` - onSubmit form (gitBranch, gitCommit) when provided

## Notes
- Status badge colors: pending=gray, queued=blue, running=arcoblue, succeeded=green, failed=red, cancelled=orange
- Log viewer placeholder (xterm.js planned for future)
- Backend PipelineRunSummary extended with TriggeredBy, GitBranch for list columns
