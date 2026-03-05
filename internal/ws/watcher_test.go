package ws

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
)

// testK8sWatcher is a mock that invokes the handler after Start.
type testK8sWatcher struct {
	handler func(runName, status string, stepStatuses []StepStatus)
	ctx     context.Context
}

func (t *testK8sWatcher) WatchPipelineRuns(ctx context.Context, namespace string, handler func(runName, status string, stepStatuses []StepStatus)) {
	t.ctx = ctx
	t.handler = handler
}

func (t *testK8sWatcher) emit(runName, status string, stepStatuses []StepStatus) {
	if t.handler != nil {
		t.handler(runName, status, stepStatuses)
	}
}

func TestMockWatcherSendsStatus(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	mock := &testK8sWatcher{}
	watcher := NewPipelineWatcher(hub, mock)
	watcher.RegisterNamespaceProject("ns1", "proj-1")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go watcher.Start(ctx)

	// Give Start time to call WatchPipelineRuns
	time.Sleep(50 * time.Millisecond)

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

	time.Sleep(50 * time.Millisecond)

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
