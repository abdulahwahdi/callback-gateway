// Package usecase contains the webhook module's business logic: ingest an
// inbound callback (persist + verify + publish), and serve the dashboard
// queries (list/detail/stats/replay/retry).
package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"webhook-middleware/internal/modules/webhook/domain"
	"webhook-middleware/internal/modules/webhook/repository"
)

// IngestRequest is everything the REST layer extracts from an inbound HTTP
// callback before handing it to the usecase.
type IngestRequest struct {
	Source   string
	Method   string
	Endpoint string
	SourceIP string
	Headers  http.Header
	Query    map[string][]string
	Body     []byte
}

// WebhookUsecase is the webhook module's business logic contract.
type WebhookUsecase interface {
	// Ingest persists an inbound gateway callback and forwards it to Kafka.
	// It never returns an error for a downstream publish failure -- that is
	// recorded on the log itself so the gateway still gets a fast 2xx and
	// doesn't retry-storm us. It only returns an error if we failed to even
	// persist the record (i.e. something the caller must 5xx for, so the
	// gateway retries and we don't silently drop the callback).
	Ingest(ctx context.Context, req IngestRequest) (*domain.WebhookLog, error)

	List(ctx context.Context, filter repository.Filter) ([]domain.WebhookLog, int64, error)
	Detail(ctx context.Context, id uuid.UUID) (*domain.WebhookLog, error)
	Replay(ctx context.Context, id uuid.UUID) (*domain.WebhookLog, error)
	RetryFailed(ctx context.Context, limit int) (retried int, failed int, err error)
	Sources(ctx context.Context) ([]repository.SourceCount, error)
	Stats(ctx context.Context, from, to time.Time) (repository.Stats, error)
}
