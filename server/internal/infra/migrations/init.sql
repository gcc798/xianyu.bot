-- xianyu.bot PostgreSQL 19 空库初始化脚本。

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: jwt_blacklists; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.jwt_blacklists (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    jwt text
);


--
-- Name: jwt_blacklists_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.jwt_blacklists_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: jwt_blacklists_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.jwt_blacklists_id_seq OWNED BY public.jwt_blacklists.id;


--
-- Name: s_dictionaries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.s_dictionaries (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    type text,
    status boolean,
    "desc" text,
    parent_id bigint
);


--
-- Name: s_dictionaries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.s_dictionaries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: s_dictionaries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.s_dictionaries_id_seq OWNED BY public.s_dictionaries.id;


--
-- Name: s_dictionary_details; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.s_dictionary_details (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    label text,
    value text,
    extend text,
    status boolean,
    sort bigint,
    sys_dictionary_id bigint,
    parent_id bigint,
    level bigint,
    path text
);


--
-- Name: s_dictionary_details_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.s_dictionary_details_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: s_dictionary_details_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.s_dictionary_details_id_seq OWNED BY public.s_dictionary_details.id;


--
-- Name: s_login_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.s_login_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    username text,
    ip text,
    status boolean,
    error_message text,
    agent text,
    user_id bigint
);


--
-- Name: s_login_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.s_login_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: s_login_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.s_login_logs_id_seq OWNED BY public.s_login_logs.id;


--
-- Name: s_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.s_users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    uuid text,
    username text,
    password text,
    nick_name text DEFAULT '系统用户'::text,
    header_img text DEFAULT ''::text,
    authority_id bigint DEFAULT 888,
    phone text,
    email text,
    enable bigint DEFAULT 1
);


--
-- Name: s_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.s_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: s_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.s_users_id_seq OWNED BY public.s_users.id;


--
-- Name: tel_acquisition_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tel_acquisition_tasks (
    id bigint NOT NULL,
    resource_task_id bigint NOT NULL,
    candidate_id text NOT NULL,
    idempotency_key text NOT NULL,
    state text DEFAULT 'queued'::text NOT NULL,
    stage text,
    bytes_completed bigint DEFAULT 0 NOT NULL,
    bytes_total bigint DEFAULT 0 NOT NULL,
    result_json text,
    error_code text,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    CONSTRAINT tel_acquisition_tasks_state_check CHECK ((state = ANY (ARRAY['queued'::text, 'running'::text, 'completed'::text, 'failed'::text, 'cancelled'::text])))
);


--
-- Name: tel_acquisition_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tel_acquisition_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_acquisition_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tel_acquisition_tasks_id_seq OWNED BY public.tel_acquisition_tasks.id;


--
-- Name: tel_dialogs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tel_dialogs (
    id bigint NOT NULL,
    telegram_id bigint NOT NULL,
    access_hash bigint,
    name text NOT NULL,
    username text,
    dialog_type text NOT NULL,
    member_count integer,
    is_channel boolean DEFAULT false NOT NULL,
    is_group boolean DEFAULT false NOT NULL,
    is_megagroup boolean DEFAULT false NOT NULL,
    is_broadcast boolean DEFAULT false NOT NULL,
    is_archived boolean DEFAULT false NOT NULL,
    is_muted boolean DEFAULT false NOT NULL,
    unread_count integer DEFAULT 0 NOT NULL,
    last_message_id bigint,
    last_message_at timestamp with time zone,
    last_read_message_id bigint,
    status text DEFAULT 'normal'::text NOT NULL,
    is_creator boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_compliance_status text DEFAULT 'allowed'::text NOT NULL
);


--
-- Name: tel_dialogs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tel_dialogs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_dialogs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tel_dialogs_id_seq OWNED BY public.tel_dialogs.id;


--
-- Name: tel_leave_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tel_leave_history (
    id bigint NOT NULL,
    dialog_id bigint,
    telegram_id bigint NOT NULL,
    name text NOT NULL,
    left_at timestamp with time zone NOT NULL,
    result text NOT NULL,
    error_message text
);


--
-- Name: tel_leave_history_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tel_leave_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_leave_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tel_leave_history_id_seq OWNED BY public.tel_leave_history.id;


--
-- Name: tel_resource_files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tel_resource_files (
    id bigint NOT NULL,
    task_id bigint,
    telegram_chat_id bigint NOT NULL,
    telegram_message_id bigint NOT NULL,
    telegram_file_id text,
    original_filename text NOT NULL,
    file_type text NOT NULL,
    file_size bigint DEFAULT 0 NOT NULL,
    mime_type text,
    object_key text,
    download_status text DEFAULT 'queued'::text NOT NULL,
    upload_status text DEFAULT 'pending'::text NOT NULL,
    bytes_downloaded bigint DEFAULT 0 NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone,
    CONSTRAINT tel_resource_files_download_status_check CHECK ((download_status = ANY (ARRAY['queued'::text, 'downloading'::text, 'success'::text, 'failed'::text, 'cancelled'::text, 'duplicate'::text]))),
    CONSTRAINT tel_resource_files_upload_status_check CHECK ((upload_status = ANY (ARRAY['pending'::text, 'uploading'::text, 'success'::text, 'failed'::text, 'cancelled'::text])))
);


--
-- Name: tel_resource_files_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tel_resource_files_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_resource_files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tel_resource_files_id_seq OWNED BY public.tel_resource_files.id;


--
-- Name: tel_resource_sources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tel_resource_sources (
    id bigint NOT NULL,
    file_id bigint NOT NULL,
    dialog_id bigint,
    telegram_chat_id bigint NOT NULL,
    telegram_message_id bigint NOT NULL,
    source_name text NOT NULL,
    source_username text,
    message_text text,
    compliance_status text DEFAULT 'allowed'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: tel_resource_sources_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tel_resource_sources_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_resource_sources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tel_resource_sources_id_seq OWNED BY public.tel_resource_sources.id;


--
-- Name: tel_resource_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tel_resource_tasks (
    id bigint NOT NULL,
    keyword text NOT NULL,
    file_types text[] DEFAULT '{}'::text[] NOT NULL,
    source_dialog_ids bigint[] DEFAULT '{}'::bigint[] NOT NULL,
    max_downloads integer NOT NULL,
    status text DEFAULT 'queued'::text NOT NULL,
    files_found integer DEFAULT 0 NOT NULL,
    files_completed integer DEFAULT 0 NOT NULL,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    cancel_requested boolean DEFAULT false NOT NULL,
    source_mode text DEFAULT 'manual'::text NOT NULL,
    source_strategy text DEFAULT 'manual'::text NOT NULL,
    CONSTRAINT tel_resource_tasks_max_downloads_check CHECK (((max_downloads >= 1) AND (max_downloads <= 100))),
    CONSTRAINT tel_resource_tasks_status_check CHECK ((status = ANY (ARRAY['queued'::text, 'running'::text, 'completed'::text, 'failed'::text, 'cancelled'::text])))
);


--
-- Name: tel_resource_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tel_resource_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_resource_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tel_resource_tasks_id_seq OWNED BY public.tel_resource_tasks.id;


--
-- Name: tel_scan_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tel_scan_history (
    id bigint NOT NULL,
    started_at timestamp with time zone NOT NULL,
    finished_at timestamp with time zone,
    status text NOT NULL,
    dialogs_total integer DEFAULT 0 NOT NULL,
    dialogs_success integer DEFAULT 0 NOT NULL,
    dialogs_failed integer DEFAULT 0 NOT NULL,
    error_message text
);


--
-- Name: tel_scan_history_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tel_scan_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_scan_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tel_scan_history_id_seq OWNED BY public.tel_scan_history.id;


--
-- Name: tel_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tel_settings (
    key text NOT NULL,
    value text
);


--
-- Name: xy_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_accounts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    external_user_id character varying(64),
    device_id character varying(128),
    label character varying(80) NOT NULL,
    note character varying(500),
    status character varying(32) DEFAULT 'unconfigured'::character varying NOT NULL,
    enabled boolean DEFAULT true NOT NULL
);


