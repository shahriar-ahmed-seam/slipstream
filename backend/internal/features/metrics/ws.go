package metrics

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/eventdriven/metrics/internal/platform/logger"
	"github.com/gorilla/websocket"
)

// WSHandler provides a WebSocket transport for the snapshot stream.
type WSHandler struct {
	log   *logger.Logger
	svc   *Service
	up    websocket.Upgrader
	path  string
}

// NewWSHandler creates a WebSocket handler bound to the supplied path.
func NewWSHandler(log *logger.Logger, svc *Service, path string) *WSHandler {
	return &WSHandler{
		log:  log,
		svc:  svc,
		path: path,
		up: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 4096,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
}

// Register attaches the WebSocket route to the provided mux.
func (h *WSHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc(h.path, h.serveWS)
}

func (h *WSHandler) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := h.up.Upgrade(w, r, nil)
	if err != nil {
		h.log.Warn("ws_upgrade_failed", map[string]any{"err": err.Error()})
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	sub := h.svc.Subscribe(ctx)

	go func() {
		defer cancel()
		conn.SetReadLimit(4096)
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		})
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	pingTicker := time.NewTicker(20 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case snap, ok := <-sub:
			if !ok {
				return
			}
			data, err := json.Marshal(snap)
			if err != nil {
				continue
			}
			_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-pingTicker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
