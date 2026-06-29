"use client";

import { useCallback, useMemo } from "react";

import { useMetricsStream } from "../hooks/useMetricsStream";
import { useThrottledValue } from "../hooks/useThrottledValue";
import { topNMetrics } from "../lib/aggregator";
import { formatPercent, formatTimestamp } from "../lib/format";
import type { MetricStats, Snapshot } from "../lib/types";

import { ErrorBoundary } from "./ErrorBoundary";
import { KpiCard } from "./KpiCard";
import { LatencyHistogram } from "./LatencyHistogram";
import { TimeSeriesChart } from "./TimeSeriesChart";

const TOP_N = 5;
const HISTORY_CAP = 240;

interface SeriesPoint {
  t: number;
  v: number;
}

export function MetricsDashboard() {
  const { snapshot, history, status, lastError, refresh, reconnect } =
    useMetricsStream({ historyLimit: HISTORY_CAP });

  // Throttle the snapshot to avoid re-render storms during a burst.
  const throttled = useThrottledValue<Snapshot | null>(snapshot, 250);

  const top = useMemo<ReadonlyArray<MetricStats>>(
    () => (throttled ? topNMetrics(throttled, TOP_N) : []),
    [throttled],
  );

  const series = useMemo<ReadonlyArray<SeriesPoint>>(() => {
    if (!throttled) return [];
    const v: number = throttled.byMetric
      ? Object.values(throttled.byMetric).reduce(
          (acc: number, m: MetricStats) => acc + m.count,
          0,
        )
      : throttled.totalEvents;
    const point: SeriesPoint = { t: Date.now(), v };
    return [...history.map((h) => ({ t: h.ts, v: h.total })), point].slice(
      -HISTORY_CAP,
    );
  }, [history, throttled]);

  const latencyFor = useCallback(
    (name: string): number => {
      if (!throttled) return 0;
      const m = throttled.byMetric[name];
      return m ? m.percentileQ : 0;
    },
    [throttled],
  );

  const headline = throttled?.byMetric["http.request.duration_ms"];
  const buckets = headline?.buckets ?? [];

  const errorRatio = throttled ? throttled.global.errorRate : 0;

  const statusColor =
    status === "open"
      ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/40"
      : status === "connecting"
        ? "bg-amber-500/20 text-amber-300 border-amber-500/40"
        : "bg-rose-500/20 text-rose-300 border-rose-500/40";

  return (
    <ErrorBoundary>
      <section className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-white">Live Metrics</h2>
          <p className="text-xs text-slate-400">
            Snapshot @ {throttled ? formatTimestamp(throttled.generatedAt) : "—"} ·{" "}
            {throttled
              ? `${Object.keys(throttled.byMetric).length} metrics`
              : "awaiting first frame"}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span
            className={`rounded-full border px-2 py-1 text-[11px] uppercase tracking-wider ${statusColor}`}
          >
            {status}
          </span>
          <button
            type="button"
            onClick={refresh}
            className="rounded-md border border-ink-700 bg-ink-900 px-3 py-1 text-xs text-slate-300 hover:border-neon-cyan/60"
          >
            Refresh
          </button>
          <button
            type="button"
            onClick={reconnect}
            className="rounded-md border border-ink-700 bg-ink-900 px-3 py-1 text-xs text-slate-300 hover:border-neon-violet/60"
          >
            Reconnect
          </button>
        </div>
      </section>

      {lastError ? (
        <div className="surface border-rose-500/40 p-3 text-sm text-rose-300">
          {lastError}
        </div>
      ) : null}

      <section className="grid grid-cols-1 gap-4 md:grid-cols-4">
        <KpiCard
          label="Events / window"
          value={throttled?.totalEvents ?? 0}
          accent="cyan"
          hint={`window ${(throttled?.windowSeconds ?? 0).toFixed(0)}s`}
        />
        <KpiCard
          label="Distinct metrics"
          value={throttled ? Object.keys(throttled.byMetric).length : 0}
          accent="violet"
        />
        <KpiCard
          label="p95 latency"
          value={latencyFor("http.request.duration_ms")}
          unit="ms"
          accent="lime"
        />
        <KpiCard
          label="Error ratio"
          value={errorRatio * 100}
          unit="%"
          accent="rose"
          hint={`${throttled?.global.errorCount ?? 0} errors`}
        />
      </section>

      <section className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <TimeSeriesChart data={series} label="Events / window" />
        </div>
        <LatencyHistogram buckets={buckets} unit="ms" />
      </section>

      <section className="surface p-4">
        <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-slate-400">
          Top {TOP_N} metrics
        </h3>
        <div className="overflow-x-auto">
          <table className="min-w-full text-left text-sm">
            <thead className="text-xs uppercase tracking-wider text-slate-500">
              <tr>
                <th className="px-3 py-2">Metric</th>
                <th className="px-3 py-2 text-right">Count</th>
                <th className="px-3 py-2 text-right">Rate / s</th>
                <th className="px-3 py-2 text-right">Mean</th>
                <th className="px-3 py-2 text-right">p95</th>
                <th className="px-3 py-2 text-right">StdDev</th>
                <th className="px-3 py-2 text-right">Err %</th>
              </tr>
            </thead>
            <tbody>
              {top.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-3 py-6 text-center text-slate-500">
                    No metrics ingested yet.
                  </td>
                </tr>
              ) : (
                top.map((m: MetricStats) => (
                  <tr
                    key={m.metric}
                    className="border-t border-ink-800 text-slate-200"
                  >
                    <td className="px-3 py-2 font-mono text-xs">{m.metric}</td>
                    <td className="px-3 py-2 text-right">{m.count}</td>
                    <td className="px-3 py-2 text-right">
                      {m.perSecond.toFixed(2)}
                    </td>
                    <td className="px-3 py-2 text-right">{m.mean.toFixed(2)}</td>
                    <td className="px-3 py-2 text-right">
                      {m.percentileQ.toFixed(2)}
                    </td>
                    <td className="px-3 py-2 text-right text-rose-300">
                      {m.stdDev.toFixed(2)}
                    </td>
                    <td className="px-3 py-2 text-right text-rose-300">
                      {formatPercent(m.errorRate)}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </section>
    </ErrorBoundary>
  );
}