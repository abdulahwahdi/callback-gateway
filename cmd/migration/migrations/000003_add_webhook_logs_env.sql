-- +goose Up
-- +goose StatementBegin
-- Deployment env taken from POST /webhooks/:env/:source. Empty for rows
-- ingested via the env-less POST /webhooks/:source (service APP_ENV applies).
ALTER TABLE webhook_logs ADD COLUMN IF NOT EXISTS env VARCHAR(50) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE webhook_logs DROP COLUMN IF EXISTS env;
-- +goose StatementEnd
