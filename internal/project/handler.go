package project

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

func (h *Handler) RegisterCollectionRoutes(router gin.IRoutes) {
	router.POST("", h.CreateProject)
	router.GET("", h.ListProjects)
}

func (h *Handler) RegisterResourceRoutes(router gin.IRoutes) {
	router.GET("", h.GetProject)
	router.PUT("", h.UpdateProject)
	router.DELETE("", h.DeleteProject)
}

func (h *Handler) RegisterMemberRoutes(router gin.IRoutes) {
	router.GET("", h.ListMembers)
	router.POST("", h.AddMember)
	router.PUT("/:uid", h.UpdateMemberRole)
	router.DELETE("/:uid", h.RemoveMember)
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

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project (admin only)
// @Tags projects
// @Accept json
// @Produce json
// @Param request body CreateProjectRequest true "Project creation payload"
// @Success 200 {object} response.Response{data=ProjectResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/projects [post]
func (h *Handler) CreateProject(c *gin.Context) {
	userRole, _ := c.Get(middleware.ContextKeyRole)
	if role, ok := userRole.(string); !ok || role != "admin" {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员可创建项目", ""))
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}

	userID, _ := c.Get(middleware.ContextKeyUserID)
	ownerID, _ := userID.(string)

	p, err := h.service.CreateProject(c.Request.Context(), req.Name, req.Description, ownerID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToProjectResponse(p))
}

// GetProject godoc
// @Summary Get a project
// @Description Retrieve a project by its ID
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} response.Response{data=ProjectResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id} [get]
func (h *Handler) GetProject(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "id is required"))
		return
	}

	p, err := h.service.GetProject(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToProjectResponse(p))
}

// ListProjects godoc
// @Summary List projects
// @Description Retrieve a paginated list of projects the current user has access to
// @Tags projects
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=ProjectListResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/projects [get]
func (h *Handler) ListProjects(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	userRole, _ := c.Get(middleware.ContextKeyRole)

	uid, _ := userID.(string)
	role, _ := userRole.(string)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	projects, total, err := h.service.ListProjects(c.Request.Context(), uid, role, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	items := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		items[i] = ToProjectResponse(p)
	}

	response.Success(c, ProjectListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// UpdateProject godoc
// @Summary Update a project
// @Description Update an existing project's name or description
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body UpdateProjectRequest true "Project update payload"
// @Success 200 {object} response.Response{data=ProjectResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id} [put]
func (h *Handler) UpdateProject(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "id is required"))
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}

	p, err := h.service.UpdateProject(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToProjectResponse(p))
}

// DeleteProject godoc
// @Summary Delete a project
// @Description Delete a project by its ID (admin only)
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} response.Response{data=object{deleted=bool}}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/projects/{id} [delete]
func (h *Handler) DeleteProject(c *gin.Context) {
	userRole, _ := c.Get(middleware.ContextKeyRole)
	if role, ok := userRole.(string); !ok || role != "admin" {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员可删除项目", ""))
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "id is required"))
		return
	}

	if err := h.service.DeleteProject(c.Request.Context(), id); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// ListMembers godoc
// @Summary List project members
// @Description Retrieve all members of a project
// @Tags members
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} response.Response{data=MemberListResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{id}/members [get]
func (h *Handler) ListMembers(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id is required"))
		return
	}

	members, err := h.service.ListMembers(c.Request.Context(), projectID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	items := make([]MemberResponse, len(members))
	for i, m := range members {
		items[i] = MemberResponse{
			UserID:   m.UserID,
			Username: m.Username,
			Role:     string(m.Role),
			JoinedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	response.Success(c, MemberListResponse{
		Items: items,
		Total: int64(len(items)),
	})
}

// AddMember godoc
// @Summary Add a member to a project
// @Description Add a user as a member of a project (admin or project admin only)
// @Tags members
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body AddMemberRequest true "Member addition payload"
// @Success 200 {object} response.Response{data=object{added=bool}}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/projects/{id}/members [post]
func (h *Handler) AddMember(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id is required"))
		return
	}

	if !isAdminOrProjectAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可添加成员", ""))
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}

	if err := h.service.AddMember(c.Request.Context(), projectID, req.UserID, req.Role); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"added": true})
}

// UpdateMemberRole godoc
// @Summary Update a member's role
// @Description Update the role of a project member (admin or project admin only)
// @Tags members
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param uid path string true "User ID"
// @Param request body UpdateMemberRoleRequest true "Role update payload"
// @Success 200 {object} response.Response{data=object{updated=bool}}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/projects/{id}/members/{uid} [put]
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	userID := strings.TrimSpace(c.Param("uid"))
	if projectID == "" || userID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and user id are required"))
		return
	}

	if !isAdminOrProjectAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可修改成员角色", ""))
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}

	if err := h.service.UpdateMemberRole(c.Request.Context(), projectID, userID, req.Role); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}

// RemoveMember godoc
// @Summary Remove a member from a project
// @Description Remove a user from a project's membership (admin or project admin only)
// @Tags members
// @Produce json
// @Param id path string true "Project ID"
// @Param uid path string true "User ID"
// @Success 200 {object} response.Response{data=object{removed=bool}}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/projects/{id}/members/{uid} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	userID := strings.TrimSpace(c.Param("uid"))
	if projectID == "" || userID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and user id are required"))
		return
	}

	if !isAdminOrProjectAdmin(c) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "仅管理员或项目管理员可移除成员", ""))
		return
	}

	if err := h.service.RemoveMember(c.Request.Context(), projectID, userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"removed": true})
}
