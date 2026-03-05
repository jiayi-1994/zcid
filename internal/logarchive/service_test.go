package logarchive

import (
	"context"
	"testing"
	"time"

	"github.com/xjy/zcid/internal/ws"
)

func TestArchiveRunLogs(t *testing.T) {
	ctx := context.Background()
	mock := NewMockStorage()
	svc := NewService(mock, "zcid-logs")

	logs := []ws.LogLine{
		{Seq: 1, StepID: "step1", Content: "line 1", Level: "info", Timestamp: time.Now()},
		{Seq: 2, StepID: "step1", Content: "line 2", Level: "info", Timestamp: time.Now()},
		{Seq: 3, StepID: "step2", Content: "line 3", Level: "error", Timestamp: time.Now()},
	}

	err := svc.ArchiveRunLogs(ctx, "run-123", logs)
	if err != nil {
		t.Fatalf("ArchiveRunLogs: %v", err)
	}

	entries, total, err := svc.GetArchivedLogs(ctx, "run-123", 1, 10)
	if err != nil {
		t.Fatalf("GetArchivedLogs: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(entries) != 3 {
		t.Errorf("len(entries) = %d, want 3", len(entries))
	}
	if entries[0].Content != "line 1" {
		t.Errorf("entries[0].Content = %q, want \"line 1\"", entries[0].Content)
	}
}

func TestArchiveRunLogs_Empty(t *testing.T) {
	ctx := context.Background()
	mock := NewMockStorage()
	svc := NewService(mock, "zcid-logs")

	err := svc.ArchiveRunLogs(ctx, "run-empty", nil)
	if err != nil {
		t.Fatalf("ArchiveRunLogs empty: %v", err)
	}

	entries, total, err := svc.GetArchivedLogs(ctx, "run-empty", 1, 10)
	if err != nil {
		t.Fatalf("GetArchivedLogs: %v", err)
	}
	if total != 0 || len(entries) != 0 {
		t.Errorf("expected empty, got total=%d len=%d", total, len(entries))
	}
}

func TestGetArchivedLogs_Pagination(t *testing.T) {
	ctx := context.Background()
	mock := NewMockStorage()
	svc := NewService(mock, "zcid-logs")

	var logs []ws.LogLine
	for i := 0; i < 5; i++ {
		logs = append(logs, ws.LogLine{
			Seq: int64(i + 1), StepID: "s1", Content: "x", Level: "info", Timestamp: time.Now(),
		})
	}
	if err := svc.ArchiveRunLogs(ctx, "run-pag", logs); err != nil {
		t.Fatal(err)
	}

	entries, total, err := svc.GetArchivedLogs(ctx, "run-pag", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(entries) != 2 {
		t.Errorf("page 1 pageSize 2: len = %d, want 2", len(entries))
	}

	entries2, _, err := svc.GetArchivedLogs(ctx, "run-pag", 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries2) != 2 {
		t.Errorf("page 2 pageSize 2: len = %d, want 2", len(entries2))
	}

	entries3, _, err := svc.GetArchivedLogs(ctx, "run-pag", 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries3) != 1 {
		t.Errorf("page 3 pageSize 2: len = %d, want 1", len(entries3))
	}
}
