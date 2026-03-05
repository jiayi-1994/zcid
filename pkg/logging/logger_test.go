package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestMaskingHandlerMasksSensitiveData(t *testing.T) {
	var buf bytes.Buffer
	h := NewMaskingHandler(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger := slog.New(h)

	logger.Info("password=abc123 token=xyz", slog.String("secret_key", "k1"), slog.String("safe", "ok"))

	line := strings.TrimSpace(buf.String())
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("failed to parse log json: %v", err)
	}

	msg := m["msg"].(string)
	if strings.Contains(msg, "abc123") || strings.Contains(msg, "xyz") {
		t.Fatalf("expected sensitive values masked in message, got %q", msg)
	}
	if m["secret_key"] != "***" {
		t.Fatalf("expected masked secret_key, got %v", m["secret_key"])
	}
	if m["safe"] != "ok" {
		t.Fatalf("expected safe field unchanged, got %v", m["safe"])
	}
}

func TestSetLevel(t *testing.T) {
	Init("info")
	if err := SetLevel("debug"); err != nil {
		t.Fatalf("expected set debug success, got %v", err)
	}
	if got := CurrentLevel(); got != "debug" {
		t.Fatalf("expected debug, got %s", got)
	}

	if err := SetLevel("bad"); err == nil {
		t.Fatal("expected invalid level error")
	}
}
