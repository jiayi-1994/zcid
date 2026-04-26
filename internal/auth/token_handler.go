package auth

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/pkg/middleware"
	"github.com/xjy/zcid/pkg/response"
)

type TokenHandler struct {
	service *TokenService
}

func NewTokenHandler(service *TokenService) *TokenHandler {
	return &TokenHandler{service: service}
}

func (h *TokenHandler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/access-tokens", h.List)
	router.POST("/access-tokens", h.Create)
	router.POST("/access-tokens/:tokenId/revoke", h.Revoke)
}

func (h *TokenHandler) List(c *gin.Context) {
	actorID := contextUserID(c)
	includeProject := contextRole(c) == string(SystemRoleAdmin)
	tokens, err := h.service.List(c.Request.Context(), actorID, includeProject)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	items := make([]AccessTokenResponse, len(tokens))
	for i, token := range tokens {
		items[i] = toAccessTokenResponse(token)
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

func (h *TokenHandler) Create(c *gin.Context) {
	var req CreateAccessTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}
	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", "expiresAt must be RFC3339"))
		return
	}
	tokenType := AccessTokenType(strings.TrimSpace(req.Type))
	actorID := contextUserID(c)
	if tokenType == AccessTokenTypeProject && contextRole(c) != string(SystemRoleAdmin) {
		response.HandleError(c, response.NewBizError(response.CodeForbidden, "forbidden", "admin role required for project tokens"))
		return
	}
	created, err := h.service.Create(c.Request.Context(), CreateAccessTokenInput{
		Name:      req.Name,
		Type:      tokenType,
		Scopes:    req.Scopes,
		ExpiresAt: expiresAt,
		UserID:    actorID,
		ProjectID: req.ProjectID,
		ActorID:   actorID,
		IP:        c.ClientIP(),
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, CreateAccessTokenResponse{Token: toAccessTokenResponse(created.Token), Raw: created.Raw})
}

func (h *TokenHandler) Revoke(c *gin.Context) {
	tokenID := strings.TrimSpace(c.Param("tokenId"))
	if tokenID == "" {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", "token id is required"))
		return
	}
	if err := h.service.Revoke(c.Request.Context(), tokenID, contextUserID(c), contextRole(c), c.ClientIP()); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"revoked": true})
}

func toAccessTokenResponse(token *AccessToken) AccessTokenResponse {
	resp := AccessTokenResponse{
		ID:          token.ID,
		Type:        string(token.TokenType),
		Name:        token.Name,
		TokenPrefix: token.TokenPrefix,
		Scopes:      DecodeTokenScopes(token.Scopes),
		UserID:      token.UserID,
		ProjectID:   token.ProjectID,
		CreatedBy:   token.CreatedBy,
		ExpiresAt:   token.ExpiresAt.Format(time.RFC3339),
		CreatedAt:   token.CreatedAt.Format(time.RFC3339),
	}
	if token.LastUsedAt != nil {
		value := token.LastUsedAt.Format(time.RFC3339)
		resp.LastUsedAt = &value
	}
	if token.RevokedAt != nil {
		value := token.RevokedAt.Format(time.RFC3339)
		resp.RevokedAt = &value
	}
	return resp
}

func contextUserID(c *gin.Context) string {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := userID.(string)
	return uid
}

func contextRole(c *gin.Context) string {
	role, _ := c.Get(middleware.ContextKeyRole)
	value, _ := role.(string)
	return value
}
