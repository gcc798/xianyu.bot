-- +goose Up
ALTER TABLE tel_resource_tasks
    ADD COLUMN error_code text NOT NULL DEFAULT '',
    ADD COLUMN error_stage text NOT NULL DEFAULT '',
    ADD COLUMN failure_count integer NOT NULL DEFAULT 0,
    ADD COLUMN failure_details jsonb NOT NULL DEFAULT '[]'::jsonb;

-- +goose Down
ALTER TABLE tel_resource_tasks
    DROP COLUMN failure_details,
    DROP COLUMN failure_count,
    DROP COLUMN error_stage,
    DROP COLUMN error_code;
