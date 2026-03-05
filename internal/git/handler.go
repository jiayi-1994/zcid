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

func (h *Handler) GetConnection(c *gin.Context) {
	connID := c.Param("connId")
	conn, err := h.service.GetConnection(connID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, ToConnectionResponse(conn))
}

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

func (h *Handler) DeleteConnection(c *gin.Context) {
	connID := c.Param("connId")
	if err := h.service.DeleteConnection(connID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

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

func (h *Handler) GetWebhookSecret(c *gin.Context) {
	connID := c.Param("connId")
	secret, err := h.service.GetWebhookSecret(c.Request.Context(), connID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"webhookSecret": secret})
}

func (h *Handler) TestConnection(c *gin.Context) {
	connID := c.Param("connId")
	result, err := h.service.TestConnection(c.Request.Context(), connID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}
