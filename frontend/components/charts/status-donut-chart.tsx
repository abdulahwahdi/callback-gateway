"use client";

import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";

import type { StatusCount } from "@/lib/types";
import { Skeleton } from "@/components/ui/skeleton";
import { numberCompact } from "@/lib/utils";

const COLORS: Record<string, string> = {
  published: "hsl(var(--success))",
  pending: "hsl(var(--warning))",
  failed: "hsl(var(--destructive))",
};

export function StatusDonutChart({
  data,
  loading,
}: {
  data: StatusCount[];
  loading?: boolean;
}) {
  if (loading) return <Skeleton className="h-[220px] w-full" />;

  const total = data.reduce((sum, d) => sum + d.total, 0);

  if (!total) {
    return (
      <div className="flex h-[220px] w-full items-center justify-center text-sm text-muted-foreground">
        No data yet.
      </div>
    );
  }

  return (
    <div className="relative">
      <ResponsiveContainer width="100%" height={220}>
        <PieChart>
          <Pie
            data={data}
            dataKey="total"
            nameKey="publish_status"
            innerRadius={62}
            outerRadius={92}
            paddingAngle={3}
            strokeWidth={0}
          >
            {data.map((entry) => (
              <Cell
                key={entry.publish_status}
                fill={COLORS[entry.publish_status] ?? "hsl(var(--muted-foreground))"}
              />
            ))}
          </Pie>
          <Tooltip
            formatter={(value: number, name: string) => [numberCompact(value), name]}
            contentStyle={{
              background: "hsl(var(--popover))",
              border: "1px solid hsl(var(--border))",
              borderRadius: "0.5rem",
              fontSize: "0.75rem",
              textTransform: "capitalize",
              color: "hsl(var(--popover-foreground))",
            }}
          />
        </PieChart>
      </ResponsiveContainer>
      <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
        <span className="text-2xl font-semibold tabular-nums">{numberCompact(total)}</span>
        <span className="text-[11px] uppercase tracking-wide text-muted-foreground">
          total events
        </span>
      </div>
    </div>
  );
}
