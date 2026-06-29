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

const STATUS_STYLE: Record<string, string> = {
  open: "bg-emerald-50 text-emerald-700 border-emerald-200",
  connecting: "bg-amber-50 text-amber-700 border-amber-200",
  idle: "bg-zinc-50 text-zinc-600 border-zinc-200",
  closed: "bg-zinc-50 text-zinc-600 border-zinc-200",
  error: "bg-rose-50 text-rose-700 border-rose-200",
};

export function MetricsDashboard() {
  const { snapshot, history, status, lastError, refresh, reconnect } =
    useMetricsStream({ historyLimit: HISTORY_CAP });

  const throttled = useThrottledValue<Snapshot | null>(snapshot, 250);

  const top = useMemo<ReadonlyArray<MetricStats>>(
    () => (throttled ? topNMetrics(throttled, TOP_N) : []),
    [throttled],
  );

  const series = useMemo<ReadonlyArray<SeriesPoint>>(() => {
    if (!throttled) return [];
    const point: SeriesPoint = { t: Date.now(), v: throttled.totalEvents };
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
  const statusStyle = STATUS_STYLE[status] ?? STATUS_STYLE.idle;

  return (
    <ErrorBoundary>
      {/* Hero header */}
      <section className="border-b border-line bg-white">
        <div className="mx-auto flex max-w-7xl flex-col gap-6 px-8 py-12 md:flex-row md:items-end md:justify-between">
          <div>
            <span className="text-[11px] font-medium uppercase tracking-wider text-zinc-500">
              Live operational view
            </span>
            <h1 className="mt-2 text-4xl font-semibold tracking-tight text-zinc-900">
              Live Metrics
            </h1>
            <p className="mt-3 text-sm text-zinc-500">
              Snapshot @ {throttled ? formatTimestamp(throttled.generatedAt) : "—"}
              <span className="mx-2 text-zinc-300">·</span>
              {throttled
                ? `${Object.keys(throttled.byMetric).length} metrics tracked`
                : "awaiting first frame"}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <span
              className={`inline-flex items-center gap-2 border px-3 py-1.5 text-[11px] font-medium uppercase tracking-wider ${statusStyle}`}
            >
              <span className="h-1.5 w-1.5 rounded-full bg-current" aria-hidden />
              {status}
            </span>
            <button
              type="button"
              onClick={refresh}
              className="border border-zinc-300 bg-white px-4 py-1.5 text-xs font-medium text-zinc-700 transition hover:border-zinc-400 hover:bg-zinc-50"
            >
              Refresh
            </button>
            <button
              type="button"
              onClick={reconnect}
              className="border border-zinc-300 bg-white px-4 py-1.5 text-xs font-medium text-zinc-700 transition hover:border-zinc-400 hover:bg-zinc-50"
            >
              Reconnect
            </button>
          </div>
        </div>
      </section>

      <div className="mx-auto max-w-7xl space-y-6 px-8 py-10">
        {lastError ? (
          <div className="border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
            {lastError}
          </div>
        ) : null}

        {/* KPI grid */}
        <section className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
          <KpiCard
            label="Events / window"
            value={throttled?.totalEvents ?? 0}
            accent="indigo"
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
            accent="emerald"
          />
          <KpiCard
            label="Error ratio"
            value={errorRatio * 100}
            unit="%"
            accent="rose"
            hint={`${throttled?.global.errorCount ?? 0} errors in window`}
          />
        </section>

        {/* Charts */}
        <section className="grid grid-cols-1 gap-5 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <TimeSeriesChart data={series} label="Events / window" />
          </div>
          <LatencyHistogram buckets={buckets} unit="ms" />
        </section>

        {/* Top metrics table */}
        <section className="surface">
          <div className="border-b border-line px-6 py-4">
            <h2 className="text-[11px] font-medium uppercase tracking-wider text-zinc-500">
              Top {TOP_N} metrics by rate
            </h2>
          </div>
          <div className="overflow-x-auto">
            <table className="min-w-full text-left text-sm">
              <thead>
                <tr className="text-[11px] uppercase tracking-wider text-zinc-400">
                  <th className="px-6 py-3 font-medium">Metric</th>
                  <th className="px-6 py-3 text-right font-medium">Count</th>
                  <th className="px-6 py-3 text-right font-medium">Rate / s</th>
                  <th className="px-6 py-3 text-right font-medium">Mean</th>
                  <th className="px-6 py-3 text-right font-medium">p95</th>
                  <th className="px-6 py-3 text-right font-medium">StdDev</th>
                  <th className="px-6 py-3 text-right font-medium">Err %</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-line">
                {top.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="px-6 py-10 text-center text-zinc-400">
                      No metrics ingested yet.
                    </td>
                  </tr>
                ) : (
                  top.map((m: MetricStats) => (
                    <tr key={m.metric} className="transition-colors hover:bg-zinc-50">
                      <td className="px-6 py-3 font-mono text-xs text-zinc-700">
                        {m.metric}
                      </td>
                      <td className="px-6 py-3 text-right tabular-nums text-zinc-900">
                        {m.count}
                      </td>
                      <td className="px-6 py-3 text-right tabular-nums text-zinc-900">
                        {m.perSecond.toFixed(2)}
                      </td>
                      <td className="px-6 py-3 text-right tabular-nums text-zinc-900">
                        {m.mean.toFixed(2)}
                      </td>
                      <td className="px-6 py-3 text-right tabular-nums text-zinc-900">
                        {m.percentileQ.toFixed(2)}
                      </td>
                      <td className="px-6 py-3 text-right tabular-nums text-zinc-500">
                        {m.stdDev.toFixed(2)}
                      </td>
                      <td className="px-6 py-3 text-right tabular-nums text-rose-600">
                        {formatPercent(m.errorRate)}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </ErrorBoundary>
  );
}
