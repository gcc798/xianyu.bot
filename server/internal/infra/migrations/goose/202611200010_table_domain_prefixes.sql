-- +goose Up
ALTER TABLE xy_agent_config RENAME TO agent_config;
ALTER TABLE xy_agent_events RENAME TO agent_events;
ALTER TABLE xy_agent_messages RENAME TO agent_messages;
ALTER TABLE xy_agent_run_tools RENAME TO agent_run_tools;
ALTER TABLE xy_agent_runs RENAME TO agent_runs;
ALTER TABLE xy_agent_sessions RENAME TO agent_sessions;
ALTER TABLE xy_ai_models RENAME TO agent_models;

ALTER TABLE xy_mcp_servers RENAME TO mcp_servers;
ALTER TABLE xy_mcp_tools RENAME TO mcp_tools;

ALTER TABLE xy_resource_candidates RENAME TO resource_candidates;
ALTER TABLE xy_resource_searches RENAME TO resource_searches;
ALTER TABLE xy_resource_sources RENAME TO resource_sources;
ALTER TABLE xy_resource_tasks RENAME TO resource_tasks;
ALTER TABLE xy_resources RENAME TO resources;

ALTER TABLE s_telegram_api_configs RENAME TO tg_api_configs;
ALTER TABLE tel_acquisition_tasks RENAME TO tg_acquisition_tasks;
ALTER TABLE tel_dialogs RENAME TO tg_dialogs;
ALTER TABLE tel_leave_history RENAME TO tg_leave_history;
ALTER TABLE tel_resource_files RENAME TO tg_resource_files;
ALTER TABLE tel_resource_sources RENAME TO tg_resource_sources;
ALTER TABLE tel_resource_tasks RENAME TO tg_resource_tasks;
ALTER TABLE tel_scan_history RENAME TO tg_scan_history;
ALTER TABLE tel_settings RENAME TO tg_settings;

-- +goose Down
ALTER TABLE tg_settings RENAME TO tel_settings;
ALTER TABLE tg_scan_history RENAME TO tel_scan_history;
ALTER TABLE tg_resource_tasks RENAME TO tel_resource_tasks;
ALTER TABLE tg_resource_sources RENAME TO tel_resource_sources;
ALTER TABLE tg_resource_files RENAME TO tel_resource_files;
ALTER TABLE tg_leave_history RENAME TO tel_leave_history;
ALTER TABLE tg_dialogs RENAME TO tel_dialogs;
ALTER TABLE tg_acquisition_tasks RENAME TO tel_acquisition_tasks;
ALTER TABLE tg_api_configs RENAME TO s_telegram_api_configs;

ALTER TABLE resources RENAME TO xy_resources;
ALTER TABLE resource_tasks RENAME TO xy_resource_tasks;
ALTER TABLE resource_sources RENAME TO xy_resource_sources;
ALTER TABLE resource_searches RENAME TO xy_resource_searches;
ALTER TABLE resource_candidates RENAME TO xy_resource_candidates;

ALTER TABLE mcp_tools RENAME TO xy_mcp_tools;
ALTER TABLE mcp_servers RENAME TO xy_mcp_servers;

ALTER TABLE agent_models RENAME TO xy_ai_models;
ALTER TABLE agent_sessions RENAME TO xy_agent_sessions;
ALTER TABLE agent_runs RENAME TO xy_agent_runs;
ALTER TABLE agent_run_tools RENAME TO xy_agent_run_tools;
ALTER TABLE agent_messages RENAME TO xy_agent_messages;
ALTER TABLE agent_events RENAME TO xy_agent_events;
ALTER TABLE agent_config RENAME TO xy_agent_config;
