-- +goose Up
-- The product keeps only authentication, users, dictionaries, login audit and domain tables.
DROP TABLE IF EXISTS
    casbin_rule,
    exa_attachment_category,
    exa_customers,
    exa_file_chunks,
    exa_file_upload_and_downloads,
    exa_files,
    gva_announcements_info,
    sys_ai_workflow_sessions,
    sys_api_tokens,
    sys_apis,
    sys_authorities,
    sys_authority_btns,
    sys_authority_menus,
    sys_auto_code_histories,
    sys_auto_code_packages,
    sys_base_menu_btns,
    sys_base_menu_parameters,
    sys_base_menus,
    sys_data_authority_id,
    sys_error,
    sys_export_template_condition,
    sys_export_template_join,
    sys_export_templates,
    sys_ignore_apis,
    sys_operation_records,
    sys_params,
    sys_user_authority,
    sys_versions
CASCADE;

-- +goose Down
-- Removed scaffold data cannot be reconstructed safely.
SELECT 1;
