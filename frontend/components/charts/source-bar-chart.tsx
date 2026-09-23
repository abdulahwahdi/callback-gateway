"use client";

import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

import type { SourceCount } from "@/lib/types";
import { Skeleton } from "@/components/ui/skeleton";

const PALETTE = [
  "hsl(var(--chart-1))",
  "hsl(var(--chart-2))",
  "hsl(var(--chart-3))",
  "hsl(var(--chart-4))",
  "hsl(var(--chart-5))",
];

export function SourceBarChart({
  data,
  loading,
}: {
  data: SourceCount[];
  loading?: boolean;
}) {
  if (loading) return <Skeleton className="h-[280px] w-full" />;

  if (!data.length) {
    return (
      <div className="flex h-[280px] w-full items-center justify-center text-sm text-muted-foreground">
        No sources seen yet.
      </div>
    );
  }

  const top = [...data].sort((a, b) => b.total - a.total).slice(0, 8);

  return (
    <ResponsiveContainer width="100%" height={280}>
      <BarChart data={top} layout="vertical" margin={{ top: 4, right: 16, left: 0, bottom: 4 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" horizontal={false} />
        <XAxis
          type="number"
          allowDecimals={false}
          tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
          axisLine={false}
          tickLine={false}
        />
        <YAxis
          type="category"
          dataKey="source"
          width={90}
          tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
          axisLine={false}
          tickLine={false}
        />
        <Tooltip
          cursor={{ fill: "hsl(var(--muted))" }}
          contentStyle={{
            background: "hsl(var(--popover))",
            border: "1px solid hsl(var(--border))",
            borderRadius: "0.5rem",
            fontSize: "0.75rem",
            color: "hsl(var(--popover-foreground))",
          }}
        />
        <Bar dataKey="total" name="Events" radius={[0, 6, 6, 0]} barSize={16}>
          {top.map((entry, i) => (
            <Cell key={entry.source} fill={PALETTE[i % PALETTE.length]} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
}
