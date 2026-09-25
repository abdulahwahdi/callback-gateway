package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"webhook-middleware/internal/modules/webhook/domain"
	shareddomain "webhook-middleware/pkg/shared/domain"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/google/uuid"
)

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

func (uc *webhookUsecaseImpl) Ingest(ctx context.Context, req *domain.IngestRequest) (result *shareddomain.WebhookLog, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:Ingest")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	headers := flattenHeaders(req.Headers)
	query := flattenQuery(req.Query)

	log := shareddomain.NewWebhookLog(req.Source, req.Method, req.Endpoint, req.SourceIP, headers, query, req.Body)
	log.Env = normalizeRouteField(req.Env)
	log.EventType = extractEventType(req.Body)

	if res := uc.verifiers.Verify(req.Source, req.Headers, req.Body); res.Checked {
		v := res.Verified
		log.SignatureVerified = &v
		log.SignatureHeader = res.Header
		if !v {
			logger.LogYellow(fmt.Sprintf("webhook signature verification failed: source=%s header=%s", req.Source, res.Header))
		}
	}

	// Persist first: once this succeeds, the callback is durable even if
	// Kafka is completely unreachable.
	if err := uc.repoSQL.WebhookLogRepo().Create(ctx, log); err != nil {
		return nil, fmt.Errorf("persist webhook log: %w", err)
	}

	uc.publishAndRecord(ctx, log)

	return log, nil
}

// publishAndRecord publishes the stored log to Kafka and writes the result
// (published/failed + topics + error) back onto the same row. Errors here
// are logged, not returned -- the gateway already has its 2xx.
func (uc *webhookUsecaseImpl) publishAndRecord(ctx context.Context, log *shareddomain.WebhookLog) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookUsecase:PublishAndRecord")
	defer trace.Finish()

	topics := []string{uc.resolver.CommonTopic(log.Env)}
	if sourceTopic := uc.resolver.SourceTopic(log.Env, log.Source); sourceTopic != topics[0] {
		topics = append(topics, sourceTopic)
	}

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
		uc.markPublishResult(ctx, log.ID, shareddomain.PublishStatusFailed, "marshal envelope: "+err.Error(), nil)
		return
	}

	kafkaHeaders := map[string]any{
		"source":     log.Source,
		"event_type": log.EventType,
		"log_id":     log.ID.String(),
	}

	var publishErrs []string
	for _, topic := range topics {
		if err := uc.publisher.PublishMessage(ctx, &candishared.PublisherArgument{
			Topic:   topic,
			Key:     log.Source,
			Message: payload,
			Header:  kafkaHeaders,
		}); err != nil {
			publishErrs = append(publishErrs, fmt.Sprintf("%s: %v", topic, err))
		}
	}

	if len(publishErrs) > 0 {
		logger.LogEf("failed to publish webhook event to kafka: id=%s errors=%v", log.ID, publishErrs)
		uc.markPublishResult(ctx, log.ID, shareddomain.PublishStatusFailed, strings.Join(publishErrs, "; "), topics)
		return
	}

	uc.markPublishResult(ctx, log.ID, shareddomain.PublishStatusPublished, "", topics)
}

func (uc *webhookUsecaseImpl) markPublishResult(ctx context.Context, id uuid.UUID, status shareddomain.PublishStatus, publishErr string, topics []string) {
	if err := uc.repoSQL.WebhookLogRepo().UpdatePublishResult(ctx, id, status, publishErr, topics); err != nil {
		logger.LogEf("failed to persist publish result: id=%s error=%v", id, err)
	}
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

func flattenHeaders(h map[string][]string) shareddomain.JSONMap {
	out := make(shareddomain.JSONMap, len(h))
	for k, v := range h {
		out[k] = strings.Join(v, ", ")
	}
	return out
}

func flattenQuery(q map[string][]string) shareddomain.JSONMap {
	out := make(shareddomain.JSONMap, len(q))
	for k, v := range q {
		out[k] = strings.Join(v, ", ")
	}
	return out
}