--
-- Name: xy_accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_accounts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_accounts_id_seq OWNED BY public.xy_accounts.id;


--
-- Name: xy_agent_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_agent_config (
    id bigint NOT NULL,
    max_wall_time_seconds bigint NOT NULL,
    max_iterations bigint NOT NULL,
    max_model_calls bigint NOT NULL,
    max_tool_calls bigint NOT NULL,
    max_search_calls bigint NOT NULL,
    max_write_tool_calls bigint NOT NULL,
    max_output_tokens bigint NOT NULL,
    conversation_history_limit bigint NOT NULL,
    summary_trigger_tokens bigint NOT NULL,
    max_concurrent_tool_calls bigint DEFAULT 3 NOT NULL
);


--
-- Name: xy_agent_config_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_agent_config_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_agent_config_id_seq OWNED BY public.xy_agent_config.id;


--
-- Name: xy_agent_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_agent_events (
    id bigint NOT NULL,
    owner_id bigint NOT NULL,
    session_id bigint NOT NULL,
    run_id bigint,
    turn_no bigint,
    step_no bigint,
    event_type character varying(80) NOT NULL,
    call_id character varying(200),
    parent_call_id character varying(200),
    payload text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: xy_agent_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_agent_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_agent_events_id_seq OWNED BY public.xy_agent_events.id;


--
-- Name: xy_agent_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_agent_messages (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    session_id bigint NOT NULL,
    run_id bigint,
    sequence bigint NOT NULL,
    role character varying(32) NOT NULL,
    content text NOT NULL,
    directives text
);


--
-- Name: xy_agent_messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_agent_messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_agent_messages_id_seq OWNED BY public.xy_agent_messages.id;


--
-- Name: xy_agent_run_tools; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_agent_run_tools (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    run_id bigint NOT NULL,
    call_id character varying(200) NOT NULL,
    mcp_server_code character varying(64),
    tool_name character varying(120) NOT NULL,
    arguments text,
    status character varying(32) NOT NULL,
    duration_millis bigint DEFAULT 0 NOT NULL,
    error_code character varying(64),
    error_message character varying(1000),
    started_at timestamp with time zone NOT NULL,
    completed_at timestamp with time zone,
    result text,
    tool_code character varying(120) NOT NULL,
    tool_source character varying(16) DEFAULT 'mcp'::character varying NOT NULL
);


--
-- Name: xy_agent_run_tools_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_agent_run_tools_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_run_tools_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_agent_run_tools_id_seq OWNED BY public.xy_agent_run_tools.id;


--
-- Name: xy_agent_runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_agent_runs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    session_id bigint NOT NULL,
    idempotency_key character varying(100),
    model_id bigint NOT NULL,
    model_name character varying(120) NOT NULL,
    status character varying(32) DEFAULT 'running'::character varying NOT NULL,
    end_reason character varying(64),
    iterations bigint DEFAULT 0 NOT NULL,
    model_calls bigint DEFAULT 0 NOT NULL,
    tool_calls bigint DEFAULT 0 NOT NULL,
    input_tokens bigint DEFAULT 0 NOT NULL,
    output_tokens bigint DEFAULT 0 NOT NULL,
    wall_time_millis bigint DEFAULT 0 NOT NULL,
    config_snapshot text NOT NULL,
    error_message character varying(2000),
    started_at timestamp with time zone NOT NULL,
    completed_at timestamp with time zone
);


--
-- Name: xy_agent_runs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_agent_runs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_runs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_agent_runs_id_seq OWNED BY public.xy_agent_runs.id;


--
-- Name: xy_agent_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_agent_sessions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    title character varying(200) NOT NULL,
    summary text,
    last_model_id bigint,
    last_message_at timestamp with time zone
);


--
-- Name: xy_agent_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_agent_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_agent_sessions_id_seq OWNED BY public.xy_agent_sessions.id;


--
-- Name: xy_ai_models; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_ai_models (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(80) NOT NULL,
    provider character varying(40) NOT NULL,
    base_url character varying(500) NOT NULL,
    api_key text,
    model character varying(120) NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    supports_tools boolean DEFAULT true NOT NULL,
    supports_streaming boolean DEFAULT true NOT NULL,
    context_window bigint DEFAULT 128000 NOT NULL,
    max_output_tokens bigint DEFAULT 2000 NOT NULL,
    request_timeout_seconds bigint DEFAULT 60 NOT NULL,
    extra_config text
);


--
-- Name: xy_ai_models_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_ai_models_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_ai_models_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_ai_models_id_seq OWNED BY public.xy_ai_models.id;


--
-- Name: xy_business_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_business_events (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    account_id bigint,
    product_id bigint,
    order_id bigint,
    type character varying(32) NOT NULL,
    status character varying(16) NOT NULL,
    summary character varying(255) NOT NULL,
    error character varying(500)
);


--
-- Name: xy_business_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_business_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_business_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_business_events_id_seq OWNED BY public.xy_business_events.id;


--
-- Name: xy_conversation_takeovers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_conversation_takeovers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    account_id bigint NOT NULL,
    session_id character varying(255) NOT NULL,
    active boolean DEFAULT true NOT NULL
);


--
-- Name: xy_conversation_takeovers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_conversation_takeovers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_conversation_takeovers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_conversation_takeovers_id_seq OWNED BY public.xy_conversation_takeovers.id;


--
-- Name: xy_credentials; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_credentials (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    cookie_ciphertext text NOT NULL,
    token_ciphertext text,
    token_expires_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: xy_credentials_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_credentials_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_credentials_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_credentials_id_seq OWNED BY public.xy_credentials.id;


--
-- Name: xy_delivery_resources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_delivery_resources (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    rule_id bigint NOT NULL,
    value_hash character varying(64) NOT NULL,
    value_ciphertext text NOT NULL,
    status character varying(16) NOT NULL,
    reserved_task_id bigint,
    consumed_at timestamp with time zone
);


--
-- Name: xy_delivery_resources_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_delivery_resources_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_delivery_resources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_delivery_resources_id_seq OWNED BY public.xy_delivery_resources.id;


--
-- Name: xy_delivery_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_delivery_rules (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    product_id bigint NOT NULL,
    sku_key character varying(128) NOT NULL,
    sku_name character varying(255),
    mode character varying(16) NOT NULL,
    template text NOT NULL,
    enabled boolean NOT NULL
);


--
-- Name: xy_delivery_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_delivery_rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_delivery_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_delivery_rules_id_seq OWNED BY public.xy_delivery_rules.id;


--
-- Name: xy_delivery_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_delivery_tasks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    account_id bigint NOT NULL,
    order_id bigint NOT NULL,
    rule_id bigint NOT NULL,
    resource_id bigint,
    payload_ciphertext text NOT NULL,
    status character varying(32) NOT NULL,
    attempts bigint DEFAULT 0 NOT NULL,
    next_attempt_at timestamp with time zone NOT NULL,
    last_error character varying(500),
    sent_at timestamp with time zone,
    confirmed_at timestamp with time zone
);


