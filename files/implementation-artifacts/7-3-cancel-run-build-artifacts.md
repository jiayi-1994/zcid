# Story 7.3: Cancel Run & Build Artifacts

**Status:** done

## Summary
Added artifacts CRUD for pipeline runs; build steps can report artifacts via PUT.

## Deliverables
- `internal/pipelinerun/service.go` - UpdateArtifacts, GetArtifacts methods
- `internal/pipelinerun/handler.go` - GET /:runId/artifacts, PUT /:runId/artifacts
- `internal/pipelinerun/service_test.go` - TestUpdateArtifacts_Success, TestGetArtifacts_Success

## Notes
- Artifact struct: Type (image/file), Name, URL, Size
- PUT artifacts for build steps to report produced artifacts
- Stored in pipeline_runs.artifacts JSONB
