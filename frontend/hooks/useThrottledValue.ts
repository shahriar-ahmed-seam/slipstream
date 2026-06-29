"use client";

import { useEffect, useRef, useState } from "react";

// useThrottledValue delays updates of `value` so the consuming component
// does not re-render more frequently than `intervalMs`. The latest value
// is always flushed after the throttle window closes. This keeps the
// dashboard responsive under a high-frequency event stream.
export function useThrottledValue<T>(value: T, intervalMs = 200): T {
  const [throttled, setThrottled] = useState<T>(value);
  const lastFlush = useRef<number>(0);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pending = useRef<T>(value);
  const mounted = useRef<boolean>(true);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      if (timer.current) {
        clearTimeout(timer.current);
        timer.current = null;
      }
    };
  }, []);

  useEffect(() => {
    pending.current = value;
    const now = Date.now();
    const elapsed = now - lastFlush.current;
    if (elapsed >= intervalMs) {
      lastFlush.current = now;
      setThrottled(value);
      return;
    }
    if (timer.current) return;
    timer.current = setTimeout(() => {
      timer.current = null;
      lastFlush.current = Date.now();
      if (mounted.current) {
        setThrottled(pending.current);
      }
    }, intervalMs - elapsed);
  }, [value, intervalMs]);

  return throttled;
}
