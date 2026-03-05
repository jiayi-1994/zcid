package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// LogLine represents a single log line with sequence for ordering and replay.
type LogLine struct {
	Seq       int64     `json:"seq"`
	StepID    string    `json:"stepId"`
	Content   string    `json:"content"`
	Level     string    `json:"level"`
	Timestamp time.Time `json:"timestamp"`
}

// LogBuffer stores recent log lines with a max size for reconnection replay.
type LogBuffer struct {
	mu       sync.RWMutex
	lines    []LogLine
	maxLines int
	nextSeq  int64
}

// NewLogBuffer creates a LogBuffer with the given max line count.
func NewLogBuffer(maxLines int) *LogBuffer {
	if maxLines <= 0 {
		maxLines = 1000
	}
	return &LogBuffer{
		lines:    make([]LogLine, 0, maxLines*2),
		maxLines: maxLines,
	}
}

// Append adds a log line. Seq is auto-incremented. Returns the assigned seq.
func (b *LogBuffer) Append(line LogLine) int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	seq := atomic.AddInt64(&b.nextSeq, 1)
	line.Seq = seq
	b.lines = append(b.lines, line)
	if len(b.lines) > b.maxLines {
		b.lines = b.lines[len(b.lines)-b.maxLines:]
	}
	return seq
}

// GetSince returns all lines with Seq > seq, for reconnection replay.
func (b *LogBuffer) GetSince(seq int64) []LogLine {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]LogLine, 0)
	for _, l := range b.lines {
		if l.Seq > seq {
			out = append(out, l)
		}
	}
	return out
}

// LogCollector abstracts log streaming from a pod.
// TODO: Replace with real implementation that streams from K8s pod logs.
type LogCollector interface {
	StreamLogs(ctx context.Context, namespace, podName string, handler func(line string))
}

// SecretMasker provides secret values to mask in log output (FR14).
// For MVP, returns a list of values to replace with "***" before broadcasting.
type SecretMasker interface {
	GetSecretsToMask(runID string) []string
}

// PlaceholderSecretMasker returns no secrets for MVP; wire variable service later.
type PlaceholderSecretMasker struct {
	Values []string
}

func (p *PlaceholderSecretMasker) GetSecretsToMask(runID string) []string {
	_ = runID
	return p.Values
}

// MockLogCollector implements LogCollector with no-op behavior.
type MockLogCollector struct{}
// TODO: Replace with real K8s log streaming (e.g. corev1.PodInterface.GetLogs with stream).
func (m *MockLogCollector) StreamLogs(ctx context.Context, namespace, podName string, handler func(line string)) {
	slog.Info("MOCK: StreamLogs", slog.String("namespace", namespace), slog.String("podName", podName))
	<-ctx.Done()
}

// LogStreamManager manages log buffers and streams logs to the hub.
type LogStreamManager struct {
	hub        *Hub
	buffers    map[string]*LogBuffer
	cancels    map[string]context.CancelFunc
	collector  LogCollector
	masker     SecretMasker
	mu         sync.RWMutex
}

// NewLogStreamManager creates a LogStreamManager.
func NewLogStreamManager(hub *Hub, collector LogCollector, masker SecretMasker) *LogStreamManager {
	return &LogStreamManager{
		hub:       hub,
		buffers:   make(map[string]*LogBuffer),
		cancels:   make(map[string]context.CancelFunc),
		collector: collector,
		masker:    masker,
	}
}

// GetBuffer returns or creates the log buffer for a run.
func (m *LogStreamManager) GetBuffer(runID string, maxLines int) *LogBuffer {
	m.mu.Lock()
	defer m.mu.Unlock()
	if b, ok := m.buffers[runID]; ok {
		return b
	}
	b := NewLogBuffer(maxLines)
	m.buffers[runID] = b
	return b
}

// StopStreaming stops log streaming for a run and removes its buffer.
func (m *LogStreamManager) StopStreaming(runID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cancel, ok := m.cancels[runID]; ok {
		cancel()
		delete(m.cancels, runID)
	}
	delete(m.buffers, runID)
}

// StartStreaming begins streaming logs from the pod to the hub and buffer.
func (m *LogStreamManager) StartStreaming(runID, namespace, podName string) {
	buffer := m.GetBuffer(runID, 5000)
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancels[runID] = cancel
	m.mu.Unlock()
	go m.collector.StreamLogs(ctx, namespace, podName, func(line string) {
		masked := m.maskLine(runID, line)
		ll := LogLine{
			StepID:    "default",
			Content:   masked,
			Level:     "info",
			Timestamp: time.Now(),
		}
		seq := buffer.Append(ll)
		msg := WSMessage{
			Type: MsgTypeLog,
			Seq:  seq,
			Data: LogData{Line: masked, StepID: ll.StepID, Level: ll.Level},
			Timestamp: ll.Timestamp,
		}
		buf, err := json.Marshal(msg)
		if err != nil {
			slog.Warn("marshal log msg failed", slog.Any("error", err))
			return
		}
		m.hub.BroadcastToRun(runID, buf)
	})
}

// ReplayFn returns a function that provides replay for the handler.
func (m *LogStreamManager) ReplayFn() ReplayFn {
	return func(runID string, lastSeq int64) [][]byte {
		m.mu.RLock()
		b, ok := m.buffers[runID]
		m.mu.RUnlock()
		if !ok {
			return nil
		}
		lines := b.GetSince(lastSeq)
		out := make([][]byte, 0, len(lines))
		for _, ll := range lines {
			// Content already masked when appended
			msg := WSMessage{
				Type: MsgTypeLog,
				Seq:  ll.Seq,
				Data: LogData{Line: ll.Content, StepID: ll.StepID, Level: ll.Level},
				Timestamp: ll.Timestamp,
			}
			buf, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			out = append(out, buf)
		}
		return out
	}
}

func (m *LogStreamManager) maskLine(runID, line string) string {
	if m.masker == nil {
		return line
	}
	out := line
	for _, v := range m.masker.GetSecretsToMask(runID) {
		if v != "" {
			out = strings.ReplaceAll(out, v, "***")
		}
	}
	return out
}
