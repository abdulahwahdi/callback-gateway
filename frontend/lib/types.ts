// Mirrors internal/modules/webhook/domain + repository JSON shapes, and the
// internal/pkg/response.Envelope wrapper every endpoint returns.

export type PublishStatus = "pending" | "published" | "failed";

export interface WebhookLog {
  id: string;
  source: string;
  event_type: string;
  method: string;
  endpoint: string;
  source_ip: string;
  headers: Record<string, unknown>;
  query_params: Record<string, unknown>;
  body: unknown;
  signature_header?: string;
  signature_verified?: boolean | null;
  kafka_topics: string[];
  publish_status: PublishStatus;
  publish_error?: string;
  published_at?: string | null;
  retry_count: number;
  response_status_code: number;
  created_at: string;
  updated_at: string;
}

export interface SourceCount {
  source: string;
  total: number;
}

export interface StatusCount {
  publish_status: PublishStatus | string;
  total: number;
}

export interface DailyCount {
  day: string;
  total: number;
}

export interface Stats {
  total_events: number;
  by_source: SourceCount[];
  by_status: StatusCount[];
  by_day: DailyCount[];
  failed_pending: number;
}

export interface Pagination {
  page: number;
  limit: number;
  total_data: number;
  total_pages: number;
}

export interface Envelope<T> {
  success: boolean;
  message?: string;
  data: T;
  meta?: Pagination;
}

export interface WebhookListFilter {
  source?: string;
  status?: PublishStatus | "";
  event_type?: string;
  q?: string;
  from?: string;
  to?: string;
  page?: number;
  limit?: number;
  order?: "asc" | "desc";
}

export interface RetryFailedResult {
  retried: number;
  failed: number;
}

export interface TopicRoute {
  id: string;
  env: string;
  source: string;
  topic: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface TopicRouteInput {
  env: string;
  source: string;
  topic: string;
}

export type TopicRoutePatch = Partial<TopicRouteInput & { enabled: boolean }>;
