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
	router.GET("/:runId/step-executions", h.GetStepExecutions)
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

// TriggerRun godoc
// @Summary Trigger a pipeline run
// @Description Trigger a new run for a pipeline within a project
// @Tags pipeline-runs
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param pipelineId path string true "Pipeline ID"
// @Param request body TriggerRunRequest false "Run trigger parameters"
// @Success 200 {object} response.Response{data=PipelineRunResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId}/runs [post]
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

// ListRuns godoc
// @Summary List pipeline runs
// @Description Retrieve a paginated list of runs for a pipeline
// @Tags pipeline-runs
// @Produce json
// @Param id path string true "Project ID"
// @Param pipelineId path string true "Pipeline ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=PipelineRunListResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId}/runs [get]
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

// GetRun godoc
// @Summary Get a pipeline run
// @Description Retrieve details of a specific pipeline run
// @Tags pipeline-runs
// @Produce json
// @Param id path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} response.Response{data=PipelineRunResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId}/runs/{runId} [get]
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

// GetStepExecutions godoc
// @Summary Get run step executions
// @Description Retrieve persisted per-step execution records for a pipeline run
// @Tags pipeline-runs
// @Produce json
// @Param id path string true "Project ID"
// @Param pipelineId path string true "Pipeline ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} response.Response{data=StepExecutionListResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId}/runs/{runId}/step-executions [get]
func (h *Handler) GetStepExecutions(c *gin.Context) {
	projectID := getProjectID(c)
	pipelineID := getPipelineID(c)
	runID := getRunID(c)
	if projectID == "" || pipelineID == "" || runID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id, pipeline id and run id are required"))
		return
	}

	result, err := h.service.GetStepExecutions(c.Request.Context(), projectID, pipelineID, runID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, result)
}

// CancelRun godoc
// @Summary Cancel a pipeline run
// @Description Cancel a running pipeline execution
// @Tags pipeline-runs
// @Produce json
// @Param id path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} response.Response{data=object{cancelled=bool}}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId}/runs/{runId}/cancel [post]
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

// GetArtifacts godoc
// @Summary Get run artifacts
// @Description Retrieve artifacts produced by a pipeline run
// @Tags pipeline-runs
// @Produce json
// @Param id path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} response.Response{data=object{artifacts=[]Artifact}}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId}/runs/{runId}/artifacts [get]
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

// UpdateArtifacts godoc
// @Summary Update run artifacts
// @Description Update or replace artifacts for a pipeline run
// @Tags pipeline-runs
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param request body object{artifacts=[]Artifact} true "Artifacts payload"
// @Success 200 {object} response.Response{data=object{updated=bool}}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId}/runs/{runId}/artifacts [put]
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
