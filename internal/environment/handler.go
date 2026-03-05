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

	response.Success(c, ToEnvironmentResponse(env))
}

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
		items[i] = ToEnvironmentResponse(e)
	}

	response.Success(c, EnvironmentListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

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
