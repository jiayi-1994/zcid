package middleware

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xjy/zcid/pkg/response"
)

const (
	ContextKeyUserID   = "userId"
	ContextKeyUsername = "username"
	ContextKeyRole     = "role"
)

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
	authHeader = strings.TrimSpace(authHeader)
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return nil, false
	}
	tokenString := strings.TrimSpace(authHeader[len("Bearer "):])
	return parseTokenFromString(tokenString, secret)
}
