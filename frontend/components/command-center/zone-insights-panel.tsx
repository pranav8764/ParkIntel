"use client";

// ============================================================================
// Zone Insights Panel — ML-driven zone drill-down with charts and stats.
// ============================================================================

import { useCommandStore } from "@/store/command-store";
import { useZoneInsights } from "@/hooks/use-zone-insights";
import { Panel } from "@/components/ui/panel";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/ui/empty-state";
import { SkeletonLines, SkeletonMetric } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { formatScore, formatPercent } from "@/lib/format";
import { priorityHex, isCritical } from "@/lib/priority";
import {
  Crosshair,
  Brain,
  AlertTriangle,
  TrendingUp,
  Car,
  FileWarning,
  Lightbulb,
  Target,
} from "lucide-react";

export function ZoneInsightsPanel() {
  const selectedZoneId = useCommandStore((s) => s.selectedZoneId);
  const selectedHour = useCommandStore((s) => s.selectedHour);
  const forecastMode = useCommandStore((s) => s.forecastMode);

  const { insights, isLoading, error } = useZoneInsights(
    selectedZoneId ? { zone_id: selectedZoneId, hour: selectedHour } : null
  );

  if (!selectedZoneId) {
    return (
      <Panel title="ZONE INSIGHTS" icon={<Crosshair className="h-4 w-4" />}>
        <EmptyState
          icon={<Target className="h-5 w-5" />}
          title="No zone selected"
          description="Click a hotspot on the map or a row in the ranking table to view insights."
        />
      </Panel>
    );
  }

  if (isLoading) {
    return (
      <Panel title="ZONE INSIGHTS" icon={<Crosshair className="h-4 w-4" />}>
        <div className="flex flex-col gap-3">
          <SkeletonMetric />
          <SkeletonMetric />
          <SkeletonLines lines={4} />
        </div>
      </Panel>
    );
  }

  if (error || !insights) {
    return (
      <Panel title="ZONE INSIGHTS" icon={<Crosshair className="h-4 w-4" />}>
        <EmptyState
          icon={<AlertTriangle className="h-5 w-5 text-red-400" />}
          title="Failed to load insights"
          description="Could not retrieve data for this zone."
        />
      </Panel>
    );
  }

  const critical = isCritical(insights.priority_score);

  return (
    <Panel
      title="ZONE INSIGHTS"
      icon={<Crosshair className="h-4 w-4" />}
      forecast={forecastMode}
      headerRight={
        <Badge variant={critical ? "gold" : "muted"} className="text-[10px] mono-value">
          {insights.zone_id}
        </Badge>
      }
    >
      <div className="flex flex-col gap-4">
        {/* ── Score Row ── */}
        <div className="grid grid-cols-2 gap-3">
          {/* Priority Score */}
          <div className="flex flex-col items-center gap-1.5 p-3 rounded-lg bg-white/[0.03] border border-white/5">
            <span className="text-[10px] uppercase tracking-widest text-muted">
              Priority
            </span>
            <span
              className={cn(
                "mono-value text-3xl font-bold",
                critical ? "gold-shimmer" : "text-foreground"
              )}
            >
              {formatScore(insights.priority_score)}
            </span>
            <Badge
              variant={
                insights.priority_level === "CRITICAL" || insights.priority_level === "HIGH"
                  ? "gold"
                  : insights.priority_level === "MEDIUM"
                  ? "blue"
                  : "emerald"
              }
              className="text-[10px]"
            >
              {insights.priority_level}
            </Badge>
          </div>

          {/* Impact Score */}
          <div className="flex flex-col items-center gap-1.5 p-3 rounded-lg bg-white/[0.03] border border-white/5">
            <span className="text-[10px] uppercase tracking-widest text-muted">
              Impact
            </span>
            <span className="mono-value text-3xl font-bold text-foreground">
              {formatScore(insights.parking_congestion_impact_score)}
            </span>
            <Badge variant="muted" className="text-[10px]">
              {insights.predicted_hotspot_risk} RISK
            </Badge>
          </div>
        </div>

        {/* ── Class Probabilities ── */}
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-1.5">
            <Brain className="h-3.5 w-3.5 text-gold/60" />
            <span className="text-[10px] uppercase tracking-widest text-muted">
              Model Probabilities
            </span>
            <span className="text-[9px] text-muted/50 ml-auto">
              Confidence: {insights.model_confidence}
            </span>
          </div>
          {(["HIGH", "MEDIUM", "LOW"] as const).map((level) => {
            const value = insights.class_probabilities[level];
            const pct = (value * 100);
            return (
              <div key={level} className="flex items-center gap-2">
                <span className="text-[10px] text-muted w-14 text-right uppercase">
                  {level}
                </span>
                <div className="flex-1 h-2.5 rounded-full bg-white/5 overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all duration-500"
                    style={{
                      width: `${pct}%`,
                      backgroundColor: priorityHex(level),
                    }}
                  />
                </div>
                <span className="mono-value text-[11px] text-foreground/70 w-12 text-right">
                  {formatPercent(pct, 1)}
                </span>
              </div>
            );
          })}
        </div>

        {/* ── Recommended Action ── */}
        <div className="flex items-start gap-2 p-3 rounded-lg bg-gold/[0.04] border border-gold/10">
          <TrendingUp className="h-4 w-4 text-gold/60 mt-0.5 flex-shrink-0" />
          <div className="flex flex-col gap-0.5">
            <span className="text-[10px] uppercase tracking-widest text-muted">
              Recommended Action
            </span>
            <span className="text-sm text-foreground/90">
              {insights.recommended_action}
            </span>
          </div>
        </div>

        {/* ── Zone Stats ── */}
        <div className="grid grid-cols-2 gap-2">
          <div className="flex flex-col gap-1 p-2.5 rounded-lg bg-white/[0.02] border border-white/5">
            <span className="text-[9px] uppercase tracking-widest text-muted">
              Historical Violations
            </span>
            <span className="mono-value text-sm font-semibold text-foreground">
              {insights.zone_stats.total_historical_violations.toLocaleString()}
            </span>
          </div>
          <div className="flex flex-col gap-1 p-2.5 rounded-lg bg-white/[0.02] border border-white/5">
            <span className="text-[9px] uppercase tracking-widest text-muted">
              Repeat Hotspot Score
            </span>
            <span className="mono-value text-sm font-semibold text-foreground">
              {formatScore(insights.zone_stats.repeat_hotspot_score)}
            </span>
          </div>
        </div>

        {/* ── Top Violation Types ── */}
        {insights.zone_stats.top_violation_types.length > 0 && (
          <div className="flex flex-col gap-1.5">
            <div className="flex items-center gap-1.5">
              <FileWarning className="h-3.5 w-3.5 text-muted/50" />
              <span className="text-[10px] uppercase tracking-widest text-muted">
                Top Violation Types
              </span>
            </div>
            <div className="flex flex-wrap gap-1">
              {insights.zone_stats.top_violation_types.map((type) => (
                <Badge key={type} variant="muted" className="text-[10px]">
                  {type}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* ── Top Vehicle Types ── */}
        {insights.zone_stats.top_vehicle_types.length > 0 && (
          <div className="flex flex-col gap-1.5">
            <div className="flex items-center gap-1.5">
              <Car className="h-3.5 w-3.5 text-muted/50" />
              <span className="text-[10px] uppercase tracking-widest text-muted">
                Top Vehicle Types
              </span>
            </div>
            <div className="flex flex-wrap gap-1">
              {insights.zone_stats.top_vehicle_types.map((type) => (
                <Badge key={type} variant="muted" className="text-[10px]">
                  {type}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* ── Logic Transparency / Reasons ── */}
        {insights.reasons.length > 0 && (
          <div className="flex flex-col gap-2 p-3 rounded-lg bg-white/[0.02] border border-white/5">
            <div className="flex items-center gap-1.5">
              <Lightbulb className="h-3.5 w-3.5 text-gold/50" />
              <span className="text-[10px] uppercase tracking-widest text-gold/60 font-semibold">
                Logic Transparency
              </span>
            </div>
            <ul className="flex flex-col gap-1">
              {insights.reasons.map((reason, i) => (
                <li
                  key={i}
                  className="text-xs text-foreground/70 pl-4 relative before:absolute before:left-1 before:top-[7px] before:h-1 before:w-1 before:rounded-full before:bg-gold/30"
                >
                  {reason}
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* ── Note ── */}
        {insights.note && (
          <p className="text-[11px] text-muted/60 italic">
            {insights.note}
          </p>
        )}
      </div>
    </Panel>
  );
}
