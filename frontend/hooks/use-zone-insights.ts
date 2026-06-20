"use client";

// ============================================================================
// useZoneInsights – SWR hook for zone-level ML insights.
// ============================================================================

import useSWR from "swr";
import { fetchZoneInsights } from "@/lib/api-client";
import type { ZoneInsightsParams } from "@/types/api";

export function useZoneInsights(params: ZoneInsightsParams | null) {
  const key =
    params?.zone_id && params?.hour != null
      ? ["zone-insights", params.zone_id, params.hour]
      : null;

  const { data, error, isLoading, isValidating, mutate } = useSWR(
    key,
    () => fetchZoneInsights(params!),
    {
      refreshInterval: 30_000,
      shouldRetryOnError: true,
      errorRetryCount: 2,
      revalidateOnFocus: false,
    }
  );

  return {
    insights: data ?? null,
    isLoading,
    isValidating,
    error,
    mutate,
  };
}