--
-- Name: xy_delivery_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_delivery_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_delivery_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_delivery_tasks_id_seq OWNED BY public.xy_delivery_tasks.id;


--
-- Name: xy_mcp_servers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_mcp_servers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(80) NOT NULL,
    code character varying(64) NOT NULL,
    base_url character varying(500) NOT NULL,
    auth_header character varying(100),
    auth_token text,
    enabled boolean DEFAULT true NOT NULL,
    request_timeout_seconds bigint DEFAULT 15 NOT NULL,
    task_timeout_seconds bigint DEFAULT 300 NOT NULL,
    last_status character varying(32) DEFAULT 'unknown'::character varying NOT NULL,
    last_checked_at timestamp with time zone,
    last_error character varying(1000)
);


--
-- Name: xy_mcp_servers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_mcp_servers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_mcp_servers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_mcp_servers_id_seq OWNED BY public.xy_mcp_servers.id;


--
-- Name: xy_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_messages (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    account_id bigint NOT NULL,
    pnm_id character varying(255) NOT NULL,
    session_id character varying(255),
    content_type bigint DEFAULT 0 NOT NULL,
    content text,
    sender_user_name character varying(255),
    sender_user_id character varying(128),
    sender_app_version character varying(64),
    sender_platform character varying(64),
    reminder_url character varying(2048),
    external_product_id character varying(128),
    external_order_id character varying(128),
    message_time bigint DEFAULT 0 NOT NULL,
    direction character varying(16) DEFAULT 'inbound'::character varying NOT NULL,
    source character varying(32) DEFAULT 'remote'::character varying NOT NULL
);


--
-- Name: xy_messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_messages_id_seq OWNED BY public.xy_messages.id;


--
-- Name: xy_order_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_order_events (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    account_id bigint NOT NULL,
    order_id bigint NOT NULL,
    pnm_id character varying(255) NOT NULL,
    type character varying(32) NOT NULL,
    event_time bigint NOT NULL
);


--
-- Name: xy_order_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_order_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_order_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_order_events_id_seq OWNED BY public.xy_order_events.id;


--
-- Name: xy_orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_orders (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    account_id bigint NOT NULL,
    external_order_id character varying(128) NOT NULL,
    product_id bigint,
    external_product_id character varying(128),
    product_title character varying(255),
    buyer_user_id character varying(128),
    buyer_user_name character varying(255),
    session_id character varying(255),
    sku_key character varying(128),
    status character varying(32) NOT NULL,
    amount_cent bigint DEFAULT 0 NOT NULL
);


--
-- Name: xy_orders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_orders_id_seq OWNED BY public.xy_orders.id;


--
-- Name: xy_product_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_product_settings (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id bigint NOT NULL,
    auto_reply_enabled boolean DEFAULT false NOT NULL,
    auto_delivery_enabled boolean DEFAULT false NOT NULL,
    auto_relist_enabled boolean DEFAULT false NOT NULL,
    reply_delay_seconds bigint DEFAULT 0 NOT NULL,
    relist_delay_seconds bigint DEFAULT 60 NOT NULL,
    relist_description text,
    relist_image_urls text,
    relist_address character varying(255),
    virtual_product boolean DEFAULT false NOT NULL
);


--
-- Name: xy_product_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_product_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_product_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_product_settings_id_seq OWNED BY public.xy_product_settings.id;


--
-- Name: xy_products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_products (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    account_id bigint NOT NULL,
    external_product_id character varying(64) NOT NULL,
    title character varying(255) NOT NULL,
    price_cent bigint DEFAULT 0 NOT NULL,
    status character varying(32) DEFAULT 'on_sale'::character varying NOT NULL,
    image_url character varying(2048),
    publish_snapshot text,
    source_json text
);


--
-- Name: xy_products_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_products_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_products_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_products_id_seq OWNED BY public.xy_products.id;


--
-- Name: xy_relist_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_relist_tasks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    account_id bigint NOT NULL,
    order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    status character varying(32) NOT NULL,
    attempts bigint DEFAULT 0 NOT NULL,
    next_attempt_at timestamp with time zone NOT NULL,
    new_external_product_id character varying(128),
    last_error character varying(500)
);


--
-- Name: xy_relist_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_relist_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_relist_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_relist_tasks_id_seq OWNED BY public.xy_relist_tasks.id;


--
-- Name: xy_reply_contents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_reply_contents (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    rule_id bigint NOT NULL,
    text text NOT NULL
);


--
-- Name: xy_reply_contents_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_reply_contents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_reply_contents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_reply_contents_id_seq OWNED BY public.xy_reply_contents.id;


--
-- Name: xy_reply_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_reply_rules (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    product_id bigint NOT NULL,
    type character varying(16) NOT NULL,
    keyword character varying(255) NOT NULL,
    priority bigint DEFAULT 0 NOT NULL,
    enabled boolean NOT NULL
);


--
-- Name: xy_reply_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_reply_rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_reply_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_reply_rules_id_seq OWNED BY public.xy_reply_rules.id;


--
-- Name: xy_reply_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_reply_tasks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    account_id bigint NOT NULL,
    message_id bigint NOT NULL,
    product_id bigint NOT NULL,
    external_product_id character varying(128) NOT NULL,
    rule_id bigint NOT NULL,
    session_id character varying(255) NOT NULL,
    recipient_id character varying(128) NOT NULL,
    reply_text text NOT NULL,
    status character varying(16) NOT NULL,
    error character varying(500),
    scheduled_at timestamp with time zone NOT NULL,
    completed_at timestamp with time zone
);


--
-- Name: xy_reply_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_reply_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_reply_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_reply_tasks_id_seq OWNED BY public.xy_reply_tasks.id;


--
-- Name: xy_resource_candidates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_resource_candidates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    search_id bigint NOT NULL,
    mcp_server_code character varying(64) NOT NULL,
    external_id character varying(1000) NOT NULL,
    name character varying(500) NOT NULL,
    size bigint DEFAULT 0 NOT NULL,
    mime_type character varying(200),
    source text NOT NULL,
    preview text,
    license character varying(100) DEFAULT 'unknown'::character varying NOT NULL,
    risk character varying(1000),
    expires_at timestamp with time zone
);


--
-- Name: xy_resource_candidates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_resource_candidates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_resource_candidates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_resource_candidates_id_seq OWNED BY public.xy_resource_candidates.id;


--
-- Name: xy_resource_searches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_resource_searches (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    query character varying(500) NOT NULL,
    status character varying(32) DEFAULT 'running'::character varying NOT NULL,
    directives text
);


--
-- Name: xy_resource_searches_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_resource_searches_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_resource_searches_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_resource_searches_id_seq OWNED BY public.xy_resource_searches.id;


