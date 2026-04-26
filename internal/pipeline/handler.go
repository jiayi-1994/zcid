package pipeline

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
	router.POST("", h.CreatePipeline)
	router.GET("", h.ListPipelines)
	router.POST("/from-template", h.CreateFromTemplate)
	router.GET("/:pipelineId", h.GetPipeline)
	router.PUT("/:pipelineId", h.UpdatePipeline)
	router.DELETE("/:pipelineId", h.DeletePipeline)
	router.POST("/:pipelineId/copy", h.CopyPipeline)
}

// CreateFromTemplate godoc
// @Summary Create a pipeline from template
// @Description Instantiate a pipeline template with user-provided parameter values
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body FromTemplateRequest true "Template instantiation payload"
// @Success 200 {object} response.Response{data=PipelineResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/from-template [post]
func (h *Handler) CreateFromTemplate(c *gin.Context) {
	projectID := getProjectID(c)
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id is required"))
		return
	}
	uid := getUserID(c)
	if uid == "" {
		response.HandleError(c, response.NewBizError(response.CodeUnauthorized, "用户未认证", ""))
		return
	}
	var req FromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}
	p, err := h.service.CreateFromTemplate(c.Request.Context(), projectID, req, uid)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, ToPipelineResponse(p))
}

func (h *Handler) RegisterTemplateRoutes(router gin.IRoutes) {
	router.GET("", h.ListTemplates)
	router.GET("/:templateId", h.GetTemplate)
}

// ListTemplates godoc
// @Summary List pipeline templates
// @Description Retrieve all available pipeline templates
// @Tags pipeline-templates
// @Produce json
// @Success 200 {object} response.Response{data=object{items=[]TemplateSummary,total=int}}
// @Router /api/v1/pipeline-templates [get]
func (h *Handler) ListTemplates(c *gin.Context) {
	templates := h.service.ListTemplates()
	items := make([]TemplateSummary, len(templates))
	for i, t := range templates {
		items[i] = ToTemplateSummary(t)
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

// GetTemplate godoc
// @Summary Get a pipeline template
// @Description Retrieve a pipeline template by its ID
// @Tags pipeline-templates
// @Produce json
// @Param templateId path string true "Template ID"
// @Success 200 {object} response.Response{data=PipelineTemplate}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/pipeline-templates/{templateId} [get]
func (h *Handler) GetTemplate(c *gin.Context) {
	templateID := strings.TrimSpace(c.Param("templateId"))
	if templateID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "template id is required"))
		return
	}

	t, err := h.service.GetTemplate(templateID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, t)
}

func getProjectID(c *gin.Context) string {
	return strings.TrimSpace(c.Param("id"))
}

func getPipelineID(c *gin.Context) string {
	return strings.TrimSpace(c.Param("pipelineId"))
}

func getUserID(c *gin.Context) string {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := userID.(string)
	return uid
}

// CreatePipeline godoc
// @Summary Create a pipeline
// @Description Create a new pipeline within a project
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body CreatePipelineRequest true "Pipeline creation payload"
// @Success 200 {object} response.Response{data=PipelineResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines [post]
func (h *Handler) CreatePipeline(c *gin.Context) {
	projectID := getProjectID(c)
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id is required"))
		return
	}

	var req CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}

	uid := getUserID(c)
	if uid == "" {
		response.HandleError(c, response.NewBizError(response.CodeUnauthorized, "用户未认证", ""))
		return
	}

	p, err := h.service.CreatePipeline(c.Request.Context(), projectID, req, uid)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToPipelineResponse(p))
}

// GetPipeline godoc
// @Summary Get a pipeline
// @Description Retrieve a pipeline by its ID within a project
// @Tags pipelines
// @Produce json
// @Param id path string true "Project ID"
// @Param pipelineId path string true "Pipeline ID"
// @Success 200 {object} response.Response{data=PipelineResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId} [get]
func (h *Handler) GetPipeline(c *gin.Context) {
	projectID := getProjectID(c)
	pipelineID := getPipelineID(c)
	if projectID == "" || pipelineID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and pipeline id are required"))
		return
	}

	p, err := h.service.GetPipeline(c.Request.Context(), pipelineID, projectID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToPipelineResponse(p))
}

// ListPipelines godoc
// @Summary List pipelines
// @Description Retrieve a paginated list of pipelines in a project
// @Tags pipelines
// @Produce json
// @Param id path string true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=PipelineListResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines [get]
func (h *Handler) ListPipelines(c *gin.Context) {
	projectID := getProjectID(c)
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id is required"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	pipelines, total, err := h.service.ListPipelines(c.Request.Context(), projectID, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	items := make([]PipelineSummaryResponse, len(pipelines))
	for i, p := range pipelines {
		items[i] = ToPipelineSummaryResponse(p)
	}

	response.Success(c, PipelineListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// UpdatePipeline godoc
// @Summary Update a pipeline
// @Description Update an existing pipeline's configuration
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param pipelineId path string true "Pipeline ID"
// @Param request body UpdatePipelineRequest true "Pipeline update payload"
// @Success 200 {object} response.Response{data=PipelineResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId} [put]
func (h *Handler) UpdatePipeline(c *gin.Context) {
	projectID := getProjectID(c)
	pipelineID := getPipelineID(c)
	if projectID == "" || pipelineID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and pipeline id are required"))
		return
	}

	var req UpdatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}

	p, err := h.service.UpdatePipeline(c.Request.Context(), pipelineID, projectID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToPipelineResponse(p))
}

// DeletePipeline godoc
// @Summary Delete a pipeline
// @Description Delete a pipeline from a project
// @Tags pipelines
// @Produce json
// @Param id path string true "Project ID"
// @Param pipelineId path string true "Pipeline ID"
// @Success 200 {object} response.Response{data=object{deleted=bool}}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId} [delete]
func (h *Handler) DeletePipeline(c *gin.Context) {
	projectID := getProjectID(c)
	pipelineID := getPipelineID(c)
	if projectID == "" || pipelineID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and pipeline id are required"))
		return
	}

	if err := h.service.DeletePipeline(c.Request.Context(), pipelineID, projectID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// CopyPipeline godoc
// @Summary Copy a pipeline
// @Description Create a copy of an existing pipeline within the same project
// @Tags pipelines
// @Produce json
// @Param id path string true "Project ID"
// @Param pipelineId path string true "Pipeline ID to copy"
// @Success 200 {object} response.Response{data=PipelineResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId}/copy [post]
func (h *Handler) CopyPipeline(c *gin.Context) {
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

	p, err := h.service.CopyPipeline(c.Request.Context(), pipelineID, projectID, uid)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToPipelineResponse(p))
}
