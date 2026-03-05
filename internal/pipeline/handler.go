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
	router.GET("/:pipelineId", h.GetPipeline)
	router.PUT("/:pipelineId", h.UpdatePipeline)
	router.DELETE("/:pipelineId", h.DeletePipeline)
	router.POST("/:pipelineId/copy", h.CopyPipeline)
}

func (h *Handler) RegisterTemplateRoutes(router gin.IRoutes) {
	router.GET("", h.ListTemplates)
	router.GET("/:templateId", h.GetTemplate)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	templates := h.service.ListTemplates()
	items := make([]TemplateSummary, len(templates))
	for i, t := range templates {
		items[i] = ToTemplateSummary(t)
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

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