--
-- Name: xy_resource_sources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_resource_sources (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    resource_id bigint NOT NULL,
    mcp_server_code character varying(64) NOT NULL,
    external_id character varying(1000) NOT NULL,
    source_name character varying(500),
    source_reference character varying(1000),
    metadata text
);


--
-- Name: xy_resource_sources_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_resource_sources_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_resource_sources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_resource_sources_id_seq OWNED BY public.xy_resource_sources.id;


--
-- Name: xy_resource_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_resource_tasks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    agent_run_id bigint,
    candidate_id bigint NOT NULL,
    mcp_server_code character varying(64) NOT NULL,
    external_task_id character varying(200),
    idempotency_key character varying(100) NOT NULL,
    state character varying(32) DEFAULT 'queued'::character varying NOT NULL,
    stage character varying(64),
    progress_percent bigint DEFAULT 0 NOT NULL,
    progress_done bigint DEFAULT 0 NOT NULL,
    progress_total bigint DEFAULT 0 NOT NULL,
    progress_unit character varying(32),
    result text,
    error_code character varying(64),
    error_message character varying(2000),
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    delivery_token_hash character varying(64),
    delivery_expires_at timestamp with time zone,
    upload_object_key character varying(1000)
);


--
-- Name: xy_resource_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_resource_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_resource_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_resource_tasks_id_seq OWNED BY public.xy_resource_tasks.id;


--
-- Name: xy_resources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_resources (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_id bigint NOT NULL,
    task_id bigint NOT NULL,
    name character varying(500) NOT NULL,
    size bigint NOT NULL,
    mime_type character varying(200),
    sha256 character varying(64),
    bucket character varying(200) NOT NULL,
    object_key character varying(1000) NOT NULL,
    state character varying(32) DEFAULT 'available'::character varying NOT NULL
);


--
-- Name: xy_resources_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.xy_resources_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_resources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.xy_resources_id_seq OWNED BY public.xy_resources.id;


--
-- Name: xy_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.xy_settings (
    key character varying(128) NOT NULL,
    value text NOT NULL
);


--
-- Name: jwt_blacklists id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jwt_blacklists ALTER COLUMN id SET DEFAULT nextval('public.jwt_blacklists_id_seq'::regclass);


--
-- Name: s_dictionaries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.s_dictionaries ALTER COLUMN id SET DEFAULT nextval('public.s_dictionaries_id_seq'::regclass);


--
-- Name: s_dictionary_details id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.s_dictionary_details ALTER COLUMN id SET DEFAULT nextval('public.s_dictionary_details_id_seq'::regclass);


--
-- Name: s_login_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.s_login_logs ALTER COLUMN id SET DEFAULT nextval('public.s_login_logs_id_seq'::regclass);


--
-- Name: s_users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.s_users ALTER COLUMN id SET DEFAULT nextval('public.s_users_id_seq'::regclass);


--
-- Name: tel_acquisition_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_acquisition_tasks ALTER COLUMN id SET DEFAULT nextval('public.tel_acquisition_tasks_id_seq'::regclass);


--
-- Name: tel_dialogs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_dialogs ALTER COLUMN id SET DEFAULT nextval('public.tel_dialogs_id_seq'::regclass);


--
-- Name: tel_leave_history id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_leave_history ALTER COLUMN id SET DEFAULT nextval('public.tel_leave_history_id_seq'::regclass);


--
-- Name: tel_resource_files id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_resource_files ALTER COLUMN id SET DEFAULT nextval('public.tel_resource_files_id_seq'::regclass);


--
-- Name: tel_resource_sources id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_resource_sources ALTER COLUMN id SET DEFAULT nextval('public.tel_resource_sources_id_seq'::regclass);


--
-- Name: tel_resource_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_resource_tasks ALTER COLUMN id SET DEFAULT nextval('public.tel_resource_tasks_id_seq'::regclass);


--
-- Name: tel_scan_history id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_scan_history ALTER COLUMN id SET DEFAULT nextval('public.tel_scan_history_id_seq'::regclass);


--
-- Name: xy_accounts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_accounts ALTER COLUMN id SET DEFAULT nextval('public.xy_accounts_id_seq'::regclass);


--
-- Name: xy_agent_config id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_config ALTER COLUMN id SET DEFAULT nextval('public.xy_agent_config_id_seq'::regclass);


--
-- Name: xy_agent_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_events ALTER COLUMN id SET DEFAULT nextval('public.xy_agent_events_id_seq'::regclass);


--
-- Name: xy_agent_messages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_messages ALTER COLUMN id SET DEFAULT nextval('public.xy_agent_messages_id_seq'::regclass);


--
-- Name: xy_agent_run_tools id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_run_tools ALTER COLUMN id SET DEFAULT nextval('public.xy_agent_run_tools_id_seq'::regclass);


--
-- Name: xy_agent_runs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_runs ALTER COLUMN id SET DEFAULT nextval('public.xy_agent_runs_id_seq'::regclass);


--
-- Name: xy_agent_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_sessions ALTER COLUMN id SET DEFAULT nextval('public.xy_agent_sessions_id_seq'::regclass);


--
-- Name: xy_ai_models id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_ai_models ALTER COLUMN id SET DEFAULT nextval('public.xy_ai_models_id_seq'::regclass);


--
-- Name: xy_business_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_business_events ALTER COLUMN id SET DEFAULT nextval('public.xy_business_events_id_seq'::regclass);


--
-- Name: xy_conversation_takeovers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_conversation_takeovers ALTER COLUMN id SET DEFAULT nextval('public.xy_conversation_takeovers_id_seq'::regclass);


--
-- Name: xy_credentials id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_credentials ALTER COLUMN id SET DEFAULT nextval('public.xy_credentials_id_seq'::regclass);


--
-- Name: xy_delivery_resources id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_delivery_resources ALTER COLUMN id SET DEFAULT nextval('public.xy_delivery_resources_id_seq'::regclass);


--
-- Name: xy_delivery_rules id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_delivery_rules ALTER COLUMN id SET DEFAULT nextval('public.xy_delivery_rules_id_seq'::regclass);


--
-- Name: xy_delivery_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_delivery_tasks ALTER COLUMN id SET DEFAULT nextval('public.xy_delivery_tasks_id_seq'::regclass);


--
-- Name: xy_mcp_servers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_mcp_servers ALTER COLUMN id SET DEFAULT nextval('public.xy_mcp_servers_id_seq'::regclass);


--
-- Name: xy_messages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_messages ALTER COLUMN id SET DEFAULT nextval('public.xy_messages_id_seq'::regclass);


--
-- Name: xy_order_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_order_events ALTER COLUMN id SET DEFAULT nextval('public.xy_order_events_id_seq'::regclass);


--
-- Name: xy_orders id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_orders ALTER COLUMN id SET DEFAULT nextval('public.xy_orders_id_seq'::regclass);


--
-- Name: xy_product_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_product_settings ALTER COLUMN id SET DEFAULT nextval('public.xy_product_settings_id_seq'::regclass);


--
-- Name: xy_products id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_products ALTER COLUMN id SET DEFAULT nextval('public.xy_products_id_seq'::regclass);


--
-- Name: xy_relist_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_relist_tasks ALTER COLUMN id SET DEFAULT nextval('public.xy_relist_tasks_id_seq'::regclass);


