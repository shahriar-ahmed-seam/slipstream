"use client";

import { memo } from "react";

import { formatNumber } from "../lib/format";

interface Props {
  label: string;
  value: number;
  hint?: string;
  accent?: "indigo" | "violet" | "emerald" | "amber" | "rose";
  unit?: string;
}

const ACCENT: Record<NonNullable<Props["accent"]>, string> = {
  indigo: "bg-indigo-500",
  violet: "bg-violet-500",
  emerald: "bg-emerald-500",
  amber: "bg-amber-500",
  rose: "bg-rose-500",
};

function KpiCardImpl({ label, value, hint, accent = "indigo", unit }: Props) {
  return (
    <div className="surface relative p-6">
      <span className={`absolute inset-x-0 top-0 h-0.5 ${ACCENT[accent]}`} aria-hidden />
      <p className="text-[11px] font-medium uppercase tracking-wider text-zinc-500">
        {label}
      </p>
      <p className="mt-3 text-3xl font-semibold tracking-tight tabular-nums text-zinc-900">
        {formatNumber(value)}
        {unit ? <span className="ml-1 text-base font-normal text-zinc-400">{unit}</span> : null}
      </p>
      {hint ? <p className="mt-2 text-xs text-zinc-500">{hint}</p> : null}
    </div>
  );
}

export const KpiCard = memo(KpiCardImpl);
