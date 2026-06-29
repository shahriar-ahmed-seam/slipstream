// Package httpserver wraps net/http with a graceful shutdown lifecycle
// and middleware orchestration. It is deliberately small to keep the
// production surface area limited.
package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/eventdriven/metrics/internal/platform/logger"
)

// Server wraps an http.Server with a logger and a shutdown timeout.
type Server struct {
	httpServer *http.Server
	log        *logger.Logger
}

// New constructs a Server around a fully built http.Handler.
func New(addr string, handler http.Handler, log *logger.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			// WriteTimeout is intentionally 0 (no per-request deadline) so
			// long-lived SSE / WebSocket connections are not torn down by
			// the server. IdleTimeout still bounds dormant connections.
			WriteTimeout: 0,
			IdleTimeout:  120 * time.Second,
		},
		log: log,
	}
}

// Start runs the HTTP server in a blocking fashion until the provided
// context is canceled or the server returns a non-nil error.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("http_listen", map[string]any{"addr": s.httpServer.Addr})
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// Shutdown gracefully stops the server using the supplied timeout.
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	s.log.Info("http_shutdown", map[string]any{"timeout": timeout.String()})
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("http_shutdown_failed", map[string]any{"err": err.Error()})
		return err
	}
	return nil
}
