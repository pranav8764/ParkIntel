// ============================================================================
// ParkIntel API Response Types
// Matches the actual Go backend contracts (not the PRD idealizations).
// ============================================================================

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------
export type HealthResponse = {
  status: string;
};

// ---------------------------------------------------------------------------
// Hotspots
// ---------------------------------------------------------------------------
export type Hotspot = {
  zone_id: string;
  lat: number;
  lng: number;
  police_station: string;
  priority_score: number;
  priority_level: string;
  impact_score: number;
  expected_violations: number;
  predicted_hotspot_risk: string;
  model_confidence: string;
};

export type HotspotsResponse = {
  hotspots: Hotspot[];
  count: number;
  filters_applied: Record<string, unknown>;
};

export type HotspotsParams = {
  hour: number;
  date?: string;
  police_station?: string;
  risk_level?: string;
};

// ---------------------------------------------------------------------------
// Enforcement Ranking
// ---------------------------------------------------------------------------
export type Ranking = {
  rank: number;
  zone_id: string;
  police_station: string;
  junction_name: string;
  expected_violations: number;
  impact_score: number;
  priority_score: number;
  priority_level: string;
  model_confidence: string;
  recommended_action: string;
};

export type RankingsResponse = {
  rankings: Ranking[];
  total: number;
  hour: number;
};

export type RankingParams = {
  hour: number;
  date?: string;
  limit?: number;
};

// ---------------------------------------------------------------------------
// Zone Insights
// ---------------------------------------------------------------------------
export type ClassProbabilities = {
  LOW: number;
  MEDIUM: number;
  HIGH: number;
};

export type ZoneStats = {
  total_historical_violations: number;
  repeat_hotspot_score: number;
  top_violation_types: string[];
  top_vehicle_types: string[];
};

export type ZoneInsightsResponse = {
  zone_id: string;
  predicted_hotspot_risk: string;
  model_confidence: string;
  class_probabilities: ClassProbabilities;
  parking_congestion_impact_score: number;
  priority_score: number;
  priority_level: string;
  recommended_action: string;
  reasons: string[];
  note: string;
  zone_stats: ZoneStats;
};

export type ZoneInsightsParams = {
  zone_id: string;
  hour: number;
};

// ---------------------------------------------------------------------------
// Simulation
// ---------------------------------------------------------------------------
export type SimulateRequest = {
  zone_id: string;
  violation_reduction_percent: number;
};

export type SimulateResponse = {
  zone_id: string;
  violation_reduction_percent: number;
  current_priority_score: number;
  current_priority_level: string;
  simulated_priority_score: number;
  simulated_priority_level: string;
  priority_change: string;
  estimated_impact_reduction: number;
  note: string;
};

// ---------------------------------------------------------------------------
// Shared / Derived
// ---------------------------------------------------------------------------
export type PriorityLevel = "CRITICAL" | "HIGH" | "MEDIUM" | "LOW";
export type RiskLevel = "HIGH" | "MEDIUM" | "LOW";
export type Confidence = "HIGH" | "MODERATE" | "LOW";

export type ApiError = {
  message: string;
  status?: number;
};
