import type { MetricStats, Snapshot } from "./types";

// Reduce a sequence of snapshots into a compact history useful for
// the time-series chart. Each entry collapses all per-metric stats
// into a single point in time.
export interface HistoryPoint {
  ts: number;
  total: number;
  perSecond: number;
  mean: number;
  p95: number;
  byMetric: Record<string, number>;
}

const HISTORY_LIMIT = 240;

export function pushHistory(
  history: readonly HistoryPoint[],
  snap: Snapshot,
): HistoryPoint[] {
  const point: HistoryPoint = {
    ts: new Date(snap.generatedAt).getTime() || Date.now(),
    total: snap.totalEvents,
    perSecond: snap.global.perSecond,
    mean: snap.global.mean,
    p95: snap.global.percentileQ,
    byMetric: Object.fromEntries(
      Object.entries(snap.byMetric).map(([k, v]) => [k, v.mean]),
    ),
  };
  const next = history.length >= HISTORY_LIMIT ? history.slice(1) : history.slice();
  next.push(point);
  return next;
}

export function topNMetrics(
  snap: Snapshot,
  n: number,
): ReadonlyArray<MetricStats> {
  return Object.values(snap.byMetric)
    .sort((a, b) => b.perSecond - a.perSecond)
    .slice(0, n);
}
