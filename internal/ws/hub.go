package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	pingInterval    = 30 * time.Second
	pongWait        = 60 * time.Second
	writeWait       = 10 * time.Second
	maxMessageSize  = 512 * 1024
	maxConnsPerUser = 10
)

type Client struct {
	ID        string
	UserID    string
	RunID     string // for log subscription; empty for status subscription
	ProjectID string // for status subscription; empty for log subscription
	Conn      *websocket.Conn
	Send      chan []byte
	LastSeq   int64
	hub       *Hub
	mu        sync.Mutex
}

type Hub struct {
	clients    map[string]*Client // clientID -> Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *broadcastMsg
	mu         sync.RWMutex
}

type broadcastMsg struct {
	runID     string
	projectID string
	msg       []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *broadcastMsg, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			slog.Debug("ws client registered", slog.String("clientId", client.ID), slog.String("userId", client.UserID))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()

		case bm := <-h.broadcast:
			h.mu.RLock()
			for _, c := range h.clients {
				if bm.runID != "" && c.RunID == bm.runID {
					select {
					case c.Send <- bm.msg:
					default:
						// buffer full, skip
					}
				}
				if bm.projectID != "" && c.ProjectID == bm.projectID {
					select {
					case c.Send <- bm.msg:
					default:
						// buffer full, skip
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) BroadcastToRun(runID string, msg []byte) {
	h.broadcast <- &broadcastMsg{runID: runID, msg: msg}
}

func (h *Hub) BroadcastToProject(projectID string, msg []byte) {
	h.broadcast <- &broadcastMsg{projectID: projectID, msg: msg}
}

func (h *Hub) CountUserConnections(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	n := 0
	for _, c := range h.clients {
		if c.UserID == userID {
			n++
		}
	}
	return n
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Debug("ws read error", slog.Any("error", err))
			}
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func NewClient(conn *websocket.Conn, userID, runID, projectID string, hub *Hub) *Client {
	return &Client{
		ID:        uuid.New().String(),
		UserID:    userID,
		RunID:     runID,
		ProjectID: projectID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		hub:       hub,
	}
}

func EncodeMessage(msg *WSMessage) ([]byte, error) {
	return json.Marshal(msg)
}
