"use client";

// ============================================================================
// Simulation Panel — Policy simulation engine for selected zones.
// ============================================================================

import { useCommandStore } from "@/store/command-store";
import { useSimulation } from "@/hooks/use-simulation";
import { Panel } from "@/components/ui/panel";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Slider } from "@/components/ui/slider";
import { EmptyState } from "@/components/ui/empty-state";
import { cn } from "@/lib/utils";
import { formatScore, formatPercent } from "@/lib/format";
import { isCritical } from "@/lib/priority";
import {
  Gauge,
  Play,
  RotateCcw,
  ArrowRight,
  TrendingDown,
  FlaskConical,
  Target,
  AlertTriangle,
} from "lucide-react";

export function SimulationPanel() {
  const selectedZoneId = useCommandStore((s) => s.selectedZoneId);
  const reductionPercent = useCommandStore((s) => s.simulationReductionPercent);
  const setReductionPercent = useCommandStore(
    (s) => s.setSimulationReductionPercent
  );
  const activeSimulation = useCommandStore((s) => s.activeSimulation);
  const forecastMode = useCommandStore((s) => s.forecastMode);

  const { simulate, reset, isSubmitting, error } = useSimulation();

  if (!selectedZoneId) {
    return (
      <Panel
        title="POLICY SIMULATION"
        icon={<Gauge className="h-4 w-4" />}
      >
        <EmptyState
          icon={<Target className="h-5 w-5" />}
          title="No zone selected"
          description="Select a zone to run policy simulations."
        />
      </Panel>
    );
  }

  return (
    <Panel
      title="POLICY SIMULATION"
      icon={<Gauge className="h-4 w-4" />}
      forecast={forecastMode}
      headerRight={
        <div className="flex items-center gap-2">
          {forecastMode && (
            <Badge variant="gold" className="text-[10px] animate-pulse">
              FORECAST MODE
            </Badge>
          )}
          <Badge variant="muted" className="text-[10px] mono-value">
            {selectedZoneId}
          </Badge>
        </div>
      }
    >
      <div className="flex flex-col gap-4">
        {/* ── Reduction Slider ── */}
        <Slider
          id="violation-reduction"
          label="Violation Reduction"
          value={reductionPercent}
          min={0}
          max={100}
          step={5}
          displayValue={`${reductionPercent}%`}
          onChange={setReductionPercent}
          disabled={isSubmitting}
        />

        {/* ── Run / Reset Buttons ── */}
        <div className="flex items-center gap-2">
          <Button
            variant="primary"
            size="sm"
            onClick={simulate}
            disabled={isSubmitting}
            className="flex-1"
          >
            {isSubmitting ? (
              <>
                <FlaskConical className="h-3.5 w-3.5 animate-spin" />
                Simulating…
              </>
            ) : (
              <>
                <Play className="h-3.5 w-3.5" />
                Run Simulation
              </>
            )}
          </Button>

          {activeSimulation && (
            <Button variant="ghost" size="sm" onClick={reset}>
              <RotateCcw className="h-3.5 w-3.5" />
              Reset
            </Button>
          )}
        </div>

        {/* ── Error ── */}
        {error && (
          <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-red-500/10 border border-red-500/15 text-xs text-red-400">
            <AlertTriangle className="h-3.5 w-3.5 flex-shrink-0" />
            {error.message}
          </div>
        )}

        {/* ── Simulation Results ── */}
        {activeSimulation && (
          <div className="flex flex-col gap-3 pt-2 border-t border-white/5">
            {/* Current → Simulated */}
            <div className="flex items-center justify-between gap-2">
              {/* Current */}
              <div className="flex flex-col items-center gap-1 flex-1 p-3 rounded-lg bg-white/[0.03] border border-white/5">
                <span className="text-[9px] uppercase tracking-widest text-muted">
                  Current
                </span>
                <span
                  className={cn(
                    "mono-value text-2xl font-bold",
                    isCritical(activeSimulation.current_priority_score)
                      ? "text-gold"
                      : "text-foreground"
                  )}
                >
                  {formatScore(activeSimulation.current_priority_score)}
                </span>
                <Badge
                  variant={
                    activeSimulation.current_priority_level === "CRITICAL" ||
                    activeSimulation.current_priority_level === "HIGH"
                      ? "gold"
                      : "muted"
                  }
                  className="text-[9px]"
                >
                  {activeSimulation.current_priority_level}
                </Badge>
              </div>

              <ArrowRight className="h-5 w-5 text-gold/40 flex-shrink-0" />

              {/* Simulated */}
              <div className="flex flex-col items-center gap-1 flex-1 p-3 rounded-lg bg-emerald/[0.04] border border-emerald/15">
                <span className="text-[9px] uppercase tracking-widest text-emerald/70">
                  Simulated
                </span>
                <span
                  className={cn(
                    "mono-value text-2xl font-bold",
                    isCritical(activeSimulation.simulated_priority_score)
                      ? "text-gold"
                      : "text-emerald"
                  )}
                >
                  {formatScore(activeSimulation.simulated_priority_score)}
                </span>
                <Badge
                  variant={
                    activeSimulation.simulated_priority_level === "CRITICAL" ||
                    activeSimulation.simulated_priority_level === "HIGH"
                      ? "gold"
                      : "emerald"
                  }
                  className="text-[9px]"
                >
                  {activeSimulation.simulated_priority_level}
                </Badge>
              </div>
            </div>

            {/* Change & Impact Reduction */}
            <div className="grid grid-cols-2 gap-2">
              <div className="flex flex-col gap-0.5 p-2.5 rounded-lg bg-white/[0.02] border border-white/5">
                <span className="text-[9px] uppercase tracking-widest text-muted">
                  Priority Change
                </span>
                <span className="mono-value text-sm font-semibold text-emerald flex items-center gap-1">
                  <TrendingDown className="h-3 w-3" />
                  {activeSimulation.priority_change}
                </span>
              </div>
              <div className="flex flex-col gap-0.5 p-2.5 rounded-lg bg-white/[0.02] border border-white/5">
                <span className="text-[9px] uppercase tracking-widest text-muted">
                  Impact Reduction
                </span>
                <span className="mono-value text-sm font-semibold text-emerald">
                  {formatPercent(activeSimulation.estimated_impact_reduction)}
                </span>
              </div>
            </div>

            {/* Note */}
            {activeSimulation.note && (
              <p className="text-[11px] text-muted/60 italic">
                {activeSimulation.note}
              </p>
            )}
          </div>
        )}
      </div>
    </Panel>
  );
}
