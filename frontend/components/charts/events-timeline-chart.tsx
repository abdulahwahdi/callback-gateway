"use client";

import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { format, parseISO } from "date-fns";

import type { DailyCount } from "@/lib/types";
import { Skeleton } from "@/components/ui/skeleton";

function tickFormatter(value: string) {
  try {
    return format(parseISO(value), "MMM d");
  } catch {
    return value;
  }
}

export function EventsTimelineChart({
  data,
  loading,
}: {
  data: DailyCount[];
  loading?: boolean;
}) {
  if (loading) {
    return <Skeleton className="h-[280px] w-full" />;
  }

  if (!data.length) {
    return (
      <div className="flex h-[280px] w-full items-center justify-center text-sm text-muted-foreground">
        No events in this range yet.
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={280}>
      <AreaChart data={data} margin={{ top: 8, right: 12, left: -12, bottom: 0 }}>
        <defs>
          <linearGradient id="eventsFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="hsl(var(--chart-1))" stopOpacity={0.45} />
            <stop offset="95%" stopColor="hsl(var(--chart-1))" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" vertical={false} />
        <XAxis
          dataKey="day"
          tickFormatter={tickFormatter}
          tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
          axisLine={false}
          tickLine={false}
        />
        <YAxis
          allowDecimals={false}
          tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
          axisLine={false}
          tickLine={false}
          width={36}
        />
        <Tooltip
          labelFormatter={(v) => tickFormatter(String(v))}
          contentStyle={{
            background: "hsl(var(--popover))",
            border: "1px solid hsl(var(--border))",
            borderRadius: "0.5rem",
            fontSize: "0.75rem",
            color: "hsl(var(--popover-foreground))",
          }}
        />
        <Area
          type="monotone"
          dataKey="total"
          name="Events"
          stroke="hsl(var(--chart-1))"
          strokeWidth={2.5}
          fill="url(#eventsFill)"
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
