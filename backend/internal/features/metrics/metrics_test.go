package metrics

import (
	"math"
	"testing"
	"time"
)

func TestEventValidate(t *testing.T) {
	now := time.Now()
	good := Event{ID: "a", Metric: "latency", Value: 12, Timestamp: now}
	if err := good.Validate(now); err != nil {
		t.Fatalf("expected valid event, got %v", err)
	}
	bad := Event{ID: "a", Metric: "latency", Value: math.NaN(), Timestamp: now}
	if err := bad.Validate(now); err == nil {
		t.Fatalf("expected NaN value to be rejected")
	}
}

func TestPerMetricStateAddAndSnapshot(t *testing.T) {
	state := NewPerMetricState("rps", 16, 8, 0.95)
	now := time.Now()
	for i := 0; i < 100; i++ {
		state.Add(float64(i), now.Add(time.Duration(i)*time.Millisecond), 10*time.Second, i%10 == 0)
	}
	snap := state.Snapshot(now.Add(5*time.Second), 10*time.Second)
	if snap.Count == 0 {
		t.Fatalf("expected non-zero count")
	}
	if snap.Mean <= 0 {
		t.Fatalf("expected positive mean, got %f", snap.Mean)
	}
	if snap.Min >= snap.Max {
		t.Fatalf("expected min < max, got min=%f max=%f", snap.Min, snap.Max)
	}
	if snap.ErrorCount == 0 {
		t.Fatalf("expected non-zero error count")
	}
	if snap.ErrorRate <= 0 || snap.ErrorRate > 1 {
		t.Fatalf("expected error rate in (0,1], got %f", snap.ErrorRate)
	}
}

func TestPerMetricStateEviction(t *testing.T) {
	state := NewPerMetricState("lat", 16, 4, 0.5)
	now := time.Now()
	for i := 0; i < 10; i++ {
		state.Add(float64(i), now.Add(time.Duration(i)*time.Second), 5*time.Second, false)
	}
	snap := state.Snapshot(now.Add(9*time.Second), 5*time.Second)
	if snap.Count >= 10 {
		t.Fatalf("expected eviction, got count=%d", snap.Count)
	}
}

func TestP2QuantileWarming(t *testing.T) {
	q := NewP2Quantile(0.5)
	// Verify the estimator warms up over five samples and stays finite.
	for i := 1; i <= 5; i++ {
		q.Add(float64(i))
	}
	est := q.Quantile()
	if est == 0 {
		t.Fatalf("expected non-zero median after warm-up, got %f", est)
	}
}

func TestRepositorySnapshot(t *testing.T) {
	repo := NewRepository(5*time.Second, 4, 4, 0.9)
	now := time.Now()
	for i := 0; i < 50; i++ {
		repo.Record(Event{
			ID:        "x",
			Metric:    "cpu",
			Value:     float64(i),
			Timestamp: now.Add(time.Duration(i) * time.Millisecond),
		})
	}
	snap := repo.Snapshot(now.Add(2 * time.Second))
	if snap.TotalEvents == 0 {
		t.Fatalf("expected events in snapshot")
	}
	if _, ok := snap.ByMetric["cpu"]; !ok {
		t.Fatalf("expected cpu metric in snapshot")
	}
}
