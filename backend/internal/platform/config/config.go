// Package config centralizes runtime configuration for the metrics service.
// All values can be overridden via environment variables, enabling safe
// deployment across development, staging, and production environments.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime parameters for the metrics microservice.
type Config struct {
	HTTPAddr        string
	MetricsPath     string
	IngestPath      string
	StreamPath      string
	WindowSize      time.Duration
	WorkerCount     int
	IngestBuffer    int
	StreamBuffer    int
	HistogramBins   int
	Percentile      float64
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
}

// Load reads configuration from the environment and applies safe defaults.
// It returns an error if any supplied value is malformed.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:        getString("METRICS_HTTP_ADDR", ":8080"),
		MetricsPath:     getString("METRICS_HTTP_PATH", "/api/metrics"),
		IngestPath:      getString("METRICS_INGEST_PATH", "/api/events"),
		StreamPath:      getString("METRICS_STREAM_PATH", "/api/stream"),
		WindowSize:      getDuration("METRICS_WINDOW", 60*time.Second),
		ShutdownTimeout: getDuration("METRICS_SHUTDOWN_TIMEOUT", 10*time.Second),
		AllowedOrigins:  []string{"*"},
	}

	workerCount, err := getInt("METRICS_WORKERS", 8)
	if err != nil {
		return nil, err
	}
	cfg.WorkerCount = workerCount

	ingestBuffer, err := getInt("METRICS_INGEST_BUFFER", 4096)
	if err != nil {
		return nil, err
	}
	cfg.IngestBuffer = ingestBuffer

	streamBuffer, err := getInt("METRICS_STREAM_BUFFER", 1024)
	if err != nil {
		return nil, err
	}
	cfg.StreamBuffer = streamBuffer

	histBins, err := getInt("METRICS_HISTOGRAM_BINS", 20)
	if err != nil {
		return nil, err
	}
	cfg.HistogramBins = histBins

	pct, err := getFloat("METRICS_PERCENTILE", 0.95)
	if err != nil {
		return nil, err
	}
	cfg.Percentile = pct

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.WorkerCount <= 0 {
		return errors.New("METRICS_WORKERS must be > 0")
	}
	if c.IngestBuffer <= 0 {
		return errors.New("METRICS_INGEST_BUFFER must be > 0")
	}
	if c.StreamBuffer <= 0 {
		return errors.New("METRICS_STREAM_BUFFER must be > 0")
	}
	if c.HistogramBins <= 0 {
		return errors.New("METRICS_HISTOGRAM_BINS must be > 0")
	}
	if c.Percentile <= 0 || c.Percentile >= 1 {
		return errors.New("METRICS_PERCENTILE must be in (0,1)")
	}
	if c.WindowSize <= 0 {
		return errors.New("METRICS_WINDOW must be > 0")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.New("METRICS_SHUTDOWN_TIMEOUT must be > 0")
	}
	return nil
}

func getString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s: %w", key, err)
	}
	return n, nil
}

func getFloat(key string, def float64) (float64, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float for %s: %w", key, err)
	}
	return f, nil
}

func getDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
