package usecase

import (
	"context"

	"webhook-middleware/internal/modules/webhook/domain"
	shareddomain "webhook-middleware/pkg/shared/domain"

	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/google/uuid"
)

// ReplayWebhook re-publishes an already-stored webhook's original body to
// Kafka again -- useful from the dashboard when a downstream consumer
// missed it or needs a manual redelivery.
func (uc *webhookUsecaseImpl) ReplayWebhook(ctx context.Context, id uuid.UUID) (*shareddomain.WebhookLog, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:ReplayWebhook")
	defer trace.Finish()

	repo := uc.repoSQL.WebhookLogRepo()
	log, err := repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := repo.IncrementRetry(ctx, id); err != nil {
		logger.LogEf("failed to increment retry count: id=%s error=%v", id, err)
	}
	uc.publishAndRecord(ctx, log)
	return repo.FindByID(ctx, id)
}

// RetryFailedWebhooks re-publishes every log currently in "failed" state,
// up to limit rows, skipping any that already hit maxRetries.
func (uc *webhookUsecaseImpl) RetryFailedWebhooks(ctx context.Context, limit int) (result domain.ResponseRetryFailed, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:RetryFailedWebhooks")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	if limit <= 0 {
		limit = 50
	}
	repo := uc.repoSQL.WebhookLogRepo()
	logs, err := repo.FindForRetry(ctx, []shareddomain.PublishStatus{shareddomain.PublishStatusFailed}, uc.maxRetries, limit)
	if err != nil {
		return result, err
	}

	for i := range logs {
		if err := repo.IncrementRetry(ctx, logs[i].ID); err != nil {
			logger.LogEf("failed to increment retry count: id=%s error=%v", logs[i].ID, err)
		}
		uc.publishAndRecord(ctx, &logs[i])

		refreshed, ferr := repo.FindByID(ctx, logs[i].ID)
		if ferr == nil && refreshed.PublishStatus == shareddomain.PublishStatusPublished {
			result.Retried++
		} else {
			result.Failed++
		}
	}

	return result, nil
}
