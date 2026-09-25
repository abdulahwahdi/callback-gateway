package repository

import (
	"context"
	"time"

	"webhook-middleware/internal/modules/webhook/domain"
	shareddomain "webhook-middleware/pkg/shared/domain"

	"github.com/google/uuid"
)

// WebhookLogRepository persists and queries webhook callback records.
type WebhookLogRepository interface {
	Create(ctx context.Context, log *shareddomain.WebhookLog) error
	UpdatePublishResult(ctx context.Context, id uuid.UUID, status shareddomain.PublishStatus, publishErr string, topics []string) error
	IncrementRetry(ctx context.Context, id uuid.UUID) error

	FindByID(ctx context.Context, id uuid.UUID) (*shareddomain.WebhookLog, error)
	Find(ctx context.Context, filter *domain.FilterWebhookLog) ([]shareddomain.WebhookLog, int64, error)
	FindForRetry(ctx context.Context, statuses []shareddomain.PublishStatus, maxRetry, limit int) ([]shareddomain.WebhookLog, error)

	Sources(ctx context.Context) ([]domain.SourceCount, error)
	Stats(ctx context.Context, from, to time.Time) (domain.Stats, error)
}

// TopicRouteRepository persists DB-configurable overrides of which Kafka
// topic a given (env, source) pair should publish to.
type TopicRouteRepository interface {
	Create(ctx context.Context, route *shareddomain.TopicRoute) error
	Update(ctx context.Context, id uuid.UUID, in *domain.RequestUpdateTopicRoute) (*shareddomain.TopicRoute, error)
	Delete(ctx context.Context, id uuid.UUID) error

	FindByID(ctx context.Context, id uuid.UUID) (*shareddomain.TopicRoute, error)
	// List returns every route, enabled or not, for the dashboard's management UI.
	List(ctx context.Context) ([]shareddomain.TopicRoute, error)
	// ListEnabled returns only enabled routes -- what TopicResolver caches
	// for the hot publish path.
	ListEnabled(ctx context.Context) ([]shareddomain.TopicRoute, error)
}
