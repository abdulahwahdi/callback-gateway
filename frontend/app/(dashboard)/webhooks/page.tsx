"use client";

import * as React from "react";
import { Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { RotateCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { WebhookFilters } from "@/components/webhooks/webhook-filters";
import { WebhookTable } from "@/components/webhooks/webhook-table";
import { WebhookDetailSheet } from "@/components/webhooks/webhook-detail-sheet";
import { useRetryFailed, useWebhooks } from "@/hooks/use-webhooks";
import type { WebhookListFilter } from "@/lib/types";

const DEFAULT_FILTER: WebhookListFilter = {
  page: 1,
  limit: 20,
  order: "desc",
};

export default function WebhooksPage() {
  return (
    <Suspense fallback={null}>
      <WebhooksPageInner />
    </Suspense>
  );
}

function WebhooksPageInner() {
  const searchParams = useSearchParams();
  const [filter, setFilter] = React.useState<WebhookListFilter>(DEFAULT_FILTER);
  const [selectedId, setSelectedId] = React.useState<string | null>(null);
  const appliedPrefill = React.useRef(false);

  React.useEffect(() => {
    if (appliedPrefill.current) return;
    const source = searchParams.get("source");
    if (source) {
      setFilter((f) => ({ ...f, source }));
    }
    appliedPrefill.current = true;
  }, [searchParams]);

  const { data, isLoading, isFetching } = useWebhooks(filter);
  const retryFailed = useRetryFailed();

  const logs = data?.data ?? [];
  const pagination = data?.meta;

  return (
    <div className="space-y-4">
      <Card>
        <CardContent className="space-y-4 p-4">
          <div className="flex items-center justify-between gap-3">
            <WebhookFilters filter={filter} onChange={setFilter} />
          </div>
          <div className="flex items-center justify-between border-t pt-3 text-xs text-muted-foreground">
            <span>{isFetching ? "Refreshing…" : "Up to date"}</span>
            <Button
              variant="outline"
              size="sm"
              disabled={retryFailed.isPending}
              onClick={() => retryFailed.mutate(50)}
            >
              <RotateCw className={retryFailed.isPending ? "h-3.5 w-3.5 animate-spin" : "h-3.5 w-3.5"} />
              Retry all failed
            </Button>
          </div>
        </CardContent>
      </Card>

      <WebhookTable
        logs={logs}
        loading={isLoading}
        pagination={pagination}
        onPageChange={(page) => setFilter((f) => ({ ...f, page }))}
        onSelect={setSelectedId}
      />

      <WebhookDetailSheet id={selectedId} onOpenChange={(open) => !open && setSelectedId(null)} />
    </div>
  );
}
