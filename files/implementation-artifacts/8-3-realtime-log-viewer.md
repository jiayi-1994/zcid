# Story 8.3: Realtime Log Viewer

**Status:** done

## Summary
Implemented log streaming with buffer for reconnection replay and mock log collector.

## Deliverables
- `internal/ws/logstream.go` - LogBuffer, LogLine, LogCollector interface, MockLogCollector, LogStreamManager
- `internal/ws/logstream_test.go` - TestLogBufferAppendAndGetSince, TestLogBufferMaxLines, TestGetSinceReplay
- LogStreamManager integrates with hub for BroadcastToRun
- ReplayFn provides reconnection replay from GetSince(lastSeq)

## Notes
- LogCollector is external dependency; MockLogCollector used with TODO for real K8s pod log streaming
- LogBuffer caps at maxLines (default 5000 per run) for memory
- StartStreaming(runID, namespace, podName) spawns collector goroutine
- TODO: Replace MockLogCollector with real K8s corev1.PodInterface.GetLogs stream
