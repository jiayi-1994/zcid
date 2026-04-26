package environment

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
	router.GET("/:eid", h.Get)
	router.PUT("/:eid", h.Update)
	router.DELETE("/:eid", h.Delete)
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
// @Summary Create an environment
// @Description Create a new environment within a project (admin or project admin only)
// @Tags environments
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body CreateEnvironmentRequest true "Environment creation payload"
// @Success 200 {object} response.Response{data=EnvironmentResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/projects/{id}/environments [post]
func (h *Handler) Create(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id is required"))
		return
	}

	if !isAdminOrProjectAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可创建环境", ""))
		return
	}

	var req CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}

	env, err := h.service.Create(c.Request.Context(), projectID, req.Name, req.Namespace, req.Description)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToEnvironmentResponse(env))
}

// Get godoc
// @Summary Get an environment
// @Description Retrieve an environment by its ID within a project
// @Tags environments
// @Produce json
// @Param id path string true "Project ID"
// @Param eid path string true "Environment ID"
// @Success 200 {object} response.Response{data=EnvironmentResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/environments/{eid} [get]
func (h *Handler) Get(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	envID := strings.TrimSpace(c.Param("eid"))
	if projectID == "" || envID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and environment id are required"))
		return
	}

	env, err := h.service.Get(c.Request.Context(), envID, projectID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToEnvironmentResponseWithHealth(env, h.service.Health(c.Request.Context(), projectID, env.ID)))
}

// List godoc
// @Summary List environments
// @Description Retrieve a paginated list of environments in a project
// @Tags environments
// @Produce json
// @Param id path string true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=EnvironmentListResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{id}/environments [get]
func (h *Handler) List(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id is required"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	envs, total, err := h.service.List(c.Request.Context(), projectID, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	items := make([]EnvironmentResponse, len(envs))
	for i, e := range envs {
		items[i] = ToEnvironmentResponseWithHealth(e, h.service.Health(c.Request.Context(), projectID, e.ID))
	}

	response.Success(c, EnvironmentListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// Update godoc
// @Summary Update an environment
// @Description Update an existing environment's configuration
// @Tags environments
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param eid path string true "Environment ID"
// @Param request body UpdateEnvironmentRequest true "Environment update payload"
// @Success 200 {object} response.Response{data=EnvironmentResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/environments/{eid} [put]
func (h *Handler) Update(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	envID := strings.TrimSpace(c.Param("eid"))
	if projectID == "" || envID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and environment id are required"))
		return
	}

	var req UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}

	env, err := h.service.Update(c.Request.Context(), envID, projectID, req.Name, req.Namespace, req.Description)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToEnvironmentResponse(env))
}

// Delete godoc
// @Summary Delete an environment
// @Description Delete an environment from a project (admin or project admin only)
// @Tags environments
// @Produce json
// @Param id path string true "Project ID"
// @Param eid path string true "Environment ID"
// @Success 200 {object} response.Response{data=object{deleted=bool}}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/projects/{id}/environments/{eid} [delete]
func (h *Handler) Delete(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	envID := strings.TrimSpace(c.Param("eid"))
	if projectID == "" || envID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and environment id are required"))
		return
	}

	if !isAdminOrProjectAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可删除环境", ""))
		return
	}

	if err := h.service.Delete(c.Request.Context(), envID, projectID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}
