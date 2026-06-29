"use client";

import { memo, useMemo } from "react";

interface SeriesPoint {
  t: number;
  v: number;
}

interface Props {
  data: ReadonlyArray<SeriesPoint>;
  height?: number;
  color?: string;
  label?: string;
}

function TimeSeriesChartImpl({
  data,
  height = 160,
  color = "#22D3EE",
  label,
}: Props) {
  const path = useMemo(() => {
    if (data.length === 0) return "";
    const w = 600;
    const h = height;
    const min = data[0]?.t ?? 0;
    const max = data[data.length - 1]?.t ?? 1;
    const span = Math.max(1, max - min);
    let minV = Infinity;
    let maxV = -Infinity;
    for (const p of data) {
      if (p.v < minV) minV = p.v;
      if (p.v > maxV) maxV = p.v;
    }
    if (!isFinite(minV) || !isFinite(maxV)) return "";
    if (minV === maxV) {
      minV = minV - 1;
      maxV = maxV + 1;
    }
    const stepX = w / Math.max(1, data.length - 1);
    return data
      .map((p, i) => {
        const x = i * stepX;
        const y = h - ((p.v - minV) / (maxV - minV)) * h;
        return `${i === 0 ? "M" : "L"}${x.toFixed(2)},${y.toFixed(2)}`;
      })
      .join(" ");
  }, [data, height]);

  return (
    <div className="surface p-4">
      <div className="mb-2 flex items-center justify-between">
        <p className="text-xs uppercase tracking-wider text-slate-400">
          {label ?? "Throughput"}
        </p>
        <p className="text-xs text-slate-500">{data.length} pts</p>
      </div>
      <svg
        viewBox={`0 0 600 ${height}`}
        preserveAspectRatio="none"
        className="h-40 w-full"
      >
        <defs>
          <linearGradient id="ts-fill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={color} stopOpacity="0.35" />
            <stop offset="100%" stopColor={color} stopOpacity="0" />
          </linearGradient>
        </defs>
        {data.length > 1 ? (
          <>
            <path
              d={`${path} L600,${height} L0,${height} Z`}
              fill="url(#ts-fill)"
              stroke="none"
            />
            <path d={path} fill="none" stroke={color} strokeWidth={1.5} />
          </>
        ) : (
          <text
            x="300"
            y={height / 2}
            textAnchor="middle"
            fill="#475569"
            fontSize="12"
          >
            Waiting for data…
          </text>
        )}
      </svg>
    </div>
  );
}

export const TimeSeriesChart = memo(TimeSeriesChartImpl);
