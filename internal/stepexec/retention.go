package stepexec

import (
	"context"
	"log/slog"
	"time"
)

type RetentionWorker struct {
	repo          Repository
	retentionDays int
	interval      time.Duration
	batchSize     int
	maxBatches    int
}

type RetentionOption func(*RetentionWorker)

func WithInterval(interval time.Duration) RetentionOption {
	return func(w *RetentionWorker) { w.interval = interval }
}
func WithBatchSize(size int) RetentionOption { return func(w *RetentionWorker) { w.batchSize = size } }
func WithMaxBatches(max int) RetentionOption { return func(w *RetentionWorker) { w.maxBatches = max } }

func NewRetentionWorker(repo Repository, retentionDays int, opts ...RetentionOption) *RetentionWorker {
	if retentionDays <= 0 {
		retentionDays = 90
	}
	w := &RetentionWorker{repo: repo, retentionDays: retentionDays, interval: time.Hour, batchSize: 1000, maxBatches: 100}
	for _, opt := range opts {
		opt(w)
	}
	if w.interval <= 0 {
		w.interval = time.Hour
	}
	if w.batchSize <= 0 {
		w.batchSize = 1000
	}
	if w.maxBatches <= 0 {
		w.maxBatches = 100
	}
	return w
}

func (w *RetentionWorker) Run(ctx context.Context) {
	slog.Info("step execution retention worker started", slog.Int("retention_days", w.retentionDays), slog.Duration("interval", w.interval))
	defer slog.Info("step execution retention worker stopped")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	w.runOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *RetentionWorker) runOnce(ctx context.Context) {
	cutoff := time.Now().Add(-time.Duration(w.retentionDays) * 24 * time.Hour)
	deleted, truncated, err := w.repo.DeleteExpired(ctx, cutoff, w.batchSize, w.maxBatches)
	if err != nil {
		slog.Error("step execution retention failed", slog.Any("error", err))
		return
	}
	attrs := []slog.Attr{slog.Int("deleted", deleted), slog.Time("cutoff", cutoff)}
	if truncated {
		slog.Warn("step execution retention batch cap reached", attrsToAny(attrs)...)
		return
	}
	slog.Info("step execution retention completed", attrsToAny(attrs)...)
}

func attrsToAny(attrs []slog.Attr) []any {
	out := make([]any, 0, len(attrs))
	for _, a := range attrs {
		out = append(out, a)
	}
	return out
}
