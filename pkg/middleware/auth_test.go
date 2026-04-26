package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProgrammaticTokenValidator struct {
	requiredScope string
}

func (f *fakeProgrammaticTokenValidator) ValidateProgrammaticToken(ctx context.Context, raw string, requiredScope string, ip string) (*ProgrammaticTokenPrincipal, error) {
	f.requiredScope = requiredScope
	return &ProgrammaticTokenPrincipal{
		TokenID:   "tok_1",
		TokenType: "personal",
		UserID:    "user_1",
		Scopes:    []string{requiredScope},
	}, nil
}

func TestAdminJWTOrTokenReadAuth_AllowsReadOnlyAdminToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator := &fakeProgrammaticTokenValidator{}
	router := gin.New()
	router.Use(AdminJWTOrTokenReadAuth("secret", validator, "admin:read"))
	router.GET("/api/v1/admin/health", func(c *gin.Context) {
		assert.Equal(t, PrincipalTypeUserToken, c.GetString(ContextKeyPrincipalType))
		assert.Equal(t, "admin_token", c.GetString(ContextKeyRole))
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/health", nil)
	req.Header.Set("Authorization", "Bearer zcid_pat_test")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "admin:read", validator.requiredScope)
}

func TestAdminJWTOrTokenReadAuth_RejectsTokenWrites(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AdminJWTOrTokenReadAuth("secret", &fakeProgrammaticTokenValidator{}, "admin:read"))
	router.POST("/api/v1/admin/settings", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings", nil)
	req.Header.Set("Authorization", "Bearer zcid_pat_test")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAdminJWTOrTokenReadAuth_RequiresAdminJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AdminJWTOrTokenReadAuth("secret", &fakeProgrammaticTokenValidator{}, "admin:read"))
	router.GET("/api/v1/admin/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	memberReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/health", nil)
	memberReq.Header.Set("Authorization", "Bearer "+signedAccessToken(t, "secret", "member"))
	memberRec := httptest.NewRecorder()
	router.ServeHTTP(memberRec, memberReq)
	assert.Equal(t, http.StatusForbidden, memberRec.Code)

	adminReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/health", nil)
	adminReq.Header.Set("Authorization", "Bearer "+signedAccessToken(t, "secret", "admin"))
	adminRec := httptest.NewRecorder()
	router.ServeHTTP(adminRec, adminReq)
	assert.Equal(t, http.StatusNoContent, adminRec.Code)
}

func signedAccessToken(t *testing.T, secret string, role string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &accessTokenClaims{
		Username:  "tester",
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user_1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}
