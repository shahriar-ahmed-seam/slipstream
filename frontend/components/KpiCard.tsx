"use client";

import { memo } from "react";

import { formatNumber } from "../lib/format";

interface Props {
  label: string;
  value: number;
  hint?: string;
  accent?: "cyan" | "violet" | "lime" | "amber" | "rose";
  unit?: string;
}

const ACCENT: Record<NonNullable<Props["accent"]>, string> = {
  cyan: "from-cyan-400/30 to-cyan-500/0 text-cyan-300",
  violet: "from-violet-400/30 to-violet-500/0 text-violet-300",
  lime: "from-lime-400/30 to-lime-500/0 text-lime-300",
  amber: "from-amber-400/30 to-amber-500/0 text-amber-300",
  rose: "from-rose-400/30 to-rose-500/0 text-rose-300",
};

function KpiCardImpl({ label, value, hint, accent = "cyan", unit }: Props) {
  return (
    <div className="surface relative overflow-hidden p-4">
      <div
        className={`pointer-events-none absolute inset-x-0 -top-10 h-24 bg-gradient-to-b ${ACCENT[accent]}`}
      />
      <div className="relative">
        <p className="text-xs uppercase tracking-wider text-slate-400">{label}</p>
        <p className="mt-2 text-2xl font-semibold text-white">
          {formatNumber(value)}
          {unit ? <span className="ml-1 text-sm text-slate-400">{unit}</span> : null}
        </p>
        {hint ? <p className="mt-1 text-xs text-slate-500">{hint}</p> : null}
      </div>
    </div>
  );
}

export const KpiCard = memo(KpiCardImpl);
