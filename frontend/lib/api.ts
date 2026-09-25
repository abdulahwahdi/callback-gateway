import type {
  Envelope,
  RetryFailedResult,
  SourceCount,
  Stats,
  TopicRoute,
  TopicRouteInput,
  TopicRoutePatch,
  WebhookListFilter,
  WebhookLog,
} from "@/lib/types";
import type { DashboardSettings } from "@/lib/settings";

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

function buildUrl(base: string, path: string, params?: Record<string, unknown>) {
  const url = new URL(path.replace(/^\//, ""), base.endsWith("/") ? base : `${base}/`);
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value === undefined || value === null || value === "") continue;
      url.searchParams.set(key, String(value));
    }
  }
  return url.toString();
}

async function request<T>(
  cfg: DashboardSettings,
  path: string,
  init?: RequestInit & { params?: Record<string, unknown> }
): Promise<Envelope<T>> {
  const { params, ...rest } = init ?? {};
  const url = buildUrl(cfg.apiBaseUrl, path, params);

  const headers = new Headers(rest.headers);
  headers.set("Accept", "application/json");
  if (rest.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (cfg.apiKey) headers.set("X-API-Key", cfg.apiKey);
  if (cfg.token) headers.set("Authorization", `Bearer ${cfg.token}`);

  let res: Response;
  try {
    res = await fetch(url, { ...rest, headers, cache: "no-store" });
  } catch (err) {
    throw new ApiError(
      `Could not reach ${cfg.apiBaseUrl}. Is the backend running and CORS-enabled? (${
        err instanceof Error ? err.message : "network error"
      })`,
      0
    );
  }

  let json: Envelope<T> | null = null;
  try {
    json = await res.json();
  } catch {
    // non-JSON body, fall through to status-based error below
  }

  if (res.status === 401 && path !== "/auth/login" && typeof window !== "undefined") {
    // Missing, expired or revoked session: let the AuthGate send the user back to /login.
    window.dispatchEvent(new Event("auth:unauthorized"));
  }

  if (!res.ok || (json && json.success === false)) {
    throw new ApiError(
      json?.message || `Request failed with status ${res.status}`,
      res.status
    );
  }

  if (!json) {
    throw new ApiError("Empty response from server", res.status);
  }

  return json;
}

export const api = {
  authStatus: (cfg: DashboardSettings) =>
    request<{ login_enabled: boolean; auth_required: boolean }>(cfg, "/auth/status"),

  login: (cfg: DashboardSettings, username: string, password: string) =>
    request<{ token: string; expires_at: string; username: string }>(cfg, "/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    }),

  health: (cfg: DashboardSettings) =>
    request<unknown>(cfg, "/healthz"),

  listWebhooks: (cfg: DashboardSettings, filter: WebhookListFilter) =>
    request<WebhookLog[]>(cfg, "/dashboard/webhooks", {
      params: {
        source: filter.source,
        status: filter.status,
        event_type: filter.event_type,
        q: filter.q,
        from: filter.from,
        to: filter.to,
        page: filter.page,
        limit: filter.limit,
        order: filter.order,
      },
    }),

  getWebhook: (cfg: DashboardSettings, id: string) =>
    request<WebhookLog>(cfg, `/dashboard/webhooks/${id}`),

  getSources: (cfg: DashboardSettings) =>
    request<SourceCount[]>(cfg, "/dashboard/webhooks/sources"),

  getStats: (cfg: DashboardSettings, from?: string, to?: string) =>
    request<Stats>(cfg, "/dashboard/webhooks/stats", { params: { from, to } }),

  replay: (cfg: DashboardSettings, id: string) =>
    request<WebhookLog>(cfg, `/dashboard/webhooks/${id}/replay`, {
      method: "POST",
    }),

  retryFailed: (cfg: DashboardSettings, limit = 50) =>
    request<RetryFailedResult>(cfg, "/dashboard/webhooks/retry-failed", {
      method: "POST",
      params: { limit },
    }),

  listTopicRoutes: (cfg: DashboardSettings) =>
    request<TopicRoute[]>(cfg, "/dashboard/topic-routes"),

  createTopicRoute: (cfg: DashboardSettings, input: TopicRouteInput) =>
    request<TopicRoute>(cfg, "/dashboard/topic-routes", {
      method: "POST",
      body: JSON.stringify(input),
    }),

  updateTopicRoute: (cfg: DashboardSettings, id: string, patch: TopicRoutePatch) =>
    request<TopicRoute>(cfg, `/dashboard/topic-routes/${id}`, {
      method: "PUT",
      body: JSON.stringify(patch),
    }),

  deleteTopicRoute: (cfg: DashboardSettings, id: string) =>
    request<unknown>(cfg, `/dashboard/topic-routes/${id}`, { method: "DELETE" }),
};
