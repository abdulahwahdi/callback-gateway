"use client";

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { StatusBadge } from "@/components/status-badge";
import { JsonViewer } from "@/components/json-viewer";
import { useReplay, useWebhook } from "@/hooks/use-webhooks";
import { formatDateTime } from "@/lib/utils";
import { RotateCw, ShieldCheck, ShieldQuestion, ShieldX } from "lucide-react";

export function WebhookDetailSheet({
  id,
  onOpenChange,
}: {
  id: string | null;
  onOpenChange: (open: boolean) => void;
}) {
  const { data, isLoading } = useWebhook(id);
  const replay = useReplay();
  const log = data?.data;

  return (
    <Sheet open={Boolean(id)} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="flex w-full flex-col p-0 sm:max-w-2xl">
        <SheetHeader>
          <div className="flex items-start justify-between gap-4">
            <div>
              <SheetTitle className="capitalize">
                {log?.source ?? (isLoading ? "Loading…" : "Event detail")}
              </SheetTitle>
              <SheetDescription className="font-mono text-xs">{id}</SheetDescription>
            </div>
            {log && <StatusBadge status={log.publish_status} />}
          </div>
        </SheetHeader>

        <div className="scrollbar-thin flex-1 space-y-6 overflow-y-auto p-6">
          {isLoading || !log ? (
            <div className="space-y-3">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : (
            <>
              <section className="grid grid-cols-2 gap-4 text-sm">
                <Field label="Event type" value={log.event_type || "—"} />
                <Field label="Method / Endpoint" value={`${log.method} ${log.endpoint}`} mono />
                <Field label="Source IP" value={log.source_ip || "—"} mono />
                <Field label="Received at" value={formatDateTime(log.created_at)} />
                <Field label="Published at" value={formatDateTime(log.published_at)} />
                <Field label="Retry count" value={String(log.retry_count)} />
                <Field
                  label="Response status"
                  value={String(log.response_status_code || "—")}
                />
                <Field label="Signature">
                  <SignatureIndicator verified={log.signature_verified} />
                </Field>
              </section>

              {log.publish_error && (
                <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
                  <p className="mb-1 font-medium">Publish error</p>
                  <p className="font-mono text-xs">{log.publish_error}</p>
                </div>
              )}

              <section>
                <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Kafka topics
                </p>
                {log.kafka_topics?.length ? (
                  <div className="flex flex-wrap gap-1.5">
                    {log.kafka_topics.map((t) => (
                      <Badge key={t} variant="secondary" className="font-mono">
                        {t}
                      </Badge>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">Not yet published.</p>
                )}
              </section>

              <Separator />

              <section>
                <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Raw body
                </p>
                <JsonViewer value={log.body} emptyLabel="Empty body" />
              </section>

              <section>
                <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Headers
                </p>
                <JsonViewer value={log.headers} emptyLabel="No headers recorded" />
              </section>

              <section>
                <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Query params
                </p>
                <JsonViewer value={log.query_params} emptyLabel="No query params" />
              </section>
            </>
          )}
        </div>

        {log && (
          <div className="flex justify-end gap-2 border-t p-4">
            <Button
              onClick={() => replay.mutate(log.id)}
              disabled={replay.isPending}
              variant={log.publish_status === "failed" ? "default" : "outline"}
            >
              <RotateCw className={replay.isPending ? "h-4 w-4 animate-spin" : "h-4 w-4"} />
              Replay to Kafka
            </Button>
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}

function Field({
  label,
  value,
  mono,
  children,
}: {
  label: string;
  value?: string;
  mono?: boolean;
  children?: React.ReactNode;
}) {
  return (
    <div className="space-y-0.5">
      <p className="text-xs text-muted-foreground">{label}</p>
      {children ?? (
        <p className={mono ? "break-all font-mono text-xs" : "text-sm font-medium"}>{value}</p>
      )}
    </div>
  );
}

function SignatureIndicator({ verified }: { verified?: boolean | null }) {
  if (verified === true) {
    return (
      <p className="flex items-center gap-1.5 text-sm font-medium text-success">
        <ShieldCheck className="h-4 w-4" /> Verified
      </p>
    );
  }
  if (verified === false) {
    return (
      <p className="flex items-center gap-1.5 text-sm font-medium text-destructive">
        <ShieldX className="h-4 w-4" /> Failed
      </p>
    );
  }
  return (
    <p className="flex items-center gap-1.5 text-sm font-medium text-muted-foreground">
      <ShieldQuestion className="h-4 w-4" /> Not checked
    </p>
  );
}
