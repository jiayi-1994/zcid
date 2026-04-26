package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjy/zcid/pkg/response"
)

func TestHandlerListPassesCategoryFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var got ListOpts
	repo := &mockAuditRepo{
		list: func(ctx context.Context, opts ListOpts) ([]*AuditLog, int64, error) {
			got = opts
			return []*AuditLog{{ID: "1", Action: "auth.login", ResourceType: ResourceTypeAuthSecurity, Result: ResultSuccess}}, 1, nil
		},
	}

	r := gin.New()
	NewHandler(NewService(repo)).RegisterRoutes(r.Group("/api/v1/admin/audit-logs"))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs?category=auth_security&page=2&pageSize=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, got.Category)
	assert.Equal(t, ResourceTypeAuthSecurity, *got.Category)
	assert.Equal(t, 2, got.Page)
	assert.Equal(t, 10, got.PageSize)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(response.CodeSuccess), body["code"])
}
