package usecase

import (
	"context"
	"strings"

	"webhook-middleware/internal/modules/webhook/domain"
	shareddomain "webhook-middleware/pkg/shared/domain"

	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/google/uuid"
)

func (uc *webhookUsecaseImpl) ListTopicRoutes(ctx context.Context) ([]shareddomain.TopicRoute, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:ListTopicRoutes")
	defer trace.Finish()

	return uc.repoSQL.TopicRouteRepo().List(ctx)
}

// CreateTopicRoute adds a new (env, source) -> topic override and refreshes
// the live resolver so the very next publish already sees it.
func (uc *webhookUsecaseImpl) CreateTopicRoute(ctx context.Context, req *domain.RequestCreateTopicRoute) (*shareddomain.TopicRoute, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:CreateTopicRoute")
	defer trace.Finish()

	env := normalizeRouteField(req.Env)
	source := normalizeRouteField(req.Source)
	topic := strings.TrimSpace(req.Topic)
	if env == "" || source == "" || topic == "" {
		return nil, domain.ErrInvalidTopicRoute
	}

	route := shareddomain.NewTopicRoute(env, source, topic)
	if err := uc.repoSQL.TopicRouteRepo().Create(ctx, route); err != nil {
		return nil, err
	}

	uc.refreshResolver(ctx)
	return route, nil
}

func (uc *webhookUsecaseImpl) UpdateTopicRoute(ctx context.Context, id uuid.UUID, req *domain.RequestUpdateTopicRoute) (*shareddomain.TopicRoute, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:UpdateTopicRoute")
	defer trace.Finish()

	if req.Env != nil {
		v := normalizeRouteField(*req.Env)
		req.Env = &v
	}
	if req.Source != nil {
		v := normalizeRouteField(*req.Source)
		req.Source = &v
	}
	if req.Topic != nil {
		v := strings.TrimSpace(*req.Topic)
		req.Topic = &v
	}

	route, err := uc.repoSQL.TopicRouteRepo().Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	uc.refreshResolver(ctx)
	return route, nil
}

func (uc *webhookUsecaseImpl) DeleteTopicRoute(ctx context.Context, id uuid.UUID) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:DeleteTopicRoute")
	defer trace.Finish()

	if err := uc.repoSQL.TopicRouteRepo().Delete(ctx, id); err != nil {
		return err
	}
	uc.refreshResolver(ctx)
	return nil
}

func (uc *webhookUsecaseImpl) RefreshTopicRoutes(ctx context.Context) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:RefreshTopicRoutes")
	defer trace.Finish()

	return uc.resolver.Refresh(ctx)
}

func (uc *webhookUsecaseImpl) refreshResolver(ctx context.Context) {
	if err := uc.resolver.Refresh(ctx); err != nil {
		logger.LogYellow("failed to refresh topic routes after change: " + err.Error())
	}
}
