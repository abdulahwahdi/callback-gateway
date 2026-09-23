package resthandler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"webhook-middleware/internal/modules/webhook/repository"
	"webhook-middleware/internal/pkg/response"
)

// List handles GET /dashboard/webhooks -- every incoming webhook, filterable
// by source, publish_status, event_type, free-text search, and a
// created_at date range, paginated.
//
// Query params: source, status, event_type, q, from, to, page, limit, sort, order(asc|desc)
func (h *RestHandler) List(c echo.Context) error {
	filter := repository.Filter{
		Source:        c.QueryParam("source"),
		PublishStatus: c.QueryParam("status"),
		EventType:     c.QueryParam("event_type"),
		Search:        c.QueryParam("q"),
		Page:          atoiDefault(c.QueryParam("page"), 1),
		Limit:         atoiDefault(c.QueryParam("limit"), 20),
		SortBy:        "created_at",
		SortDesc:      c.QueryParam("order") != "asc",
	}

	if v := c.QueryParam("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.From = &t
		} else if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.From = &t
		}
	}
	if v := c.QueryParam("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.To = &t
		} else if t, err := time.Parse("2006-01-02", v); err == nil {
			t = t.Add(24*time.Hour - time.Second)
			filter.To = &t
		}
	}

	logs, total, err := h.uc.List(c.Request().Context(), filter)
	if err != nil {
		return response.Err(c, http.StatusInternalServerError, "failed to list webhook events")
	}

	filter.Normalize()
	return response.OKPaginated(c, http.StatusOK, "ok", logs, filter.Page, filter.Limit, total)
}

// Detail handles GET /dashboard/webhooks/:id -- full record for one event,
// including headers, raw body, and Kafka publish outcome.
func (h *RestHandler) Detail(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Err(c, http.StatusBadRequest, "invalid id")
	}

	log, err := h.uc.Detail(c.Request().Context(), id)
	if err != nil {
		return response.Err(c, http.StatusNotFound, "webhook event not found")
	}

	return response.OK(c, http.StatusOK, "ok", log)
}

// Sources handles GET /dashboard/webhooks/sources -- distinct payment
// gateways seen so far, with event counts, for populating dashboard filters.
func (h *RestHandler) Sources(c echo.Context) error {
	sources, err := h.uc.Sources(c.Request().Context())
	if err != nil {
		return response.Err(c, http.StatusInternalServerError, "failed to load sources")
	}
	return response.OK(c, http.StatusOK, "ok", sources)
}

// Stats handles GET /dashboard/webhooks/stats -- aggregate counts (by
// source, by publish status, per day) for the dashboard overview/charts.
// Query params: from, to (default: last 7 days).
func (h *RestHandler) Stats(c echo.Context) error {
	to := time.Now()
	from := to.AddDate(0, 0, -7)

	if v := c.QueryParam("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = t
		}
	}
	if v := c.QueryParam("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = t.Add(24*time.Hour - time.Second)
		}
	}

	stats, err := h.uc.Stats(c.Request().Context(), from, to)
	if err != nil {
		return response.Err(c, http.StatusInternalServerError, "failed to compute stats")
	}
	return response.OK(c, http.StatusOK, "ok", stats)
}

// Replay handles POST /dashboard/webhooks/:id/replay -- manually re-publish
// a previously received event's original body to Kafka again.
func (h *RestHandler) Replay(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Err(c, http.StatusBadRequest, "invalid id")
	}

	log, err := h.uc.Replay(c.Request().Context(), id)
	if err != nil {
		return response.Err(c, http.StatusNotFound, "webhook event not found")
	}

	return response.OK(c, http.StatusOK, "replayed", log)
}

// RetryFailed handles POST /dashboard/webhooks/retry-failed -- bulk retry
// every event currently stuck in "failed" publish status.
// Query param: limit (default 50).
func (h *RestHandler) RetryFailed(c echo.Context) error {
	limit := atoiDefault(c.QueryParam("limit"), 50)

	retried, failed, err := h.uc.RetryFailed(c.Request().Context(), limit)
	if err != nil {
		return response.Err(c, http.StatusInternalServerError, "failed to retry events")
	}

	return response.OK(c, http.StatusOK, "retry completed", map[string]any{
		"retried": retried,
		"failed":  failed,
	})
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
