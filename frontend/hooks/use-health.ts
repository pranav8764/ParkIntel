"use client";

// ============================================================================
// useHealth – SWR hook for backend health polling.
// ============================================================================

import useSWR from "swr";
import { fetchHealth } from "@/lib/api-client";

export function useHealth() {
  const { data, error, isLoading, isValidating, mutate } = useSWR(
    "/health",
    () => fetchHealth(),
    {
      refreshInterval: 30_000,
      shouldRetryOnError: true,
      errorRetryCount: 3,
      errorRetryInterval: 5_000,
      revalidateOnFocus: false,
    }
  );

  return {
    health: data,
    isOnline: !!data && !error,
    isOffline: !!error,
    isLoading,
    isValidating,
    error,
    mutate,
  };
}
