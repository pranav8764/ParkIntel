"use client";

// ============================================================================
// Command Map — Leaflet integration with hotspot visualization.
// Completely free and open source, removing proprietary Mapbox token requirements.
// ============================================================================

import { useEffect, useRef } from "react";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { useCommandStore } from "@/store/command-store";
import { useHotspots } from "@/hooks/use-hotspots";
import { MAP_DEFAULTS } from "@/lib/map";
import { riskHex } from "@/lib/priority";

export function CommandMap() {
  const mapContainer = useRef<HTMLDivElement>(null);
  const mapRef = useRef<L.Map | null>(null);
  const markersGroupRef = useRef<L.LayerGroup | null>(null);

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

  // ── Initialize Leaflet Map ──
  useEffect(() => {
    if (!mapContainer.current || mapRef.current) return;

    // Create Leaflet map instance centered on defaults
    const map = L.map(mapContainer.current, {
      zoomControl: false,
    }).setView(MAP_DEFAULTS.center, MAP_DEFAULTS.zoom);

    // Add dark CartoDB tiles to match the Aurelian theme
    L.tileLayer(MAP_DEFAULTS.tileUrl, {
      attribution: MAP_DEFAULTS.attribution,
      maxZoom: 19,
    }).addTo(map);

    // Position zoom control in the top-right
    L.control.zoom({ position: "topright" }).addTo(map);

    // Create and add LayerGroup to hold dynamic hotspot markers
    const markersGroup = L.layerGroup().addTo(map);
    markersGroupRef.current = markersGroup;

    mapRef.current = map;

    // Setup ResizeObserver to dynamically update Leaflet's viewport on size changes
    const resizeObserver = new ResizeObserver(() => {
      map.invalidateSize();
    });
    
    if (mapContainer.current) {
      resizeObserver.observe(mapContainer.current);
    }

    return () => {
      resizeObserver.disconnect();
      map.remove();
      mapRef.current = null;
      markersGroupRef.current = null;
    };
  }, []);

  // ── Update Hotspot Circle Markers ──
  useEffect(() => {
    const map = mapRef.current;
    const markersGroup = markersGroupRef.current;
    if (!map || !markersGroup) return;

    // Clear old markers
    markersGroup.clearLayers();

    // Add new circle markers for current hotspots
    hotspots.forEach((h) => {
      const isSelected = h.zone_id === selectedZoneId;
      const color = riskHex(h.predicted_hotspot_risk);
      const radius = Math.max(6, (h.priority_score / 100) * 16);

      const marker = L.circleMarker([h.lat, h.lng], {
        radius: radius,
        fillColor: color,
        fillOpacity: 0.65,
        color: isSelected ? "#d4af37" : "rgba(255, 255, 255, 0.2)",
        weight: isSelected ? 3 : 1,
      });

      // Immersion tooltips styled natively
      marker.bindTooltip(
        `<div class="mono-value text-[10px] p-0.5 leading-normal">
          <strong class="text-gold font-bold">${h.police_station}</strong><br/>
          <span class="text-muted">Priority:</span> <span class="font-semibold text-foreground">${h.priority_score}</span><br/>
          <span class="text-muted">Violations:</span> <span class="font-semibold text-foreground">${h.expected_violations}</span>
         </div>`,
        {
          direction: "top",
          className: "leaflet-dark-tooltip",
          opacity: 0.9,
          sticky: true,
        }
      );

      // Select hotspot on click
      marker.on("click", () => {
        selectZone(h.zone_id);
      });

      // Hover pointer styling on marker canvas
      marker.on("mouseover", (e) => {
        const target = e.target as L.CircleMarker;
        const elem = target.getElement();
        if (elem) (elem as HTMLElement).style.cursor = "pointer";
      });

      markersGroup.addLayer(marker);
    });
  }, [hotspots, selectedZoneId, selectZone]);

  return (
    <div className="w-full h-full min-h-[400px] relative flex-1 flex flex-col">
      <div ref={mapContainer} className="w-full h-full min-h-[400px] flex-1 bg-[#101415]" />
      
      {/* Custom tooltips style injector to match Aurelian Dark theme */}
      <style jsx global>{`
        .leaflet-dark-tooltip {
          background: rgba(18, 22, 23, 0.9) !important;
          border: 1px solid rgba(212, 175, 55, 0.3) !important;
          color: #f3f4f6 !important;
          border-radius: 4px !important;
          box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.5) !important;
          padding: 6px 8px !important;
        }
        .leaflet-tooltip-top:before {
          border-top-color: rgba(212, 175, 55, 0.3) !important;
        }
      `}</style>
    </div>
  );
}
