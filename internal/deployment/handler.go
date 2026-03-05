package deployment

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/internal/environment"
	"github.com/xjy/zcid/pkg/middleware"
	"github.com/xjy/zcid/pkg/response"
)

// EnvLookup gets environment by ID and project for role checks.
type EnvLookup interface {
	Get(ctx context.Context, id, projectID string) (*environment.Environment, error)
}

type Handler struct {
	service   *Service
	envLookup EnvLookup
}

func NewHandler(service *Service, envLookup EnvLookup) *Handler {
	return &Handler{service: service, envLookup: envLookup}
}

func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.POST("", h.TriggerDeploy)
	router.GET("", h.List)
	router.GET("/environments/:envId/deploy-history", h.GetDeployHistory)
	router.GET("/:deployId", h.Get)
	router.GET("/:deployId/status", h.GetStatus)
	router.POST("/:deployId/resync", h.Resync)
	router.POST("/:deployId/rollback", h.Rollback)
}

func getProjectID(c *gin.Context) string {
	return strings.TrimSpace(c.Param("id"))
}

func getUserID(c *gin.Context) string {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := userID.(string)
	return uid
}

func getUserRole(c *gin.Context) string {
	role, _ := c.Get(middleware.ContextKeyRole)
	r, _ := role.(string)
	if r != "" {
		return r
	}
	projRole, _ := c.Get(middleware.ContextKeyProjectRole)
	pr, _ := projRole.(string)
	return pr
}

func (h *Handler) TriggerDeploy(c *gin.Context) {
	projectID := getProjectID(c)
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id is required"))
		return
	}
	uid := getUserID(c)
	if uid == "" {
		response.HandleError(c, response.NewBizError(response.CodeUnauthorized, "用户未认证", ""))
		return
	}
	var req TriggerDeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", err.Error()))
		return
	}
	if h.envLookup != nil {
		env, err := h.envLookup.Get(c.Request.Context(), req.EnvironmentID, projectID)
		if err != nil {
			response.HandleError(c, response.NewBizError(response.CodeNotFound, "环境不存在", ""))
			return
		}
		role := getUserRole(c)
		if !canDeployToEnv(role, env.Name) {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "无权部署到该环境", ""))
			return
		}
	}
	d, err := h.service.TriggerDeploy(c.Request.Context(), projectID, uid, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, ToDeploymentResponse(d))
}

func (h *Handler) List(c *gin.Context) {
	projectID := getProjectID(c)
	if projectID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id is required"))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := h.service.ListDeployments(c.Request.Context(), projectID, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	items := make([]DeploymentSummary, len(list))
	for i, d := range list {
		items[i] = ToDeploymentSummary(d)
	}
	response.Success(c, DeploymentListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *Handler) Get(c *gin.Context) {
	projectID := getProjectID(c)
	deployID := strings.TrimSpace(c.Param("deployId"))
	if projectID == "" || deployID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and deploy id are required"))
		return
	}
	d, err := h.service.GetDeployment(c.Request.Context(), projectID, deployID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, ToDeploymentResponse(d))
}

func (h *Handler) GetStatus(c *gin.Context) {
	projectID := getProjectID(c)
	deployID := strings.TrimSpace(c.Param("deployId"))
	if projectID == "" || deployID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and deploy id are required"))
		return
	}
	d, err := h.service.RefreshStatus(c.Request.Context(), projectID, deployID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, ToDeploymentResponse(d))
}

func (h *Handler) Resync(c *gin.Context) {
	projectID := getProjectID(c)
	deployID := strings.TrimSpace(c.Param("deployId"))
	if projectID == "" || deployID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and deploy id are required"))
		return
	}
	if h.envLookup != nil {
		d, getErr := h.service.GetDeployment(c.Request.Context(), projectID, deployID)
		if getErr != nil {
			response.HandleError(c, getErr)
			return
		}
		env, err := h.envLookup.Get(c.Request.Context(), d.EnvironmentID, projectID)
		if err != nil {
			response.HandleError(c, response.NewBizError(response.CodeNotFound, "环境不存在", ""))
			return
		}
		role := getUserRole(c)
		if !canDeployToEnv(role, env.Name) {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "无权操作该环境的部署", ""))
			return
		}
	}
	d, err := h.service.ResyncDeploy(c.Request.Context(), projectID, deployID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, ToDeploymentResponse(d))
}

func (h *Handler) Rollback(c *gin.Context) {
	projectID := getProjectID(c)
	deployID := strings.TrimSpace(c.Param("deployId"))
	if projectID == "" || deployID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and deploy id are required"))
		return
	}
	uid := getUserID(c)
	if uid == "" {
		response.HandleError(c, response.NewBizError(response.CodeUnauthorized, "用户未认证", ""))
		return
	}
	if h.envLookup != nil {
		d, getErr := h.service.GetDeployment(c.Request.Context(), projectID, deployID)
		if getErr != nil {
			response.HandleError(c, getErr)
			return
		}
		env, err := h.envLookup.Get(c.Request.Context(), d.EnvironmentID, projectID)
		if err != nil {
			response.HandleError(c, response.NewBizError(response.CodeNotFound, "环境不存在", ""))
			return
		}
		role := getUserRole(c)
		if !canDeployToEnv(role, env.Name) {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "无权回滚该环境的部署", ""))
			return
		}
	}
	d, err := h.service.RollbackDeploy(c.Request.Context(), projectID, deployID, uid)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, ToDeploymentResponse(d))
}

func (h *Handler) GetDeployHistory(c *gin.Context) {
	projectID := getProjectID(c)
	envID := strings.TrimSpace(c.Param("envId"))
	if projectID == "" || envID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "请求参数错误", "project id and environment id are required"))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := h.service.GetDeployHistory(c.Request.Context(), projectID, envID, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	items := make([]DeploymentSummary, len(list))
	for i, d := range list {
		items[i] = ToDeploymentSummary(d)
	}
	response.Success(c, DeploymentListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
