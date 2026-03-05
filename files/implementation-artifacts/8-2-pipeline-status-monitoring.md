# Story 8.2: Pipeline Status Monitoring

**Status:** done

## Summary
Implemented pipeline status monitoring via WebSocket with mock K8s watcher.

## Deliverables
- `internal/ws/watcher.go` - PipelineWatcher, K8sWatcher interface, MockK8sWatcher with TODO for real K8s Informer
- `internal/ws/watcher_test.go` - TestMockWatcherSendsStatus
- Status messages broadcast to project subscribers via hub.BroadcastToProject

## Notes
- K8sWatcher is external dependency; MockK8sWatcher used with TODO for real Informer
- PipelineWatcher maps namespace -> projectID for routing
- Start(ctx) spawns watchers per registered namespace
- TODO: Wire real K8s Informer when cluster available
