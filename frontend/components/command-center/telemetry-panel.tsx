"use client";

// ============================================================================
// Telemetry Panel — Real-time system status and data overview.
// ============================================================================

import { useHealth } from "@/hooks/use-health";
import { useHotspots } from "@/hooks/use-hotspots";
import { useRanking } from "@/hooks/use-ranking";
import { useCommandStore } from "@/store/command-store";
import { Panel } from "@/components/ui/panel";
import { MetricCard } from "@/components/ui/metric-card";
import { Badge } from "@/components/ui/badge";
import { formatHour, formatTimestamp } from "@/lib/format";
import {
  Activity,
  MapPin,
  BarChart3,
  AlertTriangle,
  Gauge,
  Cpu,
  HardDrive,
  Timer,
} from "lucide-react";
import { useState, useEffect } from "react";

export function TelemetryPanel() {
  const selectedHour = useCommandStore((s) => s.selectedHour);
  const selectedDate = useCommandStore((s) => s.selectedDate);
  const selectedPoliceStation = useCommandStore((s) => s.selectedPoliceStation);
  const selectedRiskLevel = useCommandStore((s) => s.selectedRiskLevel);
  const rankingLimit = useCommandStore((s) => s.rankingLimit);
  const forecastMode = useCommandStore((s) => s.forecastMode);

  const { isOnline, isValidating } = useHealth();

  const { hotspots, count: hotspotCount } = useHotspots({
    hour: selectedHour,
    date: selectedDate ?? undefined,
    police_station: selectedPoliceStation ?? undefined,
    risk_level: selectedRiskLevel ?? undefined,
  });

  const { total: rankingTotal } = useRanking({
    hour: selectedHour,
    date: selectedDate ?? undefined,
    limit: rankingLimit,
  });

  // Derived stats
  const criticalCount = hotspots.filter((h) => h.priority_score >= 72.0).length;
  const avgPriority =
    hotspots.length > 0
      ? hotspots.reduce((sum, h) => sum + h.priority_score, 0) / hotspots.length
      : 0;

  // Last refresh time
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());
  useEffect(() => {
    if (!isValidating) {
      const timer = setTimeout(() => {
        setLastRefresh(new Date());
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [isValidating]);

  // Active filters display
  const activeFilters: string[] = [];
  if (selectedPoliceStation) activeFilters.push(`Station: ${selectedPoliceStation}`);
  if (selectedRiskLevel) activeFilters.push(`Risk: ${selectedRiskLevel}`);

  return (
    <Panel
      title="TELEMETRY"
      icon={<Activity className="h-4 w-4" />}
      forecast={forecastMode}
      headerRight={
        <div className="flex items-center gap-2">
          {isValidating && (
            <Badge variant="gold" className="animate-pulse text-[10px]">
              SYNCING
            </Badge>
          )}
          <span className="mono-value text-[10px] text-muted">
            {formatTimestamp(lastRefresh)}
          </span>
        </div>
      }
    >
      <div className="flex flex-col gap-3">
        {/* Status Row */}
        <div className="flex items-center gap-2 text-xs">
          <div
            className={`flex items-center gap-1.5 px-2 py-1 rounded-md ${
              isOnline
                ? "text-emerald bg-emerald/10"
                : "text-red-400 bg-red-500/10"
            }`}
          >
            <span className={`h-1.5 w-1.5 rounded-full ${isOnline ? "bg-emerald animate-pulse" : "bg-red-400"}`} />
            {isOnline ? "BACKEND ONLINE" : "BACKEND OFFLINE"}
          </div>
          <span className="text-muted text-[10px] mono-value">
            {formatHour(selectedHour)}
          </span>
        </div>

        {/* Active Filters */}
        {activeFilters.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {activeFilters.map((f) => (
              <Badge key={f} variant="muted" className="text-[10px]">
                {f}
              </Badge>
            ))}
          </div>
        )}

        {/* Metric Grid */}
        <div className="grid grid-cols-2 gap-2">
          <MetricCard
            label="Hotspots"
            value={hotspotCount}
            icon={<MapPin className="h-3.5 w-3.5" />}
          />
          <MetricCard
            label="Critical"
            value={criticalCount}
            icon={<AlertTriangle className="h-3.5 w-3.5" />}
            highlight={criticalCount > 0}
          />
          <MetricCard
            label="Avg Priority"
            value={avgPriority.toFixed(1)}
            icon={<Gauge className="h-3.5 w-3.5" />}
          />
          <MetricCard
            label="Rankings"
            value={rankingTotal}
            icon={<BarChart3 className="h-3.5 w-3.5" />}
          />
        </div>

        {/* ONNX / System Metrics (placeholders) */}
        <div className="border-t border-white/5 pt-3">
          <p className="text-[10px] uppercase tracking-widest text-muted/50 mb-2">
            System Metrics
          </p>
          <div className="grid grid-cols-3 gap-2">
            <div className="flex flex-col items-center gap-1 py-2 rounded-md bg-white/[0.02] border border-white/5">
              <Timer className="h-3.5 w-3.5 text-muted/30" />
              <span className="mono-value text-xs text-muted/40">— ms</span>
              <span className="text-[9px] text-muted/30">ONNX</span>
            </div>
            <div className="flex flex-col items-center gap-1 py-2 rounded-md bg-white/[0.02] border border-white/5">
              <Cpu className="h-3.5 w-3.5 text-muted/30" />
              <span className="mono-value text-xs text-muted/40">— %</span>
              <span className="text-[9px] text-muted/30">CPU</span>
            </div>
            <div className="flex flex-col items-center gap-1 py-2 rounded-md bg-white/[0.02] border border-white/5">
              <HardDrive className="h-3.5 w-3.5 text-muted/30" />
              <span className="mono-value text-xs text-muted/40">— MB</span>
              <span className="text-[9px] text-muted/30">MEM</span>
            </div>
          </div>
        </div>
      </div>
    </Panel>
  );
}
