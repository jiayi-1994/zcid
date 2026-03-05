package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"
)

// K8sWatcher abstracts Kubernetes PipelineRun watching.
// TODO: Replace with real K8s Informer when cluster is available.
type K8sWatcher interface {
	WatchPipelineRuns(ctx context.Context, namespace string, handler func(runName, status string, stepStatuses []StepStatus))
}

// MockK8sWatcher implements K8sWatcher with stub behavior for local development.
type MockK8sWatcher struct {
	mu sync.Mutex
	// TODO: Replace with real K8s Informer when cluster is available.
	// Real implementation would use client-go informers to watch PipelineRun CRs.
}

// WatchPipelineRuns starts a mock watch that does nothing.
// TODO: Implement with K8s Informer to watch PipelineRun resources and invoke handler on status changes.
func (m *MockK8sWatcher) WatchPipelineRuns(ctx context.Context, namespace string, handler func(runName, status string, stepStatuses []StepStatus)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	slog.Info("MOCK: WatchPipelineRuns started", slog.String("namespace", namespace))
	// In real impl: set up informer, on add/update call handler(runName, status, stepStatuses)
	<-ctx.Done()
}

// PipelineWatcher watches pipeline runs and broadcasts status updates via the hub.
type PipelineWatcher struct {
	hub        *Hub
	k8sWatcher K8sWatcher
	projectMap map[string]string // namespace -> projectID; for routing status to project subscribers
	mu         sync.RWMutex
}

// NewPipelineWatcher creates a new PipelineWatcher.
func NewPipelineWatcher(hub *Hub, k8sWatcher K8sWatcher) *PipelineWatcher {
	return &PipelineWatcher{
		hub:        hub,
		k8sWatcher: k8sWatcher,
		projectMap: make(map[string]string),
	}
}

// RegisterNamespaceProject maps a namespace to a project ID for status routing.
func (w *PipelineWatcher) RegisterNamespaceProject(namespace, projectID string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.projectMap[namespace] = projectID
}

// Start begins watching. Call with context for cancellation.
func (w *PipelineWatcher) Start(ctx context.Context) {
	// TODO: In real impl, get namespaces from project config and watch each.
	w.mu.RLock()
	namespaces := make([]string, 0, len(w.projectMap))
	for ns := range w.projectMap {
		namespaces = append(namespaces, ns)
	}
	w.mu.RUnlock()

	for _, ns := range namespaces {
		ns := ns
		go func() {
			w.k8sWatcher.WatchPipelineRuns(ctx, ns, func(runName, status string, stepStatuses []StepStatus) {
				w.mu.RLock()
				projectID := w.projectMap[ns]
				w.mu.RUnlock()
				if projectID == "" {
					return
				}
				now := time.Now()
				msg := WSMessage{
					Type: MsgTypeStatus,
					Seq:  time.Now().UnixNano(),
					Data: StatusData{
						RunID:        runName,
						Status:       status,
						StepStatuses: stepStatuses,
					},
					Timestamp: now,
				}
				buf, err := json.Marshal(msg)
				if err != nil {
					slog.Warn("failed to marshal status msg", slog.Any("error", err))
					return
				}
				w.hub.BroadcastToProject(projectID, buf)
			})
		}()
	}
}
