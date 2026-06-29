package metrics

import (
	"context"
	"errors"
	"sync"

	"slipstream/internal/platform/logger"
)

// Job is a unit of work submitted to the worker pool. Each Job wraps a
// single Event together with the sink that will receive aggregated state.
type Job struct {
	Event Event
}

// WorkerPool consumes Jobs from a channel and applies them to a
// Repository. It honors context cancellation and uses sync.WaitGroup
// to track in-flight workers, ensuring no goroutine leaks.
type WorkerPool struct {
	log     *logger.Logger
	repo    *Repository
	jobs    chan Job
	wg      sync.WaitGroup
	workers int
	once    sync.Once
	stopped chan struct{}
}

// NewWorkerPool constructs a worker pool with the configured capacity.
func NewWorkerPool(repo *Repository, workers int, buffer int, log *logger.Logger) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}
	if buffer <= 0 {
		buffer = 1
	}
	return &WorkerPool{
		log:     log,
		repo:    repo,
		jobs:    make(chan Job, buffer),
		workers: workers,
		stopped: make(chan struct{}),
	}
}

// Start spawns the configured number of worker goroutines. It returns an
// error if the pool is already started.
func (p *WorkerPool) Start(ctx context.Context) error {
	if p == nil {
		return errors.New("nil worker pool")
	}
	started := false
	p.once.Do(func() { started = true })
	if !started {
		return errors.New("worker pool already started")
	}
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.runWorker(ctx, i)
	}
	go func() {
		p.wg.Wait()
		close(p.stopped)
	}()
	return nil
}

func (p *WorkerPool) runWorker(ctx context.Context, id int) {
	defer p.wg.Done()
	p.log.Debug("worker_started", map[string]any{"id": id})
	for {
		select {
		case <-ctx.Done():
			p.log.Debug("worker_ctx_done", map[string]any{"id": id})
			return
		case job, ok := <-p.jobs:
			if !ok {
				p.log.Debug("worker_channel_closed", map[string]any{"id": id})
				return
			}
			p.repo.Record(job.Event)
		}
	}
}

// Submit enqueues a Job. It respects context cancellation.
func (p *WorkerPool) Submit(ctx context.Context, job Job) error {
	if p == nil {
		return errors.New("nil worker pool")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.jobs <- job:
		return nil
	}
}

// Close stops accepting new Jobs, drains the channel, and waits for all
// workers to exit. Safe to call multiple times.
func (p *WorkerPool) Close() {
	if p == nil {
		return
	}
	defer func() {
		// Recover from potential double-close of stopped channel.
		_ = recover()
	}()
	select {
	case <-p.stopped:
		return
	default:
	}
	close(p.jobs)
	<-p.stopped
	p.log.Info("worker_pool_stopped", map[string]any{"workers": p.workers})
}
