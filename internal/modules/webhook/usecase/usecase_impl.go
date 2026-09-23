package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"webhook-middleware/internal/modules/webhook/domain"
	"webhook-middleware/internal/modules/webhook/repository"
	"webhook-middleware/internal/pkg/broker"
	"webhook-middleware/internal/pkg/logger"
	"webhook-middleware/internal/pkg/verifier"
)

// KafkaConfig controls how ingested events are published.
type KafkaConfig struct {
	// CommonTopic receives every event, from every source, with the source
	// name as the message key -- so any downstream service can subscribe
	// once and filter, or use it for a generic audit stream.
	CommonTopic string
	// TopicPrefix produces a per-source topic ("<prefix>.<source>") so a
	// service that only cares about e.g. "midtrans" can subscribe narrowly
	// without consuming everything.
	TopicPrefix string
}

func (c KafkaConfig) sourceTopic(source string) string {
	return fmt.Sprintf("%s.%s", c.TopicPrefix, source)
}

type webhookUsecase struct {
	repo       repository.WebhookLogRepository
	publisher  broker.Publisher
	verifiers  *verifier.Registry
	kafkaCfg   KafkaConfig
	maxRetries int
}

// New builds the webhook module's usecase implementation.
func New(repo repository.WebhookLogRepository, publisher broker.Publisher, verifiers *verifier.Registry, kafkaCfg KafkaConfig, maxRetries int) WebhookUsecase {
	if maxRetries <= 0 {
		maxRetries = 5
	}
	return &webhookUsecase{
		repo:       repo,
		publisher:  publisher,
		verifiers:  verifiers,
		kafkaCfg:   kafkaCfg,
		maxRetries: maxRetries,
	}
}

// eventEnvelope is the message shape every downstream service receives from
// Kafka -- it wraps the gateway's raw body with routing/tracing metadata.
type eventEnvelope struct {
	ID                string          `json:"id"`
	Source            string          `json:"source"`
	EventType         string          `json:"event_type,omitempty"`
	ReceivedAt        time.Time       `json:"received_at"`
	Endpoint          string          `json:"endpoint"`
	SignatureVerified *bool           `json:"signature_verified,omitempty"`
	Headers           map[string]any  `json:"headers"`
	Body              json.RawMessage `json:"body"`
}

func (u *webhookUsecase) Ingest(ctx context.Context, req IngestRequest) (*domain.WebhookLog, error) {
	headers := flattenHeaders(req.Headers)
	query := flattenQuery(req.Query)

	log := domain.NewWebhookLog(req.Source, req.Method, req.Endpoint, req.SourceIP, headers, query, req.Body)
	log.EventType = extractEventType(req.Body)

	if res := u.verifiers.Verify(req.Source, req.Headers, req.Body); res.Checked {
		v := res.Verified
		log.SignatureVerified = &v
		log.SignatureHeader = res.Header
		if !v {
			logger.Warn("webhook signature verification failed",
				"source", req.Source, "header", res.Header)
		}
	}

	// Persist first: once this succeeds, the callback is durable even if
	// Kafka is completely unreachable. This is the "save to database too
	// what coming from and what body they send" requirement.
	if err := u.repo.Create(ctx, log); err != nil {
		return nil, fmt.Errorf("persist webhook log: %w", err)
	}

	u.publishAndRecord(ctx, log)

	return log, nil
}

// publishAndRecord publishes the stored log to Kafka and writes the result
// (published/failed + topics + error) back onto the same row. Errors here
// are logged, not returned -- the gateway already has its 2xx.
func (u *webhookUsecase) publishAndRecord(ctx context.Context, log *domain.WebhookLog) {
	topics := []string{u.kafkaCfg.CommonTopic, u.kafkaCfg.sourceTopic(log.Source)}

	envelope := eventEnvelope{
		ID:                log.ID.String(),
		Source:            log.Source,
		EventType:         log.EventType,
		ReceivedAt:        log.CreatedAt,
		Endpoint:          log.Endpoint,
		SignatureVerified: log.SignatureVerified,
		Headers:           log.Headers,
		Body:              json.RawMessage(log.Body),
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		u.markPublishResult(ctx, log.ID, domain.PublishStatusFailed, "marshal envelope: "+err.Error(), nil)
		return
	}

	kafkaHeaders := map[string]string{
		"source":     log.Source,
		"event_type": log.EventType,
		"log_id":     log.ID.String(),
	}

	var publishErrs []string
	for _, topic := range topics {
		if err := u.publisher.Publish(ctx, topic, log.Source, payload, kafkaHeaders); err != nil {
			publishErrs = append(publishErrs, fmt.Sprintf("%s: %v", topic, err))
		}
	}

	if len(publishErrs) > 0 {
		logger.Error("failed to publish webhook event to kafka", "id", log.ID, "errors", publishErrs)
		u.markPublishResult(ctx, log.ID, domain.PublishStatusFailed, strings.Join(publishErrs, "; "), topics)
		return
	}

	u.markPublishResult(ctx, log.ID, domain.PublishStatusPublished, "", topics)
}

