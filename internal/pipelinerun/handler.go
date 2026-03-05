package pipelinerun

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/pkg/middleware"
	"github.com/xjy/zcid/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.POST("", h.TriggerRun)
	router.GET("", h.ListRuns)
	router.GET("/:runId", h.GetRun)
	router.POST("/:runId/cancel", h.CancelRun)
	router.GET("/:runId/artifacts", h.GetArtifacts)
	router.PUT("/:runId/artifacts", h.UpdateArtifacts)
}

func getProjectID(c *gin.Context) string {
	return strings.TrimSpace(c.Param("id"))
}

func getPipelineID(c *gin.Context) string {
	return strings.TrimSpace(c.Param("pipelineId"))
}

func getRunID(c *gin.Context) string {
	return strings.TrimSpace(c.Param("runId"))
}

func getUserID(c *gin.Context) string {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := userID.(string)
	return uid
}

func (h *Handler) TriggerRun(c *gin.Context) {
	projectID := getProjectID(c)
	pipelineID := getPipelineID(c)
	if projectID == "" || pipelineID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and pipeline id are required"))
		return
	}

	uid := getUserID(c)
	if uid == "" {
		response.HandleError(c, response.NewBizError(response.CodeUnauthorized, "用户未认证", ""))
		return
	}

	var req TriggerRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = TriggerRunRequest{}
	}

	run, err := h.service.TriggerRun(c.Request.Context(), projectID, pipelineID, uid, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, run)
}

func (h *Handler) ListRuns(c *gin.Context) {
	projectID := getProjectID(c)
	pipelineID := getPipelineID(c)
	if projectID == "" || pipelineID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and pipeline id are required"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	result, err := h.service.ListRuns(c.Request.Context(), projectID, pipelineID, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *Handler) GetRun(c *gin.Context) {
	projectID := getProjectID(c)
	runID := getRunID(c)
	if projectID == "" || runID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and run id are required"))
		return
	}

	run, err := h.service.GetRun(c.Request.Context(), projectID, runID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, run)
}

func (h *Handler) CancelRun(c *gin.Context) {
	projectID := getProjectID(c)
	runID := getRunID(c)
	if projectID == "" || runID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and run id are required"))
		return
	}

	if err := h.service.CancelRun(c.Request.Context(), projectID, runID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"cancelled": true})
}

func (h *Handler) GetArtifacts(c *gin.Context) {
	projectID := getProjectID(c)
	runID := getRunID(c)
	if projectID == "" || runID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and run id are required"))
		return
	}

	artifacts, err := h.service.GetArtifacts(c.Request.Context(), projectID, runID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"artifacts": artifacts})
}

func (h *Handler) UpdateArtifacts(c *gin.Context) {
	projectID := getProjectID(c)
	runID := getRunID(c)
	if projectID == "" || runID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and run id are required"))
		return
	}

	var req struct {
		Artifacts []Artifact `json:"artifacts"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}

	if err := h.service.UpdateArtifacts(c.Request.Context(), projectID, runID, req.Artifacts); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}
