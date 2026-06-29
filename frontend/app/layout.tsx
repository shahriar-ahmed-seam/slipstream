import type { Metadata } from "next";
import type { ReactNode } from "react";
import Link from "next/link";

import "./globals.css";

export const metadata: Metadata = {
  title: "Slipstream — Real-Time Metrics",
  description:
    "Real-time, production-grade visualization for high-throughput event streams.",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-canvas font-sans antialiased text-zinc-900">
        <header className="sticky top-0 z-40 border-b border-line bg-white/90 backdrop-blur">
          <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-8">
            <div className="flex items-center gap-3">
              <span className="h-5 w-5 bg-accent" aria-hidden />
              <Link
                href="/"
                className="text-sm font-semibold tracking-tight text-zinc-900"
              >
                Slipstream
              </Link>
            </div>
            <nav className="flex items-center gap-6 text-sm">
              <Link
                href="/metrics"
                className="text-zinc-600 transition hover:text-zinc-900"
              >
                Dashboard
              </Link>
              <a
                href="https://github.com/shahriar-ahmed-seam/slipstream"
                target="_blank"
                rel="noreferrer"
                className="text-zinc-600 transition hover:text-zinc-900"
              >
                GitHub
              </a>
              <span className="border border-line px-2 py-0.5 text-[11px] font-medium uppercase tracking-wider text-zinc-500">
                v1.0.0
              </span>
            </nav>
          </div>
        </header>
        {children}
      </body>
    </html>
  );
}
