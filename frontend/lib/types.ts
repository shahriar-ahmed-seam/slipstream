// Strict TypeScript types mirroring the Go backend's JSON contracts.
// Field names and shapes must remain synchronized with internal/features/metrics/model.go.
//
// Backend JSON keys we mirror:
//   Event:        id, source, metric, value, timestamp, tags
//   Snapshot:     generatedAt, windowSeconds, totalEvents, byMetric, global
//   MetricStats:  metric, count, errorCount, errorRate, sum, mean, min, max,
//                 variance, stdDev, percentile, percentileQ, buckets, perSecond
//   Bucket:       lower, upper, count
//   IngestResp:   accepted, rejected, errors?, serverTs

export interface MetricEvent {
  id: string;
  source: string;
  metric: string;
  value: number;
  timestamp: string; // RFC3339 from Go time.Time
  tags?: Record<string, string>;
}

export interface IngestRequest {
  events: MetricEvent[];
}

export interface IngestResponse {
  accepted: number;
  rejected: number;
  errors?: string[];
  serverTs: string;
}

export interface Bucket {
  lower: number;
  upper: number;
  count: number;
}

export interface MetricStats {
  metric: string;
  count: number;
  errorCount: number;
  errorRate: number;
  sum: number;
  mean: number;
  min: number;
  max: number;
  variance: number;
  stdDev: number;
  percentile: number;
  percentileQ: number;
  buckets: Bucket[];
  perSecond: number;
}

export interface Snapshot {
  generatedAt: string;
  windowSeconds: number;
  totalEvents: number;
  byMetric: Record<string, MetricStats>;
  global: MetricStats;
}

export interface HealthResponse {
  status: "ok" | string;
  ts: string;
  accepted: number;
  rejected: number;
}

export type StreamEvent =
  | { type: "snapshot"; payload: Snapshot }
  | { type: "error"; payload: { message: string } }
  | { type: "open" }
  | { type: "close" };

export interface IngestOptions {
  baseUrl?: string;
  signal?: AbortSignal;
}
