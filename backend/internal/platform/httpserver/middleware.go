package httpserver

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/eventdriven/metrics/internal/platform/logger"
)

// Chain composes multiple middleware functions around a terminal handler.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Recover catches panics in downstream handlers and converts them to
// 500 responses without crashing the process.
func Recover(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic", map[string]any{
						"err":   rec,
						"stack": string(debug.Stack()),
					})
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Logger emits one structured access log per request.
func Logger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)
			log.Info("http_request", map[string]any{
				"method":   r.Method,
				"path":     r.URL.Path,
				"status":   ww.status,
				"duration": time.Since(start).String(),
			})
		})
	}
}

// CORS sets permissive CORS headers for development. Tighten allowed
// origins in production by setting them in config.
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.wrote {
		return
	}
	s.status = code
	s.ResponseWriter.WriteHeader(code)
	s.wrote = true
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if !s.wrote {
		s.status = http.StatusOK
		s.wrote = true
	}
	return s.ResponseWriter.Write(b)
}

// Flush forwards to the underlying ResponseWriter if it implements
// http.Flusher. This is required so that SSE handlers can flush frames
// to the client when wrapped by middleware.
func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack forwards to the underlying ResponseWriter if it implements
// http.Hijacker. Required so that WebSocket upgrade requests work
// when the connection is wrapped by middleware.
func (s *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := s.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijacker not supported by underlying ResponseWriter")
	}
	return h.Hijack()
}
