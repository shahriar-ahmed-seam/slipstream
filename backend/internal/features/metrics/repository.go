package metrics

import (
	"sync"
	"time"
)

// Repository is the in-memory storage layer for per-metric state.
// It is safe for concurrent access and supports read snapshots.
type Repository struct {
	mu          sync.RWMutex
	states      map[string]*PerMetricState
	expectedCap int
	binCount    int
	percentile  float64
	window      time.Duration
	globalP2    *P2Quantile
	globalP2Mu  sync.Mutex
}

// NewRepository constructs a Repository sized for the expected high-water
// cardinality of metric names.
func NewRepository(window time.Duration, expectedCap int, binCount int, percentile float64) *Repository {
	if expectedCap < 8 {
		expectedCap = 8
	}
	return &Repository{
		states:      make(map[string]*PerMetricState, expectedCap),
		expectedCap: expectedCap,
		binCount:    binCount,
		percentile:  percentile,
		window:      window,
		globalP2:    NewP2Quantile(percentile),
	}
}

// stateFor returns or lazily creates a per-metric state.
func (r *Repository) stateFor(metric string) *PerMetricState {
	r.mu.RLock()
	st, ok := r.states[metric]
	r.mu.RUnlock()
	if ok {
		return st
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if st, ok = r.states[metric]; ok {
		return st
	}
	st = NewPerMetricState(metric, 1024, r.binCount, r.percentile)
	r.states[metric] = st
	return st
}

// Record adds a new event to the appropriate per-metric state.
func (r *Repository) Record(ev Event) {
	if ev.Metric == "" {
		return
	}
	ts := ev.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	state := r.stateFor(ev.Metric)
	state.Add(ev.Value, ts, r.window, ev.IsError)

	r.globalP2Mu.Lock()
	r.globalP2.Add(ev.Value)
	r.globalP2Mu.Unlock()
}

// Snapshot returns the aggregate snapshot for all tracked metrics.
func (r *Repository) Snapshot(now time.Time) Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	global := MetricStats{Metric: "_global", Min: 1e308, Max: -1e308}
	by := make(map[string]MetricStats, len(r.states))
	var totalSum float64
	var totalSumSq float64
	var totalCount uint64
	var totalErrors uint64

	for name, st := range r.states {
		s := st.Snapshot(now, r.window)
		by[name] = s
		totalCount += s.Count
		totalErrors += s.ErrorCount
		totalSum += s.Sum
		totalSumSq += s.Sum * s.Sum
		if s.Count > 0 {
			if s.Min < global.Min {
				global.Min = s.Min
			}
			if s.Max > global.Max {
				global.Max = s.Max
			}
		}
	}
	global.Count = totalCount
	global.ErrorCount = totalErrors
	global.Sum = totalSum
	if totalCount > 0 {
		global.Mean = totalSum / float64(totalCount)
		diff := totalSumSq/float64(totalCount) - global.Mean*global.Mean
		if diff > 0 {
			global.Variance = diff
			global.StdDev = sqrt(diff)
		}
		global.PerSecond = float64(totalCount) / r.window.Seconds()
		global.ErrorRate = float64(totalErrors) / float64(totalCount)
	} else {
		global.Min = 0
		global.Max = 0
	}
	r.globalP2Mu.Lock()
	gq := r.globalP2.Quantile()
	r.globalP2Mu.Unlock()
	global.Percentile = gq
	global.PercentileQ = gq
	return NewSnapshot(now, r.window, global, by)
}

// Names returns the set of known metric names.
func (r *Repository) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.states))
	for k := range r.states {
		out = append(out, k)
	}
	return out
}

// sqrt avoids importing math at file scope; kept as a private helper to
// keep package imports tidy.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 16; i++ {
		z = (z + x/z) / 2
	}
	return z
}
