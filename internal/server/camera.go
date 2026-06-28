package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"RoboPanel-backend/internal/camera"
)

// cameraHub captures frames from the RealSense and broadcasts JPEG bytes
// to all connected frontend WebSocket clients.
type cameraHub struct {
	cfg camera.Config

	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func newCameraHub(cfg camera.Config) *cameraHub {
	return &cameraHub{
		cfg:     cfg,
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (h *cameraHub) add(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *cameraHub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

func (h *cameraHub) broadcast(frame []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

func (h *cameraHub) clientCount() int {
	h.mu.Lock()
	n := len(h.clients)
	h.mu.Unlock()
	return n
}

// captureLoop opens the RealSense camera, captures frames, and broadcasts.
// When no clients are connected it stops the camera to free the USB device,
// and restarts when clients reconnect.
func (h *cameraHub) captureLoop() {
	const retryInterval = 3 * time.Second

	for {
		if h.clientCount() == 0 {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		reader := camera.NewReader(h.cfg)
		if err := reader.Start(); err != nil {
			fmt.Printf("[camera] warn: cannot open camera: %v, retry in %s\n", err, retryInterval)
			time.Sleep(retryInterval)
			continue
		}
		fmt.Printf("[camera] camera opened (serial=%s, %dx%d@%dfps)\n",
			h.cfg.Serial, h.cfg.Width, h.cfg.Height, h.cfg.FPS)

		for h.clientCount() > 0 {
			frame, err := reader.ReadJPEG()
			if err != nil {
				fmt.Printf("[camera] read error: %v\n", err)
				break
			}
			h.broadcast(frame)
		}

		reader.Stop()
		fmt.Printf("[camera] camera closed (no clients)\n")
	}
}

// cameraWsHandler upgrades to WebSocket and registers the client with the hub.
func cameraWsHandler(hub *cameraHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		conn.SetReadLimit(512)
		hub.add(conn)
		defer func() {
			hub.remove(conn)
			conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}
