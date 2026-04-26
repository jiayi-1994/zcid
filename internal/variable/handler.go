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

// CreateProjectVariable godoc
// @Summary Create a project variable
// @Description Create a new variable scoped to a project (admin or project admin only)
// @Tags variables
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body CreateVariableRequest true "Variable creation payload"
// @Success 200 {object} response.Response{data=VariableResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/projects/{id}/variables [post]
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
	v, err := h.service.CreateVariable(c.Request.Context(), ScopeProject, &projectID, nil, req, userID.(string))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToVariableResponse(v, true))
}

// CreateGlobalVariable godoc
// @Summary Create a global variable
// @Description Create a new global variable (admin only)
// @Tags variables
// @Accept json
// @Produce json
// @Param request body CreateVariableRequest true "Variable creation payload"
// @Success 200 {object} response.Response{data=VariableResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/variables [post]
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
	v, err := h.service.CreateVariable(c.Request.Context(), ScopeGlobal, nil, nil, req, userID.(string))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToVariableResponse(v, true))
}

// ListProjectVariables godoc
// @Summary List project variables
// @Description Retrieve all variables scoped to a project
// @Tags variables
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} response.Response{data=VariableListResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/projects/{id}/variables [get]
func (h *Handler) ListProjectVariables(c *gin.Context) {
	projectID := c.Param("id")
	vars, _, err := h.service.ListProjectVariables(c.Request.Context(), projectID)
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

// ListGlobalVariables godoc
// @Summary List global variables
// @Description Retrieve all global variables
// @Tags variables
// @Produce json
// @Success 200 {object} response.Response{data=VariableListResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/variables [get]
func (h *Handler) ListGlobalVariables(c *gin.Context) {
	vars, total, err := h.service.ListGlobalVariables(c.Request.Context())
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

// ListMergedVariables godoc
// @Summary List merged variables
// @Description Retrieve merged variables (global + project scope) for a project
// @Tags variables
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} response.Response{data=VariableListResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/projects/{id}/variables/merged [get]
func (h *Handler) ListMergedVariables(c *gin.Context) {
	projectID := c.Param("id")
	vars, err := h.service.GetMergedVariables(c.Request.Context(), projectID)
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

// UpdateVariable godoc
// @Summary Update a variable
// @Description Update an existing variable's value or description
// @Tags variables
// @Accept json
// @Produce json
// @Param vid path string true "Variable ID"
// @Param request body UpdateVariableRequest true "Variable update payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/variables/{vid} [put]
func (h *Handler) UpdateVariable(c *gin.Context) {
	vid := c.Param("vid")
	var req UpdateVariableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "请求参数错误", err.Error()))
		return
	}

	existing, err := h.service.GetVariable(c.Request.Context(), vid)
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

	if err := h.service.UpdateVariable(c.Request.Context(), vid, req, existing.VarType == TypeSecret); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListPipelineVariables godoc
// @Summary List pipeline variables
// @Description Retrieve all variables scoped to a specific pipeline
// @Tags variables
// @Produce json
// @Param id path string true "Project ID"
// @Param pipelineId path string true "Pipeline ID"
// @Success 200 {object} response.Response{data=VariableListResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId}/variables [get]
func (h *Handler) ListPipelineVariables(c *gin.Context) {
	projectID := c.Param("id")
	pipelineID := c.Param("pipelineId")
	if projectID == "" || pipelineID == "" {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "请求参数错误", "project id and pipeline id are required"))
		return
	}
	vars, total, err := h.service.ListPipelineVariables(c.Request.Context(), projectID, pipelineID)
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

// CreatePipelineVariable godoc
// @Summary Create a pipeline variable
// @Description Create a new variable scoped to a specific pipeline (admin or project admin only)
// @Tags variables
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param pipelineId path string true "Pipeline ID"
// @Param request body CreateVariableRequest true "Variable creation payload"
// @Success 200 {object} response.Response{data=VariableResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/projects/{id}/pipelines/{pipelineId}/variables [post]
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
	v, err := h.service.CreateVariable(c.Request.Context(), ScopePipeline, &projectID, &pipelineID, req, userID.(string))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, ToVariableResponse(v, true))
}

// DeleteVariable godoc
// @Summary Delete a variable
// @Description Delete a variable by its ID (admin or project admin only)
// @Tags variables
// @Produce json
// @Param vid path string true "Variable ID"
// @Success 200 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/variables/{vid} [delete]
func (h *Handler) DeleteVariable(c *gin.Context) {
	if !isAdminOrProjectAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可删除变量", ""))
		return
	}

	vid := c.Param("vid")
	if err := h.service.DeleteVariable(c.Request.Context(), vid); err != nil {
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
	projRole, _ := c.Get(middleware.ContextKeyProjectRole)
	if role, ok := projRole.(string); ok && role == "project_token" {
		return "project_token"
	}
	userRole, _ := c.Get(middleware.ContextKeyRole)
	role, _ := userRole.(string)
	return role
}
