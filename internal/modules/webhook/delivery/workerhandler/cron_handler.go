package workerhandler

import (
	"fmt"

	"webhook-middleware/pkg/shared"
	"webhook-middleware/pkg/shared/usecase"

	"github.com/golangid/candi/candishared"
	cronworker "github.com/golangid/candi/codebase/app/cron_worker"
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

// CronHandler struct
type CronHandler struct {
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewCronHandler constructor
func NewCronHandler(uc usecase.Usecase, deps dependency.Dependency) *CronHandler {
	return &CronHandler{
		uc:        uc,
		validator: deps.GetValidator(),
	}
}

// MountHandlers mount handler group
func (h *CronHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	env := shared.GetEnv()

	// Retry publishing any webhook stuck in "failed" state, so a temporary
	// Kafka outage doesn't lose events -- they were already durably saved
	// to the database on ingest.
	group.Add(cronworker.CreateCronJobKey("webhook-retry-failed", "", env.RetryWorkerInterval.String()), h.retryFailedWebhooks)

	// Keep the in-memory kafka_topic_routes cache fresh across replicas.
	group.Add(cronworker.CreateCronJobKey("topic-route-refresh", "", env.TopicRouteRefreshInterval.String()), h.refreshTopicRoutes)
}

func (h *CronHandler) retryFailedWebhooks(eventContext *candishared.EventContext) error {
	trace, ctx := tracer.StartTraceWithContext(eventContext.Context(), "WebhookDeliveryCron:RetryFailedWebhooks")
	defer trace.Finish()

	result, err := h.uc.Webhook().RetryFailedWebhooks(ctx, shared.GetEnv().RetryWorkerBatch)
	if err != nil {
		trace.SetError(err)
		return fmt.Errorf("retry worker sweep failed: %w", err)
	}
	if result.Retried > 0 || result.Failed > 0 {
		logger.LogIf("retry worker sweep completed: retried=%d still_failed=%d", result.Retried, result.Failed)
	}
	return nil
}

func (h *CronHandler) refreshTopicRoutes(eventContext *candishared.EventContext) error {
	trace, ctx := tracer.StartTraceWithContext(eventContext.Context(), "WebhookDeliveryCron:RefreshTopicRoutes")
	defer trace.Finish()

	if err := h.uc.Webhook().RefreshTopicRoutes(ctx); err != nil {
		trace.SetError(err)
		return fmt.Errorf("failed to refresh kafka topic routes: %w", err)
	}
	return nil
}
