-- +goose Up
ALTER TABLE tel_resource_tasks
    ADD COLUMN updated_at timestamp with time zone NOT NULL DEFAULT now();

-- +goose Down
ALTER TABLE tel_resource_tasks
    DROP COLUMN updated_at;
