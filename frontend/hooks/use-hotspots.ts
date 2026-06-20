"use client";

// ============================================================================
// useHotspots – SWR hook for parking hotspot data.
// ============================================================================

import useSWR from "swr";
import { fetchHotspots } from "@/lib/api-client";
import type { HotspotsParams } from "@/types/api";

export function useHotspots(params: HotspotsParams) {
  const key = params.hour != null
    ? ["hotspots", params.hour, params.date, params.police_station, params.risk_level]
    : null;

  const { data, error, isLoading, isValidating, mutate } = useSWR(
    key,
    () => fetchHotspots(params),
    {
      refreshInterval: 30_000,
      shouldRetryOnError: true,
      errorRetryCount: 2,
      revalidateOnFocus: false,
    }
  );

  return {
    hotspots: data?.hotspots ?? [],
    count: data?.count ?? 0,
    filtersApplied: data?.filters_applied ?? {},
    isLoading,
    isValidating,
    error,
    mutate,
  };
}
