package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequestIDUsesIncomingHeader(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"requestId": GetRequestID(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "req-from-client")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("X-Request-ID"); got != "req-from-client" {
		t.Fatalf("expected response header req-from-client, got %q", got)
	}
	if !strings.Contains(w.Body.String(), "req-from-client") {
		t.Fatalf("expected response body to include requestId, got %s", w.Body.String())
	}
}

func TestRequestIDGeneratesWhenMissing(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"requestId": GetRequestID(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Fatal("expected generated X-Request-ID header")
	}
	if !strings.HasPrefix(requestID, "req-") {
		t.Fatalf("expected request id prefix req-, got %q", requestID)
	}
	if !strings.Contains(w.Body.String(), requestID) {
		t.Fatalf("expected response body to include generated requestId %q, got %s", requestID, w.Body.String())
	}
}
