"use client";

import { cn } from "@/lib/utils";

interface SkeletonProps {
  className?: string;
}

function Skeleton({ className }: SkeletonProps) {
  return (
    <div
      className={cn(
        "animate-pulse rounded-md bg-white/[0.06]",
        className
      )}
    />
  );
}

/**
 * Multiple skeleton lines for text-like loading states.
 */
function SkeletonLines({
  lines = 3,
  className,
}: {
  lines?: number;
  className?: string;
}) {
  return (
    <div className={cn("flex flex-col gap-2", className)}>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton
          key={i}
          className={cn(
            "h-4",
            i === lines - 1 ? "w-2/3" : "w-full"
          )}
        />
      ))}
    </div>
  );
}

/**
 * Skeleton shaped like a metric card.
 */
function SkeletonMetric({ className }: SkeletonProps) {
  return (
    <div
      className={cn(
        "rounded-lg p-3 bg-white/[0.03] border border-white/5",
        className
      )}
    >
      <Skeleton className="h-3 w-20 mb-2" />
      <Skeleton className="h-7 w-16" />
    </div>
  );
}

export { Skeleton, SkeletonLines, SkeletonMetric };
