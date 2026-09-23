"use client";

import { ChevronLeft, ChevronRight, RotateCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { StatusBadge } from "@/components/status-badge";
import { useReplay } from "@/hooks/use-webhooks";
import { formatDateTime, formatRelativeTime } from "@/lib/utils";
import type { Pagination, WebhookLog } from "@/lib/types";

export function WebhookTable({
  logs,
  loading,
  pagination,
  onPageChange,
  onSelect,
}: {
  logs: WebhookLog[];
  loading: boolean;
  pagination?: Pagination;
  onPageChange: (page: number) => void;
  onSelect: (id: string) => void;
}) {
  const replay = useReplay();

  return (
    <div className="overflow-hidden rounded-xl border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Source</TableHead>
            <TableHead>Event type</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Signature</TableHead>
            <TableHead>Retries</TableHead>
            <TableHead>Received</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {loading ? (
            Array.from({ length: 8 }).map((_, i) => (
              <TableRow key={i}>
                {Array.from({ length: 7 }).map((__, j) => (
                  <TableCell key={j}>
                    <Skeleton className="h-4 w-full max-w-[120px]" />
                  </TableCell>
                ))}
              </TableRow>
            ))
          ) : logs.length === 0 ? (
            <TableRow>
              <TableCell colSpan={7} className="h-40 text-center text-sm text-muted-foreground">
                No webhook events match these filters.
              </TableCell>
            </TableRow>
          ) : (
            logs.map((log) => (
              <TableRow
                key={log.id}
                className="cursor-pointer"
                onClick={() => onSelect(log.id)}
              >
                <TableCell className="font-medium capitalize">{log.source}</TableCell>
                <TableCell className="max-w-[160px] truncate text-muted-foreground">
                  {log.event_type || "—"}
                </TableCell>
                <TableCell>
                  <StatusBadge status={log.publish_status} />
                </TableCell>
                <TableCell>
                  <SignatureDot verified={log.signature_verified} />
                </TableCell>
                <TableCell className="tabular-nums text-muted-foreground">
                  {log.retry_count}
                </TableCell>
                <TableCell
                  className="text-xs text-muted-foreground"
                  title={formatDateTime(log.created_at)}
                >
                  {formatRelativeTime(log.created_at)}
                </TableCell>
                <TableCell className="text-right">
                  {log.publish_status === "failed" && (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={replay.isPending}
                      onClick={(e) => {
                        e.stopPropagation();
                        replay.mutate(log.id);
                      }}
                    >
                      <RotateCw className="h-3.5 w-3.5" />
                      Replay
                    </Button>
                  )}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>

      {pagination && pagination.total_data > 0 && (
        <div className="flex flex-col items-center justify-between gap-3 border-t px-4 py-3 sm:flex-row">
          <p className="text-xs text-muted-foreground">
            Page <span className="font-medium text-foreground">{pagination.page}</span> of{" "}
            <span className="font-medium text-foreground">{pagination.total_pages || 1}</span>{" "}
            · {pagination.total_data} events total
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={pagination.page <= 1}
              onClick={() => onPageChange(pagination.page - 1)}
            >
              <ChevronLeft className="h-3.5 w-3.5" />
              Prev
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={pagination.page >= pagination.total_pages}
              onClick={() => onPageChange(pagination.page + 1)}
            >
              Next
              <ChevronRight className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

function SignatureDot({ verified }: { verified?: boolean | null }) {
  if (verified === true)
    return <span className="text-xs font-medium text-success">Verified</span>;
  if (verified === false)
    return <span className="text-xs font-medium text-destructive">Failed</span>;
  return <span className="text-xs text-muted-foreground">—</span>;
}
