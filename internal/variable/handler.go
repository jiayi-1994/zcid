package variable

import (
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

func (h *Handler) RegisterProjectRoutes(router gin.IRoutes) {
	router.GET("", h.ListProjectVariables)
	router.GET("/merged", h.ListMergedVariables)
	router.POST("", h.CreateProjectVariable)
	router.PUT("/:vid", h.UpdateVariable)
	router.DELETE("/:vid", h.DeleteVariable)
}

func (h *Handler) RegisterPipelineRoutes(router gin.IRoutes) {
	router.GET("", h.ListPipelineVariables)
	router.POST("", h.CreatePipelineVariable)
	router.PUT("/:vid", h.UpdateVariable)
	router.DELETE("/:vid", h.DeleteVariable)
}

func (h *Handler) RegisterGlobalRoutes(router gin.IRoutes) {
	router.GET("", h.ListGlobalVariables)
	router.POST("", h.CreateGlobalVariable)
	router.PUT("/:vid", h.UpdateVariable)
	router.DELETE("/:vid", h.DeleteVariable)
}

func (h *Handler) CreateProjectVariable(c *gin.Context) {
	if !isAdminOrProjectAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可创建变量", ""))
		return
	}

	projectID := c.Param("id")
	var req CreateVariableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "请求参数错误", err.Error()))
		return
	}

	userID, _ := c.Get(middleware.ContextKeyUserID)
	v, err := h.service.CreateVariable(ScopeProject, &projectID, nil, req, userID.(string))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToVariableResponse(v, true))
}

func (h *Handler) CreateGlobalVariable(c *gin.Context) {
	if !isAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员可创建全局变量", ""))
		return
	}

	var req CreateVariableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "请求参数错误", err.Error()))
		return
	}

	userID, _ := c.Get(middleware.ContextKeyUserID)
	v, err := h.service.CreateVariable(ScopeGlobal, nil, nil, req, userID.(string))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToVariableResponse(v, true))
}

func (h *Handler) ListProjectVariables(c *gin.Context) {
	projectID := c.Param("id")
	vars, _, err := h.service.ListProjectVariables(projectID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	role := getUserRole(c)
	vars = FilterForRole(vars, role)

	items := make([]VariableResponse, len(vars))
	for i, v := range vars {
		items[i] = ToVariableResponse(&v, true)
	}
	response.Success(c, VariableListResponse{Items: items, Total: int64(len(items))})
}

func (h *Handler) ListGlobalVariables(c *gin.Context) {
	vars, total, err := h.service.ListGlobalVariables()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	items := make([]VariableResponse, len(vars))
	for i, v := range vars {
		items[i] = ToVariableResponse(&v, true)
	}
	response.Success(c, VariableListResponse{Items: items, Total: total})
}

func (h *Handler) ListMergedVariables(c *gin.Context) {
	projectID := c.Param("id")
	vars, err := h.service.GetMergedVariables(projectID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	role := getUserRole(c)
	vars = FilterForRole(vars, role)

	items := make([]VariableResponse, len(vars))
	for i, v := range vars {
		items[i] = ToVariableResponse(&v, true)
	}
	response.Success(c, VariableListResponse{Items: items, Total: int64(len(items))})
}

func (h *Handler) UpdateVariable(c *gin.Context) {
	vid := c.Param("vid")
	var req UpdateVariableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "请求参数错误", err.Error()))
		return
	}

	existing, err := h.service.GetVariable(vid)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if existing.Scope == ScopeGlobal {
		if !isAdmin(c) {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员可修改全局变量", ""))
			return
		}
	} else {
		if !isAdminOrProjectAdmin(c) {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可修改变量", ""))
			return
		}
	}

	if err := h.service.UpdateVariable(vid, req, existing.VarType == TypeSecret); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ListPipelineVariables(c *gin.Context) {
	projectID := c.Param("id")
	pipelineID := c.Param("pipelineId")
	if projectID == "" || pipelineID == "" {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "请求参数错误", "project id and pipeline id are required"))
		return
	}
	vars, total, err := h.service.ListPipelineVariables(projectID, pipelineID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	role := getUserRole(c)
	vars = FilterForRole(vars, role)
	items := make([]VariableResponse, len(vars))
	for i, v := range vars {
		items[i] = ToVariableResponse(&v, true)
	}
	response.Success(c, VariableListResponse{Items: items, Total: total})
}

func (h *Handler) CreatePipelineVariable(c *gin.Context) {
	if !isAdminOrProjectAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可创建流水线变量", ""))
		return
	}
	projectID := c.Param("id")
	pipelineID := c.Param("pipelineId")
	if projectID == "" || pipelineID == "" {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "请求参数错误", "project id and pipeline id are required"))
		return
	}
	var req CreateVariableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "请求参数错误", err.Error()))
		return
	}
	userID, _ := c.Get(middleware.ContextKeyUserID)
	v, err := h.service.CreateVariable(ScopePipeline, &projectID, &pipelineID, req, userID.(string))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, ToVariableResponse(v, true))
}

func (h *Handler) DeleteVariable(c *gin.Context) {
	if !isAdminOrProjectAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可删除变量", ""))
		return
	}

	vid := c.Param("vid")
	if err := h.service.DeleteVariable(vid); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

func isAdmin(c *gin.Context) bool {
	userRole, _ := c.Get(middleware.ContextKeyRole)
	role, ok := userRole.(string)
	return ok && role == "admin"
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

func getUserRole(c *gin.Context) string {
	userRole, _ := c.Get(middleware.ContextKeyRole)
	role, _ := userRole.(string)
	return role
}
