package usecase

import (
	"context"
	"time"

	"webhook-middleware/internal/modules/webhook/domain"
	shareddomain "webhook-middleware/pkg/shared/domain"

	"github.com/golangid/candi/tracer"
	"github.com/google/uuid"
)

func (uc *webhookUsecaseImpl) ListWebhook(ctx context.Context, filter *domain.FilterWebhookLog) ([]shareddomain.WebhookLog, int64, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:ListWebhook")
	defer trace.Finish()

	return uc.repoSQL.WebhookLogRepo().Find(ctx, filter)
}

func (uc *webhookUsecaseImpl) GetDetailWebhook(ctx context.Context, id uuid.UUID) (*shareddomain.WebhookLog, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:GetDetailWebhook")
	defer trace.Finish()

	return uc.repoSQL.WebhookLogRepo().FindByID(ctx, id)
}

func (uc *webhookUsecaseImpl) GetSources(ctx context.Context) ([]domain.SourceCount, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:GetSources")
	defer trace.Finish()

	return uc.repoSQL.WebhookLogRepo().Sources(ctx)
}

func (uc *webhookUsecaseImpl) GetStats(ctx context.Context, from, to time.Time) (domain.Stats, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:GetStats")
	defer trace.Finish()

	return uc.repoSQL.WebhookLogRepo().Stats(ctx, from, to)
}
