package usecase

import (
	"context"
	"time"

	"webhook-middleware/internal/modules/webhook/domain"
	"webhook-middleware/pkg/shared"
	shareddomain "webhook-middleware/pkg/shared/domain"
	"webhook-middleware/pkg/shared/repository"
	"webhook-middleware/pkg/shared/usecase/common"
	"webhook-middleware/pkg/verifier"

	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/logger"
	"github.com/google/uuid"
)

// WebhookUsecase abstraction
type WebhookUsecase interface {
	// Ingest persists an inbound gateway callback and forwards it to Kafka.
	// It never returns an error for a downstream publish failure -- that is
	// recorded on the log itself so the gateway still gets a fast 2xx and
	// doesn't retry-storm us. It only returns an error if we failed to even
	// persist the record (i.e. something the caller must 5xx for, so the
	// gateway retries and we don't silently drop the callback).
	Ingest(ctx context.Context, req *domain.IngestRequest) (*shareddomain.WebhookLog, error)

	ListWebhook(ctx context.Context, filter *domain.FilterWebhookLog) ([]shareddomain.WebhookLog, int64, error)
	GetDetailWebhook(ctx context.Context, id uuid.UUID) (*shareddomain.WebhookLog, error)
	ReplayWebhook(ctx context.Context, id uuid.UUID) (*shareddomain.WebhookLog, error)
	RetryFailedWebhooks(ctx context.Context, limit int) (result domain.ResponseRetryFailed, err error)
	GetSources(ctx context.Context) ([]domain.SourceCount, error)
	GetStats(ctx context.Context, from, to time.Time) (domain.Stats, error)

	// ListTopicRoutes returns every DB-configured (env, source) -> topic
	// override, enabled or not, for the dashboard's management UI.
	ListTopicRoutes(ctx context.Context) ([]shareddomain.TopicRoute, error)
	// CreateTopicRoute adds a new override and immediately refreshes the
	// live resolver so it takes effect without a restart.
	CreateTopicRoute(ctx context.Context, req *domain.RequestCreateTopicRoute) (*shareddomain.TopicRoute, error)
	UpdateTopicRoute(ctx context.Context, id uuid.UUID, req *domain.RequestUpdateTopicRoute) (*shareddomain.TopicRoute, error)
	DeleteTopicRoute(ctx context.Context, id uuid.UUID) error
	// RefreshTopicRoutes reloads DB overrides into the resolver's in-memory
	// cache (called on a schedule by the cron handler).
	RefreshTopicRoutes(ctx context.Context) error
}

type webhookUsecaseImpl struct {
	deps          dependency.Dependency
	sharedUsecase common.Usecase
	repoSQL       repository.RepoSQL
	publisher     interfaces.Publisher
	verifiers     *verifier.Registry
	resolver      TopicResolver
	maxRetries    int
}

// NewWebhookUsecase usecase impl constructor
func NewWebhookUsecase(deps dependency.Dependency) (WebhookUsecase, func(sharedUsecase common.Usecase)) {
	env := shared.GetEnv()
	repoSQL := repository.GetSharedRepoSQL()

	resolver := NewTopicResolver(repoSQL.TopicRouteRepo(), env.AppEnv, env.KafkaCommonTopic, env.KafkaSourceTopicTemplate)

	uc := &webhookUsecaseImpl{
		deps:       deps,
		repoSQL:    repoSQL,
		publisher:  deps.GetBroker(types.Kafka).GetPublisher(),
		verifiers:  verifier.NewRegistryFromEnv(),
		resolver:   resolver,
		maxRetries: env.MaxPublishRetries,
	}

	// Warm the cache so the very first publish already sees DB overrides;
	// afterwards the cron handler keeps it fresh.
	if err := resolver.Refresh(context.Background()); err != nil {
		logger.LogYellow("failed initial load of kafka topic route overrides, using template only: " + err.Error())
	}

	return uc, func(sharedUsecase common.Usecase) {
		uc.sharedUsecase = sharedUsecase
	}
}
