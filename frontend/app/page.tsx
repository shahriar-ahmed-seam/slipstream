import Link from "next/link";

export default function HomePage() {
  return (
    <main className="surface p-8">
      <h2 className="text-xl font-semibold text-white">Welcome</h2>
      <p className="mt-2 max-w-2xl text-sm text-slate-300">
        This is the operational dashboard for the Event-Driven Metrics microservice. The
        backend aggregates time-series events over a sliding window and pushes snapshots
        via Server-Sent Events.
      </p>
      <Link
        href="/metrics"
        className="mt-6 inline-flex items-center rounded-lg border border-neon-cyan/40 bg-ink-900 px-4 py-2 text-sm font-medium text-neon-cyan shadow-glow transition hover:border-neon-cyan"
      >
        Open Metrics Dashboard →
      </Link>
    </main>
  );
}