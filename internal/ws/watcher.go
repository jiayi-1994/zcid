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
	WatchPipelineRuns(ctx context.Context, namespace string, handler func(runName, projectID, status string, stepStatuses []StepStatus))
}

// MockK8sWatcher implements K8sWatcher with stub behavior for local development.
type MockK8sWatcher struct {
	mu sync.Mutex
	// TODO: Replace with real K8s Informer when cluster is available.
	// Real implementation would use client-go informers to watch PipelineRun CRs.
}

// WatchPipelineRuns starts a mock watch that does nothing.
// TODO: Implement with K8s Informer to watch PipelineRun resources and invoke handler on status changes.
func (m *MockK8sWatcher) WatchPipelineRuns(ctx context.Context, namespace string, handler func(runName, projectID, status string, stepStatuses []StepStatus)) {
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
	watched    map[string]bool
	ctx        context.Context
	mu         sync.RWMutex
}

// NewPipelineWatcher creates a new PipelineWatcher.
func NewPipelineWatcher(hub *Hub, k8sWatcher K8sWatcher) *PipelineWatcher {
	return &PipelineWatcher{
		hub:        hub,
		k8sWatcher: k8sWatcher,
		projectMap: make(map[string]string),
		watched:    make(map[string]bool),
	}
}

// RegisterNamespaceProject maps a namespace to a project ID for status routing.
func (w *PipelineWatcher) RegisterNamespaceProject(namespace, projectID string) {
	w.mu.Lock()
	w.projectMap[namespace] = projectID
	ctx := w.ctx
	alreadyWatching := w.watched[namespace]
	if ctx != nil && !alreadyWatching {
		w.watched[namespace] = true
	}
	w.mu.Unlock()
	if ctx != nil && !alreadyWatching {
		w.watchNamespace(ctx, namespace)
	}
}

// DeregisterNamespaceProject removes a namespace mapping. Existing watch goroutines
// may keep running until process shutdown, but status events are dropped once the
// mapping is gone.
func (w *PipelineWatcher) DeregisterNamespaceProject(namespace string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.projectMap, namespace)
}

// Start begins watching. Call with context for cancellation.
func (w *PipelineWatcher) Start(ctx context.Context) {
	w.mu.Lock()
	w.ctx = ctx
	namespaces := make([]string, 0, len(w.projectMap))
	for ns := range w.projectMap {
		if !w.watched[ns] {
			w.watched[ns] = true
			namespaces = append(namespaces, ns)
		}
	}
	w.mu.Unlock()

	for _, ns := range namespaces {
		w.watchNamespace(ctx, ns)
	}
}

func (w *PipelineWatcher) watchNamespace(ctx context.Context, ns string) {
	go func() {
		w.k8sWatcher.WatchPipelineRuns(ctx, ns, func(runName, eventProjectID, status string, stepStatuses []StepStatus) {
			w.mu.RLock()
			projectID := w.projectMap[ns]
			w.mu.RUnlock()
			if eventProjectID != "" {
				projectID = eventProjectID
			}
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
