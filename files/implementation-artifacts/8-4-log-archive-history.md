# Story 8.4: Log Archival & History

**Status:** done

## Summary
Implemented log archival to MinIO (JSONL chunks) and paginated API for archived log retrieval.

## Deliverables
- `internal/logarchive/model.go` - LogEntry struct
- `internal/logarchive/storage.go` - StorageClient interface
- `internal/logarchive/minio_adapter.go` - MinIOStorageClient (MinIOAdapter)
- `internal/logarchive/mock_storage.go` - MockStorage for tests
- `internal/logarchive/service.go` - ArchiveRunLogs (1MB JSONL chunks), GetArchivedLogs (paginated)
- `internal/logarchive/handler.go` - GET `/api/v1/projects/:id/pipeline-runs/:runId/logs`
- `internal/logarchive/service_test.go` - TestArchiveRunLogs, TestArchiveRunLogs_Empty, TestGetArchivedLogs_Pagination
- Routes registered in cmd/server/main.go

## Notes
- Logs stored at `logs/{runID}/chunk-{n}.jsonl` in zcid-logs bucket
- ArchiveRunLogs accepts []ws.LogLine; use after run completion to persist buffer
- GetArchivedLogs returns []LogEntry with pagination (page, pageSize)
