package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xjy/zcid/pkg/response"
)

const (
	ContextKeyUserID          = "userId"
	ContextKeyUsername        = "username"
	ContextKeyRole            = "role"
	ContextKeyPrincipalType   = "principalType"
	ContextKeyTokenID         = "tokenId"
	ContextKeyTokenScopes     = "tokenScopes"
	ContextKeyTokenProjectID  = "tokenProjectId"
	PrincipalTypeUser         = "user"
	PrincipalTypeUserToken    = "user_token"
	PrincipalTypeProjectToken = "project_token"
)

type ProgrammaticTokenPrincipal struct {
	TokenID   string
	TokenType string
	UserID    string
	ProjectID string
	Scopes    []string
}

type ProgrammaticTokenValidator interface {
	ValidateProgrammaticToken(ctx context.Context, raw string, requiredScope string, ip string) (*ProgrammaticTokenPrincipal, error)
}

type TokenScopeResolver func(c *gin.Context) (scope string, ok bool)

type accessTokenClaims struct {
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

func JWTAuth(jwtSecret string) gin.HandlerFunc {
	secret := []byte(strings.TrimSpace(jwtSecret))
	return func(c *gin.Context) {
		claims, ok := parseAccessToken(c.GetHeader("Authorization"), secret)
		if !ok {
			response.HandleError(c, response.NewBizError(response.CodeUnauthorized, "unauthorized", ""))
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.Subject)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

func JWTOrTokenAuth(jwtSecret string, tokenService ProgrammaticTokenValidator, requiredScope string) gin.HandlerFunc {
	return JWTOrMappedTokenAuth(jwtSecret, tokenService, func(*gin.Context) (string, bool) {
		return requiredScope, strings.TrimSpace(requiredScope) != ""
	})
}

func JWTOrMappedTokenAuth(jwtSecret string, tokenService ProgrammaticTokenValidator, resolveScope TokenScopeResolver) gin.HandlerFunc {
	secret := []byte(strings.TrimSpace(jwtSecret))
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		bearer := bearerToken(authHeader)
		if isProgrammaticToken(bearer) {
			requiredScope, ok := resolveScope(c)
			if !ok || strings.TrimSpace(requiredScope) == "" {
				response.HandleError(c, response.NewBizError(response.CodeForbidden, "forbidden", "programmatic tokens are not allowed for this route"))
				c.Abort()
				return
			}
			if tokenService == nil {
				response.HandleError(c, response.NewBizError(response.CodeInternalServerError, "internal server error", ""))
				c.Abort()
				return
			}
			principal, err := tokenService.ValidateProgrammaticToken(c.Request.Context(), bearer, requiredScope, c.ClientIP())
			if err != nil {
				response.HandleError(c, err)
				c.Abort()
				return
			}
			if principal.TokenType == "personal" {
				c.Set(ContextKeyPrincipalType, PrincipalTypeUserToken)
				c.Set(ContextKeyUserID, principal.UserID)
				c.Set(ContextKeyRole, "member")
			} else {
				c.Set(ContextKeyPrincipalType, PrincipalTypeProjectToken)
				c.Set(ContextKeyTokenProjectID, principal.ProjectID)
			}
			c.Set(ContextKeyTokenID, principal.TokenID)
			c.Set(ContextKeyTokenScopes, principal.Scopes)
			c.Next()
			return
		}

		claims, ok := parseAccessToken(authHeader, secret)
		if !ok {
			response.HandleError(c, response.NewBizError(response.CodeUnauthorized, "unauthorized", ""))
			c.Abort()
			return
		}
		c.Set(ContextKeyPrincipalType, PrincipalTypeUser)
		c.Set(ContextKeyUserID, claims.Subject)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

func AdminJWTOrTokenReadAuth(jwtSecret string, tokenService ProgrammaticTokenValidator, requiredScope string) gin.HandlerFunc {
	secret := []byte(strings.TrimSpace(jwtSecret))
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		bearer := bearerToken(authHeader)
		if isProgrammaticToken(bearer) {
			if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
				response.HandleError(c, response.NewBizError(response.CodeForbidden, "forbidden", "admin read token cannot modify resources"))
				c.Abort()
				return
			}
			if strings.TrimSpace(requiredScope) == "" {
				response.HandleError(c, response.NewBizError(response.CodeForbidden, "forbidden", "programmatic tokens are not allowed for this route"))
				c.Abort()
				return
			}
			if tokenService == nil {
				response.HandleError(c, response.NewBizError(response.CodeInternalServerError, "internal server error", ""))
				c.Abort()
				return
			}
			principal, err := tokenService.ValidateProgrammaticToken(c.Request.Context(), bearer, requiredScope, c.ClientIP())
			if err != nil {
				response.HandleError(c, err)
				c.Abort()
				return
			}
			if principal.TokenType != "personal" {
				response.HandleError(c, response.NewBizError(response.CodeForbidden, "forbidden", "admin read requires a personal access token"))
				c.Abort()
				return
			}
			c.Set(ContextKeyPrincipalType, PrincipalTypeUserToken)
			c.Set(ContextKeyUserID, principal.UserID)
			c.Set(ContextKeyRole, "admin_token")
			c.Set(ContextKeyTokenID, principal.TokenID)
			c.Set(ContextKeyTokenScopes, principal.Scopes)
			c.Next()
			return
		}

		claims, ok := parseAccessToken(authHeader, secret)
		if !ok || !strings.EqualFold(claims.Role, "admin") {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "forbidden", "admin role required"))
			c.Abort()
			return
		}

		c.Set(ContextKeyPrincipalType, PrincipalTypeUser)
		c.Set(ContextKeyUserID, claims.Subject)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

func RequireCasbinRBAC(jwtSecret string, enforcer *casbin.SyncedEnforcer) gin.HandlerFunc {
	secret := []byte(strings.TrimSpace(jwtSecret))
	return func(c *gin.Context) {
		claims, ok := parseAccessToken(c.GetHeader("Authorization"), secret)
		if !ok {
			response.HandleError(c, response.NewBizError(response.CodeUnauthorized, "unauthorized", ""))
			c.Abort()
			return
		}
		if enforcer == nil {
			response.HandleError(c, response.NewBizError(response.CodeInternalServerError, "internal server error", ""))
			c.Abort()
			return
		}

		projectID := strings.TrimSpace(c.GetHeader("X-Project-ID"))
		if projectID == "" {
			projectID = "*"
		}

		allowed, err := enforcer.Enforce(claims.Subject, projectID, c.Request.URL.Path, c.Request.Method)
		if err != nil {
			response.HandleError(c, response.NewBizError(response.CodeInternalServerError, "internal server error", ""))
			c.Abort()
			return
		}
		if !allowed {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "forbidden", "insufficient permissions"))
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.Subject)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

func RequireAdminRBAC(jwtSecret string) gin.HandlerFunc {
	secret := []byte(strings.TrimSpace(jwtSecret))
	return func(c *gin.Context) {
		claims, ok := parseAccessToken(c.GetHeader("Authorization"), secret)
		if !ok || !strings.EqualFold(claims.Role, "admin") {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "forbidden", "admin role required"))
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.Subject)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

// ParseTokenForWebSocket parses JWT from raw token string (e.g. query param).
// Returns userID and ok. Use for WebSocket handshakes that pass token via ?token=xxx.
func ParseTokenForWebSocket(tokenString string, secret []byte) (userID string, ok bool) {
	claims, ok := parseTokenFromString(tokenString, secret)
	if !ok {
		return "", false
	}
	return claims.Subject, true
}

func parseTokenFromString(tokenString string, secret []byte) (*accessTokenClaims, bool) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" || len(secret) == 0 {
		return nil, false
	}
	claims := &accessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, false
	}
	if claims.TokenType != "access" || claims.Subject == "" {
		return nil, false
	}
	return claims, true
}

func parseAccessToken(authHeader string, secret []byte) (*accessTokenClaims, bool) {
	tokenString := bearerToken(authHeader)
	return parseTokenFromString(tokenString, secret)
}

func bearerToken(authHeader string) string {
	authHeader = strings.TrimSpace(authHeader)
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return ""
	}
	return strings.TrimSpace(authHeader[len("Bearer "):])
}

func isProgrammaticToken(token string) bool {
	return strings.HasPrefix(token, "zcid_pat_") || strings.HasPrefix(token, "zcid_proj_")
}
