package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.POST("/login", h.Login)
	router.POST("/refresh", h.Refresh)
	router.POST("/logout", h.Logout)
}

func (h *Handler) RegisterAdminUserRoutes(router gin.IRoutes) {
	router.GET("/users", h.ListUsers)
	router.POST("/users", h.CreateUser)
	router.PUT("/users/:uid", h.UpdateUser)
	router.PUT("/users/:uid/role", h.AssignSystemRole)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}

	pair, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, pair)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}

	accessToken, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"accessToken": accessToken})
}

func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}

	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"loggedOut": true})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req.Username, req.Password, req.Status, req.Role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, UserResponse{ID: user.ID, Username: user.Username, Status: string(user.Status), Role: string(user.Role)})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}

	uid := strings.TrimSpace(c.Param("uid"))
	if uid == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", "uid is required"))
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), uid, req.Username, req.Password, req.Status, req.Role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, UserResponse{ID: user.ID, Username: user.Username, Status: string(user.Status), Role: string(user.Role)})
}

func (h *Handler) AssignSystemRole(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}

	uid := strings.TrimSpace(c.Param("uid"))
	if uid == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", "uid is required"))
		return
	}

	user, err := h.service.AssignSystemRole(c.Request.Context(), uid, req.Role)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, UserResponse{ID: user.ID, Username: user.Username, Status: string(user.Status), Role: string(user.Role)})
}

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.service.ListUsers(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result := make([]UserResponse, len(users))
	for i, user := range users {
		result[i] = UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Status:    string(user.Status),
			Role:      string(user.Role),
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	response.Success(c, result)
}
