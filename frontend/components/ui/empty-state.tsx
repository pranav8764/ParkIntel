"use client";

import { type ReactNode } from "react";
import { cn } from "@/lib/utils";
import { Inbox } from "lucide-react";

interface EmptyStateProps {
  icon?: ReactNode;
  title?: string;
  description?: string;
  action?: ReactNode;
  className?: string;
}

function EmptyState({
  icon,
  title = "No data available",
  description,
  action,
  className,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-3 py-10 text-center",
        className
      )}
    >
      <div className="flex h-12 w-12 items-center justify-center rounded-full bg-white/5 text-muted">
        {icon ?? <Inbox className="h-6 w-6" />}
      </div>
      <div className="flex flex-col gap-1">
        <p className="text-sm font-medium text-foreground/70">{title}</p>
        {description && (
          <p className="text-xs text-muted max-w-[240px]">{description}</p>
        )}
      </div>
      {action && <div className="mt-1">{action}</div>}
    </div>
  );
}

export { EmptyState };
export type { EmptyStateProps };
