"use client";

import { type HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

type BadgeVariant = "gold" | "blue" | "emerald" | "muted" | "danger";

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant;
}

const variantStyles: Record<BadgeVariant, string> = {
  gold: "bg-gold/15 text-gold border-gold/25",
  blue: "bg-synthetic-blue/15 text-synthetic-blue border-synthetic-blue/25",
  emerald: "bg-emerald/15 text-emerald border-emerald/25",
  muted: "bg-white/5 text-muted border-white/10",
  danger: "bg-red-500/10 text-red-400 border-red-500/20",
};

function Badge({ className, variant = "muted", ...props }: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md border px-2 py-0.5",
        "text-xs font-medium tracking-wide",
        "transition-colors duration-300",
        variantStyles[variant],
        className
      )}
      {...props}
    />
  );
}

export { Badge };
export type { BadgeProps, BadgeVariant };
