"use client";

import { type ReactNode } from "react";
import { cn } from "@/lib/utils";

interface MetricCardProps {
  label: string;
  value: string | number;
  icon?: ReactNode;
  suffix?: string;
  trend?: "up" | "down" | "neutral";
  highlight?: boolean;
  className?: string;
}

function MetricCard({
  label,
  value,
  icon,
  suffix,
  trend,
  highlight,
  className,
}: MetricCardProps) {
  return (
    <div
      className={cn(
        "flex flex-col gap-1.5 rounded-lg p-3",
        "bg-white/[0.03] border border-white/5",
        "transition-all duration-300",
        "hover:bg-white/[0.05] hover:border-white/8",
        highlight && "border-gold/20 bg-gold/[0.04]",
        className
      )}
    >
      <div className="flex items-center justify-between">
        <span className="text-[11px] font-medium uppercase tracking-widest text-muted">
          {label}
        </span>
        {icon && <span className="text-gold/60">{icon}</span>}
      </div>
      <div className="flex items-baseline gap-1">
        <span
          className={cn(
            "mono-value text-2xl font-bold",
            highlight ? "text-gold" : "text-foreground"
          )}
        >
          {value}
        </span>
        {suffix && (
          <span className="text-xs text-muted font-medium">{suffix}</span>
        )}
        {trend && trend !== "neutral" && (
          <span
            className={cn(
              "text-xs font-medium ml-1",
              trend === "up" ? "text-red-400" : "text-emerald"
            )}
          >
            {trend === "up" ? "↑" : "↓"}
          </span>
        )}
      </div>
    </div>
  );
}

export { MetricCard };
export type { MetricCardProps };
