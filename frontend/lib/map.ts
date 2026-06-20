// ============================================================================
// Map Utilities
// Helpers for Mapbox GL integration and GeoJSON conversion.
// ============================================================================

import type { Hotspot } from "@/types/api";
import { riskHex } from "@/lib/priority";

/** Default map configuration from PRD. */
export const MAP_DEFAULTS = {
  /** Map center — Bengaluru, India. Adjust to your deployment region. */
  center: [77.5946, 12.9716] as [number, number],
  zoom: 12,
  pitch: 45,
  bearing: -17,
  style: "mapbox://styles/mapbox/dark-v11",
} as const;

/** GeoJSON feature for a hotspot point. */
export type HotspotFeature = {
  type: "Feature";
  geometry: {
    type: "Point";
    coordinates: [number, number];
  };
  properties: Record<string, unknown>;
};

/** GeoJSON FeatureCollection of hotspot points. */
export type HotspotFeatureCollection = {
  type: "FeatureCollection";
  features: HotspotFeature[];
};

/**
 * Convert hotspot array to a GeoJSON FeatureCollection
 * for use as a Mapbox source.
 */
export function hotspotsToGeoJSON(
  hotspots: Hotspot[]
): HotspotFeatureCollection {
  return {
    type: "FeatureCollection",
    features: hotspots.map((h) => ({
      type: "Feature" as const,
      geometry: {
        type: "Point" as const,
        coordinates: [h.lng, h.lat] as [number, number],
      },
      properties: {
        zone_id: h.zone_id,
        police_station: h.police_station,
        priority_score: h.priority_score,
        priority_level: h.priority_level,
        impact_score: h.impact_score,
        expected_violations: h.expected_violations,
        predicted_hotspot_risk: h.predicted_hotspot_risk,
        model_confidence: h.model_confidence,
        color: riskHex(h.predicted_hotspot_risk),
        /** Circle radius proportional to priority score. */
        radius: Math.max(4, (h.priority_score / 100) * 20),
      },
    })),
  };
}

/**
 * Check if a Mapbox token is available.
 */
export function hasMapboxToken(): boolean {
  const token = process.env.NEXT_PUBLIC_MAPBOX_TOKEN;
  return !!token && token.length > 0;
}