func (u *webhookUsecase) markPublishResult(ctx context.Context, id uuid.UUID, status domain.PublishStatus, publishErr string, topics []string) {
	if err := u.repo.UpdatePublishResult(ctx, id, status, publishErr, topics); err != nil {
		logger.Error("failed to persist publish result", "id", id, "error", err)
	}
}

func (u *webhookUsecase) List(ctx context.Context, filter repository.Filter) ([]domain.WebhookLog, int64, error) {
	return u.repo.Find(ctx, filter)
}

func (u *webhookUsecase) Detail(ctx context.Context, id uuid.UUID) (*domain.WebhookLog, error) {
	return u.repo.FindByID(ctx, id)
}

// Replay re-publishes an already-stored webhook's original body to Kafka
// again -- useful from the dashboard when a downstream consumer missed it
// or needs a manual redelivery.
func (u *webhookUsecase) Replay(ctx context.Context, id uuid.UUID) (*domain.WebhookLog, error) {
	log, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.repo.IncrementRetry(ctx, id); err != nil {
		logger.Error("failed to increment retry count", "id", id, "error", err)
	}
	u.publishAndRecord(ctx, log)
	return u.repo.FindByID(ctx, id)
}

// RetryFailed re-publishes every log currently in "failed" state, up to
// limit rows, skipping any that already hit maxRetries.
func (u *webhookUsecase) RetryFailed(ctx context.Context, limit int) (retried int, failed int, err error) {
	if limit <= 0 {
		limit = 50
	}
	logs, err := u.repo.FindForRetry(ctx, []domain.PublishStatus{domain.PublishStatusFailed}, u.maxRetries, limit)
	if err != nil {
		return 0, 0, err
	}

	for i := range logs {
		if err := u.repo.IncrementRetry(ctx, logs[i].ID); err != nil {
			logger.Error("failed to increment retry count", "id", logs[i].ID, "error", err)
		}
		u.publishAndRecord(ctx, &logs[i])

		refreshed, ferr := u.repo.FindByID(ctx, logs[i].ID)
		if ferr == nil && refreshed.PublishStatus == domain.PublishStatusPublished {
			retried++
		} else {
			failed++
		}
	}

	return retried, failed, nil
}

func (u *webhookUsecase) Sources(ctx context.Context) ([]repository.SourceCount, error) {
	return u.repo.Sources(ctx)
}

func (u *webhookUsecase) Stats(ctx context.Context, from, to time.Time) (repository.Stats, error) {
	return u.repo.Stats(ctx, from, to)
}

// extractEventType makes a best-effort guess at the gateway's event/status
// field so the dashboard can filter/group by it, without knowing about
// every gateway's schema up front.
func extractEventType(body []byte) string {
	var generic map[string]any
	if err := json.Unmarshal(body, &generic); err != nil {
		return ""
	}
	for _, key := range []string{"event", "event_type", "eventType", "type", "transaction_status", "status"} {
		if v, ok := generic[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func flattenHeaders(h map[string][]string) domain.JSONMap {
	out := make(domain.JSONMap, len(h))
	for k, v := range h {
		out[k] = strings.Join(v, ", ")
	}
	return out
}

func flattenQuery(q map[string][]string) domain.JSONMap {
	out := make(domain.JSONMap, len(q))
	for k, v := range q {
		out[k] = strings.Join(v, ", ")
	}
	return out
}
