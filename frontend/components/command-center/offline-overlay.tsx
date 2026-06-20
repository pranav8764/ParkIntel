"use client";

import { useState } from "react";
import { WifiOff, RefreshCw, Eye } from "lucide-react";
import { cn } from "@/lib/utils";

interface OfflineOverlayProps {
  visible: boolean;
  onRetry?: () => void;
}

function OfflineOverlay({ visible, onRetry }: OfflineOverlayProps) {
  const [dismissed, setDismissed] = useState(false);

  // Reset dismissed state when visibility changes (backend comes back then goes offline again)
  if (!visible && dismissed) {
    setDismissed(false);
  }

  if (!visible || dismissed) return null;

  return (
    <div
      className={cn(
        "fixed inset-0 z-50",
        "flex items-center justify-center",
        "bg-background/90 backdrop-blur-md",
        "animate-in fade-in duration-300"
      )}
    >
      <div className="flex flex-col items-center gap-6 text-center max-w-sm px-6">
        {/* Pulsing icon */}
        <div className="relative">
          <div className="absolute inset-0 rounded-full bg-red-500/20 animate-ping" />
          <div className="relative flex h-20 w-20 items-center justify-center rounded-full bg-red-500/10 border border-red-500/20">
            <WifiOff className="h-9 w-9 text-red-400" />
          </div>
        </div>

        {/* Title */}
        <div className="flex flex-col gap-2">
          <h1 className="display-title text-lg tracking-wider text-foreground">
            SYSTEM OFFLINE
          </h1>
          <p className="text-sm text-muted leading-relaxed">
            Unable to reach the ParkIntel inference service. The command center
            will automatically reconnect when the backend becomes available.
          </p>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-3">
          {onRetry && (
            <button
              onClick={onRetry}
              className={cn(
                "inline-flex items-center gap-2 px-5 py-2.5 rounded-lg",
                "bg-white/5 border border-white/10 text-sm font-medium text-foreground",
                "transition-all duration-300",
                "hover:bg-white/10 hover:border-white/15",
                "active:scale-95"
              )}
            >
              <RefreshCw className="h-4 w-4" />
              Retry Connection
            </button>
          )}

          <button
            onClick={() => setDismissed(true)}
            className={cn(
              "inline-flex items-center gap-2 px-4 py-2.5 rounded-lg",
              "text-sm text-muted",
              "transition-all duration-300",
              "hover:text-foreground hover:bg-white/5",
              "active:scale-95"
            )}
          >
            <Eye className="h-4 w-4" />
            View Dashboard
          </button>
        </div>

        {/* Status bar */}
        <div className="flex items-center gap-2 text-xs text-muted/60">
          <span className="h-1.5 w-1.5 rounded-full bg-red-400 animate-pulse" />
          Polling every 30 seconds
        </div>
      </div>
    </div>
  );
}

export { OfflineOverlay };
export type { OfflineOverlayProps };
