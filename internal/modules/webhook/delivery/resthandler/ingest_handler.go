package resthandler

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"webhook-middleware/internal/pkg/response"
	"webhook-middleware/internal/modules/webhook/usecase"
)

const maxBodyBytes = 5 << 20 // 5MB, generous for a payment gateway callback payload

// Ingest is the single catch-all endpoint every payment gateway calls:
//
//	POST /webhooks/:source
//
// e.g. POST /webhooks/midtrans, POST /webhooks/xendit, POST /webhooks/stripe.
// It always answers fast with 2xx (once the callback is durably persisted)
// so gateways don't go into aggressive retry loops; Kafka publish failures
// are recorded on the log and handled by the retry worker / dashboard, not
// surfaced as an error to the caller.
func (h *RestHandler) Ingest(c echo.Context) error {
	source := c.Param("source")
	if source == "" {
		return response.Err(c, http.StatusBadRequest, "missing source in path, expected /webhooks/:source")
	}

	req := c.Request()
	body, err := io.ReadAll(io.LimitReader(req.Body, maxBodyBytes))
	if err != nil {
		return response.Err(c, http.StatusBadRequest, "failed to read request body")
	}

	log, err := h.uc.Ingest(req.Context(), usecase.IngestRequest{
		Source:   source,
		Method:   req.Method,
		Endpoint: req.URL.Path,
		SourceIP: c.RealIP(),
		Headers:  req.Header,
		Query:    map[string][]string(req.URL.Query()),
		Body:     body,
	})
	if err != nil {
		// We failed to even persist the callback -- ask the gateway to
		// retry rather than silently dropping it.
		return response.Err(c, http.StatusInternalServerError, "failed to record webhook event, please retry")
	}

	return response.OK(c, http.StatusOK, "webhook received", map[string]any{
		"id":             log.ID,
		"source":         log.Source,
		"publish_status": log.PublishStatus,
	})
}

// Health is a trivial liveness probe.
func (h *RestHandler) Health(c echo.Context) error {
	return response.OK(c, http.StatusOK, "ok", nil)
}
