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
  height = 180,
  color = "#4F46E5",
  label,
}: Props) {
  const { path, area } = useMemo(() => {
    if (data.length === 0) return { path: "", area: "" };
    const w = 600;
    const h = height;
    let minV = Infinity;
    let maxV = -Infinity;
    for (const p of data) {
      if (p.v < minV) minV = p.v;
      if (p.v > maxV) maxV = p.v;
    }
    if (!isFinite(minV) || !isFinite(maxV)) return { path: "", area: "" };
    if (minV === maxV) {
      minV -= 1;
      maxV += 1;
    }
    const pad = 8;
    const stepX = w / Math.max(1, data.length - 1);
    const d = data
      .map((p, i) => {
        const x = i * stepX;
        const y =
          pad + (h - pad * 2) - ((p.v - minV) / (maxV - minV)) * (h - pad * 2);
        return `${i === 0 ? "M" : "L"}${x.toFixed(2)},${y.toFixed(2)}`;
      })
      .join(" ");
    return { path: d, area: `${d} L600,${h} L0,${h} Z` };
  }, [data, height]);

  return (
    <div className="surface p-6">
      <div className="mb-4 flex items-center justify-between">
        <p className="text-[11px] font-medium uppercase tracking-wider text-zinc-500">
          {label ?? "Throughput"}
        </p>
        <p className="text-[11px] tabular-nums text-zinc-400">{data.length} pts</p>
      </div>
      <svg
        viewBox={`0 0 600 ${height}`}
        preserveAspectRatio="none"
        className="w-full"
        style={{ height }}
      >
        <defs>
          <linearGradient id="ts-fill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={color} stopOpacity="0.14" />
            <stop offset="100%" stopColor={color} stopOpacity="0" />
          </linearGradient>
        </defs>
        {[0.25, 0.5, 0.75].map((g) => (
          <line
            key={g}
            x1="0"
            x2="600"
            y1={height * g}
            y2={height * g}
            stroke="#EEF0F3"
            strokeWidth={1}
          />
        ))}
        {data.length > 1 ? (
          <>
            <path d={area} fill="url(#ts-fill)" stroke="none" />
            <path
              d={path}
              fill="none"
              stroke={color}
              strokeWidth={2}
              vectorEffect="non-scaling-stroke"
            />
          </>
        ) : (
          <text
            x="300"
            y={height / 2}
            textAnchor="middle"
            fill="#9CA3AF"
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
