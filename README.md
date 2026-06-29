# Event-Driven Metrics Dashboard

Real-time ingestion and visualization of high-throughput time-series events.
A **Go** microservice ingests events through a worker pool, aggregates them over a
sliding window with online statistics (mean, variance, percentiles, histograms),
and streams snapshots to a **Next.js / TypeScript** dashboard over Server-Sent Events.

<p>
  <img alt="Go" src="https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white">
  <img alt="Next.js" src="https://img.shields.io/badge/Next.js-14-000000?logo=next.js&logoColor=white">
  <img alt="TypeScript" src="https://img.shields.io/badge/TypeScript-strict-3178C6?logo=typescript&logoColor=white">
  <img alt="License" src="https://img.shields.io/badge/license-MIT-green">
</p>

```
event-driven/
├── backend/        # Go microservice (sliding window + worker pool + SSE/WS)
├── frontend/       # Next.js 14 / React 18 dashboard (App Router, strict TS)
├── LICENSE
└── README.md
```

---

## 1. Architecture

```
   ┌──────────────────┐   POST /api/events    ┌────────────────────────┐
   │  Event Producers │ ───────────────────▶  │   Go Metrics Service   │
   └──────────────────┘                       │  worker pool → repo    │
                                              │  sliding window + P²   │
                                              └──────────┬─────────────┘
                                                         │ SSE  /api/stream
                                                         │ WS   /api/ws
                                                         │ JSON /api/metrics
                                                         ▼
                                              ┌────────────────────────┐
                                              │  Next.js Dashboard     │
                                              │  (React 18, TS strict) │
                                              └────────────────────────┘
```

### Algorithmic core

For a metric `m`, the service keeps events in a per-metric time series ordered by
timestamp. The sliding window at time `t` is:

```
W_m(t) = { e ∈ E_m | t - w ≤ timestamp(e) ≤ t }
```

where `w` is the window size (default 60s). Aggregates computed over `W_m(t)`:

```
count   = |W_m(t)|
sum     = Σ value(e)
mean    = sum / count
var     = Σ (value(e) - mean)² / count
p95     = quantile_α(values)         exact, over the window
p95~    = P²-estimated quantile_α    online, O(1) memory
errRate = errors / count
```

Eviction is amortized `O(1)`: events are appended in timestamp order, so expired
samples are dropped from the head whenever `timestamp(head) < t - w`. An equi-width
histogram bucketizes the window for the latency distribution chart.

| Operation     | Time (per event)              | Space (per metric) |
|---------------|-------------------------------|--------------------|
| `Add(event)`  | `O(b)` rebuild of `b` buckets | `O(w)`             |
| `Snapshot()`  | `O(k·w)` (k active metrics)   | `O(w + b)`         |
| Eviction head | `O(1)` amortized              | —                  |
| P² quantile   | `O(1)`                        | `O(1)`             |

### Concurrency model

* **Worker pool** — a fixed-size goroutine set drains a bounded channel of jobs.
  Submission is `context`-aware and the pool tracks in-flight work with a
  `sync.WaitGroup`, so shutdown drains cleanly with no goroutine leaks.
* **Repository** — per-metric state behind a `sync.RWMutex`; readers take the read
  lock for snapshots, writers take the write lock on `Add`. A separate mutex guards
  the global P² estimator so records never serialize on a single lock.
* **Service** — orchestrates the pool and repo, keeps atomic accepted/rejected
  counters, and fans snapshots out to subscribers via per-subscriber buffered
  channels (slow consumers drop frames instead of blocking the producer).

---

## 2. Backend

### Requirements

* Go ≥ 1.22

### Configuration (environment variables)

