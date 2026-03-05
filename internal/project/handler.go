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