--
-- Name: xy_reply_contents id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_reply_contents ALTER COLUMN id SET DEFAULT nextval('public.xy_reply_contents_id_seq'::regclass);


--
-- Name: xy_reply_rules id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_reply_rules ALTER COLUMN id SET DEFAULT nextval('public.xy_reply_rules_id_seq'::regclass);


--
-- Name: xy_reply_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_reply_tasks ALTER COLUMN id SET DEFAULT nextval('public.xy_reply_tasks_id_seq'::regclass);


--
-- Name: xy_resource_candidates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_resource_candidates ALTER COLUMN id SET DEFAULT nextval('public.xy_resource_candidates_id_seq'::regclass);


--
-- Name: xy_resource_searches id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_resource_searches ALTER COLUMN id SET DEFAULT nextval('public.xy_resource_searches_id_seq'::regclass);


--
-- Name: xy_resource_sources id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_resource_sources ALTER COLUMN id SET DEFAULT nextval('public.xy_resource_sources_id_seq'::regclass);


--
-- Name: xy_resource_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_resource_tasks ALTER COLUMN id SET DEFAULT nextval('public.xy_resource_tasks_id_seq'::regclass);


--
-- Name: xy_resources id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_resources ALTER COLUMN id SET DEFAULT nextval('public.xy_resources_id_seq'::regclass);


--
-- Name: jwt_blacklists jwt_blacklists_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jwt_blacklists
    ADD CONSTRAINT jwt_blacklists_pkey PRIMARY KEY (id);


--
-- Name: s_dictionaries s_dictionaries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.s_dictionaries
    ADD CONSTRAINT s_dictionaries_pkey PRIMARY KEY (id);


--
-- Name: s_dictionary_details s_dictionary_details_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.s_dictionary_details
    ADD CONSTRAINT s_dictionary_details_pkey PRIMARY KEY (id);


--
-- Name: s_login_logs s_login_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.s_login_logs
    ADD CONSTRAINT s_login_logs_pkey PRIMARY KEY (id);


--
-- Name: s_users s_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.s_users
    ADD CONSTRAINT s_users_pkey PRIMARY KEY (id);


--
-- Name: tel_acquisition_tasks tel_acquisition_tasks_resource_task_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_acquisition_tasks
    ADD CONSTRAINT tel_acquisition_tasks_resource_task_id_key UNIQUE (resource_task_id);


--
-- Name: tel_acquisition_tasks tel_acquisition_tasks_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_acquisition_tasks
    ADD CONSTRAINT tel_acquisition_tasks_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: tel_acquisition_tasks tel_acquisition_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_acquisition_tasks
    ADD CONSTRAINT tel_acquisition_tasks_pkey PRIMARY KEY (id);


--
-- Name: tel_dialogs tel_dialogs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_dialogs
    ADD CONSTRAINT tel_dialogs_pkey PRIMARY KEY (id);


--
-- Name: tel_dialogs tel_dialogs_telegram_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_dialogs
    ADD CONSTRAINT tel_dialogs_telegram_id_key UNIQUE (telegram_id);


--
-- Name: tel_leave_history tel_leave_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_leave_history
    ADD CONSTRAINT tel_leave_history_pkey PRIMARY KEY (id);


--
-- Name: tel_resource_files tel_resource_files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_resource_files
    ADD CONSTRAINT tel_resource_files_pkey PRIMARY KEY (id);


--
-- Name: tel_resource_files tel_resource_files_telegram_chat_id_telegram_message_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_resource_files
    ADD CONSTRAINT tel_resource_files_telegram_chat_id_telegram_message_id_key UNIQUE (telegram_chat_id, telegram_message_id);


--
-- Name: tel_resource_sources tel_resource_sources_file_id_telegram_chat_id_telegram_mess_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_resource_sources
    ADD CONSTRAINT tel_resource_sources_file_id_telegram_chat_id_telegram_mess_key UNIQUE (file_id, telegram_chat_id, telegram_message_id);


--
-- Name: tel_resource_sources tel_resource_sources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_resource_sources
    ADD CONSTRAINT tel_resource_sources_pkey PRIMARY KEY (id);


--
-- Name: tel_resource_tasks tel_resource_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_resource_tasks
    ADD CONSTRAINT tel_resource_tasks_pkey PRIMARY KEY (id);


--
-- Name: tel_scan_history tel_scan_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_scan_history
    ADD CONSTRAINT tel_scan_history_pkey PRIMARY KEY (id);


--
-- Name: tel_settings tel_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_settings
    ADD CONSTRAINT tel_settings_pkey PRIMARY KEY (key);


--
-- Name: xy_accounts xy_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_accounts
    ADD CONSTRAINT xy_accounts_pkey PRIMARY KEY (id);


--
-- Name: xy_agent_config xy_agent_config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_config
    ADD CONSTRAINT xy_agent_config_pkey PRIMARY KEY (id);


--
-- Name: xy_agent_events xy_agent_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_events
    ADD CONSTRAINT xy_agent_events_pkey PRIMARY KEY (id);


--
-- Name: xy_agent_messages xy_agent_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_messages
    ADD CONSTRAINT xy_agent_messages_pkey PRIMARY KEY (id);


--
-- Name: xy_agent_run_tools xy_agent_run_tools_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_run_tools
    ADD CONSTRAINT xy_agent_run_tools_pkey PRIMARY KEY (id);


--
-- Name: xy_agent_runs xy_agent_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_runs
    ADD CONSTRAINT xy_agent_runs_pkey PRIMARY KEY (id);


--
-- Name: xy_agent_sessions xy_agent_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_agent_sessions
    ADD CONSTRAINT xy_agent_sessions_pkey PRIMARY KEY (id);


--
-- Name: xy_ai_models xy_ai_models_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_ai_models
    ADD CONSTRAINT xy_ai_models_pkey PRIMARY KEY (id);


--
-- Name: xy_business_events xy_business_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_business_events
    ADD CONSTRAINT xy_business_events_pkey PRIMARY KEY (id);


--
-- Name: xy_conversation_takeovers xy_conversation_takeovers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_conversation_takeovers
    ADD CONSTRAINT xy_conversation_takeovers_pkey PRIMARY KEY (id);


--
-- Name: xy_credentials xy_credentials_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_credentials
    ADD CONSTRAINT xy_credentials_pkey PRIMARY KEY (id);


--
-- Name: xy_delivery_resources xy_delivery_resources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_delivery_resources
    ADD CONSTRAINT xy_delivery_resources_pkey PRIMARY KEY (id);


--
-- Name: xy_delivery_rules xy_delivery_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_delivery_rules
    ADD CONSTRAINT xy_delivery_rules_pkey PRIMARY KEY (id);


--
-- Name: xy_delivery_tasks xy_delivery_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_delivery_tasks
    ADD CONSTRAINT xy_delivery_tasks_pkey PRIMARY KEY (id);


--
-- Name: xy_mcp_servers xy_mcp_servers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_mcp_servers
    ADD CONSTRAINT xy_mcp_servers_pkey PRIMARY KEY (id);


--
-- Name: xy_messages xy_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_messages
    ADD CONSTRAINT xy_messages_pkey PRIMARY KEY (id);


--
-- Name: xy_order_events xy_order_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_order_events
    ADD CONSTRAINT xy_order_events_pkey PRIMARY KEY (id);


