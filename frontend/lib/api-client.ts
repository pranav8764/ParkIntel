// ============================================================================
// ParkIntel API Client
// Typed fetch wrappers for the Go inference service.
// ============================================================================

import type {
  HealthResponse,
  HotspotsResponse,
  HotspotsParams,
  RankingsResponse,
  RankingParams,
  ZoneInsightsResponse,
  ZoneInsightsParams,
  SimulateRequest,
  SimulateResponse,
  ApiError,
} from "@/types/api";

function normalizeBaseUrl(rawUrl: string | undefined): string {
  const url = (rawUrl || "http://localhost:8080").trim();
  const withProtocol = /^https?:\/\//i.test(url)
    ? url
    : url.startsWith("//")
      ? `https:${url}`
      : `https://${url}`;

  return withProtocol.replace(/\/+$/, "");
}

const BASE_URL = normalizeBaseUrl(process.env.NEXT_PUBLIC_API_BASE_URL);

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

class ApiClientError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiClientError";
    this.status = status;
  }

  toApiError(): ApiError {
    return { message: this.message, status: this.status };
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const url = `${BASE_URL}${path}`;

  const res = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  if (!res.ok) {
    let message = `Request failed: ${res.status}`;
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
      else if (body?.message) message = body.message;
    } catch {
      // ignore parse failures
    }
    throw new ApiClientError(message, res.status);
  }

  return res.json() as Promise<T>;
}

function toQueryString(params: Record<string, unknown>): string {
  const entries = Object.entries(params).filter(
    ([, v]) => v !== undefined && v !== null && v !== ""
  );
  if (entries.length === 0) return "";
  return "?" + entries.map(([k, v]) => `${k}=${encodeURIComponent(String(v))}`).join("&");
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/** Check backend health. */
export async function fetchHealth(): Promise<HealthResponse> {
  return request<HealthResponse>("/health");
}

/** Fetch parking hotspots for a given hour. */
export async function fetchHotspots(params: HotspotsParams): Promise<HotspotsResponse> {
  const qs = toQueryString(params as Record<string, unknown>);
  return request<HotspotsResponse>(`/api/hotspots${qs}`);
}

/** Fetch enforcement ranking for a given hour. */
export async function fetchRanking(params: RankingParams): Promise<RankingsResponse> {
  const qs = toQueryString(params as Record<string, unknown>);
  return request<RankingsResponse>(`/api/enforcement/ranking${qs}`);
}

/** Fetch zone-level ML insights. */
export async function fetchZoneInsights(
  params: ZoneInsightsParams
): Promise<ZoneInsightsResponse> {
  const { zone_id, hour } = params;
  return request<ZoneInsightsResponse>(`/api/zones/${zone_id}/insights?hour=${hour}`);
}

/** Run a policy simulation for a zone. */
export async function postSimulation(
  body: SimulateRequest
): Promise<SimulateResponse> {
  return request<SimulateResponse>("/api/simulate", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

// SWR-compatible fetcher keyed by URL string
export const swrFetcher = <T>(url: string): Promise<T> => request<T>(url);

export { ApiClientError };
