import type {
  HealthResponse,
  IngestOptions,
  IngestRequest,
  IngestResponse,
  MetricEvent,
  Snapshot,
} from "./types";

const DEFAULT_BASE = process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080";

function buildURL(path: string, base?: string): string {
  const root = (base ?? DEFAULT_BASE).replace(/\/$/, "");
  return `${root}${path.startsWith("/") ? path : `/${path}`}`;
}

async function readError(response: Response): Promise<string> {
  try {
    const data = (await response.json()) as { error?: string };
    if (data && typeof data.error === "string") {
      return data.error;
    }
  } catch {
    // fall through
  }
  return `HTTP ${response.status} ${response.statusText}`;
}

export async function fetchSnapshot(options: IngestOptions = {}): Promise<Snapshot> {
  const response = await fetch(buildURL("/api/metrics", options.baseUrl), {
    method: "GET",
    headers: { Accept: "application/json" },
    cache: "no-store",
    signal: options.signal,
  });
  if (!response.ok) {
    throw new Error(`Snapshot failed: ${await readError(response)}`);
  }
  return (await response.json()) as Snapshot;
}

export async function fetchHealth(options: IngestOptions = {}): Promise<HealthResponse> {
  const response = await fetch(buildURL("/api/healthz", options.baseUrl), {
    method: "GET",
    headers: { Accept: "application/json" },
    cache: "no-store",
    signal: options.signal,
  });
  if (!response.ok) {
    throw new Error(`Health failed: ${await readError(response)}`);
  }
  return (await response.json()) as HealthResponse;
}

export async function ingestEvents(
  events: MetricEvent[],
  options: IngestOptions = {},
): Promise<IngestResponse> {
  const body: IngestRequest = { events };
  const response = await fetch(buildURL("/api/events", options.baseUrl), {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify(body),
    signal: options.signal,
  });
  if (!response.ok) {
    throw new Error(`Ingest failed: ${await readError(response)}`);
  }
  return (await response.json()) as IngestResponse;
}
