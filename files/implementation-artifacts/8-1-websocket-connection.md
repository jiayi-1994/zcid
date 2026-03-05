# Story 8.1: WebSocket Connection Management

**Status:** done

## Summary
Implemented WebSocket connection management for real-time logs and status via gorilla/websocket.

## Deliverables
- `internal/ws/hub.go` - Client, Hub with register/unregister/broadcast, heartbeat (ping 30s, pong 60s)
- `internal/ws/message.go` - WSMessage, LogData, StatusData, StepStatus types
- `internal/ws/handler.go` - ServeWsLogs, ServeWsStatus with JWT auth from query param, connection limit (10/user), lastSeq support
- `internal/ws/hub_test.go` - TestHubRegisterUnregister, TestBroadcastToRun, TestBroadcastToProject, TestConnectionLimit, TestHeartbeat
- `pkg/middleware/auth.go` - ParseTokenForWebSocket for JWT from query
- `pkg/response/codes.go` - CodeWSConnectionLimit, CodeWSAuthFailed, CodeWSInvalidMessage

## Notes
- Routes: GET /ws/v1/logs/:runId, GET /ws/v1/pipeline-status/:projectId
- Auth via ?token=xxx (JWT)
- Reconnection: ?lastSeq=N for log replay
- Dependency: github.com/gorilla/websocket
