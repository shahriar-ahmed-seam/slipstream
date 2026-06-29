import Link from "next/link";

const STATS = [
  { label: "Ingest path", value: "Worker pool" },
  { label: "Aggregation", value: "Sliding window" },
  { label: "Percentiles", value: "P² online" },
  { label: "Transport", value: "SSE · WebSocket" },
];

export default function HomePage() {
  return (
    <main>
      {/* Hero */}
      <section className="border-b border-line bg-white">
        <div className="mx-auto max-w-7xl px-8 py-24">
          <div className="max-w-3xl">
            <span className="inline-flex items-center gap-2 border border-line px-3 py-1 text-[11px] font-medium uppercase tracking-wider text-zinc-500">
              <span className="h-1.5 w-1.5 bg-emerald-500" aria-hidden />
              Real-time event metrics
            </span>
            <h1 className="mt-6 text-5xl font-semibold leading-[1.05] tracking-tight text-zinc-900 sm:text-6xl">
              Metrics in the slipstream of your event firehose.
            </h1>
            <p className="mt-6 max-w-2xl text-lg leading-relaxed text-zinc-600">
              Slipstream ingests high-throughput time-series events through a Go engine,
              aggregates them over a sliding window with online statistics, and streams
              live snapshots to this dashboard over Server-Sent Events.
            </p>
            <div className="mt-10 flex flex-wrap items-center gap-3">
              <Link
                href="/metrics"
                className="inline-flex items-center bg-accent px-6 py-3 text-sm font-medium text-white transition hover:bg-accent-strong"
              >
                Open the dashboard
              </Link>
              <a
                href="https://github.com/shahriar-ahmed-seam/slipstream"
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center border border-zinc-300 bg-white px-6 py-3 text-sm font-medium text-zinc-800 transition hover:border-zinc-400 hover:bg-zinc-50"
              >
                View source
              </a>
            </div>
          </div>
        </div>

        {/* Stat strip */}
        <div className="border-t border-line">
          <div className="mx-auto grid max-w-7xl grid-cols-2 md:grid-cols-4">
            {STATS.map((s, i) => (
              <div
                key={s.label}
                className={`px-8 py-6 ${i !== 0 ? "border-l border-line" : ""}`}
              >
                <p className="text-[11px] uppercase tracking-wider text-zinc-500">
                  {s.label}
                </p>
                <p className="mt-1 text-base font-medium text-zinc-900">{s.value}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
