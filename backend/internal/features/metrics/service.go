package metrics

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"slipstream/internal/platform/logger"
)

// Service orchestrates the worker pool, repository, and stream
// subscribers. It is the application-layer entry point for the feature.
type Service struct {
	log    *logger.Logger
	repo   *Repository
	pool   *WorkerPool
	window time.Duration

	subsMu sync.RWMutex
	subs   map[chan Snapshot]struct{}

	accepted atomic.Uint64
	rejected atomic.Uint64
}

// NewService wires the service together from its collaborators.
func NewService(log *logger.Logger, repo *Repository, pool *WorkerPool, window time.Duration) *Service {
	return &Service{
		log:    log,
		repo:   repo,
		pool:   pool,
		window: window,
		subs:   make(map[chan Snapshot]struct{}),
	}
}

// Start launches the snapshot broadcaster. It returns when ctx is canceled.
func (s *Service) Start(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.log.Info("service_started", map[string]any{"interval": interval.String()})
	for {
		select {
		case <-ctx.Done():
			s.log.Info("service_stopping", map[string]any{"reason": ctx.Err().Error()})
			s.closeAllSubs()
			return
		case now := <-ticker.C:
			snap := s.repo.Snapshot(now)
			s.broadcast(snap)
		}
	}
}

// Ingest validates and submits a batch of events. It returns a response
// describing how many were accepted and any errors.
func (s *Service) Ingest(ctx context.Context, events []Event) (IngestResponse, error) {
	if s == nil {
		return IngestResponse{}, errors.New("nil service")
	}
	resp := IngestResponse{ServerTs: time.Now().UTC()}
	now := time.Now()
	for i := range events {
		ev := events[i]
		if err := ev.Validate(now); err != nil {
			resp.Rejected++
			resp.Errors = append(resp.Errors, err.Error())
			s.rejected.Add(1)
			continue
		}
		if err := s.pool.Submit(ctx, Job{Event: ev}); err != nil {
			resp.Rejected++
			resp.Errors = append(resp.Errors, err.Error())
			s.rejected.Add(1)
			continue
		}
		resp.Accepted++
		s.accepted.Add(1)
	}
	return resp, nil
}

// Snapshot returns the current aggregate.
func (s *Service) Snapshot() Snapshot {
	return s.repo.Snapshot(time.Now())
}

// Counters returns the cumulative ingest counters.
func (s *Service) Counters() (accepted, rejected uint64) {
	return s.accepted.Load(), s.rejected.Load()
}

// Subscribe registers a channel to receive periodic snapshots. The
// returned channel is closed when ctx is canceled or the service stops.
func (s *Service) Subscribe(ctx context.Context) <-chan Snapshot {
	ch := make(chan Snapshot, 8)
	s.subsMu.Lock()
	s.subs[ch] = struct{}{}
	s.subsMu.Unlock()

	go func() {
		<-ctx.Done()
		s.unsubscribe(ch)
	}()
	return ch
}

func (s *Service) unsubscribe(ch chan Snapshot) {
	s.subsMu.Lock()
	if _, ok := s.subs[ch]; ok {
		delete(s.subs, ch)
		close(ch)
	}
	s.subsMu.Unlock()
}

func (s *Service) closeAllSubs() {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	for ch := range s.subs {
		close(ch)
		delete(s.subs, ch)
	}
}

func (s *Service) broadcast(snap Snapshot) {
	s.subsMu.RLock()
	defer s.subsMu.RUnlock()
	for ch := range s.subs {
		select {
		case ch <- snap:
		default:
			// Drop on slow consumer; ensures producer never blocks.
		}
	}
}
