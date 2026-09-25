-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS webhook_logs (
    id                   UUID PRIMARY KEY,
    source               VARCHAR(100) NOT NULL,
    event_type           VARCHAR(150) NOT NULL DEFAULT '',
    method               VARCHAR(10)  NOT NULL DEFAULT '',
    endpoint             VARCHAR(255) NOT NULL DEFAULT '',
    source_ip            VARCHAR(64)  NOT NULL DEFAULT '',
    headers              JSONB        NOT NULL DEFAULT '{}',
    query_params         JSONB        NOT NULL DEFAULT '{}',
    body                 JSONB,
    signature_header     VARCHAR(100) NOT NULL DEFAULT '',
    signature_verified   BOOLEAN,
    kafka_topics         JSONB        NOT NULL DEFAULT '[]',
    publish_status       VARCHAR(20)  NOT NULL DEFAULT 'pending',
    publish_error        TEXT         NOT NULL DEFAULT '',
    published_at         TIMESTAMPTZ,
    retry_count          INTEGER      NOT NULL DEFAULT 0,
    response_status_code INTEGER      NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webhook_logs_source         ON webhook_logs (source);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_event_type     ON webhook_logs (event_type);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_publish_status ON webhook_logs (publish_status);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_created_at     ON webhook_logs (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS webhook_logs;
-- +goose StatementEnd
