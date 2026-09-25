"use client";

import * as React from "react";
import { Pencil, Plus, Route, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  useCreateTopicRoute,
  useDeleteTopicRoute,
  useTopicRoutes,
  useUpdateTopicRoute,
} from "@/hooks/use-webhooks";
import type { TopicRoute, TopicRouteInput } from "@/lib/types";

const EMPTY: TopicRouteInput = { env: "", source: "", topic: "" };

function RouteFormDialog({
  open,
  route,
  onOpenChange,
}: {
  open: boolean;
  route: TopicRoute | null;
  onOpenChange: (open: boolean) => void;
}) {
  const create = useCreateTopicRoute();
  const update = useUpdateTopicRoute();
  const [form, setForm] = React.useState<TopicRouteInput>(EMPTY);
  const pending = create.isPending || update.isPending;

  React.useEffect(() => {
    if (open) {
      setForm(
        route ? { env: route.env, source: route.source, topic: route.topic } : EMPTY
      );
    }
  }, [open, route]);

  const valid = form.env.trim() && form.source.trim() && form.topic.trim();

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!valid) return;
    const input = {
      env: form.env.trim(),
      source: form.source.trim(),
      topic: form.topic.trim(),
    };
    const opts = { onSuccess: () => onOpenChange(false) };
    if (route) update.mutate({ id: route.id, patch: input }, opts);
    else create.mutate(input, opts);
  }

  const field = (key: keyof TopicRouteInput) => ({
    value: form[key],
    onChange: (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [key]: e.target.value })),
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <form onSubmit={submit} className="space-y-4">
          <DialogHeader>
            <DialogTitle>{route ? "Edit topic route" : "Add topic route"}</DialogTitle>
            <DialogDescription>
              Use <code className="rounded bg-muted px-1">*</code> as env or source
              to match any value.
            </DialogDescription>
          </DialogHeader>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="route-env">Env</Label>
              <Input id="route-env" placeholder="production" {...field("env")} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="route-source">Source</Label>
              <Input id="route-source" placeholder="midtrans" {...field("source")} />
            </div>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="route-topic">Kafka topic</Label>
            <Input
              id="route-topic"
              placeholder="webhook_gateway.production.midtrans"
              {...field("topic")}
            />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={!valid || pending}>
              {pending ? "Saving…" : route ? "Save changes" : "Create route"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default function TopicsPage() {
  const { data, isLoading } = useTopicRoutes();
  const update = useUpdateTopicRoute();
  const remove = useDeleteTopicRoute();
  const [formOpen, setFormOpen] = React.useState(false);
  const [editing, setEditing] = React.useState<TopicRoute | null>(null);
  const [deleting, setDeleting] = React.useState<TopicRoute | null>(null);

  const routes = data?.data ?? [];

  function openCreate() {
    setEditing(null);
    setFormOpen(true);
  }
  function openEdit(route: TopicRoute) {
    setEditing(route);
    setFormOpen(true);
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          Overrides take precedence over the env-var topic template. Disabled routes
          fall back to the template.
        </p>
        <Button onClick={openCreate}>
          <Plus /> Add route
        </Button>
      </div>

      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : routes.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-3 py-16 text-center">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
                <Route className="h-5 w-5 text-muted-foreground" />
              </div>
              <p className="text-sm font-medium">No topic overrides yet</p>
              <p className="max-w-sm text-xs text-muted-foreground">
                All events currently publish to the default topics from the
                environment template.
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Env</TableHead>
                  <TableHead>Source</TableHead>
                  <TableHead>Topic</TableHead>
                  <TableHead>Enabled</TableHead>
                  <TableHead className="w-24 text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {routes.map((r) => (
                  <TableRow key={r.id}>
                    <TableCell>
                      <Badge variant="outline">{r.env}</Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{r.source}</Badge>
                    </TableCell>
                    <TableCell className="font-mono text-xs">{r.topic}</TableCell>
                    <TableCell>
                      <Switch
                        checked={r.enabled}
                        disabled={update.isPending}
                        onCheckedChange={(enabled) =>
                          update.mutate({ id: r.id, patch: { enabled } })
                        }
                        aria-label={`Toggle ${r.env}/${r.source}`}
                      />
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => openEdit(r)}
                        aria-label="Edit route"
                      >
                        <Pencil />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setDeleting(r)}
                        aria-label="Delete route"
                      >
                        <Trash2 className="text-destructive" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <RouteFormDialog open={formOpen} route={editing} onOpenChange={setFormOpen} />

      <Dialog open={Boolean(deleting)} onOpenChange={(o) => !o && setDeleting(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete topic route?</DialogTitle>
            <DialogDescription>
              {deleting?.env}/{deleting?.source} will fall back to the default topic
              template.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleting(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              disabled={remove.isPending}
              onClick={() =>
                deleting &&
                remove.mutate(deleting.id, { onSuccess: () => setDeleting(null) })
              }
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
