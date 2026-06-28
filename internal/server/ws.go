package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"robot-panel/internal/ros"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// wsHub broadcasts JSON state snapshots to all connected WebSocket clients.
type wsHub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func newHub() *wsHub {
	return &wsHub{clients: make(map[*websocket.Conn]struct{})}
}

func (h *wsHub) add(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *wsHub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

func (h *wsHub) broadcast(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

// wsStateHandler upgrades an HTTP connection to WebSocket and registers it with the hub.
// The client just needs to hold the connection open; the server pushes state at 20 Hz.
func wsStateHandler(hub *wsHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		hub.add(conn)
		defer func() {
			hub.remove(conn)
			conn.Close()
		}()
		// drain client frames to detect disconnection
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}

// broadcastLoop pushes the latest robot state to all WebSocket clients at 20 Hz.
func broadcastLoop(rosClient *ros.Client, hub *wsHub) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		state := rosClient.GetState()
		data, err := json.Marshal(state)
		if err != nil {
			continue
		}
		hub.broadcast(data)
	}
}