| Variable                   | Default        | Description                       |
|----------------------------|----------------|-----------------------------------|
| `METRICS_HTTP_ADDR`        | `:8080`        | HTTP listen address               |
| `METRICS_INGEST_PATH`      | `/api/events`  | Event ingestion path              |
| `METRICS_HTTP_PATH`        | `/api/metrics` | Snapshot path                     |
| `METRICS_STREAM_PATH`      | `/api/stream`  | SSE stream path                   |
| `METRICS_WINDOW`           | `60s`          | Sliding window duration           |
| `METRICS_WORKERS`          | `8`            | Worker pool size                  |
| `METRICS_INGEST_BUFFER`    | `4096`         | Inbound job queue size            |
| `METRICS_STREAM_BUFFER`    | `1024`         | Per-subscriber channel buffer     |
| `METRICS_HISTOGRAM_BINS`   | `20`           | Histogram bucket count            |
| `METRICS_PERCENTILE`       | `0.95`         | Target percentile, in `(0,1)`     |
| `METRICS_SHUTDOWN_TIMEOUT` | `10s`          | Graceful shutdown deadline        |

### Run

```bash
cd backend
make run          # or: go run ./cmd/server
make seed         # generate synthetic traffic (separate terminal)
make test         # unit tests
make build        # → bin/metrics-server, bin/metrics-seed
```

### HTTP API

| Method | Path           | Notes                                            |
|--------|----------------|--------------------------------------------------|
| GET    | `/api/healthz` | `{"status":"ok","accepted":N,"rejected":N}`      |
| POST   | `/api/events`  | `{"events":[...]}` → `202 {"accepted":N,...}`    |
| GET    | `/api/metrics` | Full `Snapshot` JSON                             |
| GET    | `/api/stream`  | `text/event-stream` SSE; emits `event: snapshot` |
| GET    | `/api/ws`      | WebSocket upgrade; same payload as SSE           |

### Event shape

Only `metric` and a finite `value` are required. `timestamp` defaults to server time
when omitted; `id`, `source`, `isError`, and `tags` are optional.

```json
{
  "metric": "http.request.duration_ms",
  "value": 42.5,
  "timestamp": "2026-01-15T12:00:00Z",
  "isError": false,
  "source": "api-gateway",
  "tags": { "route": "/api/orders", "status": "200" }
}
```

---

## 3. Frontend

### Requirements

* Node.js ≥ 18.17

### Configuration

```bash
cp frontend/.env.local.example frontend/.env.local
# NEXT_PUBLIC_API_BASE=http://localhost:8080
# NEXT_PUBLIC_WS_BASE=ws://localhost:8080
```

### Run

```bash
cd frontend
npm install
npm run dev          # http://localhost:3000  (dashboard at /metrics)
npm run build        # production build
npm run typecheck    # tsc --noEmit
npm run lint         # eslint
```

### Highlights

* **App Router** with **TypeScript strict** + `noUncheckedIndexedAccess`.
* **Custom hooks**
  * `useMetricsStream` — one-shot snapshot fetch, then SSE with exponential-backoff
    reconnect and a bounded history buffer.
  * `useThrottledValue` — coalesces high-frequency updates to avoid re-render storms.
* **Memoized panels** (`KpiCard`, `TimeSeriesChart`, `LatencyHistogram`) keyed on
  primitive props to minimize re-renders.
* **`ErrorBoundary`** isolates panel failures with a retry path.

---

## 4. End-to-end smoke test

```bash
# Terminal 1 — backend
cd backend && make run

# Terminal 2 — synthetic traffic (~200 ev/s with ~7% errors)
cd backend && make seed

# Terminal 3 — frontend
cd frontend && npm run dev
# open http://localhost:3000/metrics
```

Or push a single batch by hand:

```bash
curl -s -X POST http://localhost:8080/api/events \
  -H 'content-type: application/json' \
  -d '{"events":[
    {"metric":"http.request.duration_ms","value":42.5},
    {"metric":"http.request.duration_ms","value":510,"isError":true}
  ]}'
```

---

## 5. License

MIT — see [`LICENSE`](./LICENSE).
