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
    <div className="surface p-4">
      <div className="mb-3 flex items-center justify-between">
        <p className="text-xs uppercase tracking-wider text-slate-400">
          Latency Distribution
        </p>
        <p className="text-xs text-slate-500">{buckets.length} buckets</p>
      </div>
      <div className="flex h-40 items-end gap-1">
        {buckets.length === 0 ? (
          <p className="m-auto text-xs text-slate-500">No samples yet.</p>
        ) : (
          buckets.map((b, i) => {
            const h = max === 0 ? 0 : (b.count / max) * 100;
            return (
              <div
                key={`${b.lower}-${b.upper}-${i}`}
                className="group flex flex-1 flex-col items-center justify-end"
                title={`${formatNumber(b.lower)}–${formatNumber(b.upper)} ${unit}: ${b.count}`}
              >
                <div
                  className="w-full rounded-t bg-gradient-to-t from-violet-500/40 to-violet-300/80 transition-all group-hover:from-violet-500/60 group-hover:to-violet-200"
                  style={{ height: `${h}%` }}
                />
                <span className="mt-1 text-[10px] text-slate-500">
                  {formatNumber(b.lower)}
                </span>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}

export const LatencyHistogram = memo(LatencyHistogramImpl);
