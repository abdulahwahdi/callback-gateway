"use client";

import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { api, ApiError } from "@/lib/api";
import { useSettings } from "@/lib/settings";
import type { WebhookListFilter } from "@/lib/types";
import { toast } from "sonner";

const KEYS = {
  webhooks: (filter: WebhookListFilter) => ["webhooks", filter] as const,
  webhook: (id: string) => ["webhook", id] as const,
  sources: () => ["sources"] as const,
  stats: (from?: string, to?: string) => ["stats", from, to] as const,
  health: () => ["health"] as const,
};

export function useWebhooks(filter: WebhookListFilter) {
  const { settings } = useSettings();
  return useQuery({
    queryKey: KEYS.webhooks(filter),
    queryFn: () => api.listWebhooks(settings, filter),
    placeholderData: keepPreviousData,
  });
}

export function useWebhook(id: string | null) {
  const { settings } = useSettings();
  return useQuery({
    queryKey: KEYS.webhook(id ?? ""),
    queryFn: () => api.getWebhook(settings, id as string),
    enabled: Boolean(id),
  });
}

export function useSources() {
  const { settings } = useSettings();
  return useQuery({
    queryKey: KEYS.sources(),
    queryFn: () => api.getSources(settings),
  });
}

export function useStats(from?: string, to?: string) {
  const { settings } = useSettings();
  return useQuery({
    queryKey: KEYS.stats(from, to),
    queryFn: () => api.getStats(settings, from, to),
    refetchInterval: 30_000,
  });
}

export function useHealth() {
  const { settings } = useSettings();
  return useQuery({
    queryKey: KEYS.health(),
    queryFn: () => api.health(settings),
    retry: false,
    refetchInterval: 20_000,
  });
}

export function useReplay() {
  const { settings } = useSettings();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.replay(settings, id),
    onSuccess: (_res, id) => {
      toast.success("Event replayed", {
        description: "The stored payload was re-published to Kafka.",
      });
      qc.invalidateQueries({ queryKey: ["webhooks"] });
      qc.invalidateQueries({ queryKey: KEYS.webhook(id) });
      qc.invalidateQueries({ queryKey: ["stats"] });
    },
    onError: (err) => {
      toast.error("Replay failed", {
        description: err instanceof ApiError ? err.message : "Unknown error",
      });
    },
  });
}

export function useRetryFailed() {
  const { settings } = useSettings();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (limit?: number) => api.retryFailed(settings, limit),
    onSuccess: (res) => {
      const { retried, failed } = res.data;
      toast.success("Retry sweep completed", {
        description: `${retried} retried, ${failed} still failed.`,
      });
      qc.invalidateQueries({ queryKey: ["webhooks"] });
      qc.invalidateQueries({ queryKey: ["stats"] });
    },
    onError: (err) => {
      toast.error("Retry sweep failed", {
        description: err instanceof ApiError ? err.message : "Unknown error",
      });
    },
  });
}
