"use client";

// ============================================================================
// Command Map — Mapbox GL integration with hotspot visualization.
// Falls back to a styled placeholder when no Mapbox token is available.
// ============================================================================

import { useEffect, useRef, useState } from "react";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { useCommandStore } from "@/store/command-store";
import { useHotspots } from "@/hooks/use-hotspots";
import { hotspotsToGeoJSON, MAP_DEFAULTS, hasMapboxToken } from "@/lib/map";
import { MapPin, Layers } from "lucide-react";

const HOTSPOT_SOURCE = "hotspots-source";
const HOTSPOT_LAYER = "hotspots-layer";
const HOTSPOT_SELECTED_LAYER = "hotspots-selected-layer";

export function CommandMap() {
  const mapContainer = useRef<HTMLDivElement>(null);
  const mapRef = useRef<mapboxgl.Map | null>(null);
  const [mapLoaded, setMapLoaded] = useState(false);

  const selectedHour = useCommandStore((s) => s.selectedHour);
  const selectedDate = useCommandStore((s) => s.selectedDate);
  const selectedPoliceStation = useCommandStore((s) => s.selectedPoliceStation);
  const selectedRiskLevel = useCommandStore((s) => s.selectedRiskLevel);
  const selectedZoneId = useCommandStore((s) => s.selectedZoneId);
  const selectZone = useCommandStore((s) => s.selectZone);

  const { hotspots } = useHotspots({
    hour: selectedHour,
    date: selectedDate ?? undefined,
    police_station: selectedPoliceStation ?? undefined,
    risk_level: selectedRiskLevel ?? undefined,
  });

  const tokenAvailable = hasMapboxToken();

  // ── Initialize Mapbox ──
  useEffect(() => {
    if (!tokenAvailable || !mapContainer.current || mapRef.current) return;

    mapboxgl.accessToken = process.env.NEXT_PUBLIC_MAPBOX_TOKEN!;

    const map = new mapboxgl.Map({
      container: mapContainer.current,
      style: MAP_DEFAULTS.style,
      center: MAP_DEFAULTS.center,
      zoom: MAP_DEFAULTS.zoom,
      pitch: MAP_DEFAULTS.pitch,
      bearing: MAP_DEFAULTS.bearing,
      antialias: true,
    });

    map.addControl(new mapboxgl.NavigationControl({ showCompass: true }), "top-right");

    map.on("load", () => {
      // Add empty source
      map.addSource(HOTSPOT_SOURCE, {
        type: "geojson",
        data: { type: "FeatureCollection", features: [] },
      });

      // Hotspot circles
      map.addLayer({
        id: HOTSPOT_LAYER,
        type: "circle",
        source: HOTSPOT_SOURCE,
        paint: {
          "circle-radius": ["get", "radius"],
          "circle-color": ["get", "color"],
          "circle-opacity": 0.7,
          "circle-stroke-width": 1,
          "circle-stroke-color": "rgba(255,255,255,0.15)",
        },
      });

      // Selected zone ring
      map.addLayer({
        id: HOTSPOT_SELECTED_LAYER,
        type: "circle",
        source: HOTSPOT_SOURCE,
        paint: {
          "circle-radius": ["get", "radius"],
          "circle-color": "transparent",
          "circle-stroke-width": 3,
          "circle-stroke-color": "#d4af37",
        },
        filter: ["==", ["get", "zone_id"], ""],
      });

      setMapLoaded(true);
    });

    // Click to select
    map.on("click", HOTSPOT_LAYER, (e) => {
      if (e.features && e.features.length > 0) {
        const zoneId = e.features[0].properties?.zone_id;
        if (zoneId) selectZone(zoneId);
      }
    });

    // Cursor on hover
    map.on("mouseenter", HOTSPOT_LAYER, () => {
      map.getCanvas().style.cursor = "pointer";
    });
    map.on("mouseleave", HOTSPOT_LAYER, () => {
      map.getCanvas().style.cursor = "";
    });

    mapRef.current = map;

    return () => {
      map.remove();
      mapRef.current = null;
      setMapLoaded(false);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tokenAvailable]);

  // ── Update hotspot data ──
  useEffect(() => {
    if (!mapRef.current || !mapLoaded) return;
    const source = mapRef.current.getSource(HOTSPOT_SOURCE) as mapboxgl.GeoJSONSource | undefined;
    if (source) {
      source.setData(hotspotsToGeoJSON(hotspots));
    }
  }, [hotspots, mapLoaded]);

  // ── Update selected zone filter ──
  useEffect(() => {
    if (!mapRef.current || !mapLoaded) return;
    mapRef.current.setFilter(HOTSPOT_SELECTED_LAYER, [
      "==",
      ["get", "zone_id"],
      selectedZoneId ?? "",
    ]);
  }, [selectedZoneId, mapLoaded]);

  // ── Fallback when no Mapbox token ──
  if (!tokenAvailable) {
    return <MapFallback hotspotCount={hotspots.length} />;
  }

  return (
    <div ref={mapContainer} className="w-full h-full min-h-[400px]" />
  );
}

// ── Styled fallback for missing token ──
function MapFallback({ hotspotCount }: { hotspotCount: number }) {
  return (
    <div className="flex items-center justify-center h-full min-h-[400px] bg-surface/30">
      <div className="flex flex-col items-center gap-4 text-center px-6">
        {/* Grid pattern background */}
        <div className="absolute inset-0 opacity-[0.03]"
          style={{
            backgroundImage:
              "linear-gradient(rgba(212,175,55,0.3) 1px, transparent 1px), linear-gradient(90deg, rgba(212,175,55,0.3) 1px, transparent 1px)",
            backgroundSize: "40px 40px",
          }}
        />

        <div className="relative">
          <div className="absolute -inset-4 rounded-full bg-gold/5 animate-pulse" />
          <div className="relative flex h-16 w-16 items-center justify-center rounded-full bg-gold/10 border border-gold/20">
            <Layers className="h-7 w-7 text-gold/60" />
          </div>
        </div>

        <div className="flex flex-col gap-1.5 relative">
          <h3 className="display-title text-xs text-foreground/70">
            MAP AWAITING TOKEN
          </h3>
          <p className="text-xs text-muted max-w-[260px]">
            Set <code className="text-gold/70 text-[11px]">NEXT_PUBLIC_MAPBOX_TOKEN</code> in
            your <code className="text-gold/70 text-[11px]">.env.local</code> to enable the
            interactive command map.
          </p>
        </div>

        {hotspotCount > 0 && (
          <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-gold/5 border border-gold/10 text-xs relative">
            <MapPin className="h-3 w-3 text-gold/60" />
            <span className="mono-value text-gold/80">{hotspotCount}</span>
            <span className="text-muted">hotspots loaded</span>
          </div>
        )}
      </div>
    </div>
  );
}
