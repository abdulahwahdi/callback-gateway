"use client";

import * as React from "react";
import Link from "next/link";
import { format, subDays } from "date-fns";
import {
  AlertTriangle,
  ArrowUpRight,
  CheckCircle2,
  Clock3,
  Layers,
  RotateCw,
  XCircle,
} from "lucide-react";

import { StatCard } from "@/components/stat-card";
import { StatusBadge } from "@/components/status-badge";
import { EventsTimelineChart } from "@/components/charts/events-timeline-chart";
import { StatusDonutChart } from "@/components/charts/status-donut-chart";
import { SourceBarChart } from "@/components/charts/source-bar-chart";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useStats, useWebhooks, useRetryFailed } from "@/hooks/use-webhooks";
import { formatRelativeTime, numberCompact } from "@/lib/utils";

const RANGES = [
  { label: "7d", days: 7 },
  { label: "14d", days: 14 },
  { label: "30d", days: 30 },
] as const;

export default function OverviewPage() {
  const [rangeDays, setRangeDays] = React.useState<number>(7);

  const from = format(subDays(new Date(), rangeDays), "yyyy-MM-dd");
  const to = format(new Date(), "yyyy-MM-dd");

  const { data: statsRes, isLoading: statsLoading } = useStats(from, to);
  const { data: recentRes, isLoading: recentLoading } = useWebhooks({
    page: 1,
    limit: 8,
    order: "desc",
  });
  const retryFailed = useRetryFailed();

  const stats = statsRes?.data;
  const recent = recentRes?.data ?? [];

  const failedTotal =
    stats?.by_status?.find((s) => s.publish_status === "failed")?.total ?? 0;
  const publishedTotal =
    stats?.by_status?.find((s) => s.publish_status === "published")?.total ?? 0;
  const pendingTotal =
    stats?.by_status?.find((s) => s.publish_status === "pending")?.total ?? 0;

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="text-sm text-muted-foreground">
            Showing activity from <span className="font-medium text-foreground">{from}</span> to{" "}
            <span className="font-medium text-foreground">{to}</span>
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Tabs value={String(rangeDays)} onValueChange={(v) => setRangeDays(Number(v))}>
            <TabsList>
              {RANGES.map((r) => (
                <TabsTrigger key={r.days} value={String(r.days)}>
                  {r.label}
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
          <Button
            variant="secondary"
            size="sm"
            disabled={retryFailed.isPending}
            onClick={() => retryFailed.mutate(50)}
          >
            <RotateCw className={retryFailed.isPending ? "h-4 w-4 animate-spin" : "h-4 w-4"} />
            Retry failed
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Total events"
          value={stats ? numberCompact(stats.total_events) : "—"}
          icon={Layers}
          loading={statsLoading}
          hint={`Last ${rangeDays} days`}
        />
        <StatCard
          label="Published"
          value={stats ? numberCompact(publishedTotal) : "—"}
          icon={CheckCircle2}
          tone="success"
          loading={statsLoading}
          hint="Delivered to Kafka"
        />
        <StatCard
          label="Pending"
          value={stats ? numberCompact(pendingTotal) : "—"}
          icon={Clock3}
          tone="warning"
          loading={statsLoading}
          hint="Awaiting publish"
        />
        <StatCard
          label="Failed"
          value={stats ? numberCompact(failedTotal) : "—"}
          icon={XCircle}
          tone="destructive"
          loading={statsLoading}
          hint={
            stats && stats.failed_pending > 0
              ? `${numberCompact(stats.failed_pending)} need attention`
              : "All clear"
          }
        />
      </div>

      {stats && stats.failed_pending > 0 && (
        <Card className="border-destructive/30 bg-destructive/5">
          <CardContent className="flex flex-col items-start justify-between gap-3 p-4 sm:flex-row sm:items-center">
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-destructive/15 text-destructive">
                <AlertTriangle className="h-4 w-4" />
              </div>
              <div>
                <p className="text-sm font-medium">
                  {numberCompact(stats.failed_pending)} events stuck in a failed publish state
                </p>
                <p className="text-xs text-muted-foreground">
                  The background retry worker will sweep these automatically, or trigger it now.
                </p>
              </div>
            </div>
            <Button
              size="sm"
              variant="destructive"
              disabled={retryFailed.isPending}
              onClick={() => retryFailed.mutate(50)}
            >
              <RotateCw className={retryFailed.isPending ? "h-4 w-4 animate-spin" : "h-4 w-4"} />
              Retry now
            </Button>
          </CardContent>
        </Card>
      )}

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Card className="xl:col-span-2">
          <CardHeader>
            <CardTitle>Events over time</CardTitle>
            <CardDescription>Daily inbound webhook volume across every source.</CardDescription>
          </CardHeader>
          <CardContent>
            <EventsTimelineChart data={stats?.by_day ?? []} loading={statsLoading} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Publish status</CardTitle>
            <CardDescription>Kafka publish outcome breakdown.</CardDescription>
          </CardHeader>
          <CardContent>
            <StatusDonutChart data={stats?.by_status ?? []} loading={statsLoading} />
            <div className="mt-2 flex items-center justify-center gap-4 text-xs text-muted-foreground">
              <LegendDot color="hsl(var(--success))" label="Published" />
              <LegendDot color="hsl(var(--warning))" label="Pending" />
              <LegendDot color="hsl(var(--destructive))" label="Failed" />
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Card className="xl:col-span-1">
          <CardHeader>
            <CardTitle>Top sources</CardTitle>
            <CardDescription>Which gateways send the most traffic.</CardDescription>
          </CardHeader>
          <CardContent>
            <SourceBarChart data={stats?.by_source ?? []} loading={statsLoading} />
          </CardContent>
        </Card>

        <Card className="xl:col-span-2">
          <CardHeader className="flex flex-row items-center justify-between space-y-0">
            <div>
              <CardTitle>Recent events</CardTitle>
              <CardDescription>The latest callbacks received.</CardDescription>
            </div>
            <Button variant="ghost" size="sm" asChild>
              <Link href="/webhooks">
                View all
                <ArrowUpRight className="h-3.5 w-3.5" />
              </Link>
            </Button>
          </CardHeader>
          <CardContent className="p-0">
            {recentLoading ? (
              <div className="space-y-2 p-6 pt-0">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} className="h-10 w-full" />
                ))}
              </div>
            ) : recent.length === 0 ? (
              <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
                No events received yet.
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Source</TableHead>
                    <TableHead>Event</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Received</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {recent.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="font-medium capitalize">{log.source}</TableCell>
                      <TableCell className="text-muted-foreground">
                        {log.event_type || "—"}
                      </TableCell>
                      <TableCell>
                        <StatusBadge status={log.publish_status} />
                      </TableCell>
                      <TableCell className="text-right text-xs text-muted-foreground">
                        {formatRelativeTime(log.created_at)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function LegendDot({ color, label }: { color: string; label: string }) {
  return (
    <span className="flex items-center gap-1.5">
      <span className="h-2 w-2 rounded-full" style={{ backgroundColor: color }} />
      {label}
    </span>
  );
}
