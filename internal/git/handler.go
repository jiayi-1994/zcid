package git

import (
	"net/url"
	"strconv"

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
	router.GET("", h.ListConnections)
	router.POST("", h.CreateConnection)
	router.GET("/:connId", h.GetConnection)
	router.PUT("/:connId", h.UpdateConnection)
	router.DELETE("/:connId", h.DeleteConnection)
	router.POST("/:connId/test", h.TestConnection)
	router.GET("/:connId/webhook-secret", h.GetWebhookSecret)
	router.GET("/:connId/repos", h.ListRepos)
	router.GET("/:connId/repos/*repoPath", h.ListBranches)
}

// CreateConnection godoc
// @Summary Create a git connection
// @Description Create a new git provider connection (admin only)
// @Tags git-integrations
// @Accept json
// @Produce json
// @Param request body CreateConnectionRequest true "Connection creation payload"
// @Success 200 {object} response.Response{data=ConnectionResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/admin/integrations [post]
func (h *Handler) CreateConnection(c *gin.Context) {
	var req CreateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "请求参数错误", err.Error()))
		return
	}

	userID, _ := c.Get(middleware.ContextKeyUserID)
	conn, err := h.service.CreateConnection(req, userID.(string))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToConnectionResponse(conn))
}

// ListConnections godoc
// @Summary List git connections
// @Description Retrieve all git provider connections
// @Tags git-integrations
// @Produce json
// @Success 200 {object} response.Response{data=ConnectionListResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/integrations [get]
func (h *Handler) ListConnections(c *gin.Context) {
	conns, total, err := h.service.ListConnections()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	items := make([]ConnectionResponse, len(conns))
	for i, conn := range conns {
		items[i] = ToConnectionResponse(&conn)
	}
	response.Success(c, ConnectionListResponse{Items: items, Total: total})
}

// GetConnection godoc
// @Summary Get a git connection
// @Description Retrieve a git provider connection by its ID
// @Tags git-integrations
// @Produce json
// @Param connId path string true "Connection ID"
// @Success 200 {object} response.Response{data=ConnectionResponse}
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/integrations/{connId} [get]
func (h *Handler) GetConnection(c *gin.Context) {
	connID := c.Param("connId")
	conn, err := h.service.GetConnection(connID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToConnectionResponse(conn))
}

// UpdateConnection godoc
// @Summary Update a git connection
// @Description Update an existing git provider connection
// @Tags git-integrations
// @Accept json
// @Produce json
// @Param connId path string true "Connection ID"
// @Param request body UpdateConnectionRequest true "Connection update payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/integrations/{connId} [put]
func (h *Handler) UpdateConnection(c *gin.Context) {
	connID := c.Param("connId")
	var req UpdateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "请求参数错误", err.Error()))
		return
	}

	if err := h.service.UpdateConnection(connID, req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteConnection godoc
// @Summary Delete a git connection
// @Description Delete a git provider connection
// @Tags git-integrations
// @Produce json
// @Param connId path string true "Connection ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/integrations/{connId} [delete]
func (h *Handler) DeleteConnection(c *gin.Context) {
	connID := c.Param("connId")
	if err := h.service.DeleteConnection(connID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListRepos godoc
// @Summary List repositories
// @Description List repositories available through a git connection
// @Tags git-integrations
// @Produce json
// @Param connId path string true "Connection ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param refresh query bool false "Force refresh from provider"
// @Success 200 {object} response.Response{data=object{items=[]object,total=int,page=int,pageSize=int}}
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/integrations/{connId}/repos [get]
func (h *Handler) ListRepos(c *gin.Context) {
	connID := c.Param("connId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	refresh := c.Query("refresh") == "true"

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	repos, total, err := h.service.ListRepos(c.Request.Context(), connID, page, pageSize, refresh)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"items": repos, "total": total, "page": page, "pageSize": pageSize})
}

// ListBranches godoc
// @Summary List branches
// @Description List branches of a repository through a git connection
// @Tags git-integrations
// @Produce json
// @Param connId path string true "Connection ID"
// @Param repoPath path string true "Repository path (owner/repo/branches)"
// @Param refresh query bool false "Force refresh from provider"
// @Success 200 {object} response.Response{data=object{items=[]string}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/integrations/{connId}/repos/{repoPath} [get]
func (h *Handler) ListBranches(c *gin.Context) {
	connID := c.Param("connId")
	repoPath := c.Param("repoPath")

	// repoPath comes as "/<owner>/<repo>/branches" from the wildcard route
	// Extract the repo full name by trimming prefix "/" and suffix "/branches"
	repoFullName := extractRepoFullName(repoPath)
	if repoFullName == "" {
		response.HandleError(c, response.NewBizError(response.CodeBadRequest, "仓库名称无效", repoPath))
		return
	}

	refresh := c.Query("refresh") == "true"

	branches, err := h.service.ListBranches(c.Request.Context(), connID, repoFullName, refresh)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"items": branches})
}

// extractRepoFullName extracts "owner/repo" from path like "/owner/repo/branches"
func extractRepoFullName(path string) string {
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	// Remove trailing "/branches" if present
	const suffix = "/branches"
	if len(path) > len(suffix) && path[len(path)-len(suffix):] == suffix {
		path = path[:len(path)-len(suffix)]
	}
	decoded, err := url.PathUnescape(path)
	if err != nil {
		return path
	}
	return decoded
}

// GetWebhookSecret godoc
// @Summary Get webhook secret
// @Description Retrieve the webhook secret for a git connection
// @Tags git-integrations
// @Produce json
// @Param connId path string true "Connection ID"
// @Success 200 {object} response.Response{data=object{webhookSecret=string}}
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/integrations/{connId}/webhook-secret [get]
func (h *Handler) GetWebhookSecret(c *gin.Context) {
	connID := c.Param("connId")
	secret, err := h.service.GetWebhookSecret(c.Request.Context(), connID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"webhookSecret": secret})
}

// TestConnection godoc
// @Summary Test a git connection
// @Description Test connectivity to a git provider
// @Tags git-integrations
// @Produce json
// @Param connId path string true "Connection ID"
// @Success 200 {object} response.Response{data=TestConnectionResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/integrations/{connId}/test [post]
func (h *Handler) TestConnection(c *gin.Context) {
	connID := c.Param("connId")
	result, err := h.service.TestConnection(c.Request.Context(), connID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}
