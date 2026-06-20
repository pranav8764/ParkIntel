"use client";

// ============================================================================
// Ranking Panel — Enforcement priority ranking with TanStack Table.
// ============================================================================

import { useMemo } from "react";
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
} from "@tanstack/react-table";
import { useRanking } from "@/hooks/use-ranking";
import { useCommandStore } from "@/store/command-store";
import { Panel } from "@/components/ui/panel";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/ui/empty-state";
import { SkeletonLines } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { formatScore } from "@/lib/format";
import { isCritical } from "@/lib/priority";
import type { Ranking } from "@/types/api";
import {
  BarChart3,
  ChevronRight,
  AlertTriangle,
  Shield,
} from "lucide-react";

const columnHelper = createColumnHelper<Ranking>();

export function RankingPanel() {
  const selectedHour = useCommandStore((s) => s.selectedHour);
  const selectedDate = useCommandStore((s) => s.selectedDate);
  const rankingLimit = useCommandStore((s) => s.rankingLimit);
  const selectedZoneId = useCommandStore((s) => s.selectedZoneId);
  const selectZone = useCommandStore((s) => s.selectZone);
  const forecastMode = useCommandStore((s) => s.forecastMode);

  const { rankings, total, isLoading, error } = useRanking({
    hour: selectedHour,
    date: selectedDate ?? undefined,
    limit: rankingLimit,
  });

  const columns = useMemo(
    () => [
      columnHelper.accessor("rank", {
        header: "#",
        cell: (info) => (
          <span className="mono-value text-xs text-muted">{info.getValue()}</span>
        ),
        size: 40,
      }),
      columnHelper.accessor("zone_id", {
        header: "Zone",
        cell: (info) => (
          <span className="text-sm font-medium text-foreground truncate max-w-[100px] block">
            {info.getValue()}
          </span>
        ),
        size: 110,
      }),
      columnHelper.accessor("junction_name", {
        header: "Junction",
        cell: (info) => (
          <span className="text-xs text-muted truncate max-w-[120px] block">
            {info.getValue() || "—"}
          </span>
        ),
        size: 130,
      }),
      columnHelper.accessor("police_station", {
        header: "Station",
        cell: (info) => (
          <span className="text-xs text-muted truncate max-w-[90px] block">
            {info.getValue()}
          </span>
        ),
        size: 100,
      }),
      columnHelper.accessor("expected_violations", {
        header: "Violations",
        cell: (info) => (
          <span className="mono-value text-xs text-foreground/80">
            {info.getValue()}
          </span>
        ),
        size: 80,
      }),
      columnHelper.accessor("impact_score", {
        header: "Impact",
        cell: (info) => (
          <span className="mono-value text-xs text-foreground/80">
            {formatScore(info.getValue() as number)}
          </span>
        ),
        size: 70,
      }),
      columnHelper.accessor("priority_score", {
        header: "Priority",
        cell: (info) => {
          const val = info.getValue() as number;
          return (
            <span
              className={cn(
                "mono-value text-sm font-bold",
                isCritical(val) ? "text-gold" : "text-foreground"
              )}
            >
              {formatScore(val)}
            </span>
          );
        },
        size: 75,
      }),
      columnHelper.accessor("priority_level", {
        header: "Level",
        cell: (info) => {
          const level = info.getValue() as string;
          const variant =
            level === "CRITICAL" || level === "HIGH"
              ? "gold"
              : level === "MEDIUM"
              ? "blue"
              : "emerald";
          return (
            <Badge variant={variant} className="text-[10px]">
              {level}
            </Badge>
          );
        },
        size: 80,
      }),
      columnHelper.accessor("model_confidence", {
        header: "Confidence",
        cell: (info) => (
          <span className="text-[10px] text-muted uppercase">
            {info.getValue()}
          </span>
        ),
        size: 80,
      }),
      columnHelper.accessor("recommended_action", {
        header: "Action",
        cell: (info) => (
          <span className="text-xs text-muted truncate max-w-[130px] block">
            {info.getValue() || "—"}
          </span>
        ),
        size: 140,
      }),
    ],
    []
  );

  const table = useReactTable({
    data: rankings,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <Panel
      title="ENFORCEMENT RANKING"
      icon={<BarChart3 className="h-4 w-4" />}
      forecast={forecastMode}
      noPadding
      headerRight={
        <div className="flex items-center gap-2">
          {total > 0 && (
            <Badge variant="gold" className="text-[10px]">
              {total} ZONES
            </Badge>
          )}
        </div>
      }
    >
      {isLoading ? (
        <div className="p-4">
          <SkeletonLines lines={8} />
        </div>
      ) : error ? (
        <div className="p-4">
          <EmptyState
            icon={<AlertTriangle className="h-5 w-5 text-red-400" />}
            title="Failed to load rankings"
            description="Check backend connection and try again."
          />
        </div>
      ) : rankings.length === 0 ? (
        <div className="p-4">
          <EmptyState
            icon={<Shield className="h-5 w-5" />}
            title="No ranking data"
            description="No enforcement data available for the selected time."
          />
        </div>
      ) : (
        <div className="overflow-auto max-h-[600px]">
          <table className="w-full text-left">
            <thead className="sticky top-0 z-10 bg-surface/90 backdrop-blur-sm">
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id} className="border-b border-white/5">
                  {headerGroup.headers.map((header) => (
                    <th
                      key={header.id}
                      className="px-3 py-2 text-[10px] font-medium uppercase tracking-widest text-muted"
                      style={{ width: header.getSize() }}
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </th>
                  ))}
                  <th className="w-8" />
                </tr>
              ))}
            </thead>
            <tbody>
              {table.getRowModel().rows.map((row) => {
                const rowData = row.original;
                const critical = isCritical(rowData.priority_score);
                const selected = rowData.zone_id === selectedZoneId;

                return (
                  <tr
                    key={row.id}
                    onClick={() => selectZone(rowData.zone_id)}
                    className={cn(
                      "border-b border-white/[0.03] cursor-pointer",
                      "transition-all duration-200",
                      "hover:bg-white/[0.04]",
                      critical && "bg-gold/[0.03] hover:bg-gold/[0.06]",
                      selected &&
                        "bg-gold/[0.08] border-l-2 border-l-gold"
                    )}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id} className="px-3 py-2.5">
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </td>
                    ))}
                    <td className="px-2 py-2.5">
                      <ChevronRight
                        className={cn(
                          "h-3.5 w-3.5 transition-colors",
                          selected ? "text-gold" : "text-muted/20"
                        )}
                      />
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </Panel>
  );
}