--
-- Name: xy_orders xy_orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_orders
    ADD CONSTRAINT xy_orders_pkey PRIMARY KEY (id);


--
-- Name: xy_product_settings xy_product_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_product_settings
    ADD CONSTRAINT xy_product_settings_pkey PRIMARY KEY (id);


--
-- Name: xy_products xy_products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_products
    ADD CONSTRAINT xy_products_pkey PRIMARY KEY (id);


--
-- Name: xy_relist_tasks xy_relist_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_relist_tasks
    ADD CONSTRAINT xy_relist_tasks_pkey PRIMARY KEY (id);


--
-- Name: xy_reply_contents xy_reply_contents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_reply_contents
    ADD CONSTRAINT xy_reply_contents_pkey PRIMARY KEY (id);


--
-- Name: xy_reply_rules xy_reply_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_reply_rules
    ADD CONSTRAINT xy_reply_rules_pkey PRIMARY KEY (id);


--
-- Name: xy_reply_tasks xy_reply_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_reply_tasks
    ADD CONSTRAINT xy_reply_tasks_pkey PRIMARY KEY (id);


--
-- Name: xy_resource_candidates xy_resource_candidates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_resource_candidates
    ADD CONSTRAINT xy_resource_candidates_pkey PRIMARY KEY (id);


--
-- Name: xy_resource_searches xy_resource_searches_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_resource_searches
    ADD CONSTRAINT xy_resource_searches_pkey PRIMARY KEY (id);


--
-- Name: xy_resource_sources xy_resource_sources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_resource_sources
    ADD CONSTRAINT xy_resource_sources_pkey PRIMARY KEY (id);


--
-- Name: xy_resource_tasks xy_resource_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_resource_tasks
    ADD CONSTRAINT xy_resource_tasks_pkey PRIMARY KEY (id);


--
-- Name: xy_resources xy_resources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_resources
    ADD CONSTRAINT xy_resources_pkey PRIMARY KEY (id);


--
-- Name: xy_settings xy_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.xy_settings
    ADD CONSTRAINT xy_settings_pkey PRIMARY KEY (key);


--
-- Name: idx_jwt_blacklists_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_jwt_blacklists_deleted_at ON public.jwt_blacklists USING btree (deleted_at);


--
-- Name: idx_s_dictionaries_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_s_dictionaries_deleted_at ON public.s_dictionaries USING btree (deleted_at);


--
-- Name: idx_s_dictionary_details_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_s_dictionary_details_deleted_at ON public.s_dictionary_details USING btree (deleted_at);


--
-- Name: idx_s_login_logs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_s_login_logs_deleted_at ON public.s_login_logs USING btree (deleted_at);


--
-- Name: idx_s_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_s_users_deleted_at ON public.s_users USING btree (deleted_at);


--
-- Name: idx_s_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_s_users_username ON public.s_users USING btree (username);


--
-- Name: idx_s_users_uuid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_s_users_uuid ON public.s_users USING btree (uuid);


--
-- Name: idx_tel_resource_files_filter; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tel_resource_files_filter ON public.tel_resource_files USING btree (download_status, file_type);


--
-- Name: idx_tel_resource_files_task; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tel_resource_files_task ON public.tel_resource_files USING btree (task_id);


--
-- Name: idx_tel_resource_files_telegram_file; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tel_resource_files_telegram_file ON public.tel_resource_files USING btree (telegram_file_id) WHERE (telegram_file_id IS NOT NULL);


--
-- Name: idx_xy_accounts_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_accounts_deleted_at ON public.xy_accounts USING btree (deleted_at);


--
-- Name: idx_xy_accounts_external_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_accounts_external_user_id ON public.xy_accounts USING btree (external_user_id);


--
-- Name: idx_xy_accounts_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_accounts_owner ON public.xy_accounts USING btree (owner_id);


--
-- Name: idx_xy_agent_events_call_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_events_call_id ON public.xy_agent_events USING btree (call_id);


--
-- Name: idx_xy_agent_events_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_events_created_at ON public.xy_agent_events USING btree (created_at);


--
-- Name: idx_xy_agent_events_owner_session; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_events_owner_session ON public.xy_agent_events USING btree (owner_id, session_id);


--
-- Name: idx_xy_agent_events_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_events_run_id ON public.xy_agent_events USING btree (run_id);


--
-- Name: idx_xy_agent_events_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_events_session_id ON public.xy_agent_events USING btree (session_id);


--
-- Name: idx_xy_agent_events_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_events_type ON public.xy_agent_events USING btree (event_type);


--
-- Name: idx_xy_agent_messages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_messages_deleted_at ON public.xy_agent_messages USING btree (deleted_at);


--
-- Name: idx_xy_agent_messages_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_messages_owner ON public.xy_agent_messages USING btree (owner_id);


--
-- Name: idx_xy_agent_messages_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_messages_run_id ON public.xy_agent_messages USING btree (run_id);


--
-- Name: idx_xy_agent_messages_session_sequence; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_agent_messages_session_sequence ON public.xy_agent_messages USING btree (session_id, sequence);


--
-- Name: idx_xy_agent_run_tools_call_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_agent_run_tools_call_id ON public.xy_agent_run_tools USING btree (call_id);


--
-- Name: idx_xy_agent_run_tools_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_run_tools_deleted_at ON public.xy_agent_run_tools USING btree (deleted_at);


--
-- Name: idx_xy_agent_run_tools_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_run_tools_run_id ON public.xy_agent_run_tools USING btree (run_id);


--
-- Name: idx_xy_agent_runs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_runs_deleted_at ON public.xy_agent_runs USING btree (deleted_at);


--
-- Name: idx_xy_agent_runs_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_runs_owner ON public.xy_agent_runs USING btree (owner_id);


--
-- Name: idx_xy_agent_runs_owner_idempotency; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_agent_runs_owner_idempotency ON public.xy_agent_runs USING btree (owner_id, idempotency_key);


--
-- Name: idx_xy_agent_runs_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_runs_session_id ON public.xy_agent_runs USING btree (session_id);


--
-- Name: idx_xy_agent_runs_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_runs_status ON public.xy_agent_runs USING btree (status);


--
-- Name: idx_xy_agent_sessions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_sessions_deleted_at ON public.xy_agent_sessions USING btree (deleted_at);


--
-- Name: idx_xy_agent_sessions_last_message_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_sessions_last_message_at ON public.xy_agent_sessions USING btree (last_message_at);


--
-- Name: idx_xy_agent_sessions_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_agent_sessions_owner ON public.xy_agent_sessions USING btree (owner_id);


--
-- Name: idx_xy_ai_models_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_ai_models_deleted_at ON public.xy_ai_models USING btree (deleted_at);


--
-- Name: idx_xy_ai_models_is_default; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_ai_models_is_default ON public.xy_ai_models USING btree (is_default);


--
-- Name: idx_xy_business_events_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_business_events_account_id ON public.xy_business_events USING btree (account_id);


--
-- Name: idx_xy_business_events_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_business_events_deleted_at ON public.xy_business_events USING btree (deleted_at);


--
-- Name: idx_xy_business_events_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_business_events_order_id ON public.xy_business_events USING btree (order_id);


