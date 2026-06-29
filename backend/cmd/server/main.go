// Command server is the entry point for the Event-Driven Metrics Dashboard
// backend. It wires together configuration, logging, the metrics feature
// service, the HTTP server, and a graceful shutdown lifecycle.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"slipstream/internal/features/metrics"
	"slipstream/internal/platform/config"
	"slipstream/internal/platform/httpserver"
	"slipstream/internal/platform/logger"
)

func main() {
	log := logger.New()
	cfg, err := config.Load()
	if err != nil {
		log.Error("config_load_failed", map[string]any{"err": err.Error()})
		os.Exit(1)
	}

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := metrics.NewRepository(cfg.WindowSize, 64, cfg.HistogramBins, cfg.Percentile)
	pool := metrics.NewWorkerPool(repo, cfg.WorkerCount, cfg.IngestBuffer, log)
	if err := pool.Start(rootCtx); err != nil {
		log.Error("worker_pool_start_failed", map[string]any{"err": err.Error()})
		os.Exit(1)
	}
	svc := metrics.NewService(log, repo, pool, cfg.WindowSize)

	go svc.Start(rootCtx, time.Second)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte("event-driven-metrics backend\n"))
			return
		}
		http.NotFound(w, r)
	})
	handler := metrics.NewHandler(log, svc, cfg.IngestPath, cfg.MetricsPath, cfg.StreamPath)
	handler.Register(mux)
	wsHandler := metrics.NewWSHandler(log, svc, "/api/ws")
	wsHandler.Register(mux)

	chain := httpserver.Chain(
		mux,
		httpserver.Recover(log),
		httpserver.Logger(log),
		httpserver.CORS(),
	)
	srv := httpserver.New(cfg.HTTPAddr, chain, log)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Start(rootCtx); err != nil && !errors.Is(err, context.Canceled) {
			serverErr <- err
		}
	}()

	select {
	case sig := <-sigCh:
		log.Info("signal_received", map[string]any{"signal": sig.String()})
	case err := <-serverErr:
		log.Error("server_failed", map[string]any{"err": err.Error()})
	}

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(cfg.ShutdownTimeout); err != nil {
		log.Warn("server_shutdown_error", map[string]any{"err": err.Error()})
	}
	pool.Close()

	doneCh := make(chan struct{})
	go func() {
		// Allow goroutines a brief moment to flush.
		time.Sleep(50 * time.Millisecond)
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-shutdownCtx.Done():
	}
	log.Info("server_stopped", nil)
}
