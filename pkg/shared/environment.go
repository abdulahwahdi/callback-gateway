package shared

import (
	"time"

	"github.com/golangid/candi/config/env"
)

// Environment additional in this service (on top of candi's base environment)
type Environment struct {
	// AppEnv is the running deployment environment (e.g. "production",
	// "staging", "development"). Substituted for ":env" in the Kafka topic
	// templates below, and matched against kafka_topic_routes DB overrides.
	// Defaults to candi's ENVIRONMENT value.
	AppEnv string `env:"APP_ENV" optional:"true"`

	// KafkaCommonTopic is the template for the single topic every event,
	// from every source, publishes to. May reference ":env".
	KafkaCommonTopic string `env:"KAFKA_COMMON_TOPIC" optional:"true"`
	// KafkaTopicPrefix builds the default KafkaSourceTopicTemplate.
	KafkaTopicPrefix string `env:"KAFKA_TOPIC_PREFIX" optional:"true"`
	// KafkaSourceTopicTemplate is the per-source topic template. May
	// reference ":env" and ":source", e.g. "webhook_gateway.:env.:source".
	// Defaults to "<KAFKA_TOPIC_PREFIX>.:source" when unset.
	KafkaSourceTopicTemplate string `env:"KAFKA_SOURCE_TOPIC_TEMPLATE" optional:"true"`
	// MaxPublishRetries caps how often a failed publish is retried by the
	// cron retry job before giving up on it (still replayable from the dashboard).
	MaxPublishRetries int `env:"MAX_PUBLISH_RETRIES" optional:"true"`

	// TopicRouteRefreshInterval controls how often DB overrides
	// (kafka_topic_routes) are reloaded in the background.
	TopicRouteRefreshInterval time.Duration `env:"KAFKA_TOPIC_ROUTE_REFRESH_INTERVAL" optional:"true"`

	// RetryWorkerInterval / RetryWorkerBatch tune the cron job that retries
	// events stuck in "failed" publish status.
	RetryWorkerInterval time.Duration `env:"RETRY_WORKER_INTERVAL" optional:"true"`
	RetryWorkerBatch    int           `env:"RETRY_WORKER_BATCH" optional:"true"`

	// DashboardAPIKey, if set, is required (via X-API-Key header) to call
	// any /dashboard/* endpoint. Left empty in local/dev by default.
	DashboardAPIKey string `env:"DASHBOARD_API_KEY" optional:"true"`

	// DashboardUsername / DashboardPassword enable the dashboard login page.
	// When both are set, /dashboard/* accepts a bearer token issued by
	// POST /auth/login (in addition to X-API-Key, if configured).
	DashboardUsername string `env:"DASHBOARD_USERNAME" optional:"true"`
	DashboardPassword string `env:"DASHBOARD_PASSWORD" optional:"true"`
	// DashboardAuthSecret signs session tokens. If empty, a random secret is
	// generated at startup (sessions then end whenever the service restarts).
	DashboardAuthSecret string `env:"DASHBOARD_AUTH_SECRET" optional:"true"`
	// DashboardSessionTTL is how long a login session stays valid. Default 24h.
	DashboardSessionTTL time.Duration `env:"DASHBOARD_SESSION_TTL" optional:"true"`
}

var sharedEnv Environment

// GetEnv get global additional environment
func GetEnv() Environment {
	return sharedEnv
}

// SetEnv set global additional environment, applying defaults for anything unset
func SetEnv(e Environment) {
	if e.AppEnv == "" {
		e.AppEnv = env.BaseEnv().Environment
	}
	if e.AppEnv == "" {
		e.AppEnv = "development"
	}
	if e.KafkaCommonTopic == "" {
		e.KafkaCommonTopic = "webhook_gateway_events"
	}
	if e.KafkaTopicPrefix == "" {
		e.KafkaTopicPrefix = "webhook_gateway"
	}
	if e.KafkaSourceTopicTemplate == "" {
		e.KafkaSourceTopicTemplate = e.KafkaTopicPrefix + ".:source"
	}
	if e.MaxPublishRetries <= 0 {
		e.MaxPublishRetries = 5
	}
	if e.TopicRouteRefreshInterval <= 0 {
		e.TopicRouteRefreshInterval = 30 * time.Second
	}
	if e.RetryWorkerInterval <= 0 {
		e.RetryWorkerInterval = time.Minute
	}
	if e.RetryWorkerBatch <= 0 {
		e.RetryWorkerBatch = 50
	}
	if e.DashboardSessionTTL <= 0 {
		e.DashboardSessionTTL = 24 * time.Hour
	}
	sharedEnv = e
}
