package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccessEnvelope(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(ContextKeyRequestID, "req-test-1")

	Success(c, gin.H{"k": "v"})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if body["code"] != float64(CodeSuccess) {
		t.Fatalf("expected code %d, got %v", CodeSuccess, body["code"])
	}
	if body["message"] != "success" {
		t.Fatalf("expected message success, got %v", body["message"])
	}
	if body["requestId"] != "req-test-1" {
		t.Fatalf("expected requestId req-test-1, got %v", body["requestId"])
	}
	data, ok := body["data"].(map[string]any)
	if !ok || data["k"] != "v" {
		t.Fatalf("expected data payload, got %v", body["data"])
	}
}

func TestErrorEnvelopeWithoutDetail(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(ContextKeyRequestID, "req-test-2")

	Error(c, 0, CodeBadRequest, "bad request", "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if body["code"] != float64(CodeBadRequest) {
		t.Fatalf("expected code %d, got %v", CodeBadRequest, body["code"])
	}
	if body["requestId"] != "req-test-2" {
		t.Fatalf("expected requestId req-test-2, got %v", body["requestId"])
	}
	if _, exists := body["detail"]; exists {
		t.Fatalf("expected detail omitted when empty, got %v", body["detail"])
	}
}

func TestHandleErrorBizErrorAndUnknown(t *testing.T) {
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Set(ContextKeyRequestID, "req-biz")

	HandleError(c1, NewBizError(CodeConflict, "conflict", "already exists"))

	if w1.Code != http.StatusConflict {
		t.Fatalf("expected 409 for biz error, got %d", w1.Code)
	}

	var body1 map[string]any
	if err := json.Unmarshal(w1.Body.Bytes(), &body1); err != nil {
		t.Fatalf("failed to parse biz error json: %v", err)
	}
	if body1["code"] != float64(CodeConflict) {
		t.Fatalf("expected biz code %d, got %v", CodeConflict, body1["code"])
	}

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Set(ContextKeyRequestID, "req-unknown")

	HandleError(c2, errors.New("boom"))

	if w2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for unknown error, got %d", w2.Code)
	}
	var body2 map[string]any
	if err := json.Unmarshal(w2.Body.Bytes(), &body2); err != nil {
		t.Fatalf("failed to parse unknown error json: %v", err)
	}
	if body2["code"] != float64(CodeInternalServerError) {
		t.Fatalf("expected unknown code %d, got %v", CodeInternalServerError, body2["code"])
	}
}

func TestGetRequestIDEdgeCases(t *testing.T) {
	if got := GetRequestID(nil); got != "" {
		t.Fatalf("expected empty for nil context, got %q", got)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	if got := GetRequestID(c); got != "" {
		t.Fatalf("expected empty for missing key, got %q", got)
	}

	c.Set(ContextKeyRequestID, 123)
	if got := GetRequestID(c); got != "" {
		t.Fatalf("expected empty for non-string key, got %q", got)
	}
}
