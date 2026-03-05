package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorRecoveryHandlesPanic(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.Use(ErrorRecovery())
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set("X-Request-ID", "req-panic-test")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	if got := w.Header().Get("X-Request-ID"); got != "req-panic-test" {
		t.Fatalf("expected X-Request-ID req-panic-test, got %q", got)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if body["code"] != float64(50001) {
		t.Fatalf("expected code 50001, got %v", body["code"])
	}
	if body["message"] != "internal server error" {
		t.Fatalf("expected message internal server error, got %v", body["message"])
	}
	if body["requestId"] != "req-panic-test" {
		t.Fatalf("expected requestId req-panic-test, got %v", body["requestId"])
	}
}
