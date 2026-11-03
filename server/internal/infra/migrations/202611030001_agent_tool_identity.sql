-- +goose Up
ALTER TABLE xy_agent_run_tools
    ADD COLUMN tool_code varchar(120),
    ADD COLUMN tool_source varchar(16) NOT NULL DEFAULT 'mcp';

UPDATE xy_agent_run_tools
SET tool_code = CASE
    WHEN mcp_server_code IS NULL OR mcp_server_code = '' THEN tool_name
    ELSE mcp_server_code || '__' || tool_name
END
WHERE tool_code IS NULL;

ALTER TABLE xy_agent_run_tools ALTER COLUMN tool_code SET NOT NULL;

-- +goose Down
ALTER TABLE xy_agent_run_tools
    DROP COLUMN IF EXISTS tool_source,
    DROP COLUMN IF EXISTS tool_code;
