"use client";

// ============================================================================
// useSimulation – Imperative mutation hook for policy simulation.
// ============================================================================

import { useState, useCallback } from "react";
import { postSimulation } from "@/lib/api-client";
import { useCommandStore } from "@/store/command-store";
import type { SimulateResponse } from "@/types/api";

export function useSimulation() {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const setActiveSimulation = useCommandStore((s) => s.setActiveSimulation);
  const selectedZoneId = useCommandStore((s) => s.selectedZoneId);
  const reductionPercent = useCommandStore((s) => s.simulationReductionPercent);

  const simulate = useCallback(async (): Promise<SimulateResponse | null> => {
    if (!selectedZoneId) {
      setError(new Error("No zone selected"));
      return null;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      const result = await postSimulation({
        zone_id: selectedZoneId,
        violation_reduction_percent: reductionPercent,
      });
      setActiveSimulation(result);
      return result;
    } catch (err) {
      const e = err instanceof Error ? err : new Error(String(err));
      setError(e);
      return null;
    } finally {
      setIsSubmitting(false);
    }
  }, [selectedZoneId, reductionPercent, setActiveSimulation]);

  const reset = useCallback(() => {
    setActiveSimulation(null);
    setError(null);
  }, [setActiveSimulation]);

  return {
    simulate,
    reset,
    isSubmitting,
    error,
  };
}
