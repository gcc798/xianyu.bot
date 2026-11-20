-- +goose Up
ALTER TABLE public.xy_agent_run_tools
    ADD COLUMN catalog_hash character varying(64);

ALTER TABLE public.xy_mcp_servers
    ADD COLUMN description character varying(1000),
    ADD COLUMN tags text,
    ADD COLUMN enabled_tools text,
    ADD COLUMN disabled_tools text,
    ADD COLUMN catalog_ttl_seconds bigint NOT NULL DEFAULT 3600,
    ADD COLUMN connection_ttl_seconds bigint NOT NULL DEFAULT 600,
    ADD COLUMN connection_scope character varying(16) NOT NULL DEFAULT 'shared',
    ADD COLUMN max_concurrent_calls bigint NOT NULL DEFAULT 0,
    ADD COLUMN server_instructions text,
    ADD COLUMN catalog_status character varying(32) NOT NULL DEFAULT 'unknown',
    ADD COLUMN catalog_tool_count bigint NOT NULL DEFAULT 0,
    ADD COLUMN catalog_refreshed_at timestamp with time zone,
    ADD COLUMN catalog_error character varying(1000),
    ADD COLUMN last_check_duration_millis bigint NOT NULL DEFAULT 0;

CREATE TABLE public.xy_mcp_tools (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    mcp_server_code character varying(64) NOT NULL,
    name character varying(160) NOT NULL,
    description text,
    input_schema text NOT NULL,
    output_schema text,
    annotations text,
    definition text NOT NULL,
    search_text text NOT NULL,
    definition_hash character varying(64) NOT NULL,
    discovered_at timestamp with time zone NOT NULL
);

CREATE UNIQUE INDEX idx_xy_mcp_tools_code ON public.xy_mcp_tools (mcp_server_code, name);
CREATE INDEX idx_xy_mcp_tools_deleted_at ON public.xy_mcp_tools (deleted_at);
CREATE INDEX idx_xy_mcp_tools_server ON public.xy_mcp_tools (mcp_server_code);

-- +goose Down
DROP TABLE public.xy_mcp_tools;

ALTER TABLE public.xy_mcp_servers
    DROP COLUMN description,
    DROP COLUMN tags,
    DROP COLUMN enabled_tools,
    DROP COLUMN disabled_tools,
    DROP COLUMN catalog_ttl_seconds,
    DROP COLUMN connection_ttl_seconds,
    DROP COLUMN connection_scope,
    DROP COLUMN max_concurrent_calls,
    DROP COLUMN server_instructions,
    DROP COLUMN catalog_status,
    DROP COLUMN catalog_tool_count,
    DROP COLUMN catalog_refreshed_at,
    DROP COLUMN catalog_error,
    DROP COLUMN last_check_duration_millis;

ALTER TABLE public.xy_agent_run_tools
    DROP COLUMN catalog_hash;
