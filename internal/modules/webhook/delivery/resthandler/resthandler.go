// Package resthandler exposes the webhook module over HTTP: the public
// ingestion endpoint payment gateways call, and the dashboard API used to
// track every inbound callback.
package resthandler

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"

	"webhook-middleware/api"
	"webhook-middleware/internal/modules/webhook/domain"
	"webhook-middleware/pkg/shared/usecase"

	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/wrapper"
)

// RestHandler handler
type RestHandler struct {
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewRestHandler create new rest handler
func NewRestHandler(uc usecase.Usecase, deps dependency.Dependency) *RestHandler {
	return &RestHandler{uc: uc, validator: deps.GetValidator()}
}

// Mount handler with root "/"
//
//   - POST /webhooks/:source              -- public, called by payment gateways
//   - POST /webhooks/:env/:source         -- same, with env selecting the Kafka topic route
//   - GET  /healthz                       -- liveness probe
//   - GET  /docs, /openapi.yaml           -- Swagger UI + raw OpenAPI spec
//   - POST /auth/login                    -- exchange DASHBOARD_USERNAME/PASSWORD for a session token
//   - GET  /auth/status                   -- whether the dashboard requires login
//   - GET  /dashboard/webhooks            -- list + filter
//   - GET  /dashboard/webhooks/sources    -- distinct sources + counts
//   - GET  /dashboard/webhooks/stats      -- aggregate stats for charts
//   - GET  /dashboard/webhooks/:id        -- single event detail
//   - POST /dashboard/webhooks/:id/replay -- re-publish one event
//   - POST /dashboard/webhooks/retry-failed -- bulk retry failed publishes
//   - GET    /dashboard/topic-routes      -- list per (env, source) Kafka topic overrides
//   - POST   /dashboard/topic-routes      -- add an override
//   - PUT    /dashboard/topic-routes/:id  -- update an override
//   - DELETE /dashboard/topic-routes/:id  -- remove an override
func (h *RestHandler) Mount(root interfaces.RESTRouter) {
	root.POST("/webhooks/:source", h.ingest)
	root.POST("/webhooks/:env/:source", h.ingest)
	root.GET("/healthz", h.health)
	api.MountDocs(root)

	root.POST("/auth/login", h.login)
	root.GET("/auth/status", h.authStatus)

	dashboardAuth := dashboardAuthMiddleware()

	webhooks := root.Group("/dashboard/webhooks", dashboardAuth)
	webhooks.GET("/", h.listWebhook)
	webhooks.GET("/sources", h.getSources)
	webhooks.GET("/stats", h.getStats)
	webhooks.GET("/:id", h.getDetailWebhook)
	webhooks.POST("/:id/replay", h.replayWebhook)
	webhooks.POST("/retry-failed", h.retryFailedWebhooks)

	topicRoutes := root.Group("/dashboard/topic-routes", dashboardAuth)
	topicRoutes.GET("/", h.listTopicRoutes)
	topicRoutes.POST("/", h.createTopicRoute)
	topicRoutes.PUT("/:id", h.updateTopicRoute)
	topicRoutes.DELETE("/:id", h.deleteTopicRoute)
}

// health is a trivial liveness probe.
func (h *RestHandler) health(rw http.ResponseWriter, req *http.Request) {
	wrapper.NewHTTPResponse(http.StatusOK, "ok").JSON(rw)
}

// writeError maps usecase/repository errors onto HTTP responses.
func writeError(rw http.ResponseWriter, err error, notFoundMessage, internalMessage string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		wrapper.NewHTTPResponse(http.StatusNotFound, notFoundMessage).JSON(rw)
	case errors.Is(err, domain.ErrInvalidTopicRoute):
		wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(rw)
	default:
		wrapper.NewHTTPResponse(http.StatusInternalServerError, internalMessage).JSON(rw)
	}
}

// realIP extracts the client address the same way echo's RealIP did:
// X-Forwarded-For (first hop), then X-Real-IP, then the socket peer.
func realIP(req *http.Request) string {
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		first, _, _ := strings.Cut(xff, ",")
		return strings.TrimSpace(first)
	}
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
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
