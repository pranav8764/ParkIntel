// ============================================================================
// Map Utilities
// Configuration defaults for Leaflet maps.
// ============================================================================

/** Default map configuration. */
export const MAP_DEFAULTS = {
  /** Map center — Bengaluru, India. [Latitude, Longitude] for Leaflet. */
  center: [12.9716, 77.5946] as [number, number],
  zoom: 12,
  tileUrl: "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png",
  attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
} as const;
