-- +goose Up
ALTER TABLE sys_users DROP COLUMN IF EXISTS origin_setting;

-- +goose Down
ALTER TABLE sys_users ADD COLUMN IF NOT EXISTS origin_setting text;
