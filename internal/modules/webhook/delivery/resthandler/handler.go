// Package resthandler exposes the webhook module over HTTP: the public
// ingestion endpoint payment gateways call, and the dashboard API used to
// track every inbound callback.
package resthandler

import (
	"github.com/labstack/echo/v4"

	"webhook-middleware/internal/modules/webhook/usecase"
)

// RestHandler wires the webhook module's HTTP routes.
type RestHandler struct {
	uc usecase.WebhookUsecase
}

// New builds a RestHandler for the webhook module.
func New(uc usecase.WebhookUsecase) *RestHandler {
	return &RestHandler{uc: uc}
}

// Mount registers every route this module owns onto the given echo instance.
//
//   - POST /webhooks/:source           -- public, called by payment gateways
//   - GET  /dashboard/webhooks         -- list + filter
//   - GET  /dashboard/webhooks/sources -- distinct sources + counts
//   - GET  /dashboard/webhooks/stats   -- aggregate stats for charts
//   - GET  /dashboard/webhooks/:id     -- single event detail
//   - POST /dashboard/webhooks/:id/replay -- re-publish one event
//   - POST /dashboard/webhooks/retry-failed -- bulk retry failed publishes
func (h *RestHandler) Mount(e *echo.Echo, dashboardGroup ...echo.MiddlewareFunc) {
	e.POST("/webhooks/:source", h.Ingest)

	e.GET("/healthz", h.Health)

	dashboard := e.Group("/dashboard/webhooks", dashboardGroup...)
	dashboard.GET("", h.List)
	dashboard.GET("/sources", h.Sources)
	dashboard.GET("/stats", h.Stats)
	dashboard.GET("/:id", h.Detail)
	dashboard.POST("/:id/replay", h.Replay)
	dashboard.POST("/retry-failed", h.RetryFailed)
}
