package metrics

import (
	"math"
	"sort"
	"sync"
	"time"
)

// PerMetricState holds the running aggregate for a single metric stream
// within a sliding window. It is safe for concurrent use.
type PerMetricState struct {
	mu sync.RWMutex

	metric     string
	binCount   int
	percentile float64

	values    []float64
	times     []time.Time
	errs      []bool
	buckets   []uint64
	loBound   float64
	hiBound   float64
	p2        *P2Quantile
}

// NewPerMetricState constructs an empty state object sized for the
// expected high-water mark. binCount controls histogram resolution.
func NewPerMetricState(metric string, expectedCount int, binCount int, percentile float64) *PerMetricState {
	if expectedCount < 16 {
		expectedCount = 16
	}
	if binCount < 4 {
		binCount = 4
	}
	return &PerMetricState{
		metric:     metric,
		binCount:   binCount,
		percentile: percentile,
		values:     make([]float64, 0, expectedCount),
		times:      make([]time.Time, 0, expectedCount),
		errs:       make([]bool, 0, expectedCount),
		buckets:    make([]uint64, binCount),
		p2:         NewP2Quantile(percentile),
	}
}

// Add records a new sample and evicts any out-of-window samples. The
// caller supplies the current time, the window duration, and whether the
// observation represents an error.
func (s *PerMetricState) Add(value float64, ts time.Time, window time.Duration, isError bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := ts.Add(-window)
	// Evict expired samples from the front (sorted insertion order).
	idx := 0
	for idx < len(s.times) && s.times[idx].Before(cutoff) {
		idx++
	}
	if idx > 0 {
		s.values = s.values[idx:]
		s.times = s.times[idx:]
		s.errs = s.errs[idx:]
	}
	s.values = append(s.values, value)
	s.times = append(s.times, ts)
	s.errs = append(s.errs, isError)
	s.p2.Add(value)
	s.rebuildBucketsLocked()
}

// Snapshot computes the current statistics over the active window.
func (s *PerMetricState) Snapshot(now time.Time, window time.Duration) MetricStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoff := now.Add(-window)
	// Drop expired samples lazily.
	idx := 0
	for idx < len(s.times) && s.times[idx].Before(cutoff) {
		idx++
	}
	var vals []float64
	if idx > 0 {
		vals = append(make([]float64, 0, len(s.values)-idx), s.values[idx:]...)
	} else {
		vals = append([]float64(nil), s.values...)
	}
	var errCount uint64
	for i := idx; i < len(s.errs); i++ {
		if s.errs[i] {
			errCount++
		}
	}
	stats := computeStats(s.metric, vals, s.binCount, window, s.p2.Quantile(), s.percentile)
	stats.ErrorCount = errCount
	if stats.Count > 0 {
		stats.ErrorRate = float64(errCount) / float64(stats.Count)
	}
	stats.Buckets = append([]Bucket(nil), s.bucketBounds()...)
	return stats
}

func (s *PerMetricState) bucketBounds() []Bucket {
	out := make([]Bucket, s.binCount)
	width := (s.hiBound - s.loBound) / float64(s.binCount)
	if width == 0 {
		width = 1
	}
	for i := 0; i < s.binCount; i++ {
		out[i] = Bucket{
			Lower: s.loBound + float64(i)*width,
			Upper: s.loBound + float64(i+1)*width,
			Count: s.buckets[i],
		}
	}
	return out
}

func (s *PerMetricState) rebuildBucketsLocked() {
	if len(s.values) == 0 {
		s.loBound = 0
		s.hiBound = 0
		for i := range s.buckets {
			s.buckets[i] = 0
		}
		return
	}
	minVal, maxVal := s.values[0], s.values[0]
	for _, v := range s.values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	// Pad bounds slightly so identical values do not collapse to zero width.
	if minVal == maxVal {
		minVal = minVal - 0.5
		maxVal = maxVal + 0.5
	}
	s.loBound = minVal
	s.hiBound = maxVal
	width := (maxVal - minVal) / float64(s.binCount)
	for i := range s.buckets {
		s.buckets[i] = 0
	}
	for _, v := range s.values {
		pos := int(math.Floor((v - minVal) / width))
		if pos < 0 {
			pos = 0
		}
		if pos >= s.binCount {
			pos = s.binCount - 1
		}
		s.buckets[pos]++
	}
}

func computeStats(metric string, values []float64, binCount int, window time.Duration, quantile float64, targetPct float64) MetricStats {
	stats := MetricStats{Metric: metric, Min: math.Inf(1), Max: math.Inf(-1)}
	stats.Count = uint64(len(values))
	if stats.Count == 0 {
		stats.Min = 0
		stats.Max = 0
		stats.Buckets = make([]Bucket, binCount)
		return stats
	}
	var sum, sumSq float64
	for _, v := range values {
		sum += v
		sumSq += v * v
		if v < stats.Min {
			stats.Min = v
		}
		if v > stats.Max {
			stats.Max = v
		}
	}
	stats.Sum = sum
	stats.Mean = sum / float64(len(values))
	variance := sumSq/float64(len(values)) - stats.Mean*stats.Mean
	if variance < 0 {
		variance = 0
	}
	stats.Variance = variance
	stats.StdDev = math.Sqrt(variance)
	stats.Percentile = percentileOf(values, targetPct)
	stats.PercentileQ = quantile
	stats.PerSecond = float64(len(values)) / window.Seconds()
	return stats
}

func percentileOf(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	cp := append([]float64(nil), values...)
	sort.Float64s(cp)
	idx := int(float64(len(cp)) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
}
