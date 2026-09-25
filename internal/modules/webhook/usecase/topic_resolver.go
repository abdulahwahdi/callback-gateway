package usecase

import (
	"context"
	"strings"
	"sync"

	"webhook-middleware/internal/modules/webhook/domain"

	webhookrepo "webhook-middleware/internal/modules/webhook/repository"
)

// TopicResolver decides which Kafka topic(s) an event for a given source
// should be published to for the running environment. Resolution order
// (most specific wins):
//
//  1. A DB override (kafka_topic_routes) for the exact (env, source).
//  2. A DB override for (env, "*")    -- one topic for every source in this env.
//  3. A DB override for ("*", source) -- one topic for this source in every env.
//  4. The env-var-driven template, with ":env" and ":source" substituted.
//
// DB overrides are cached in memory and refreshed both periodically (by the
// cron handler) and eagerly after every dashboard CRUD change, so the hot
// publish path never hits the database.
type TopicResolver interface {
	// CommonTopic is the single topic every event, from every source,
	// publishes to. env comes from /webhooks/:env/:source; empty means the
	// service default (APP_ENV). An empty env means the service default (APP_ENV);
	// env otherwise comes from the request path (/webhooks/:env/:source).
	CommonTopic(env string) string
	// SourceTopic is the (possibly per-source, per-env) topic a specific
	// source's events additionally publish to.
	SourceTopic(env, source string) string
	// Refresh reloads DB overrides into the in-memory cache.
	Refresh(ctx context.Context) error
}

type topicResolver struct {
	env            string
	commonTemplate string
	sourceTemplate string
	repo           webhookrepo.TopicRouteRepository

	mu     sync.RWMutex
	routes map[string]string // key: "<env>\x00<source>" -> topic
}

// NewTopicResolver builds a resolver for the given deployment environment
// (APP_ENV, e.g. "production", "staging", "development") and topic
// templates (KAFKA_COMMON_TOPIC / KAFKA_SOURCE_TOPIC_TEMPLATE), which may
// reference ":env" and, for the source template, ":source".
func NewTopicResolver(repo webhookrepo.TopicRouteRepository, env, commonTemplate, sourceTemplate string) TopicResolver {
	return &topicResolver{
		env:            normalizeRouteField(env),
		commonTemplate: commonTemplate,
		sourceTemplate: sourceTemplate,
		repo:           repo,
		routes:         map[string]string{},
	}
}

// resolveEnv returns the per-request env, or the service default when empty.
func (t *topicResolver) resolveEnv(env string) string {
	if env = normalizeRouteField(env); env != "" {
		return env
	}
	return t.env
}

func (t *topicResolver) CommonTopic(env string) string {
	return render(t.commonTemplate, t.resolveEnv(env), "")
}

func (t *topicResolver) SourceTopic(env, source string) string {
	env = t.resolveEnv(env)
	key := normalizeRouteField(source)

	t.mu.RLock()
	defer t.mu.RUnlock()

	if topic, ok := t.routes[env+"\x00"+key]; ok {
		return topic
	}
	if topic, ok := t.routes[env+"\x00"+domain.Wildcard]; ok {
		return topic
	}
	if topic, ok := t.routes[domain.Wildcard+"\x00"+key]; ok {
		return topic
	}

	return render(t.sourceTemplate, env, source)
}

// Refresh reloads every enabled DB override into the in-memory cache. On
// error the last-known-good cache keeps serving.
func (t *topicResolver) Refresh(ctx context.Context) error {
	rows, err := t.repo.ListEnabled(ctx)
	if err != nil {
		return err
	}

	routes := make(map[string]string, len(rows))
	for _, row := range rows {
		key := normalizeRouteField(row.Env) + "\x00" + normalizeRouteField(row.Source)
		routes[key] = row.Topic
	}

	t.mu.Lock()
	t.routes = routes
	t.mu.Unlock()

	return nil
}

func render(template, env, source string) string {
	out := strings.ReplaceAll(template, ":env", env)
	out = strings.ReplaceAll(out, ":source", source)
	return out
}

func normalizeRouteField(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}
