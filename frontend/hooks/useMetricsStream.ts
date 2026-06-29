"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import { fetchSnapshot } from "../lib/api";
import { pushHistory, type HistoryPoint } from "../lib/aggregator";
import type { Snapshot } from "../lib/types";

export type StreamStatus = "idle" | "connecting" | "open" | "error" | "closed";

export interface UseMetricsStreamResult {
  snapshot: Snapshot | null;
  history: HistoryPoint[];
  status: StreamStatus;
  lastError: string | null;
  refresh: () => Promise<void>;
  reconnect: () => void;
}

export interface UseMetricsStreamOptions {
  baseUrl?: string;
  initial?: Snapshot | null;
  historyLimit?: number;
  maxRetries?: number;
}

// useMetricsStream drives the dashboard's data pipeline:
//   1. Performs a one-shot snapshot fetch on mount so the UI has data
//      before the stream connects.
//   2. Opens a Server-Sent Events (SSE) connection to the backend. SSE
//      is preferred over WebSocket for one-way fan-out from server to
//      browser; the backend exposes a `/api/stream` endpoint for it.
//   3. Maintains a bounded history buffer of HistoryPoint derived from
//      the latest snapshots.
//   4. Reconnects automatically with exponential backoff on failure.
//
// Re-renders are minimized by:
//   - Using functional state updates that compare identity
//   - Throttling external setStates in the consumer via useThrottledValue
//   - Keeping refs for mutable values that don't affect rendering
export function useMetricsStream(
  options: UseMetricsStreamOptions = {},
): UseMetricsStreamResult {
  const [snapshot, setSnapshot] = useState<Snapshot | null>(options.initial ?? null);
  const [history, setHistory] = useState<HistoryPoint[]>([]);
  const [status, setStatus] = useState<StreamStatus>("idle");
  const [lastError, setLastError] = useState<string | null>(null);

  const retryRef = useRef<number>(0);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const controllerRef = useRef<AbortController | null>(null);
  const mountedRef = useRef<boolean>(true);
  const historyRef = useRef<HistoryPoint[]>([]);
  const maxRetries = options.maxRetries ?? 8;

  const applySnapshot = useCallback((snap: Snapshot) => {
    setSnapshot(snap);
    const next = pushHistory(historyRef.current, snap);
    historyRef.current = next;
    setHistory(next);
  }, []);

  const refresh = useCallback(async () => {
    if (!mountedRef.current) return;
    try {
      const snap = await fetchSnapshot({ baseUrl: options.baseUrl });
      if (!mountedRef.current) return;
      applySnapshot(snap);
    } catch (err) {
      if (!mountedRef.current) return;
      const message = err instanceof Error ? err.message : String(err);
      setLastError(message);
    }
  }, [applySnapshot, options.baseUrl]);

  const connect = useCallback(async () => {
    if (!mountedRef.current) return;
    if (typeof window === "undefined") return;
    if (typeof EventSource === "undefined") {
      setStatus("error");
      setLastError("EventSource is not available in this environment");
      return;
    }

    setStatus("connecting");
    const base = (options.baseUrl ?? process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080")
      .replace(/\/$/, "");
    const url = `${base}/api/stream`;

    let source: EventSource;
    try {
      source = new EventSource(url, { withCredentials: false });
    } catch (err) {
      setStatus("error");
      setLastError(err instanceof Error ? err.message : String(err));
      return;
    }

    controllerRef.current = new AbortController();

    source.addEventListener("open", () => {
      if (!mountedRef.current) return;
      retryRef.current = 0;
      setStatus("open");
      setLastError(null);
    });

    source.addEventListener("snapshot", (event: MessageEvent<string>) => {
      if (!mountedRef.current) return;
      try {
        const parsed = JSON.parse(event.data) as Snapshot;
        applySnapshot(parsed);
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        setLastError(message);
      }
    });

    source.addEventListener("error", () => {
      if (!mountedRef.current) return;
      source.close();
      setStatus("error");
      const next = retryRef.current + 1;
      retryRef.current = next;
      if (next > maxRetries) {
        setLastError("Connection lost; giving up after maximum retries");
        return;
      }
      const delay = Math.min(15_000, 500 * 2 ** (next - 1));
      timerRef.current = setTimeout(() => {
        timerRef.current = null;
        if (mountedRef.current) {
          void connect();
        }
      }, delay);
    });

    return () => {
      source.close();
    };
  }, [applySnapshot, maxRetries, options.baseUrl]);

  const reconnect = useCallback(() => {
    if (controllerRef.current) {
      controllerRef.current.abort();
      controllerRef.current = null;
    }
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
    retryRef.current = 0;
    void connect();
  }, [connect]);

  useEffect(() => {
    mountedRef.current = true;
    void refresh();
    void connect();
    return () => {
      mountedRef.current = false;
      if (controllerRef.current) {
        controllerRef.current.abort();
        controllerRef.current = null;
      }
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    };
    // connect and refresh are stable; intentionally run once on mount.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return { snapshot, history, status, lastError, refresh, reconnect };
}
