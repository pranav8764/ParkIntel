"use client";

// ============================================================================
// Filter Controls — Hour, date, station, and risk level selectors.
// ============================================================================

import { useCommandStore } from "@/store/command-store";
import { formatHour } from "@/lib/format";
import { cn } from "@/lib/utils";
import { Clock, Building2, AlertTriangle, RotateCcw } from "lucide-react";

export function FilterControls() {
  const selectedHour = useCommandStore((s) => s.selectedHour);
  const setSelectedHour = useCommandStore((s) => s.setSelectedHour);
  const selectedPoliceStation = useCommandStore((s) => s.selectedPoliceStation);
  const setSelectedPoliceStation = useCommandStore(
    (s) => s.setSelectedPoliceStation
  );
  const selectedRiskLevel = useCommandStore((s) => s.selectedRiskLevel);
  const setSelectedRiskLevel = useCommandStore((s) => s.setSelectedRiskLevel);
  const resetFilters = useCommandStore((s) => s.resetFilters);

  const hours = Array.from({ length: 24 }, (_, i) => i);

  const riskLevels = [
    { value: "", label: "All Risks" },
    { value: "HIGH", label: "High" },
    { value: "MEDIUM", label: "Medium" },
    { value: "LOW", label: "Low" },
  ];

  return (
    <div className="flex items-center gap-2 flex-wrap">
      {/* Hour Selector */}
      <div className="flex items-center gap-1.5">
        <Clock className="h-3.5 w-3.5 text-muted" />
        <select
          value={selectedHour}
          onChange={(e) => setSelectedHour(Number(e.target.value))}
          className={cn(
            "h-8 rounded-md px-2 text-xs font-mono",
            "bg-white/5 text-foreground border border-white/10",
            "transition-all duration-200",
            "hover:border-white/15 focus:border-gold/40 focus:outline-none focus:ring-1 focus:ring-gold/20",
            "cursor-pointer appearance-none"
          )}
        >
          {hours.map((h) => (
            <option key={h} value={h} className="bg-surface text-foreground">
              {formatHour(h)}
            </option>
          ))}
        </select>
      </div>

      {/* Risk Level */}
      <div className="flex items-center gap-1.5">
        <AlertTriangle className="h-3.5 w-3.5 text-muted" />
        <select
          value={selectedRiskLevel ?? ""}
          onChange={(e) =>
            setSelectedRiskLevel(e.target.value || null)
          }
          className={cn(
            "h-8 rounded-md px-2 text-xs font-mono",
            "bg-white/5 text-foreground border border-white/10",
            "transition-all duration-200",
            "hover:border-white/15 focus:border-gold/40 focus:outline-none focus:ring-1 focus:ring-gold/20",
            "cursor-pointer appearance-none"
          )}
        >
          {riskLevels.map((r) => (
            <option
              key={r.value}
              value={r.value}
              className="bg-surface text-foreground"
            >
              {r.label}
            </option>
          ))}
        </select>
      </div>

      {/* Police Station (free text) */}
      <div className="flex items-center gap-1.5">
        <Building2 className="h-3.5 w-3.5 text-muted" />
        <input
          type="text"
          placeholder="Station..."
          value={selectedPoliceStation ?? ""}
          onChange={(e) =>
            setSelectedPoliceStation(e.target.value || null)
          }
          className={cn(
            "h-8 w-28 rounded-md px-2 text-xs font-mono",
            "bg-white/5 text-foreground placeholder:text-muted/50 border border-white/10",
            "transition-all duration-200",
            "hover:border-white/15 focus:border-gold/40 focus:outline-none focus:ring-1 focus:ring-gold/20"
          )}
        />
      </div>

      {/* Reset */}
      <button
        onClick={resetFilters}
        className={cn(
          "h-8 px-2 rounded-md text-xs text-muted",
          "hover:text-foreground hover:bg-white/5",
          "transition-all duration-200 active:scale-95"
        )}
        title="Reset filters"
      >
        <RotateCcw className="h-3.5 w-3.5" />
      </button>
    </div>
  );
}
