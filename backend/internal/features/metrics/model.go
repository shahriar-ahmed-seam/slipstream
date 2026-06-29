// Package metrics implements the event ingestion, sliding-window
// aggregation, persistence, and real-time streaming features of the
// dashboard backend. The package is designed to be safe for concurrent
// use by many goroutines and to honor context cancellation.
package metrics

import (
	"errors"
	"math"
	"time"
)

// Event is a single time-series observation submitted to the service.
// It is intentionally compact to minimize allocations on the hot path.
type Event struct {
	ID        string            `json:"id,omitempty"`
	Source    string            `json:"source,omitempty"`
	Metric    string            `json:"metric"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	IsError   bool              `json:"isError,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// Validate performs structural validation of an Event payload. A missing
// timestamp is tolerated and defaulted to now by the caller; the metric
// name and a finite value are the only hard requirements.
func (e *Event) Validate(now time.Time) error {
	if e == nil {
		return errors.New("nil event")
	}
	if e.Metric == "" {
		return errors.New("event metric is required")
	}
	if math.IsNaN(e.Value) || math.IsInf(e.Value, 0) {
		return errors.New("event value must be a finite number")
	}
	if !e.Timestamp.IsZero() && e.Timestamp.After(now.Add(5*time.Minute)) {
		return errors.New("event timestamp is too far in the future")
	}
	return nil
}

// Snapshot is the aggregate result computed over the active window.
type Snapshot struct {
	GeneratedAt   time.Time              `json:"generatedAt"`
	WindowSeconds float64                `json:"windowSeconds"`
	TotalEvents   uint64                 `json:"totalEvents"`
	ByMetric      map[string]MetricStats `json:"byMetric"`
	Global        MetricStats            `json:"global"`
}

// MetricStats is a dense statistical summary of one metric stream.
type MetricStats struct {
	Metric      string   `json:"metric"`
	Count       uint64   `json:"count"`
	ErrorCount  uint64   `json:"errorCount"`
	ErrorRate   float64  `json:"errorRate"`
	Sum         float64  `json:"sum"`
	Mean        float64  `json:"mean"`
	Min         float64  `json:"min"`
	Max         float64  `json:"max"`
	Variance    float64  `json:"variance"`
	StdDev      float64  `json:"stdDev"`
	Percentile  float64  `json:"percentile"`
	PercentileQ float64  `json:"percentileQ"`
	Buckets     []Bucket `json:"buckets"`
	PerSecond   float64  `json:"perSecond"`
}

// Bucket represents one bin of the equi-width histogram.
type Bucket struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
	Count uint64  `json:"count"`
}

// IngestResponse is returned to the client after a successful ingest.
type IngestResponse struct {
	Accepted int       `json:"accepted"`
	Rejected int       `json:"rejected"`
	Errors   []string  `json:"errors,omitempty"`
	ServerTs time.Time `json:"serverTs"`
}

// NewSnapshot builds a Snapshot from a set of per-metric states.
func NewSnapshot(now time.Time, window time.Duration, global MetricStats, byMetric map[string]MetricStats) Snapshot {
	return Snapshot{
		GeneratedAt:   now.UTC(),
		WindowSeconds: window.Seconds(),
		TotalEvents:   global.Count,
		ByMetric:      byMetric,
		Global:        global,
	}
}
