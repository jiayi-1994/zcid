package audit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/pkg/middleware"
)

func Middleware(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch && method != http.MethodDelete {
			return
		}
		if c.Writer.Status() >= 400 {
			return
		}
		userID, _ := c.Get(middleware.ContextKeyUserID)
		uid, _ := userID.(string)
		action := method + " " + c.FullPath()
		resourceType := inferResourceType(c)
		resourceID := inferResourceID(c)
		ip := c.ClientIP()
		service.LogAction(c.Request.Context(), uid, action, resourceType, resourceID, "success", ip, "")
	}
}

func inferResourceType(c *gin.Context) string {
	path := c.FullPath()
	switch {
	case strings.Contains(path, "/projects/"):
		return "project"
	case strings.Contains(path, "/pipelines/"):
		return "pipeline"
	case strings.Contains(path, "/deployments"):
		return "deployment"
	case strings.Contains(path, "/notification-rules"):
		return "notification_rule"
	case strings.Contains(path, "/variables"):
		return "variable"
	case strings.Contains(path, "/environments"):
		return "environment"
	default:
		return "unknown"
	}
}

func inferResourceID(c *gin.Context) string {
	if id := c.Param("id"); id != "" {
		return id
	}
	if id := c.Param("projectId"); id != "" {
		return id
	}
	if id := c.Param("deployId"); id != "" {
		return id
	}
	return ""
}
