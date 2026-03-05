package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/pkg/response"
)

func ErrorRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				requestID := GetRequestID(c)
				slog.Error("panic recovered",
					slog.String("requestId", requestID),
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
				)
				response.Error(c, http.StatusInternalServerError, response.CodeInternalServerError, "internal server error", "")
				c.Abort()
			}
		}()

		c.Next()
	}
}
