-- +goose Up
CREATE TABLE s_telegram_api_configs (
    id bigserial PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz,
    name varchar(100) NOT NULL,
    api_id bigint NOT NULL,
    api_hash text NOT NULL,
    flood_wait_until timestamptz,
    last_flood_wait_seconds bigint NOT NULL DEFAULT 0,
    CONSTRAINT uq_s_telegram_api_configs_name UNIQUE (name),
    CONSTRAINT uq_s_telegram_api_configs_api_id UNIQUE (api_id)
);
CREATE INDEX idx_s_telegram_api_configs_deleted_at ON s_telegram_api_configs (deleted_at);

-- +goose Down
DROP TABLE s_telegram_api_configs;
