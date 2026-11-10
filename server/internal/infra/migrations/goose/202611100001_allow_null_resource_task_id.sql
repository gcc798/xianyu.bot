-- +goose Up
ALTER TABLE xy_resources ALTER COLUMN task_id DROP NOT NULL;

-- +goose Down
ALTER TABLE xy_resources ALTER COLUMN task_id SET NOT NULL;
