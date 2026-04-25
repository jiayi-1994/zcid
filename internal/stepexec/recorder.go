package stepexec

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

var ErrQueueFull = errors.New("step execution recorder queue full")

type RecordEvent struct{ Row StepExecution }

type Recorder struct {
	repo  Repository
	queue chan RecordEvent
}

func NewRecorder(repo Repository, queueSize int) *Recorder {
	if queueSize <= 0 {
		queueSize = 1000
	}
	return &Recorder{repo: repo, queue: make(chan RecordEvent, queueSize)}
}

func (r *Recorder) Record(ctx context.Context, event RecordEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case r.queue <- event:
		return nil
	default:
		slog.Warn("step execution recorder queue full; dropping record",
			slog.String("pipeline_run_id", event.Row.PipelineRunID),
			slog.String("task_run", event.Row.TaskRunName),
			slog.String("step", event.Row.StepName),
		)
		return ErrQueueFull
	}
}

func (r *Recorder) RecordRows(ctx context.Context, rows []StepExecution) error {
	for i := range rows {
		if err := r.Record(ctx, RecordEvent{Row: rows[i]}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Recorder) Run(ctx context.Context) {
	slog.Info("step execution recorder started", slog.Int("queue_size", cap(r.queue)))
	defer slog.Info("step execution recorder stopped")
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-r.queue:
			r.writeWithRetry(ctx, event.Row)
		}
	}
}

func (r *Recorder) writeWithRetry(ctx context.Context, row StepExecution) {
	backoff := 100 * time.Millisecond
	for attempt := 1; attempt <= 3; attempt++ {
		if err := r.repo.Upsert(ctx, &row); err != nil {
			if attempt == 3 {
				slog.Error("step execution upsert exhausted; dropping record", slog.Any("error", err), slog.String("pipeline_run_id", row.PipelineRunID), slog.String("task_run", row.TaskRunName), slog.String("step", row.StepName))
				return
			}
			slog.Warn("step execution upsert failed; retrying", slog.Any("error", err), slog.Int("attempt", attempt), slog.String("pipeline_run_id", row.PipelineRunID), slog.String("task_run", row.TaskRunName), slog.String("step", row.StepName))
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				backoff *= 2
			}
			continue
		}
		return
	}
}

func (r *Recorder) FinalizeRun(ctx context.Context, runID, terminalStatus string) error {
	return r.repo.FinalizeRun(ctx, runID, terminalStatus)
}
