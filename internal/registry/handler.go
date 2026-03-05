package registry

import (
	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/pkg/middleware"
	"github.com/xjy/zcid/pkg/response"
)

// Handler handles registry HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new Handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers registry routes under the given router
func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("", h.List)
	router.POST("", h.Create)
	router.POST("/test-connection", h.TestConnection)
	router.GET("/:id", h.Get)
	router.PUT("/:id", h.Update)
	router.DELETE("/:id", h.Delete)
}

// List returns all registries
func (h *Handler) List(c *gin.Context) {
	regs, total, err := h.service.List()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	items := make([]RegistryResponse, len(regs))
	for i := range regs {
		items[i] = ToRegistryResponse(&regs[i])
	}
	response.Success(c, RegistryListResponse{Items: items, Total: total})
}

// Create creates a new registry
func (h *Handler) Create(c *gin.Context) {
	var req CreateRegistryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "invalid request", err.Error()))
		return
	}

	userIDVal, ok := c.Get(middleware.ContextKeyUserID)
	if !ok {
		response.HandleError(c, response.NewBizError(response.CodeUnauthorized, "未认证", ""))
		return
	}
	uid, ok := userIDVal.(string)
	if !ok || uid == "" {
		response.HandleError(c, response.NewBizError(response.CodeUnauthorized, "未认证", ""))
		return
	}
	reg, err := h.service.Create(req, uid)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToRegistryResponse(reg))
}

// Get returns a registry by ID
func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	reg, err := h.service.Get(id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToRegistryResponse(reg))
}

// Update updates a registry
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateRegistryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "invalid request", err.Error()))
		return
	}

	reg, err := h.service.Update(id, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToRegistryResponse(reg))
}

// Delete deletes a registry
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(id); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// TestConnection tests connectivity to a registry
func (h *Handler) TestConnection(c *gin.Context) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "invalid request", err.Error()))
		return
	}

	res, err := h.service.TestConnection(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, res)
}

// ToRegistryResponse converts a Registry to RegistryResponse (password never exposed)
func ToRegistryResponse(r *Registry) RegistryResponse {
	return RegistryResponse{
		ID:        r.ID,
		Name:      r.Name,
		Type:      string(r.Type),
		URL:       r.URL,
		Username:  r.Username,
		IsDefault: r.IsDefault,
		Status:    string(r.Status),
		CreatedBy: r.CreatedBy,
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
