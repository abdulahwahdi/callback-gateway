package resthandler

import (
	"io"
	"net/http"

	"webhook-middleware/internal/modules/webhook/domain"

	restserver "github.com/golangid/candi/codebase/app/rest_server"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
)

const maxBodyBytes = 5 << 20 // 5MB, generous for a payment gateway callback payload

// ingest is the single catch-all endpoint every payment gateway calls:
//
//	POST /webhooks/:source
//	POST /webhooks/:env/:source
//
// e.g. POST /webhooks/midtrans, POST /webhooks/staging/dummy. The env
// segment picks the Kafka topic routing (see /dashboard/topic-routes); the
// env-less form uses the service's APP_ENV.
// It always answers fast with 2xx (once the callback is durably persisted)
// so gateways don't go into aggressive retry loops; Kafka publish failures
// are recorded on the log and handled by the retry cron job / dashboard,
// not surfaced as an error to the caller.
func (h *RestHandler) ingest(rw http.ResponseWriter, req *http.Request) {
	trace, ctx := tracer.StartTraceWithContext(req.Context(), "WebhookDeliveryREST:Ingest")
	defer trace.Finish()

	env := restserver.URLParam(req, "env")
	source := restserver.URLParam(req, "source")
	if source == "" {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "missing source in path, expected /webhooks/:source").JSON(rw)
		return
	}

	body, err := io.ReadAll(io.LimitReader(req.Body, maxBodyBytes))
	if err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "failed to read request body").JSON(rw)
		return
	}

	log, err := h.uc.Webhook().Ingest(ctx, &domain.IngestRequest{
		Env:      env,
		Source:   source,
		Method:   req.Method,
		Endpoint: req.URL.Path,
		SourceIP: realIP(req),
		Headers:  req.Header,
		Query:    map[string][]string(req.URL.Query()),
		Body:     body,
	})
	if err != nil {
		// We failed to even persist the callback -- ask the gateway to
		// retry rather than silently dropping it.
		trace.SetError(err)
		wrapper.NewHTTPResponse(http.StatusInternalServerError, "failed to record webhook event, please retry").JSON(rw)
		return
	}

	wrapper.NewHTTPResponse(http.StatusOK, "webhook received", domain.ResponseIngest{
		ID:            log.ID,
		Source:        log.Source,
		PublishStatus: string(log.PublishStatus),
	}).JSON(rw)
}
