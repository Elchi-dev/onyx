// Package dashboard serves the Onyx live WebSocket dashboard.
package dashboard

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// Event is any message broadcast to connected dashboard clients.
type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// Hub manages all active WebSocket connections and fan-out broadcasts.
type Hub struct {
	mu      sync.RWMutex
	clients map[*wsClient]struct{}
	log     *slog.Logger
}

// wsClient wraps one WebSocket connection with a send buffer.
type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

// NewHub creates an initialized Hub.
func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		clients: make(map[*wsClient]struct{}),
		log:     log,
	}
}

// Register upgrades conn to a managed hub client and blocks until the client
// disconnects.
func (h *Hub) Register(conn *websocket.Conn) {
	c := &wsClient{conn: conn, send: make(chan []byte, 128)}
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	h.log.Info("dashboard client connected", "remote", conn.RemoteAddr())
	go c.writePump(h)
	c.readPump(h) // blocks
}

// Broadcast serializes event to JSON and sends it to all connected clients.
// Slow clients are dropped rather than blocking the broadcaster.
func (h *Hub) Broadcast(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		h.log.Error("marshaling broadcast event", "error", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.send <- data:
		default:
			// Buffer full — drop message for this slow client.
		}
	}
}

func (h *Hub) remove(c *wsClient) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	close(c.send)
	h.log.Info("dashboard client disconnected")
}

func (c *wsClient) writePump(_ *Hub) {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *wsClient) readPump(h *Hub) {
	defer h.remove(c)
	c.conn.SetReadLimit(512)
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}