--
-- Name: idx_xy_business_events_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_business_events_owner ON public.xy_business_events USING btree (owner_id);


--
-- Name: idx_xy_business_events_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_business_events_product_id ON public.xy_business_events USING btree (product_id);


--
-- Name: idx_xy_business_events_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_business_events_status ON public.xy_business_events USING btree (status);


--
-- Name: idx_xy_business_events_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_business_events_type ON public.xy_business_events USING btree (type);


--
-- Name: idx_xy_candidate_external; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_candidate_external ON public.xy_resource_candidates USING btree (search_id, mcp_server_code, external_id);


--
-- Name: idx_xy_conversation_takeovers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_conversation_takeovers_deleted_at ON public.xy_conversation_takeovers USING btree (deleted_at);


--
-- Name: idx_xy_credentials_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_credentials_account_id ON public.xy_credentials USING btree (account_id);


--
-- Name: idx_xy_credentials_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_credentials_deleted_at ON public.xy_credentials USING btree (deleted_at);


--
-- Name: idx_xy_delivery_resource_rule_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_delivery_resource_rule_hash ON public.xy_delivery_resources USING btree (rule_id, value_hash);


--
-- Name: idx_xy_delivery_resources_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_resources_deleted_at ON public.xy_delivery_resources USING btree (deleted_at);


--
-- Name: idx_xy_delivery_resources_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_resources_owner ON public.xy_delivery_resources USING btree (owner_id);


--
-- Name: idx_xy_delivery_resources_reserved_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_resources_reserved_task_id ON public.xy_delivery_resources USING btree (reserved_task_id);


--
-- Name: idx_xy_delivery_resources_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_resources_status ON public.xy_delivery_resources USING btree (status);


--
-- Name: idx_xy_delivery_rule_product_sku; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_delivery_rule_product_sku ON public.xy_delivery_rules USING btree (product_id, sku_key);


--
-- Name: idx_xy_delivery_rules_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_rules_deleted_at ON public.xy_delivery_rules USING btree (deleted_at);


--
-- Name: idx_xy_delivery_rules_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_rules_owner ON public.xy_delivery_rules USING btree (owner_id);


--
-- Name: idx_xy_delivery_tasks_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_tasks_account_id ON public.xy_delivery_tasks USING btree (account_id);


--
-- Name: idx_xy_delivery_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_tasks_deleted_at ON public.xy_delivery_tasks USING btree (deleted_at);


--
-- Name: idx_xy_delivery_tasks_next_attempt_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_tasks_next_attempt_at ON public.xy_delivery_tasks USING btree (next_attempt_at);


--
-- Name: idx_xy_delivery_tasks_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_delivery_tasks_order_id ON public.xy_delivery_tasks USING btree (order_id);


--
-- Name: idx_xy_delivery_tasks_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_tasks_owner ON public.xy_delivery_tasks USING btree (owner_id);


--
-- Name: idx_xy_delivery_tasks_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_tasks_resource_id ON public.xy_delivery_tasks USING btree (resource_id);


--
-- Name: idx_xy_delivery_tasks_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_delivery_tasks_status ON public.xy_delivery_tasks USING btree (status);


--
-- Name: idx_xy_mcp_servers_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_mcp_servers_code ON public.xy_mcp_servers USING btree (code);


--
-- Name: idx_xy_mcp_servers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_mcp_servers_deleted_at ON public.xy_mcp_servers USING btree (deleted_at);


--
-- Name: idx_xy_messages_account_pnm; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_messages_account_pnm ON public.xy_messages USING btree (account_id, pnm_id);


--
-- Name: idx_xy_messages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_messages_deleted_at ON public.xy_messages USING btree (deleted_at);


--
-- Name: idx_xy_messages_external_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_messages_external_order_id ON public.xy_messages USING btree (external_order_id);


--
-- Name: idx_xy_messages_external_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_messages_external_product_id ON public.xy_messages USING btree (external_product_id);


--
-- Name: idx_xy_messages_message_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_messages_message_time ON public.xy_messages USING btree (message_time);


--
-- Name: idx_xy_messages_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_messages_owner ON public.xy_messages USING btree (owner_id);


--
-- Name: idx_xy_messages_session; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_messages_session ON public.xy_messages USING btree (session_id);


--
-- Name: idx_xy_order_events_account_pnm; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_order_events_account_pnm ON public.xy_order_events USING btree (account_id, pnm_id);


--
-- Name: idx_xy_order_events_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_order_events_deleted_at ON public.xy_order_events USING btree (deleted_at);


--
-- Name: idx_xy_order_events_event_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_order_events_event_time ON public.xy_order_events USING btree (event_time);


--
-- Name: idx_xy_order_events_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_order_events_order_id ON public.xy_order_events USING btree (order_id);


--
-- Name: idx_xy_order_events_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_order_events_owner ON public.xy_order_events USING btree (owner_id);


--
-- Name: idx_xy_orders_account_external; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_orders_account_external ON public.xy_orders USING btree (account_id, external_order_id);


--
-- Name: idx_xy_orders_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_orders_deleted_at ON public.xy_orders USING btree (deleted_at);


--
-- Name: idx_xy_orders_external_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_orders_external_product_id ON public.xy_orders USING btree (external_product_id);


--
-- Name: idx_xy_orders_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_orders_owner ON public.xy_orders USING btree (owner_id);


--
-- Name: idx_xy_orders_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_orders_product_id ON public.xy_orders USING btree (product_id);


--
-- Name: idx_xy_orders_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_orders_session_id ON public.xy_orders USING btree (session_id);


--
-- Name: idx_xy_orders_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_orders_status ON public.xy_orders USING btree (status);


--
-- Name: idx_xy_product_settings_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_product_settings_deleted_at ON public.xy_product_settings USING btree (deleted_at);


--
-- Name: idx_xy_product_settings_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_product_settings_product_id ON public.xy_product_settings USING btree (product_id);


--
-- Name: idx_xy_products_account_external; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_products_account_external ON public.xy_products USING btree (account_id, external_product_id);


--
-- Name: idx_xy_products_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_products_deleted_at ON public.xy_products USING btree (deleted_at);


--
-- Name: idx_xy_products_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_products_owner ON public.xy_products USING btree (owner_id);


--
-- Name: idx_xy_relist_tasks_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_relist_tasks_account_id ON public.xy_relist_tasks USING btree (account_id);


--
-- Name: idx_xy_relist_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_relist_tasks_deleted_at ON public.xy_relist_tasks USING btree (deleted_at);


--
-- Name: idx_xy_relist_tasks_next_attempt_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_relist_tasks_next_attempt_at ON public.xy_relist_tasks USING btree (next_attempt_at);


--
-- Name: idx_xy_relist_tasks_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_relist_tasks_order_id ON public.xy_relist_tasks USING btree (order_id);


--
-- Name: idx_xy_relist_tasks_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_relist_tasks_owner ON public.xy_relist_tasks USING btree (owner_id);


--
-- Name: idx_xy_relist_tasks_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_relist_tasks_product_id ON public.xy_relist_tasks USING btree (product_id);


--
-- Name: idx_xy_relist_tasks_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_relist_tasks_status ON public.xy_relist_tasks USING btree (status);


--
-- Name: idx_xy_reply_contents_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_reply_contents_deleted_at ON public.xy_reply_contents USING btree (deleted_at);


