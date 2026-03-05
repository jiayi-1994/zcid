package logging

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

var sensitivePattern = regexp.MustCompile(`(?i)(password|passwd|pwd|secret|token|apikey|api_key|access_key|secret_key)\s*[:=]\s*([^\s,;]+)`)

type MaskingHandler struct {
	next slog.Handler
}

func NewMaskingHandler(next slog.Handler) slog.Handler {
	return &MaskingHandler{next: next}
}

func (h *MaskingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *MaskingHandler) Handle(ctx context.Context, r slog.Record) error {
	masked := slog.NewRecord(r.Time, r.Level, maskString(r.Message), r.PC)
	r.Attrs(func(a slog.Attr) bool {
		masked.AddAttrs(maskAttr(a))
		return true
	})
	return h.next.Handle(ctx, masked)
}

func (h *MaskingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	masked := make([]slog.Attr, 0, len(attrs))
	for _, a := range attrs {
		masked = append(masked, maskAttr(a))
	}
	return &MaskingHandler{next: h.next.WithAttrs(masked)}
}

func (h *MaskingHandler) WithGroup(name string) slog.Handler {
	return &MaskingHandler{next: h.next.WithGroup(name)}
}

func maskAttr(a slog.Attr) slog.Attr {
	a.Key = strings.TrimSpace(a.Key)
	if a.Value.Kind() == slog.KindGroup {
		group := a.Value.Group()
		masked := make([]slog.Attr, 0, len(group))
		for _, item := range group {
			masked = append(masked, maskAttr(item))
		}
		return slog.Attr{Key: a.Key, Value: slog.GroupValue(masked...)}
	}

	if isSensitiveKey(a.Key) {
		return slog.String(a.Key, "***")
	}

	switch a.Value.Kind() {
	case slog.KindString:
		return slog.String(a.Key, maskString(a.Value.String()))
	case slog.KindAny:
		return slog.String(a.Key, maskString(fmt.Sprint(a.Value.Any())))
	default:
		return a
	}
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "password") || strings.Contains(key, "passwd") || strings.Contains(key, "pwd") || strings.Contains(key, "secret") || strings.Contains(key, "token") || strings.Contains(key, "apikey") || strings.Contains(key, "api_key") || strings.Contains(key, "access_key") || strings.Contains(key, "secret_key")
}

func maskString(s string) string {
	return sensitivePattern.ReplaceAllString(s, "$1=***")
}
