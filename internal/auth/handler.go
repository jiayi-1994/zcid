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
	router.GET("/bootstrap/status", h.BootstrapStatus)
	router.POST("/bootstrap/redeem", h.RedeemBootstrap)
	router.POST("/login", h.Login)
	router.POST("/refresh", h.Refresh)
	router.POST("/logout", h.Logout)
}

func (h *Handler) BootstrapStatus(c *gin.Context) {
	required, err := h.service.BootstrapRequired(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, BootstrapStatusResponse{Required: required})
}

func (h *Handler) RedeemBootstrap(c *gin.Context) {
	var req BootstrapRedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}

	user, err := h.service.RedeemBootstrapToken(ContextWithRequestIP(c.Request.Context(), c.ClientIP()), req.Token, req.Username, req.Password)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, UserResponse{ID: user.ID, Username: user.Username, Status: string(user.Status), Role: string(user.Role)})
}

func (h *Handler) RegisterAdminUserRoutes(router gin.IRoutes) {
	router.GET("/users", h.ListUsers)
	router.POST("/users", h.CreateUser)
	router.PUT("/users/:uid", h.UpdateUser)
	router.PUT("/users/:uid/role", h.AssignSystemRole)
}

// Login godoc
// @Summary User login
// @Description Authenticate a user with username and password, returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=TokenPair}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}

	pair, err := h.service.Login(ContextWithRequestIP(c.Request.Context(), c.ClientIP()), req.Username, req.Password)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, pair)
}

// Refresh godoc
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} response.Response{data=object{accessToken=string}}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}

	accessToken, err := h.service.Refresh(ContextWithRequestIP(c.Request.Context(), c.ClientIP()), req.RefreshToken)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"accessToken": accessToken})
}

// Logout godoc
// @Summary User logout
// @Description Invalidate the given refresh token to log out the user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Refresh token to invalidate"
// @Success 200 {object} response.Response{data=object{loggedOut=bool}}
// @Failure 400 {object} response.Response
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}

	if err := h.service.Logout(ContextWithRequestIP(c.Request.Context(), c.ClientIP()), req.RefreshToken); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"loggedOut": true})
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user account (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User creation payload"
// @Success 200 {object} response.Response{data=UserResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/users [post]
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

// UpdateUser godoc
// @Summary Update a user
// @Description Update an existing user's information (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param uid path string true "User ID"
// @Param request body UpdateUserRequest true "User update payload"
// @Success 200 {object} response.Response{data=UserResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/users/{uid} [put]
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

// AssignSystemRole godoc
// @Summary Assign system role to a user
// @Description Assign a system-level role to a user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param uid path string true "User ID"
// @Param request body AssignRoleRequest true "Role assignment payload"
// @Success 200 {object} response.Response{data=UserResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/users/{uid}/role [put]
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

// ListUsers godoc
// @Summary List all users
// @Description Retrieve a list of all users (admin only)
// @Tags users
// @Produce json
// @Success 200 {object} response.Response{data=[]UserResponse}
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/users [get]
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
