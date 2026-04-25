package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testK8sWatcher struct {
	mu      sync.Mutex
	handler func(runName, projectID, status string, stepStatuses []StepStatus)
	ready   chan struct{}
}

func newTestK8sWatcher() *testK8sWatcher {
	return &testK8sWatcher{ready: make(chan struct{})}
}

func (t *testK8sWatcher) WatchPipelineRuns(ctx context.Context, namespace string, handler func(runName, projectID, status string, stepStatuses []StepStatus)) {
	t.mu.Lock()
	t.handler = handler
	t.mu.Unlock()
	close(t.ready)
	<-ctx.Done()
}

func (t *testK8sWatcher) emit(runName, status string, stepStatuses []StepStatus) {
	t.mu.Lock()
	h := t.handler
	t.mu.Unlock()
	if h != nil {
		h(runName, "", status, stepStatuses)
	}
}

func TestMockWatcherSendsStatus(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	mock := newTestK8sWatcher()
	watcher := NewPipelineWatcher(hub, mock)
	watcher.RegisterNamespaceProject("ns1", "proj-1")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go watcher.Start(ctx)

	select {
	case <-mock.ready:
	case <-time.After(2 * time.Second):
		t.Fatal("WatchPipelineRuns was not called in time")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		client := NewClient(conn, "u1", "", "proj-1", hub)
		hub.Register(client)
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+server.URL[4:], nil)
	require.NoError(t, err)
	defer conn.Close()

	time.Sleep(100 * time.Millisecond)

	mock.emit("run-abc", "Running", []StepStatus{{StepID: "step1", Name: "Build", Status: "running"}})

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)

	var out WSMessage
	require.NoError(t, json.Unmarshal(msg, &out))
	assert.Equal(t, MsgTypeStatus, out.Type)
	ds, ok := out.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "run-abc", ds["runId"])
	assert.Equal(t, "Running", ds["status"])
}
