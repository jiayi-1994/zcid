package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/pkg/middleware"
	"github.com/xjy/zcid/pkg/response"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newAdminUserRouter(repo Repository) *gin.Engine {
	handler := NewHandler(NewService(repo, "test-secret"))
	r := gin.New()
	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.RequireAdminRBAC("test-secret"))
	handler.RegisterAdminUserRoutes(admin)
	return r
}

func accessTokenForRole(role SystemRole) string {
	svc := NewService(newMockRepo(), "test-secret")
	token, _ := svc.signToken("user-id", "user", string(role), "access", AccessTokenTTL)
	return token
}

func performJSONRequest(r *gin.Engine, method string, path string, body any, admin bool) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if admin {
		req.Header.Set("Authorization", "Bearer "+accessTokenForRole(SystemRoleAdmin))
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}


func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	return body
}

func TestCreateUserAdminSuccess(t *testing.T) {
	repo := newMockRepo()
	r := newAdminUserRouter(repo)

	w := performJSONRequest(r, http.MethodPost, "/api/v1/admin/users", map[string]any{
		"username": "alice",
		"password": "pass123",
		"status":   "active",
	}, true)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := decodeBody(t, w)
	if body["code"] != float64(response.CodeSuccess) {
		t.Fatalf("expected success code, got %v", body["code"])
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", body["data"])
	}
	if data["username"] != "alice" {
		t.Fatalf("expected username alice, got %v", data["username"])
	}
	if data["status"] != string(UserStatusActive) {
		t.Fatalf("expected status active, got %v", data["status"])
	}
	if data["id"] == "" {
		t.Fatal("expected non-empty user id")
	}

	if repo.createCalls != 1 {
		t.Fatalf("expected createCalls=1, got %d", repo.createCalls)
	}
	if stored := repo.usersByName["alice"]; stored == nil || stored.PasswordHash == "pass123" {
		t.Fatal("expected stored password to be hashed")
	}
}

func TestCreateUserForbiddenWithoutAdminHeader(t *testing.T) {
	repo := newMockRepo()
	r := newAdminUserRouter(repo)

	w := performJSONRequest(r, http.MethodPost, "/api/v1/admin/users", map[string]any{
		"username": "alice",
		"password": "pass123",
	}, false)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}

	body := decodeBody(t, w)
	if body["code"] != float64(response.CodeForbidden) {
		t.Fatalf("expected forbidden code, got %v", body["code"])
	}
}

func TestCreateUserValidationError(t *testing.T) {
	repo := newMockRepo()
	r := newAdminUserRouter(repo)

	w := performJSONRequest(r, http.MethodPost, "/api/v1/admin/users", map[string]any{
		"username": "alice",
	}, true)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	body := decodeBody(t, w)
	if body["code"] != float64(response.CodeValidation) {
		t.Fatalf("expected validation code, got %v", body["code"])
	}
}

func TestUpdateUserAdminSuccess(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive})
	repo.sessions["u-1"] = "refresh-token"
	r := newAdminUserRouter(repo)

	w := performJSONRequest(r, http.MethodPut, "/api/v1/admin/users/u-1", map[string]any{
		"username": "alice-new",
		"status":   string(UserStatusDisabled),
	}, true)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := decodeBody(t, w)
	if body["code"] != float64(response.CodeSuccess) {
		t.Fatalf("expected success code, got %v", body["code"])
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", body["data"])
	}
	if data["username"] != "alice-new" {
		t.Fatalf("expected updated username, got %v", data["username"])
	}
	if data["status"] != string(UserStatusDisabled) {
		t.Fatalf("expected disabled status, got %v", data["status"])
	}
	if _, exists := repo.sessions["u-1"]; exists {
		t.Fatal("expected refresh session deleted after disabling user")
	}
}

func TestUpdateUserForbiddenWithoutAdminHeader(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive})
	r := newAdminUserRouter(repo)

	w := performJSONRequest(r, http.MethodPut, "/api/v1/admin/users/u-1", map[string]any{
		"username": "alice-new",
	}, false)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}

	body := decodeBody(t, w)
	if body["code"] != float64(response.CodeForbidden) {
		t.Fatalf("expected forbidden code, got %v", body["code"])
	}
}

func TestAssignRoleAdminSuccess(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive, Role: SystemRoleMember})
	r := newAdminUserRouter(repo)

	w := performJSONRequest(r, http.MethodPut, "/api/v1/admin/users/u-1/role", map[string]any{
		"role": string(SystemRoleAdmin),
	}, true)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := decodeBody(t, w)
	if body["code"] != float64(response.CodeSuccess) {
		t.Fatalf("expected success code, got %v", body["code"])
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", body["data"])
	}
	if data["role"] != string(SystemRoleAdmin) {
		t.Fatalf("expected role admin, got %v", data["role"])
	}
}

func TestAssignRoleValidationError(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive, Role: SystemRoleMember})
	r := newAdminUserRouter(repo)

	w := performJSONRequest(r, http.MethodPut, "/api/v1/admin/users/u-1/role", map[string]any{
		"role": "invalid-role",
	}, true)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	body := decodeBody(t, w)
	if body["code"] != float64(response.CodeValidation) {
		t.Fatalf("expected validation code, got %v", body["code"])
	}
}
