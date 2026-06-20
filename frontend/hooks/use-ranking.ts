"use client";

// ============================================================================
// useRanking – SWR hook for enforcement ranking data.
// ============================================================================

import useSWR from "swr";
import { fetchRanking } from "@/lib/api-client";
import type { RankingParams } from "@/types/api";

export function useRanking(params: RankingParams) {
  const key = params.hour != null
    ? ["ranking", params.hour, params.date, params.limit]
    : null;

  const { data, error, isLoading, isValidating, mutate } = useSWR(
    key,
    () => fetchRanking(params),
    {
      refreshInterval: 30_000,
      shouldRetryOnError: true,
      errorRetryCount: 2,
      revalidateOnFocus: false,
    }
  );

  return {
    rankings: data?.rankings ?? [],
    total: data?.total ?? 0,
    hour: data?.hour,
    isLoading,
    isValidating,
    error,
    mutate,
  };
}
