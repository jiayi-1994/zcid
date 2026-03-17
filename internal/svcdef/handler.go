package svcdef

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
	router.POST("", h.Create)
	router.GET("", h.List)
	router.GET("/:sid", h.Get)
	router.PUT("/:sid", h.Update)
	router.DELETE("/:sid", h.Delete)
}

func isAdminOrProjectAdmin(c *gin.Context) bool {
	sysRole, _ := c.Get(middleware.ContextKeyRole)
	if r, ok := sysRole.(string); ok && (r == "admin" || r == "project_admin") {
		return true
	}
	projRole, _ := c.Get(middleware.ContextKeyProjectRole)
	if r, ok := projRole.(string); ok && r == "project_admin" {
		return true
	}
	return false
}

// Create godoc
// @Summary Create a service definition
// @Description Create a new service definition within a project (admin or project admin only)
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body CreateServiceRequest true "Service creation payload"
// @Success 200 {object} response.Response{data=ServiceResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/projects/{id}/services [post]
func (h *Handler) Create(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id is required"))
		return
	}

	if !isAdminOrProjectAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可创建服务", ""))
		return
	}

	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}

	svc, err := h.service.Create(c.Request.Context(), projectID, req.Name, req.Description, req.RepoURL)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToServiceResponse(svc))
}

// Get godoc
// @Summary Get a service definition
// @Description Retrieve a service definition by its ID within a project
// @Tags services
// @Produce json
// @Param id path string true "Project ID"
// @Param sid path string true "Service ID"
// @Success 200 {object} response.Response{data=ServiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/services/{sid} [get]
func (h *Handler) Get(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	svcID := strings.TrimSpace(c.Param("sid"))
	if projectID == "" || svcID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and service id are required"))
		return
	}

	svc, err := h.service.Get(c.Request.Context(), svcID, projectID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToServiceResponse(svc))
}

// List godoc
// @Summary List service definitions
// @Description Retrieve a paginated list of service definitions in a project
// @Tags services
// @Produce json
// @Param id path string true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=ServiceListResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{id}/services [get]
func (h *Handler) List(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id is required"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	svcs, total, err := h.service.List(c.Request.Context(), projectID, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	items := make([]ServiceResponse, len(svcs))
	for i, s := range svcs {
		items[i] = ToServiceResponse(s)
	}

	response.Success(c, ServiceListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// Update godoc
// @Summary Update a service definition
// @Description Update an existing service definition's configuration
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param sid path string true "Service ID"
// @Param request body UpdateServiceRequest true "Service update payload"
// @Success 200 {object} response.Response{data=ServiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/services/{sid} [put]
func (h *Handler) Update(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	svcID := strings.TrimSpace(c.Param("sid"))
	if projectID == "" || svcID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and service id are required"))
		return
	}

	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}

	svc, err := h.service.Update(c.Request.Context(), svcID, projectID, req.Name, req.Description, req.RepoURL)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToServiceResponse(svc))
}

// Delete godoc
// @Summary Delete a service definition
// @Description Delete a service definition from a project (admin or project admin only)
// @Tags services
// @Produce json
// @Param id path string true "Project ID"
// @Param sid path string true "Service ID"
// @Success 200 {object} response.Response{data=object{deleted=bool}}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/projects/{id}/services/{sid} [delete]
func (h *Handler) Delete(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	svcID := strings.TrimSpace(c.Param("sid"))
	if projectID == "" || svcID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and service id are required"))
		return
	}

	if !isAdminOrProjectAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可删除服务", ""))
		return
	}

	if err := h.service.Delete(c.Request.Context(), svcID, projectID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}
