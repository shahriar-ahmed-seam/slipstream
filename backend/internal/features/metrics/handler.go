package metrics

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/eventdriven/metrics/internal/platform/logger"
)

// Handler exposes HTTP routes for the metrics feature.
type Handler struct {
	log        *logger.Logger
	svc        *Service
	ingestPath string
	metricsPath string
	streamPath string
}

// NewHandler constructs an HTTP handler bound to the supplied paths.
func NewHandler(log *logger.Logger, svc *Service, ingestPath, metricsPath, streamPath string) *Handler {
	return &Handler{
		log:         log,
		svc:         svc,
		ingestPath:  ingestPath,
		metricsPath: metricsPath,
		streamPath:  streamPath,
	}
}

// Register attaches the handler's routes to the provided mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(h.ingestPath, h.ingest)
	mux.HandleFunc(h.metricsPath, h.snapshot)
	mux.HandleFunc(h.streamPath, h.streamSSE)
	mux.HandleFunc("/api/healthz", h.health)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	accepted, rejected := h.svc.Counters()
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"ts":       time.Now().UTC().Format(time.RFC3339Nano),
		"accepted": accepted,
		"rejected": rejected,
	})
}

func (h *Handler) ingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	var payload struct {
		Events []Event `json:"events"`
	}
	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "decode body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(payload.Events) == 0 {
		http.Error(w, "no events", http.StatusBadRequest)
		return
	}
	resp, err := h.svc.Ingest(r.Context(), payload.Events)
	if err != nil {
		http.Error(w, "ingest: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusAccepted, resp)
}

func (h *Handler) snapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	snap := h.svc.Snapshot()
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, snap)
}

func (h *Handler) streamSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	sub := h.svc.Subscribe(ctx)

	// Send a comment frame to establish the stream immediately.
	if _, err := io.WriteString(w, ": connected\n\n"); err != nil {
		h.log.Warn("sse_write_failed", map[string]any{"err": err.Error()})
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case snap, ok := <-sub:
			if !ok {
				return
			}
			data, err := json.Marshal(snap)
			if err != nil {
				h.log.Warn("sse_encode_failed", map[string]any{"err": err.Error()})
				continue
			}
			if _, err := io.WriteString(w, "event: snapshot\ndata: "); err != nil {
				return
			}
			if _, err := w.Write(data); err != nil {
				return
			}
			if _, err := io.WriteString(w, "\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(body); err != nil {
		// We cannot recover here; the response has already started.
		_ = errors.New("encode: " + err.Error())
	}
}
