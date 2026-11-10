-- +goose Up
ALTER TABLE tel_resource_tasks
    ADD COLUMN source_selection jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN file_types_json jsonb NOT NULL DEFAULT '[]'::jsonb;
UPDATE tel_resource_tasks
SET source_selection = to_jsonb(source_dialog_ids), file_types_json = to_jsonb(file_types);

-- +goose Down
ALTER TABLE tel_resource_tasks
    DROP COLUMN file_types_json,
    DROP COLUMN source_selection;
