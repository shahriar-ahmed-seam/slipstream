import type { Metadata } from "next";
import type { ReactNode } from "react";

import "./globals.css";

export const metadata: Metadata = {
  title: "Slipstream — Real-Time Metrics",
  description:
    "Real-time, production-grade visualization for high-throughput event streams.",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body className="font-sans antialiased text-slate-200">
        <div className="mx-auto max-w-7xl px-6 py-8">
          <header className="mb-8 flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight text-white">
                Slipstream
              </h1>
              <p className="text-sm text-slate-400">
                Sliding-window metrics streamed in real time from the Go engine.
              </p>
            </div>
            <div className="rounded-full border border-ink-700 bg-ink-900 px-3 py-1 text-xs text-slate-400">
              v1.0.0
            </div>
          </header>
          {children}
        </div>
      </body>
    </html>
  );
}