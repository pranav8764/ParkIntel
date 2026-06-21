// ============================================================================
// ParkIntel Command Store (Zustand)
// Global UI state for the Aurelian Command Center.
// ============================================================================

import { create } from "zustand";
import type { SimulateResponse } from "@/types/api";

interface CommandState {
  // --- Zone Selection ---
  selectedZoneId: string | null;
  selectZone: (zoneId: string | null) => void;

  // --- Time Filters ---
  selectedHour: number;
  setSelectedHour: (hour: number) => void;

  selectedDate: string | null; // YYYY-MM-DD or null for "today"
  setSelectedDate: (date: string | null) => void;

  // --- Data Filters ---
  selectedPoliceStation: string | null;
  setSelectedPoliceStation: (station: string | null) => void;

  selectedRiskLevel: string | null;
  setSelectedRiskLevel: (risk: string | null) => void;

  rankingLimit: number;
  setRankingLimit: (limit: number) => void;

  // --- Simulation ---
  simulationReductionPercent: number;
  setSimulationReductionPercent: (percent: number) => void;

  activeSimulation: SimulateResponse | null;
  setActiveSimulation: (result: SimulateResponse | null) => void;

  forecastMode: boolean;
  setForecastMode: (active: boolean) => void;

  // --- Convenience ---
  resetFilters: () => void;
  resetSimulation: () => void;
}

const DEFAULT_SELECTED_HOUR = 18;

/** Current hour of the day (0–23) for default filter. */
function currentHour(): number {
  return new Date().getHours();
}

export function getCurrentHour(): number {
  return currentHour();
}

export const useCommandStore = create<CommandState>((set) => ({
  // --- Zone Selection ---
  selectedZoneId: null,
  selectZone: (zoneId) =>
    set({
      selectedZoneId: zoneId,
      // Clear simulation when zone changes
      activeSimulation: null,
      forecastMode: false,
    }),

  // --- Time Filters ---
  selectedHour: DEFAULT_SELECTED_HOUR,
  setSelectedHour: (hour) => set({ selectedHour: hour }),

  selectedDate: null,
  setSelectedDate: (date) => set({ selectedDate: date }),

  // --- Data Filters ---
  selectedPoliceStation: null,
  setSelectedPoliceStation: (station) =>
    set({ selectedPoliceStation: station }),

  selectedRiskLevel: null,
  setSelectedRiskLevel: (risk) => set({ selectedRiskLevel: risk }),

  rankingLimit: 50,
  setRankingLimit: (limit) => set({ rankingLimit: limit }),

  // --- Simulation ---
  simulationReductionPercent: 20,
  setSimulationReductionPercent: (percent) =>
    set({ simulationReductionPercent: percent }),

  activeSimulation: null,
  setActiveSimulation: (result) =>
    set({
      activeSimulation: result,
      forecastMode: result !== null,
    }),

  forecastMode: false,
  setForecastMode: (active) => set({ forecastMode: active }),

  // --- Convenience ---
  resetFilters: () =>
    set({
      selectedHour: currentHour(),
      selectedDate: null,
      selectedPoliceStation: null,
      selectedRiskLevel: null,
      rankingLimit: 50,
    }),

  resetSimulation: () =>
    set({
      simulationReductionPercent: 20,
      activeSimulation: null,
      forecastMode: false,
    }),
}));
