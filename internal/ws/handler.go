package ws

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/xjy/zcid/pkg/middleware"
	"github.com/xjy/zcid/pkg/response"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		// Allow localhost for development
		if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") {
			return true
		}
		// TODO: Configure allowed origins via environment variable
		return strings.HasPrefix(origin, "https://")
	},
}

// ReplayFn returns log lines (as JSON bytes) to send on reconnection.
type ReplayFn func(runID string, lastSeq int64) [][]byte

// AccessChecker verifies that a user can access the given resource.
type AccessChecker interface {
	CanAccessRun(userID, runID string) bool
	CanAccessProject(userID, projectID string) bool
}

// ServeWsLogs handles WebSocket connections for log streaming: /ws/v1/logs/:runId
func ServeWsLogs(hub *Hub, jwtSecret string, replay ReplayFn, acl AccessChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		runID := strings.TrimSpace(c.Param("runId"))
		if runID == "" {
			response.HandleError(c, response.NewBizError(response.CodeBadRequest, "runId required", ""))
			return
		}

		token := strings.TrimSpace(c.Query("token"))
		userID, ok := middleware.ParseTokenForWebSocket(token, []byte(jwtSecret))
		if !ok {
			response.HandleError(c, response.NewBizError(response.CodeWSAuthFailed, "ws auth failed", ""))
			return
		}

		if acl != nil && !acl.CanAccessRun(userID, runID) {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "无权访问该运行记录", ""))
			return
		}

		if hub.CountUserConnections(userID) >= maxConnsPerUser {
			response.HandleError(c, response.NewBizError(response.CodeWSConnectionLimit, "connection limit exceeded", ""))
			return
		}

		lastSeq := int64(0)
		if s := c.Query("lastSeq"); s != "" {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil && n >= 0 {
				lastSeq = n
			}
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			slog.Warn("ws upgrade failed", slog.Any("error", err))
			return
		}

		client := NewClient(conn, userID, runID, "", hub)
		client.LastSeq = lastSeq
		hub.Register(client)

		go client.writePump()
		go client.readPump()

		if replay != nil && lastSeq > 0 {
			for _, buf := range replay(runID, lastSeq) {
				select {
				case client.Send <- buf:
				default:
					break
				}
			}
		}
	}
}

// ServeWsStatus handles WebSocket connections for pipeline status: /ws/v1/pipeline-status/:projectId
func ServeWsStatus(hub *Hub, jwtSecret string, acl AccessChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := strings.TrimSpace(c.Param("projectId"))
		if projectID == "" {
			response.HandleError(c, response.NewBizError(response.CodeBadRequest, "projectId required", ""))
			return
		}

		token := strings.TrimSpace(c.Query("token"))
		userID, ok := middleware.ParseTokenForWebSocket(token, []byte(jwtSecret))
		if !ok {
			response.HandleError(c, response.NewBizError(response.CodeWSAuthFailed, "ws auth failed", ""))
			return
		}

		if acl != nil && !acl.CanAccessProject(userID, projectID) {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "无权访问该项目", ""))
			return
		}

		if hub.CountUserConnections(userID) >= maxConnsPerUser {
			response.HandleError(c, response.NewBizError(response.CodeWSConnectionLimit, "connection limit exceeded", ""))
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			slog.Warn("ws upgrade failed", slog.Any("error", err))
			return
		}

		client := NewClient(conn, userID, "", projectID, hub)
		hub.Register(client)

		go client.writePump()
		go client.readPump()
	}
}
