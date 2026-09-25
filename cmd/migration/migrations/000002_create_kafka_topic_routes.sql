-- +goose Up
-- +goose StatementBegin
-- Per (environment, source) Kafka topic overrides. Env or source may be
-- '*' as a wildcard: ('*', 'midtrans') applies to 'midtrans' in every
-- environment; ('production', '*') applies to every source in
-- 'production'. When no row matches, the service falls back to its
-- env-var-driven topic template (KAFKA_SOURCE_TOPIC_TEMPLATE).
CREATE TABLE IF NOT EXISTS kafka_topic_routes (
    id         UUID PRIMARY KEY,
    env        VARCHAR(50)  NOT NULL,
    source     VARCHAR(100) NOT NULL,
    topic      VARCHAR(255) NOT NULL,
    enabled    BOOLEAN      NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_topic_routes_env_source ON kafka_topic_routes (env, source);
CREATE INDEX IF NOT EXISTS idx_topic_routes_enabled ON kafka_topic_routes (enabled);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS kafka_topic_routes;
-- +goose StatementEnd
