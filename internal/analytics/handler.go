package analytics

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/pkg/response"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("", h.Get)
}

// Get godoc
// @Summary Get project pipeline analytics
// @Description Return run trends, duration percentiles, top failing steps, and top pipelines.
// @Tags analytics
// @Produce json
// @Param id path string true "Project ID"
// @Param range query string false "Range: 7d, 30d, 90d"
// @Success 200 {object} response.Response{data=Response}
// @Router /api/v1/projects/{id}/analytics [get]
func (h *Handler) Get(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "project id is required", ""))
		return
	}
	result, err := h.service.Get(c.Request.Context(), projectID, c.DefaultQuery("range", "7d"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, result)
}
