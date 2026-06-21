"use client";

import { type HTMLAttributes, type ReactNode } from "react";
import { cn } from "@/lib/utils";

interface PanelProps extends HTMLAttributes<HTMLDivElement> {
  title?: string;
  icon?: ReactNode;
  headerRight?: ReactNode;
  noPadding?: boolean;
  forecast?: boolean;
}

function Panel({
  className,
  title,
  icon,
  headerRight,
  noPadding,
  forecast,
  children,
  ...props
}: PanelProps) {
  const isFullHeight = className?.includes("h-full") || className?.includes("h-screen");

  return (
    <div
      className={cn(
        "glass-panel rounded-xl overflow-hidden",
        isFullHeight && "flex flex-col",
        "transition-all duration-300",
        forecast && "forecast-border border-2",
        className
      )}
      {...props}
    >
      {(title || headerRight) && (
        <div className="flex items-center justify-between px-4 py-3 border-b border-white/5 flex-shrink-0">
          <div className="flex items-center gap-2">
            {icon && (
              <span className="text-gold/70 flex-shrink-0">{icon}</span>
            )}
            {title && (
              <h2 className="display-title text-foreground/80">{title}</h2>
            )}
          </div>
          {headerRight && <div className="flex items-center gap-2">{headerRight}</div>}
        </div>
      )}
      <div
        className={cn(
          isFullHeight ? "flex-1 min-h-0 flex flex-col" : "",
          !noPadding && "p-4"
        )}
      >
        {children}
      </div>
    </div>
  );
}

export { Panel };
export type { PanelProps };
