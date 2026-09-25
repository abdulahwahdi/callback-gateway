package resthandler

import (
	"encoding/json"
	"io"
	"net/http"

	"webhook-middleware/internal/modules/webhook/domain"

	restserver "github.com/golangid/candi/codebase/app/rest_server"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
	"github.com/google/uuid"
)

const maxTopicRouteBodyBytes = 1 << 20

// listTopicRoutes handles GET /dashboard/topic-routes -- every DB-configured
// (env, source) -> Kafka topic override, enabled or not.
func (h *RestHandler) listTopicRoutes(rw http.ResponseWriter, req *http.Request) {
	trace, ctx := tracer.StartTraceWithContext(req.Context(), "WebhookDeliveryREST:ListTopicRoutes")
	defer trace.Finish()

	routes, err := h.uc.Webhook().ListTopicRoutes(ctx)
	if err != nil {
		trace.SetError(err)
		wrapper.NewHTTPResponse(http.StatusInternalServerError, "failed to list topic routes").JSON(rw)
		return
	}
	wrapper.NewHTTPResponse(http.StatusOK, "ok", routes).JSON(rw)
}

// createTopicRoute handles POST /dashboard/topic-routes -- add an override
// so a given (env, source) pair (either may be "*" as a wildcard) publishes
// to a specific Kafka topic instead of the env-var template's default.
func (h *RestHandler) createTopicRoute(rw http.ResponseWriter, req *http.Request) {
	trace, ctx := tracer.StartTraceWithContext(req.Context(), "WebhookDeliveryREST:CreateTopicRoute")
	defer trace.Finish()

	body, err := io.ReadAll(io.LimitReader(req.Body, maxTopicRouteBodyBytes))
	if err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "invalid request body").JSON(rw)
		return
	}
	if err := h.validator.ValidateDocument("topic_route/create", body); err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed validate payload", err).JSON(rw)
		return
	}

	var payload domain.RequestCreateTopicRoute
	if err := json.Unmarshal(body, &payload); err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "invalid request body").JSON(rw)
		return
	}

	route, err := h.uc.Webhook().CreateTopicRoute(ctx, &payload)
	if err != nil {
		trace.SetError(err)
		wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(rw)
		return
	}
	wrapper.NewHTTPResponse(http.StatusCreated, "created", route).JSON(rw)
}

// updateTopicRoute handles PUT /dashboard/topic-routes/:id -- partial update
// (e.g. flip "enabled" to fall back to the template without deleting the row).
func (h *RestHandler) updateTopicRoute(rw http.ResponseWriter, req *http.Request) {
	trace, ctx := tracer.StartTraceWithContext(req.Context(), "WebhookDeliveryREST:UpdateTopicRoute")
	defer trace.Finish()

	id, err := uuid.Parse(restserver.URLParam(req, "id"))
	if err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "invalid id").JSON(rw)
		return
	}

	body, err := io.ReadAll(io.LimitReader(req.Body, maxTopicRouteBodyBytes))
	if err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "invalid request body").JSON(rw)
		return
	}
	if err := h.validator.ValidateDocument("topic_route/update", body); err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed validate payload", err).JSON(rw)
		return
	}

	var payload domain.RequestUpdateTopicRoute
	if err := json.Unmarshal(body, &payload); err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "invalid request body").JSON(rw)
		return
	}

	route, err := h.uc.Webhook().UpdateTopicRoute(ctx, id, &payload)
	if err != nil {
		trace.SetError(err)
		writeError(rw, err, "topic route not found", "failed to update topic route")
		return
	}
	wrapper.NewHTTPResponse(http.StatusOK, "updated", route).JSON(rw)
}

// deleteTopicRoute handles DELETE /dashboard/topic-routes/:id.
func (h *RestHandler) deleteTopicRoute(rw http.ResponseWriter, req *http.Request) {
	trace, ctx := tracer.StartTraceWithContext(req.Context(), "WebhookDeliveryREST:DeleteTopicRoute")
	defer trace.Finish()

	id, err := uuid.Parse(restserver.URLParam(req, "id"))
	if err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "invalid id").JSON(rw)
		return
	}

	if err := h.uc.Webhook().DeleteTopicRoute(ctx, id); err != nil {
		trace.SetError(err)
		wrapper.NewHTTPResponse(http.StatusInternalServerError, "failed to delete topic route").JSON(rw)
		return
	}
	wrapper.NewHTTPResponse(http.StatusOK, "deleted").JSON(rw)
}
