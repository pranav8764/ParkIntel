"use client";

// ============================================================================
// Command Center Page — Main page of the ParkIntel Aurelian Command Center.
// Assembles the shell, map, telemetry, ranking, insights, and simulation.
// ============================================================================

import { CommandShell } from "@/components/command-center/command-shell";
import { CommandMap } from "@/components/command-center/command-map";
import { TelemetryPanel } from "@/components/command-center/telemetry-panel";
import { RankingPanel } from "@/components/command-center/ranking-panel";
import { ZoneInsightsPanel } from "@/components/command-center/zone-insights-panel";
import { SimulationPanel } from "@/components/command-center/simulation-panel";
import { Panel } from "@/components/ui/panel";
import { Badge } from "@/components/ui/badge";
import { useCommandStore } from "@/store/command-store";
import { formatHour } from "@/lib/format";
import { Crosshair } from "lucide-react";

export default function CommandCenterPage() {
  const selectedHour = useCommandStore((s) => s.selectedHour);

  return (
    <CommandShell>
      <div className="flex h-full overflow-hidden">
        {/* ═══ Left Column: Map + Ranking ═══ */}
        <div className="flex flex-col flex-1 min-w-0">
          {/* Map */}
          <div className="flex-1 min-h-0 relative">
            <Panel
              title="COMMAND MAP"
              icon={<Crosshair className="h-4 w-4" />}
              className="h-full rounded-none border-0"
              noPadding
              headerRight={
                <Badge variant="muted" className="mono-value text-[10px]">
                  {formatHour(selectedHour)}
                </Badge>
              }
            >
              <CommandMap />
            </Panel>
          </div>

          {/* Ranking (bottom of left column) */}
          <div className="h-[320px] min-h-[250px] border-t border-white/5 overflow-hidden">
            <RankingPanel />
          </div>
        </div>

        {/* ═══ Right Sidebar ═══ */}
        <aside className="w-[360px] flex-shrink-0 border-l border-white/5 flex flex-col overflow-y-auto">
          {/* Telemetry */}
          <div className="border-b border-white/5">
            <TelemetryPanel />
          </div>

          {/* Zone Insights */}
          <div className="border-b border-white/5">
            <ZoneInsightsPanel />
          </div>

          {/* Simulation */}
          <div>
            <SimulationPanel />
          </div>
        </aside>
      </div>
    </CommandShell>
  );
}
