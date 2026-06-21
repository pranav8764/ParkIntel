"use client";

// ============================================================================
// Command Shell — The root layout of the Aurelian Command Center.
// Owns the background glow, top status bar, and panel grid.
// ============================================================================

import { type ReactNode, useEffect } from "react";
import {
  Shield,
  Activity,
  Clock,
  Radio,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { useHealth } from "@/hooks/use-health";
import { getCurrentHour, useCommandStore } from "@/store/command-store";
import { formatHour } from "@/lib/format";
import { OfflineOverlay } from "./offline-overlay";
import { FilterControls } from "./filter-controls";

interface CommandShellProps {
  children: ReactNode;
}

export function CommandShell({ children }: CommandShellProps) {
  const { isOnline, isOffline, isValidating, mutate } = useHealth();
  const selectedHour = useCommandStore((s) => s.selectedHour);
  const setSelectedHour = useCommandStore((s) => s.setSelectedHour);
  const forecastMode = useCommandStore((s) => s.forecastMode);

  useEffect(() => {
    setSelectedHour(getCurrentHour());
  }, [setSelectedHour]);

  return (
    <div className="relative flex h-screen flex-col overflow-hidden aurelian-glow">
      {/* ─── System Offline Overlay ─── */}
      <OfflineOverlay visible={isOffline} onRetry={() => mutate()} />

      {/* ─── Top Status Bar ─── */}
      <header className="relative z-10 flex items-center justify-between px-5 py-3 border-b border-white/5">
        {/* Left: Branding */}
        <div className="flex items-center gap-3">
          <Shield className="h-5 w-5 text-gold" />
          <span className="display-title text-sm text-foreground tracking-widest">
            PARKINTEL
          </span>
          <span className="text-xs text-muted hidden sm:inline">
            AURELIAN COMMAND CENTER
          </span>
        </div>

        {/* Center: Filters */}
        <div className="hidden md:flex">
          <FilterControls />
        </div>

        {/* Right: Status Indicators */}
        <div className="flex items-center gap-3">
          {/* Forecast Mode */}
          {forecastMode && (
            <Badge variant="gold" className="animate-pulse">
              FORECAST
            </Badge>
          )}

          {/* Time */}
          <div className="flex items-center gap-1.5 text-xs text-muted">
            <Clock className="h-3.5 w-3.5" />
            <span className="mono-value">{formatHour(selectedHour)}</span>
          </div>

          {/* Connection Status */}
          <div className="flex items-center gap-1.5">
            {isValidating && (
              <Activity className="h-3.5 w-3.5 text-gold/60 animate-pulse" />
            )}
            <div
              className={cn(
                "flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium",
                isOnline
                  ? "text-emerald bg-emerald/10"
                  : "text-red-400 bg-red-500/10"
              )}
            >
              <Radio className="h-3 w-3" />
              <span className="hidden sm:inline">
                {isOnline ? "ONLINE" : "OFFLINE"}
              </span>
            </div>
          </div>
        </div>
      </header>

      {/* ─── Main Content Area ─── */}
      <main className="relative flex-1 overflow-hidden">
        {children}
      </main>
    </div>
  );
}
