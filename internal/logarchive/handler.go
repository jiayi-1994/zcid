package logarchive

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/pkg/middleware"
	"github.com/xjy/zcid/pkg/response"
)

// RunChecker verifies a run belongs to a project.
type RunChecker interface {
	RunBelongsToProject(ctx context.Context, runID, projectID string) bool
}

// Handler exposes archived logs API.
type Handler struct {
	service    *Service
	runChecker RunChecker
}

// NewHandler creates a log archive handler.
func NewHandler(service *Service, runChecker RunChecker) *Handler {
	return &Handler{service: service, runChecker: runChecker}
}

// RegisterRoutes adds log archive routes to the router.
func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/logs", h.GetArchivedLogs)
}

// GetArchivedLogs godoc
// @Summary Get archived logs
// @Description Retrieve paginated archived logs for a pipeline run
// @Tags log-archives
// @Produce json
// @Param id path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(50)
// @Success 200 {object} response.Response{data=object{items=[]object,total=int,page=int,pageSize=int}}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/projects/{id}/pipeline-runs/{runId}/logs [get]
func (h *Handler) GetArchivedLogs(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	runID := strings.TrimSpace(c.Param("runId"))
	if projectID == "" || runID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and run id are required"))
		return
	}

	_ = getUserID(c)

	if h.runChecker != nil && !h.runChecker.RunBelongsToProject(c.Request.Context(), runID, projectID) {
		response.HandleError(c, response.NewBizError(response.CodeRunNotFound, "运行记录不存在", ""))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))

	entries, total, err := h.service.GetArchivedLogs(c.Request.Context(), runID, page, pageSize)
	if err != nil {
		response.HandleError(c, response.NewBizError(response.CodeInternalServerError, "获取归档日志失败", err.Error()))
		return
	}

	response.Success(c, gin.H{
		"items":    entries,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func getUserID(c *gin.Context) string {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := userID.(string)
	return uid
}
