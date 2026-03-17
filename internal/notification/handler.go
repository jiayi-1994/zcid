package notification

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
	router.GET("/:ruleId", h.Get)
	router.PUT("/:ruleId", h.Update)
	router.DELETE("/:ruleId", h.Delete)
}

func getProjectID(c *gin.Context) string {
	return strings.TrimSpace(c.Param("id"))
}

func getUserID(c *gin.Context) string {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := userID.(string)
	return uid
}

// Create godoc
// @Summary Create a notification rule
// @Description Create a new notification rule for a project
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body CreateRuleRequest true "Notification rule creation payload"
// @Success 200 {object} response.Response{data=RuleResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/projects/{id}/notification-rules [post]
func (h *Handler) Create(c *gin.Context) {
	projectID := getProjectID(c)
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "project id is required", ""))
		return
	}
	uid := getUserID(c)
	if uid == "" {
		response.HandleError(c, response.NewBizError(response.CodeUnauthorized, "user not authenticated", ""))
		return
	}
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}
	rule, err := h.service.Create(c.Request.Context(), projectID, uid, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, ToRuleResponse(rule))
}

// Get godoc
// @Summary Get a notification rule
// @Description Retrieve a notification rule by its ID
// @Tags notifications
// @Produce json
// @Param id path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Success 200 {object} response.Response{data=RuleResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/notification-rules/{ruleId} [get]
func (h *Handler) Get(c *gin.Context) {
	projectID := getProjectID(c)
	ruleID := strings.TrimSpace(c.Param("ruleId"))
	if projectID == "" || ruleID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "project id and rule id are required", ""))
		return
	}
	rule, err := h.service.Get(c.Request.Context(), projectID, ruleID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, ToRuleResponse(rule))
}

// List godoc
// @Summary List notification rules
// @Description Retrieve a paginated list of notification rules for a project
// @Tags notifications
// @Produce json
// @Param id path string true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=object{items=[]RuleResponse,total=int,page=int,pageSize=int}}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{id}/notification-rules [get]
func (h *Handler) List(c *gin.Context) {
	projectID := getProjectID(c)
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "project id is required", ""))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := h.service.List(c.Request.Context(), projectID, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	items := make([]RuleResponse, len(list))
	for i, r := range list {
		items[i] = ToRuleResponse(r)
	}
	response.Success(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Update godoc
// @Summary Update a notification rule
// @Description Update an existing notification rule
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Param request body UpdateRuleRequest true "Notification rule update payload"
// @Success 200 {object} response.Response{data=RuleResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/notification-rules/{ruleId} [put]
func (h *Handler) Update(c *gin.Context) {
	projectID := getProjectID(c)
	ruleID := strings.TrimSpace(c.Param("ruleId"))
	if projectID == "" || ruleID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "project id and rule id are required", ""))
		return
	}
	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}
	rule, err := h.service.Update(c.Request.Context(), projectID, ruleID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, ToRuleResponse(rule))
}

// Delete godoc
// @Summary Delete a notification rule
// @Description Delete a notification rule from a project
// @Tags notifications
// @Produce json
// @Param id path string true "Project ID"
// @Param ruleId path string true "Rule ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/projects/{id}/notification-rules/{ruleId} [delete]
func (h *Handler) Delete(c *gin.Context) {
	projectID := getProjectID(c)
	ruleID := strings.TrimSpace(c.Param("ruleId"))
	if projectID == "" || ruleID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "project id and rule id are required", ""))
		return
	}
	if err := h.service.Delete(c.Request.Context(), projectID, ruleID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, nil)
}
