// Package domain holds the core entities for the webhook module: the record
// of every inbound payment-gateway callback this middleware has ever seen.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// PublishStatus tracks whether an inbound webhook has been forwarded to Kafka.
type PublishStatus string

const (
	PublishStatusPending   PublishStatus = "pending"
	PublishStatusPublished PublishStatus = "published"
	PublishStatusFailed    PublishStatus = "failed"
)

// WebhookLog is the single source of truth for "what came from where, and
// what body they sent". Every callback hitting POST /webhooks/:source is
// persisted here *before* we try to publish it to Kafka, so nothing is ever
// lost even if the broker is down.
type WebhookLog struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	// Source is the payment gateway identifier taken from the URL path,
	// e.g. "midtrans", "xendit", "doku", "gopay", "stripe".
	Source string `json:"source" gorm:"type:varchar(100);index:idx_webhook_logs_source"`
	// Env is the deployment environment from the URL path
	// (POST /webhooks/:env/:source), e.g. "staging". Empty when the
	// env-less route was used; the service's APP_ENV applies then.
	Env string `json:"env" gorm:"type:varchar(50);default:''"`
	// EventType is best-effort extracted from the payload (e.g. Xendit's
	// "event" field, Midtrans' "transaction_status"). May be empty.
	EventType string `json:"event_type" gorm:"type:varchar(150);index:idx_webhook_logs_event_type"`
	Method    string `json:"method" gorm:"type:varchar(10)"`
	Endpoint  string `json:"endpoint" gorm:"type:varchar(255)"`
	SourceIP  string `json:"source_ip" gorm:"type:varchar(64)"`

	Headers     JSONMap `json:"headers" gorm:"type:jsonb"`
	QueryParams JSONMap `json:"query_params" gorm:"type:jsonb"`
	// Body is the raw, untouched payload the gateway sent us.
	Body JSONRaw `json:"body" gorm:"type:jsonb"`

	SignatureHeader   string `json:"signature_header,omitempty" gorm:"type:varchar(100)"`
	SignatureVerified *bool  `json:"signature_verified,omitempty"`

	KafkaTopics   StringArray   `json:"kafka_topics" gorm:"type:jsonb"`
	PublishStatus PublishStatus `json:"publish_status" gorm:"type:varchar(20);index:idx_webhook_logs_publish_status"`
	PublishError  string        `json:"publish_error,omitempty" gorm:"type:text"`
	PublishedAt   *time.Time    `json:"published_at,omitempty"`
	RetryCount    int           `json:"retry_count" gorm:"default:0"`

	ResponseStatusCode int `json:"response_status_code"`

	CreatedAt time.Time `json:"created_at" gorm:"index:idx_webhook_logs_created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName pins the GORM table name explicitly.
func (WebhookLog) TableName() string { return "webhook_logs" }

// NewWebhookLog builds a pending log entry ready to be persisted.
func NewWebhookLog(source, method, endpoint, sourceIP string, headers, query JSONMap, body []byte) *WebhookLog {
	return &WebhookLog{
		ID:            uuid.New(),
		Source:        source,
		Method:        method,
		Endpoint:      endpoint,
		SourceIP:      sourceIP,
		Headers:       headers,
		QueryParams:   query,
		Body:          JSONRaw(body),
		KafkaTopics:   StringArray{},
		PublishStatus: PublishStatusPending,
	}
}
