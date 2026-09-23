"use client";

import { Search, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useSources } from "@/hooks/use-webhooks";
import type { WebhookListFilter } from "@/lib/types";

const ALL = "__all__";

interface WebhookFiltersProps {
  filter: WebhookListFilter;
  onChange: (next: WebhookListFilter) => void;
}

export function WebhookFilters({ filter, onChange }: WebhookFiltersProps) {
  const { data: sourcesRes } = useSources();
  const sources = sourcesRes?.data ?? [];

  const hasActiveFilters =
    filter.source || filter.status || filter.event_type || filter.q || filter.from || filter.to;

  return (
    <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:flex-wrap">
      <div className="relative w-full lg:max-w-xs">
        <Search className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search source, event, endpoint…"
          className="pl-8"
          value={filter.q ?? ""}
          onChange={(e) => onChange({ ...filter, q: e.target.value, page: 1 })}
        />
      </div>

      <Select
        value={filter.source || ALL}
        onValueChange={(v) => onChange({ ...filter, source: v === ALL ? "" : v, page: 1 })}
      >
        <SelectTrigger className="w-full lg:w-44">
          <SelectValue placeholder="All sources" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL}>All sources</SelectItem>
          {sources.map((s) => (
            <SelectItem key={s.source} value={s.source}>
              {s.source} · {s.total}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select
        value={filter.status || ALL}
        onValueChange={(v) =>
          onChange({
            ...filter,
            status: v === ALL ? "" : (v as WebhookListFilter["status"]),
            page: 1,
          })
        }
      >
        <SelectTrigger className="w-full lg:w-40">
          <SelectValue placeholder="All statuses" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL}>All statuses</SelectItem>
          <SelectItem value="published">Published</SelectItem>
          <SelectItem value="pending">Pending</SelectItem>
          <SelectItem value="failed">Failed</SelectItem>
        </SelectContent>
      </Select>

      <Input
        placeholder="Event type"
        className="w-full lg:w-40"
        value={filter.event_type ?? ""}
        onChange={(e) => onChange({ ...filter, event_type: e.target.value, page: 1 })}
      />

      <div className="flex items-center gap-2">
        <Input
          type="date"
          className="w-full lg:w-[150px]"
          value={filter.from ?? ""}
          onChange={(e) => onChange({ ...filter, from: e.target.value, page: 1 })}
        />
        <span className="text-xs text-muted-foreground">to</span>
        <Input
          type="date"
          className="w-full lg:w-[150px]"
          value={filter.to ?? ""}
          onChange={(e) => onChange({ ...filter, to: e.target.value, page: 1 })}
        />
      </div>

      {hasActiveFilters && (
        <Button
          variant="ghost"
          size="sm"
          onClick={() =>
            onChange({ page: 1, limit: filter.limit, order: filter.order })
          }
        >
          <X className="h-3.5 w-3.5" />
          Clear
        </Button>
      )}
    </div>
  );
}
