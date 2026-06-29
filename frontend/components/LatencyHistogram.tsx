"use client";

import { memo, useMemo } from "react";

import type { Bucket } from "../lib/types";
import { formatNumber } from "../lib/format";

interface Props {
  buckets: ReadonlyArray<Bucket>;
  unit?: string;
}

function LatencyHistogramImpl({ buckets, unit = "ms" }: Props) {
  const max = useMemo(() => {
    let m = 0;
    for (const b of buckets) if (b.count > m) m = b.count;
    return m;
  }, [buckets]);

  return (
    <div className="surface flex h-full flex-col p-6">
      <div className="mb-4 flex items-center justify-between">
        <p className="text-[11px] font-medium uppercase tracking-wider text-zinc-500">
          Latency distribution
        </p>
        <p className="text-[11px] tabular-nums text-zinc-400">
          {buckets.length} buckets
        </p>
      </div>
      <div className="flex min-h-[180px] flex-1 items-end gap-px">
        {buckets.length === 0 ? (
          <p className="m-auto text-xs text-zinc-400">No samples yet.</p>
        ) : (
          buckets.map((b, i) => {
            const h = max === 0 ? 0 : Math.max(2, (b.count / max) * 100);
            return (
              <div
                key={`${b.lower}-${b.upper}-${i}`}
                className="group flex flex-1 flex-col items-center justify-end"
                title={`${formatNumber(b.lower)}–${formatNumber(b.upper)} ${unit}: ${b.count}`}
              >
                <div
                  className="w-full bg-indigo-500/80 transition-colors group-hover:bg-indigo-600"
                  style={{ height: `${h}%` }}
                />
              </div>
            );
          })
        )}
      </div>
      {buckets.length > 0 ? (
        <div className="mt-3 flex items-center justify-between text-[10px] tabular-nums text-zinc-400">
          <span>{formatNumber(buckets[0]?.lower ?? 0)}</span>
          <span>{formatNumber(buckets[buckets.length - 1]?.upper ?? 0)} {unit}</span>
        </div>
      ) : null}
    </div>
  );
}

export const LatencyHistogram = memo(LatencyHistogramImpl);
