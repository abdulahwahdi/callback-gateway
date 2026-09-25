package resthandler

import (
	"net/http"
	"time"

	"webhook-middleware/internal/modules/webhook/domain"

	restserver "github.com/golangid/candi/codebase/app/rest_server"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
	"github.com/google/uuid"
)

// listWebhook handles GET /dashboard/webhooks -- every incoming webhook,
// filterable by source, publish_status, event_type, free-text search, and a
// created_at date range, paginated.
//
// Query params: source, status, event_type, q, from, to, page, limit, order(asc|desc)
func (h *RestHandler) listWebhook(rw http.ResponseWriter, req *http.Request) {
	trace, ctx := tracer.StartTraceWithContext(req.Context(), "WebhookDeliveryREST:ListWebhook")
	defer trace.Finish()

	q := req.URL.Query()
	filter := domain.FilterWebhookLog{
		Source:        q.Get("source"),
		PublishStatus: q.Get("status"),
		EventType:     q.Get("event_type"),
		Search:        q.Get("q"),
		Page:          atoiDefault(q.Get("page"), 1),
		Limit:         atoiDefault(q.Get("limit"), 20),
		SortBy:        "created_at",
		SortDesc:      q.Get("order") != "asc",
	}

	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.From = &t
		} else if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.From = &t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.To = &t
		} else if t, err := time.Parse("2006-01-02", v); err == nil {
			t = t.Add(24*time.Hour - time.Second)
			filter.To = &t
		}
	}

	logs, total, err := h.uc.Webhook().ListWebhook(ctx, &filter)
	if err != nil {
		trace.SetError(err)
		wrapper.NewHTTPResponse(http.StatusInternalServerError, "failed to list webhook events").JSON(rw)
		return
	}

	filter.Normalize()
	wrapper.NewHTTPResponseWithMeta(http.StatusOK, "ok",
		domain.NewPagination(filter.Page, filter.Limit, total), logs).JSON(rw)
}

// getDetailWebhook handles GET /dashboard/webhooks/:id -- full record for
// one event, including headers, raw body, and Kafka publish outcome.
func (h *RestHandler) getDetailWebhook(rw http.ResponseWriter, req *http.Request) {
	trace, ctx := tracer.StartTraceWithContext(req.Context(), "WebhookDeliveryREST:GetDetailWebhook")
	defer trace.Finish()

	id, err := uuid.Parse(restserver.URLParam(req, "id"))
	if err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "invalid id").JSON(rw)
		return
	}

	log, err := h.uc.Webhook().GetDetailWebhook(ctx, id)
	if err != nil {
		trace.SetError(err)
		writeError(rw, err, "webhook event not found", "failed to load webhook event")
		return
	}

	wrapper.NewHTTPResponse(http.StatusOK, "ok", log).JSON(rw)
}

// getSources handles GET /dashboard/webhooks/sources -- distinct payment
// gateways seen so far, with event counts, for populating dashboard filters.
func (h *RestHandler) getSources(rw http.ResponseWriter, req *http.Request) {
	trace, ctx := tracer.StartTraceWithContext(req.Context(), "WebhookDeliveryREST:GetSources")
	defer trace.Finish()

	sources, err := h.uc.Webhook().GetSources(ctx)
	if err != nil {
		trace.SetError(err)
		wrapper.NewHTTPResponse(http.StatusInternalServerError, "failed to load sources").JSON(rw)
		return
	}
	wrapper.NewHTTPResponse(http.StatusOK, "ok", sources).JSON(rw)
}

// getStats handles GET /dashboard/webhooks/stats -- aggregate counts (by
// source, by publish status, per day) for the dashboard overview/charts.
// Query params: from, to (default: last 7 days).
func (h *RestHandler) getStats(rw http.ResponseWriter, req *http.Request) {
	trace, ctx := tracer.StartTraceWithContext(req.Context(), "WebhookDeliveryREST:GetStats")
	defer trace.Finish()

	q := req.URL.Query()
	to := time.Now()
	from := to.AddDate(0, 0, -7)

	if v := q.Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = t.Add(24*time.Hour - time.Second)
		}
	}

	stats, err := h.uc.Webhook().GetStats(ctx, from, to)
	if err != nil {
		trace.SetError(err)
		wrapper.NewHTTPResponse(http.StatusInternalServerError, "failed to compute stats").JSON(rw)
		return
	}
	wrapper.NewHTTPResponse(http.StatusOK, "ok", stats).JSON(rw)
}

// replayWebhook handles POST /dashboard/webhooks/:id/replay -- manually
// re-publish a previously received event's original body to Kafka again.
func (h *RestHandler) replayWebhook(rw http.ResponseWriter, req *http.Request) {
	trace, ctx := tracer.StartTraceWithContext(req.Context(), "WebhookDeliveryREST:ReplayWebhook")
	defer trace.Finish()

	id, err := uuid.Parse(restserver.URLParam(req, "id"))
	if err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "invalid id").JSON(rw)
		return
	}

	log, err := h.uc.Webhook().ReplayWebhook(ctx, id)
	if err != nil {
		trace.SetError(err)
		writeError(rw, err, "webhook event not found", "failed to replay webhook event")
		return
	}

	wrapper.NewHTTPResponse(http.StatusOK, "replayed", log).JSON(rw)
}

// retryFailedWebhooks handles POST /dashboard/webhooks/retry-failed -- bulk
// retry every event currently stuck in "failed" publish status.
// Query param: limit (default 50).
func (h *RestHandler) retryFailedWebhooks(rw http.ResponseWriter, req *http.Request) {
	trace, ctx := tracer.StartTraceWithContext(req.Context(), "WebhookDeliveryREST:RetryFailedWebhooks")
	defer trace.Finish()

	limit := atoiDefault(req.URL.Query().Get("limit"), 50)

	result, err := h.uc.Webhook().RetryFailedWebhooks(ctx, limit)
	if err != nil {
		trace.SetError(err)
		wrapper.NewHTTPResponse(http.StatusInternalServerError, "failed to retry events").JSON(rw)
		return
	}

	wrapper.NewHTTPResponse(http.StatusOK, "retry completed", result).JSON(rw)
}