--
-- Name: idx_xy_reply_contents_rule_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_reply_contents_rule_id ON public.xy_reply_contents USING btree (rule_id);


--
-- Name: idx_xy_reply_rule_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_reply_rule_unique ON public.xy_reply_rules USING btree (product_id, type, keyword);


--
-- Name: idx_xy_reply_rules_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_reply_rules_deleted_at ON public.xy_reply_rules USING btree (deleted_at);


--
-- Name: idx_xy_reply_rules_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_reply_rules_owner ON public.xy_reply_rules USING btree (owner_id);


--
-- Name: idx_xy_reply_tasks_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_reply_tasks_account_id ON public.xy_reply_tasks USING btree (account_id);


--
-- Name: idx_xy_reply_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_reply_tasks_deleted_at ON public.xy_reply_tasks USING btree (deleted_at);


--
-- Name: idx_xy_reply_tasks_external_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_reply_tasks_external_product_id ON public.xy_reply_tasks USING btree (external_product_id);


--
-- Name: idx_xy_reply_tasks_message_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_reply_tasks_message_id ON public.xy_reply_tasks USING btree (message_id);


--
-- Name: idx_xy_reply_tasks_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_reply_tasks_owner ON public.xy_reply_tasks USING btree (owner_id);


--
-- Name: idx_xy_reply_tasks_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_reply_tasks_product_id ON public.xy_reply_tasks USING btree (product_id);


--
-- Name: idx_xy_reply_tasks_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_reply_tasks_session_id ON public.xy_reply_tasks USING btree (session_id);


--
-- Name: idx_xy_reply_tasks_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_reply_tasks_status ON public.xy_reply_tasks USING btree (status);


--
-- Name: idx_xy_resource_candidates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_candidates_deleted_at ON public.xy_resource_candidates USING btree (deleted_at);


--
-- Name: idx_xy_resource_candidates_search_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_candidates_search_id ON public.xy_resource_candidates USING btree (search_id);


--
-- Name: idx_xy_resource_searches_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_searches_deleted_at ON public.xy_resource_searches USING btree (deleted_at);


--
-- Name: idx_xy_resource_searches_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_searches_owner ON public.xy_resource_searches USING btree (owner_id);


--
-- Name: idx_xy_resource_source_external; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_resource_source_external ON public.xy_resource_sources USING btree (resource_id, mcp_server_code, external_id);


--
-- Name: idx_xy_resource_sources_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_sources_deleted_at ON public.xy_resource_sources USING btree (deleted_at);


--
-- Name: idx_xy_resource_sources_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_sources_resource_id ON public.xy_resource_sources USING btree (resource_id);


--
-- Name: idx_xy_resource_task_idempotency; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_resource_task_idempotency ON public.xy_resource_tasks USING btree (owner_id, idempotency_key);


--
-- Name: idx_xy_resource_tasks_agent_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_tasks_agent_run_id ON public.xy_resource_tasks USING btree (agent_run_id);


--
-- Name: idx_xy_resource_tasks_candidate_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_tasks_candidate_id ON public.xy_resource_tasks USING btree (candidate_id);


--
-- Name: idx_xy_resource_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_tasks_deleted_at ON public.xy_resource_tasks USING btree (deleted_at);


--
-- Name: idx_xy_resource_tasks_external_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_tasks_external_task_id ON public.xy_resource_tasks USING btree (external_task_id);


--
-- Name: idx_xy_resource_tasks_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_tasks_owner ON public.xy_resource_tasks USING btree (owner_id);


--
-- Name: idx_xy_resource_tasks_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resource_tasks_state ON public.xy_resource_tasks USING btree (state);


--
-- Name: idx_xy_resources_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resources_deleted_at ON public.xy_resources USING btree (deleted_at);


--
-- Name: idx_xy_resources_object_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_resources_object_key ON public.xy_resources USING btree (object_key);


--
-- Name: idx_xy_resources_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resources_owner ON public.xy_resources USING btree (owner_id);


--
-- Name: idx_xy_resources_sha256; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resources_sha256 ON public.xy_resources USING btree (sha256);


--
-- Name: idx_xy_resources_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_resources_task_id ON public.xy_resources USING btree (task_id);


--
-- Name: idx_xy_takeovers_account_session; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_xy_takeovers_account_session ON public.xy_conversation_takeovers USING btree (account_id, session_id);


--
-- Name: idx_xy_takeovers_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_xy_takeovers_owner ON public.xy_conversation_takeovers USING btree (owner_id);


--
-- Name: tel_resource_files tel_resource_files_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_resource_files
    ADD CONSTRAINT tel_resource_files_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tel_resource_tasks(id) ON DELETE SET NULL;


--
-- Name: tel_resource_sources tel_resource_sources_dialog_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_resource_sources
    ADD CONSTRAINT tel_resource_sources_dialog_id_fkey FOREIGN KEY (dialog_id) REFERENCES public.tel_dialogs(id) ON DELETE SET NULL;


--
-- Name: tel_resource_sources tel_resource_sources_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tel_resource_sources
    ADD CONSTRAINT tel_resource_sources_file_id_fkey FOREIGN KEY (file_id) REFERENCES public.tel_resource_files(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

-- 开源环境仅写入运行所需的最小基础数据，不包含任何业务数据。
INSERT INTO public.s_users
    (id, created_at, updated_at, uuid, username, password, nick_name, header_img, authority_id, enable)
VALUES
    (1, now(), now(), '00000000-0000-0000-0000-000000000001', 'admin', '$2y$10$tziMzps8TFC.60V6WeXnXOUzm66W5rUfVMgf4coNODM3vLDBP8rtW', '管理员', '', 888, 1);
SELECT pg_catalog.setval('public.s_users_id_seq', 1, true);

INSERT INTO public.s_dictionaries
    (id, created_at, updated_at, name, type, status, "desc")
VALUES
    (1, now(), now(), '用户角色', 'user_role', true, '固定用户角色'),
    (2, now(), now(), '用户状态', 'user_status', true, '用户启用状态');
SELECT pg_catalog.setval('public.s_dictionaries_id_seq', 2, true);

INSERT INTO public.s_dictionary_details
    (id, created_at, updated_at, label, value, status, sort, sys_dictionary_id, level, path)
VALUES
    (1, now(), now(), '管理员', '888', true, 1, 1, 0, ''),
    (2, now(), now(), '普通用户', '9528', true, 2, 1, 0, ''),
    (3, now(), now(), '正常', '1', true, 1, 2, 0, ''),
    (4, now(), now(), '冻结', '2', true, 2, 2, 0, '');
SELECT pg_catalog.setval('public.s_dictionary_details_id_seq', 4, true);

INSERT INTO public.xy_agent_config
    (id, max_wall_time_seconds, max_iterations, max_model_calls, max_tool_calls,
     max_search_calls, max_write_tool_calls, max_output_tokens,
     conversation_history_limit, summary_trigger_tokens, max_concurrent_tool_calls)
VALUES
    (1, 300, 50, 60, 200, 200, 10, 2000, 40, 24000, 3);
SELECT pg_catalog.setval('public.xy_agent_config_id_seq', 1, true);

INSERT INTO public.tel_settings (key, value)
VALUES ('adult_mode', 'false'), ('adult_dialog_ids', '[]');
