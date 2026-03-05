package ws

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHubRegisterUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer func() {
		// Give time for Run to process
		time.Sleep(10 * time.Millisecond)
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		client := NewClient(conn, "user1", "run1", "", hub)
		hub.Register(client)
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Allow time for register
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.CountUserConnections("user1"))

	conn.Close()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, hub.CountUserConnections("user1"))
}

func TestBroadcastToRun(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		client := NewClient(conn, "user1", "run-123", "", hub)
		hub.Register(client)
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToRun("run-123", []byte(`{"type":"log","seq":1}`))
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, `{"type":"log","seq":1}`, string(msg))

	// Broadcast to different run should not be received
	hub.BroadcastToRun("run-other", []byte(`{"type":"log","seq":2}`))
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, _, err = conn.ReadMessage()
	assert.Error(t, err)
}

func TestBroadcastToProject(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		client := NewClient(conn, "user1", "", "proj-456", hub)
		hub.Register(client)
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToProject("proj-456", []byte(`{"type":"status","seq":1}`))
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, `{"type":"status","seq":1}`, string(msg))
}

func TestConnectionLimit(t *testing.T) {
	hub := NewHub()
	hub.mu.Lock()
	for i := 0; i < 10; i++ {
		c := &Client{ID: "c-" + string(rune('a'+i)), UserID: "user1", hub: hub}
		hub.clients[c.ID] = c
	}
	hub.mu.Unlock()

	assert.Equal(t, 10, hub.CountUserConnections("user1"))
	assert.Equal(t, 0, hub.CountUserConnections("user2"))
}

func TestHeartbeat(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		client := NewClient(conn, "user1", "run1", "", hub)
		hub.Register(client)
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()
	conn.SetPongHandler(func(string) error { return nil })

	// Allow time for register (heartbeat uses ping every 30s, pong wait 60s)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.CountUserConnections("user1"))
}
