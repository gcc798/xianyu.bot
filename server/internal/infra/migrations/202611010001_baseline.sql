-- +goose Up
--
-- PostgreSQL database dump
--


-- Dumped from database version 17.11
-- Dumped by pg_dump version 18.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

-- *not* creating schema, since initdb creates it


SET default_tablespace = '';

SET default_table_access_method = "heap";

--
-- Name: casbin_rule; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."casbin_rule" (
    "id" bigint NOT NULL,
    "ptype" character varying(100),
    "v0" character varying(100),
    "v1" character varying(100),
    "v2" character varying(100),
    "v3" character varying(100),
    "v4" character varying(100),
    "v5" character varying(100)
);


--
-- Name: casbin_rule_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."casbin_rule_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: casbin_rule_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."casbin_rule_id_seq" OWNED BY "public"."casbin_rule"."id";


--
-- Name: exa_attachment_category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."exa_attachment_category" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "name" character varying(255) DEFAULT NULL::character varying,
    "pid" bigint DEFAULT 0
);


--
-- Name: COLUMN "exa_attachment_category"."name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."exa_attachment_category"."name" IS '分类名称';


--
-- Name: COLUMN "exa_attachment_category"."pid"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."exa_attachment_category"."pid" IS '父节点ID';


--
-- Name: exa_attachment_category_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."exa_attachment_category_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_attachment_category_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."exa_attachment_category_id_seq" OWNED BY "public"."exa_attachment_category"."id";


--
-- Name: exa_customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."exa_customers" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "customer_name" "text",
    "customer_phone_data" "text",
    "sys_user_id" bigint,
    "sys_user_authority_id" bigint
);


--
-- Name: COLUMN "exa_customers"."customer_name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."exa_customers"."customer_name" IS '客户名';


--
-- Name: COLUMN "exa_customers"."customer_phone_data"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."exa_customers"."customer_phone_data" IS '客户手机号';


--
-- Name: COLUMN "exa_customers"."sys_user_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."exa_customers"."sys_user_id" IS '管理ID';


--
-- Name: COLUMN "exa_customers"."sys_user_authority_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."exa_customers"."sys_user_authority_id" IS '管理角色ID';


--
-- Name: exa_customers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."exa_customers_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_customers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."exa_customers_id_seq" OWNED BY "public"."exa_customers"."id";


--
-- Name: exa_file_chunks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."exa_file_chunks" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "exa_file_id" bigint,
    "file_chunk_number" bigint,
    "file_chunk_path" "text"
);


--
-- Name: exa_file_chunks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."exa_file_chunks_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_file_chunks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."exa_file_chunks_id_seq" OWNED BY "public"."exa_file_chunks"."id";


--
-- Name: exa_file_upload_and_downloads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."exa_file_upload_and_downloads" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "name" "text",
    "class_id" bigint DEFAULT 0,
    "url" "text",
    "tag" "text",
    "key" "text"
);


--
-- Name: COLUMN "exa_file_upload_and_downloads"."name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."exa_file_upload_and_downloads"."name" IS '文件名';


--
-- Name: COLUMN "exa_file_upload_and_downloads"."class_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."exa_file_upload_and_downloads"."class_id" IS '分类id';


--
-- Name: COLUMN "exa_file_upload_and_downloads"."url"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."exa_file_upload_and_downloads"."url" IS '文件地址';


--
-- Name: COLUMN "exa_file_upload_and_downloads"."tag"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."exa_file_upload_and_downloads"."tag" IS '文件标签';


--
-- Name: COLUMN "exa_file_upload_and_downloads"."key"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."exa_file_upload_and_downloads"."key" IS '编号';


--
-- Name: exa_file_upload_and_downloads_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."exa_file_upload_and_downloads_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_file_upload_and_downloads_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."exa_file_upload_and_downloads_id_seq" OWNED BY "public"."exa_file_upload_and_downloads"."id";


--
-- Name: exa_files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."exa_files" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "file_name" "text",
    "file_md5" "text",
    "file_path" "text",
    "chunk_total" bigint,
    "is_finish" boolean
);


--
-- Name: exa_files_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."exa_files_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."exa_files_id_seq" OWNED BY "public"."exa_files"."id";


--
-- Name: gva_announcements_info; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."gva_announcements_info" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "title" "text",
    "content" "text",
    "user_id" bigint,
    "attachments" "jsonb"
);


--
-- Name: COLUMN "gva_announcements_info"."title"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."gva_announcements_info"."title" IS '公告标题';


--
-- Name: COLUMN "gva_announcements_info"."content"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."gva_announcements_info"."content" IS '公告内容';


--
-- Name: COLUMN "gva_announcements_info"."user_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."gva_announcements_info"."user_id" IS '发布者';


--
-- Name: COLUMN "gva_announcements_info"."attachments"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."gva_announcements_info"."attachments" IS '相关附件';


--
-- Name: gva_announcements_info_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."gva_announcements_info_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: gva_announcements_info_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."gva_announcements_info_id_seq" OWNED BY "public"."gva_announcements_info"."id";


--
-- Name: jwt_blacklists; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."jwt_blacklists" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "jwt" "text"
);


--
-- Name: COLUMN "jwt_blacklists"."jwt"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."jwt_blacklists"."jwt" IS 'jwt';


--
-- Name: jwt_blacklists_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."jwt_blacklists_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: jwt_blacklists_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."jwt_blacklists_id_seq" OWNED BY "public"."jwt_blacklists"."id";


--
-- Name: sys_ai_workflow_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_ai_workflow_sessions" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "user_id" bigint,
    "tab" character varying(32),
    "title" character varying(255),
    "summary" "text",
    "conversation_id" character varying(255),
    "message_id" character varying(255),
    "current_node_id" character varying(64),
    "settings" "jsonb",
    "form_data" "jsonb",
    "result_data" "jsonb",
    "messages" "jsonb"
);


--
-- Name: COLUMN "sys_ai_workflow_sessions"."user_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ai_workflow_sessions"."user_id" IS '用户ID';


--
-- Name: COLUMN "sys_ai_workflow_sessions"."tab"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ai_workflow_sessions"."tab" IS '会话类型';


--
-- Name: COLUMN "sys_ai_workflow_sessions"."title"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ai_workflow_sessions"."title" IS '会话标题';


--
-- Name: COLUMN "sys_ai_workflow_sessions"."summary"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ai_workflow_sessions"."summary" IS '摘要';


--
-- Name: COLUMN "sys_ai_workflow_sessions"."conversation_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ai_workflow_sessions"."conversation_id" IS 'Dify会话ID';


--
-- Name: COLUMN "sys_ai_workflow_sessions"."message_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ai_workflow_sessions"."message_id" IS 'Dify消息ID';


--
-- Name: COLUMN "sys_ai_workflow_sessions"."current_node_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ai_workflow_sessions"."current_node_id" IS '当前选中节点ID';


--
-- Name: COLUMN "sys_ai_workflow_sessions"."settings"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ai_workflow_sessions"."settings" IS '页面设置';


--
-- Name: COLUMN "sys_ai_workflow_sessions"."form_data"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ai_workflow_sessions"."form_data" IS '表单数据';


--
-- Name: COLUMN "sys_ai_workflow_sessions"."result_data"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ai_workflow_sessions"."result_data" IS '当前展示结果';


--
-- Name: COLUMN "sys_ai_workflow_sessions"."messages"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ai_workflow_sessions"."messages" IS '会话消息';


--
-- Name: sys_ai_workflow_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_ai_workflow_sessions_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_ai_workflow_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_ai_workflow_sessions_id_seq" OWNED BY "public"."sys_ai_workflow_sessions"."id";


--
-- Name: sys_api_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_api_tokens" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "user_id" bigint,
    "authority_id" bigint,
    "token" "text",
    "status" boolean DEFAULT true,
    "expires_at" timestamp with time zone,
    "remark" "text"
);


--
-- Name: COLUMN "sys_api_tokens"."user_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_api_tokens"."user_id" IS '用户ID';


--
-- Name: COLUMN "sys_api_tokens"."authority_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_api_tokens"."authority_id" IS '角色ID';


--
-- Name: COLUMN "sys_api_tokens"."token"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_api_tokens"."token" IS 'Token';


--
-- Name: COLUMN "sys_api_tokens"."status"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_api_tokens"."status" IS '状态';


--
-- Name: COLUMN "sys_api_tokens"."expires_at"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_api_tokens"."expires_at" IS '过期时间';


--
-- Name: COLUMN "sys_api_tokens"."remark"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_api_tokens"."remark" IS '备注';


--
-- Name: sys_api_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_api_tokens_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_api_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_api_tokens_id_seq" OWNED BY "public"."sys_api_tokens"."id";


--
-- Name: sys_apis; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_apis" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "path" "text",
    "description" "text",
    "api_group" "text",
    "method" "text" DEFAULT 'POST'::"text"
);


--
-- Name: COLUMN "sys_apis"."path"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_apis"."path" IS 'api路径';


--
-- Name: COLUMN "sys_apis"."description"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_apis"."description" IS 'api中文描述';


--
-- Name: COLUMN "sys_apis"."api_group"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_apis"."api_group" IS 'api组';


--
-- Name: COLUMN "sys_apis"."method"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_apis"."method" IS '方法';


--
-- Name: sys_apis_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_apis_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_apis_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_apis_id_seq" OWNED BY "public"."sys_apis"."id";


--
-- Name: sys_authorities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_authorities" (
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "authority_id" bigint NOT NULL,
    "authority_name" "text",
    "parent_id" bigint,
    "default_router" "text" DEFAULT 'dashboard'::"text"
);


--
-- Name: COLUMN "sys_authorities"."authority_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_authorities"."authority_id" IS '角色ID';


--
-- Name: COLUMN "sys_authorities"."authority_name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_authorities"."authority_name" IS '角色名';


--
-- Name: COLUMN "sys_authorities"."parent_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_authorities"."parent_id" IS '父角色ID';


--
-- Name: COLUMN "sys_authorities"."default_router"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_authorities"."default_router" IS '默认菜单';


--
-- Name: sys_authorities_authority_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_authorities_authority_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_authorities_authority_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_authorities_authority_id_seq" OWNED BY "public"."sys_authorities"."authority_id";


--
-- Name: sys_authority_btns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_authority_btns" (
    "authority_id" bigint,
    "sys_menu_id" bigint,
    "sys_base_menu_btn_id" bigint
);


--
-- Name: COLUMN "sys_authority_btns"."authority_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_authority_btns"."authority_id" IS '角色ID';


--
-- Name: COLUMN "sys_authority_btns"."sys_menu_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_authority_btns"."sys_menu_id" IS '菜单ID';


--
-- Name: COLUMN "sys_authority_btns"."sys_base_menu_btn_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_authority_btns"."sys_base_menu_btn_id" IS '菜单按钮ID';


--
-- Name: sys_authority_menus; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_authority_menus" (
    "sys_base_menu_id" bigint NOT NULL,
    "sys_authority_authority_id" bigint NOT NULL
);


--
-- Name: COLUMN "sys_authority_menus"."sys_authority_authority_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_authority_menus"."sys_authority_authority_id" IS '角色ID';


--
-- Name: sys_auto_code_histories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_auto_code_histories" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "table_name" "text",
    "package" "text",
    "request" "text",
    "struct_name" "text",
    "abbreviation" "text",
    "business_db" "text",
    "description" "text",
    "templates" "text",
    "injections" "text",
    "flag" bigint,
    "api_ids" "text",
    "menu_id" bigint,
    "export_template_id" bigint,
    "package_id" bigint
);


--
-- Name: COLUMN "sys_auto_code_histories"."table_name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."table_name" IS '表名';


--
-- Name: COLUMN "sys_auto_code_histories"."package"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."package" IS '模块名或插件名';


--
-- Name: COLUMN "sys_auto_code_histories"."request"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."request" IS '前端传入的结构化信息';


--
-- Name: COLUMN "sys_auto_code_histories"."struct_name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."struct_name" IS '结构体名称';


--
-- Name: COLUMN "sys_auto_code_histories"."abbreviation"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."abbreviation" IS '结构体简称';


--
-- Name: COLUMN "sys_auto_code_histories"."business_db"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."business_db" IS '业务库';


--
-- Name: COLUMN "sys_auto_code_histories"."description"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."description" IS '结构体中文名';


--
-- Name: COLUMN "sys_auto_code_histories"."templates"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."templates" IS '模板信息';


--
-- Name: COLUMN "sys_auto_code_histories"."injections"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."injections" IS '注入信息';


--
-- Name: COLUMN "sys_auto_code_histories"."flag"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."flag" IS '[0:创建,1:回滚]';


--
-- Name: COLUMN "sys_auto_code_histories"."api_ids"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."api_ids" IS '关联API ID';


--
-- Name: COLUMN "sys_auto_code_histories"."menu_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."menu_id" IS '菜单ID';


--
-- Name: COLUMN "sys_auto_code_histories"."export_template_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."export_template_id" IS '导出模板ID';


--
-- Name: COLUMN "sys_auto_code_histories"."package_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_histories"."package_id" IS '包ID';


--
-- Name: sys_auto_code_histories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_auto_code_histories_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_auto_code_histories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_auto_code_histories_id_seq" OWNED BY "public"."sys_auto_code_histories"."id";


--
-- Name: sys_auto_code_packages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_auto_code_packages" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "desc" "text",
    "label" "text",
    "template" "text",
    "package_name" "text",
    "module" "text"
);


--
-- Name: COLUMN "sys_auto_code_packages"."desc"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_packages"."desc" IS '描述';


--
-- Name: COLUMN "sys_auto_code_packages"."label"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_packages"."label" IS '显示名称';


--
-- Name: COLUMN "sys_auto_code_packages"."template"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_packages"."template" IS '模板';


--
-- Name: COLUMN "sys_auto_code_packages"."package_name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_auto_code_packages"."package_name" IS '包名';


--
-- Name: sys_auto_code_packages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_auto_code_packages_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_auto_code_packages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_auto_code_packages_id_seq" OWNED BY "public"."sys_auto_code_packages"."id";


--
-- Name: sys_base_menu_btns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_base_menu_btns" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "name" "text",
    "desc" "text",
    "sys_base_menu_id" bigint
);


--
-- Name: COLUMN "sys_base_menu_btns"."name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menu_btns"."name" IS '按钮关键key';


--
-- Name: COLUMN "sys_base_menu_btns"."sys_base_menu_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menu_btns"."sys_base_menu_id" IS '菜单ID';


--
-- Name: sys_base_menu_btns_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_base_menu_btns_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_base_menu_btns_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_base_menu_btns_id_seq" OWNED BY "public"."sys_base_menu_btns"."id";


--
-- Name: sys_base_menu_parameters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_base_menu_parameters" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "sys_base_menu_id" bigint,
    "type" "text",
    "key" "text",
    "value" "text"
);


--
-- Name: COLUMN "sys_base_menu_parameters"."type"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menu_parameters"."type" IS '地址栏携带参数为params还是query';


--
-- Name: COLUMN "sys_base_menu_parameters"."key"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menu_parameters"."key" IS '地址栏携带参数的key';


--
-- Name: COLUMN "sys_base_menu_parameters"."value"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menu_parameters"."value" IS '地址栏携带参数的值';


--
-- Name: sys_base_menu_parameters_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_base_menu_parameters_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_base_menu_parameters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_base_menu_parameters_id_seq" OWNED BY "public"."sys_base_menu_parameters"."id";


--
-- Name: sys_base_menus; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_base_menus" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "menu_level" bigint,
    "parent_id" bigint,
    "path" "text",
    "name" "text",
    "hidden" boolean,
    "component" "text",
    "sort" bigint,
    "active_name" "text",
    "keep_alive" boolean,
    "default_menu" boolean,
    "title" "text",
    "icon" "text",
    "close_tab" boolean,
    "transition_type" "text"
);


--
-- Name: COLUMN "sys_base_menus"."parent_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."parent_id" IS '父菜单ID';


--
-- Name: COLUMN "sys_base_menus"."path"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."path" IS '路由path';


--
-- Name: COLUMN "sys_base_menus"."name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."name" IS '路由name';


--
-- Name: COLUMN "sys_base_menus"."hidden"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."hidden" IS '是否在列表隐藏';


--
-- Name: COLUMN "sys_base_menus"."component"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."component" IS '对应前端文件路径';


--
-- Name: COLUMN "sys_base_menus"."sort"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."sort" IS '排序标记';


--
-- Name: COLUMN "sys_base_menus"."active_name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."active_name" IS '高亮菜单';


--
-- Name: COLUMN "sys_base_menus"."keep_alive"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."keep_alive" IS '是否缓存';


--
-- Name: COLUMN "sys_base_menus"."default_menu"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."default_menu" IS '是否是基础路由（开发中）';


--
-- Name: COLUMN "sys_base_menus"."title"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."title" IS '菜单名';


--
-- Name: COLUMN "sys_base_menus"."icon"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."icon" IS '菜单图标';


--
-- Name: COLUMN "sys_base_menus"."close_tab"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."close_tab" IS '自动关闭tab';


--
-- Name: COLUMN "sys_base_menus"."transition_type"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_base_menus"."transition_type" IS '路由切换动画';


--
-- Name: sys_base_menus_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_base_menus_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_base_menus_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_base_menus_id_seq" OWNED BY "public"."sys_base_menus"."id";


--
-- Name: sys_data_authority_id; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_data_authority_id" (
    "sys_authority_authority_id" bigint NOT NULL,
    "data_authority_id_authority_id" bigint NOT NULL
);


--
-- Name: COLUMN "sys_data_authority_id"."sys_authority_authority_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_data_authority_id"."sys_authority_authority_id" IS '角色ID';


--
-- Name: COLUMN "sys_data_authority_id"."data_authority_id_authority_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_data_authority_id"."data_authority_id_authority_id" IS '角色ID';


--
-- Name: sys_dictionaries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_dictionaries" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "name" "text",
    "type" "text",
    "status" boolean,
    "desc" "text",
    "parent_id" bigint
);


--
-- Name: COLUMN "sys_dictionaries"."name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionaries"."name" IS '字典名（中）';


--
-- Name: COLUMN "sys_dictionaries"."type"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionaries"."type" IS '字典名（英）';


--
-- Name: COLUMN "sys_dictionaries"."status"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionaries"."status" IS '状态';


--
-- Name: COLUMN "sys_dictionaries"."desc"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionaries"."desc" IS '描述';


--
-- Name: COLUMN "sys_dictionaries"."parent_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionaries"."parent_id" IS '父级字典ID';


--
-- Name: sys_dictionaries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_dictionaries_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_dictionaries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_dictionaries_id_seq" OWNED BY "public"."sys_dictionaries"."id";


--
-- Name: sys_dictionary_details; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_dictionary_details" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "label" "text",
    "value" "text",
    "extend" "text",
    "status" boolean,
    "sort" bigint,
    "sys_dictionary_id" bigint,
    "parent_id" bigint,
    "level" bigint,
    "path" "text"
);


--
-- Name: COLUMN "sys_dictionary_details"."label"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionary_details"."label" IS '展示值';


--
-- Name: COLUMN "sys_dictionary_details"."value"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionary_details"."value" IS '字典值';


--
-- Name: COLUMN "sys_dictionary_details"."extend"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionary_details"."extend" IS '扩展值';


--
-- Name: COLUMN "sys_dictionary_details"."status"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionary_details"."status" IS '启用状态';


--
-- Name: COLUMN "sys_dictionary_details"."sort"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionary_details"."sort" IS '排序标记';


--
-- Name: COLUMN "sys_dictionary_details"."sys_dictionary_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionary_details"."sys_dictionary_id" IS '关联标记';


--
-- Name: COLUMN "sys_dictionary_details"."parent_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionary_details"."parent_id" IS '父级字典详情ID';


--
-- Name: COLUMN "sys_dictionary_details"."level"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionary_details"."level" IS '层级深度';


--
-- Name: COLUMN "sys_dictionary_details"."path"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_dictionary_details"."path" IS '层级路径';


--
-- Name: sys_dictionary_details_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_dictionary_details_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_dictionary_details_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_dictionary_details_id_seq" OWNED BY "public"."sys_dictionary_details"."id";


--
-- Name: sys_error; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_error" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "form" "text",
    "info" "text",
    "level" "text",
    "solution" "text",
    "status" character varying(20) DEFAULT '未处理'::character varying
);


--
-- Name: COLUMN "sys_error"."form"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_error"."form" IS '错误来源';


--
-- Name: COLUMN "sys_error"."info"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_error"."info" IS '错误内容';


--
-- Name: COLUMN "sys_error"."level"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_error"."level" IS '日志等级';


--
-- Name: COLUMN "sys_error"."solution"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_error"."solution" IS '解决方案';


--
-- Name: COLUMN "sys_error"."status"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_error"."status" IS '处理状态';


--
-- Name: sys_error_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_error_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_error_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_error_id_seq" OWNED BY "public"."sys_error"."id";


--
-- Name: sys_export_template_condition; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_export_template_condition" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "template_id" "text",
    "from" "text",
    "column" "text",
    "operator" "text"
);


--
-- Name: COLUMN "sys_export_template_condition"."template_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_template_condition"."template_id" IS '模板标识';


--
-- Name: COLUMN "sys_export_template_condition"."from"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_template_condition"."from" IS '条件取的key';


--
-- Name: COLUMN "sys_export_template_condition"."column"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_template_condition"."column" IS '作为查询条件的字段';


--
-- Name: COLUMN "sys_export_template_condition"."operator"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_template_condition"."operator" IS '操作符';


--
-- Name: sys_export_template_condition_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_export_template_condition_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_export_template_condition_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_export_template_condition_id_seq" OWNED BY "public"."sys_export_template_condition"."id";


--
-- Name: sys_export_template_join; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_export_template_join" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "template_id" "text",
    "joins" "text",
    "table" "text",
    "on" "text"
);


--
-- Name: COLUMN "sys_export_template_join"."template_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_template_join"."template_id" IS '模板标识';


--
-- Name: COLUMN "sys_export_template_join"."joins"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_template_join"."joins" IS '关联';


--
-- Name: COLUMN "sys_export_template_join"."table"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_template_join"."table" IS '关联表';


--
-- Name: COLUMN "sys_export_template_join"."on"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_template_join"."on" IS '关联条件';


--
-- Name: sys_export_template_join_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_export_template_join_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_export_template_join_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_export_template_join_id_seq" OWNED BY "public"."sys_export_template_join"."id";


--
-- Name: sys_export_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_export_templates" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "db_name" "text",
    "name" "text",
    "table_name" "text",
    "template_id" "text",
    "template_info" "text",
    "sql" "text",
    "import_sql" "text",
    "limit" bigint,
    "order" "text"
);


--
-- Name: COLUMN "sys_export_templates"."db_name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_templates"."db_name" IS '数据库名称';


--
-- Name: COLUMN "sys_export_templates"."name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_templates"."name" IS '模板名称';


--
-- Name: COLUMN "sys_export_templates"."table_name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_templates"."table_name" IS '表名称';


--
-- Name: COLUMN "sys_export_templates"."template_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_templates"."template_id" IS '模板标识';


--
-- Name: COLUMN "sys_export_templates"."sql"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_templates"."sql" IS '自定义导出SQL';


--
-- Name: COLUMN "sys_export_templates"."import_sql"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_templates"."import_sql" IS '自定义导入SQL';


--
-- Name: COLUMN "sys_export_templates"."limit"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_templates"."limit" IS '导出限制';


--
-- Name: COLUMN "sys_export_templates"."order"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_export_templates"."order" IS '排序';


--
-- Name: sys_export_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_export_templates_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_export_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_export_templates_id_seq" OWNED BY "public"."sys_export_templates"."id";


--
-- Name: sys_ignore_apis; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_ignore_apis" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "path" "text",
    "method" "text" DEFAULT 'POST'::"text"
);


--
-- Name: COLUMN "sys_ignore_apis"."path"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ignore_apis"."path" IS 'api路径';


--
-- Name: COLUMN "sys_ignore_apis"."method"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_ignore_apis"."method" IS '方法';


--
-- Name: sys_ignore_apis_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_ignore_apis_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_ignore_apis_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_ignore_apis_id_seq" OWNED BY "public"."sys_ignore_apis"."id";


--
-- Name: sys_login_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_login_logs" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "username" "text",
    "ip" "text",
    "status" boolean,
    "error_message" "text",
    "agent" "text",
    "user_id" bigint
);


--
-- Name: COLUMN "sys_login_logs"."username"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_login_logs"."username" IS '用户名';


--
-- Name: COLUMN "sys_login_logs"."ip"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_login_logs"."ip" IS '请求ip';


--
-- Name: COLUMN "sys_login_logs"."status"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_login_logs"."status" IS '登录状态';


--
-- Name: COLUMN "sys_login_logs"."error_message"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_login_logs"."error_message" IS '错误信息';


--
-- Name: COLUMN "sys_login_logs"."agent"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_login_logs"."agent" IS '代理';


--
-- Name: COLUMN "sys_login_logs"."user_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_login_logs"."user_id" IS '用户id';


--
-- Name: sys_login_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_login_logs_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_login_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_login_logs_id_seq" OWNED BY "public"."sys_login_logs"."id";


--
-- Name: sys_operation_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_operation_records" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "ip" "text",
    "method" "text",
    "path" "text",
    "status" bigint,
    "latency" bigint,
    "agent" "text",
    "error_message" "text",
    "body" "text",
    "resp" "text",
    "user_id" bigint
);


--
-- Name: COLUMN "sys_operation_records"."ip"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_operation_records"."ip" IS '请求ip';


--
-- Name: COLUMN "sys_operation_records"."method"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_operation_records"."method" IS '请求方法';


--
-- Name: COLUMN "sys_operation_records"."path"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_operation_records"."path" IS '请求路径';


--
-- Name: COLUMN "sys_operation_records"."status"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_operation_records"."status" IS '请求状态';


--
-- Name: COLUMN "sys_operation_records"."latency"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_operation_records"."latency" IS '延迟';


--
-- Name: COLUMN "sys_operation_records"."agent"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_operation_records"."agent" IS '代理';


--
-- Name: COLUMN "sys_operation_records"."error_message"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_operation_records"."error_message" IS '错误信息';


--
-- Name: COLUMN "sys_operation_records"."body"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_operation_records"."body" IS '请求Body';


--
-- Name: COLUMN "sys_operation_records"."resp"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_operation_records"."resp" IS '响应Body';


--
-- Name: COLUMN "sys_operation_records"."user_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_operation_records"."user_id" IS '用户id';


--
-- Name: sys_operation_records_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_operation_records_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_operation_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_operation_records_id_seq" OWNED BY "public"."sys_operation_records"."id";


--
-- Name: sys_params; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_params" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "name" "text",
    "key" "text",
    "value" "text",
    "desc" "text"
);


--
-- Name: COLUMN "sys_params"."name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_params"."name" IS '参数名称';


--
-- Name: COLUMN "sys_params"."key"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_params"."key" IS '参数键';


--
-- Name: COLUMN "sys_params"."value"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_params"."value" IS '参数值';


--
-- Name: COLUMN "sys_params"."desc"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_params"."desc" IS '参数说明';


--
-- Name: sys_params_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_params_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_params_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_params_id_seq" OWNED BY "public"."sys_params"."id";


--
-- Name: sys_user_authority; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_user_authority" (
    "sys_user_id" bigint NOT NULL,
    "sys_authority_authority_id" bigint NOT NULL
);


--
-- Name: COLUMN "sys_user_authority"."sys_authority_authority_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_user_authority"."sys_authority_authority_id" IS '角色ID';


--
-- Name: sys_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_users" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "uuid" "text",
    "username" "text",
    "password" "text",
    "nick_name" "text" DEFAULT '系统用户'::"text",
    "header_img" "text" DEFAULT 'https://qmplusimg.henrongyi.top/gva_header.jpg'::"text",
    "authority_id" bigint DEFAULT 888,
    "phone" "text",
    "email" "text",
    "enable" bigint DEFAULT 1,
    "origin_setting" "jsonb"
);


--
-- Name: COLUMN "sys_users"."uuid"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_users"."uuid" IS '用户UUID';


--
-- Name: COLUMN "sys_users"."username"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_users"."username" IS '用户登录名';


--
-- Name: COLUMN "sys_users"."password"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_users"."password" IS '用户登录密码';


--
-- Name: COLUMN "sys_users"."nick_name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_users"."nick_name" IS '用户昵称';


--
-- Name: COLUMN "sys_users"."header_img"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_users"."header_img" IS '用户头像';


--
-- Name: COLUMN "sys_users"."authority_id"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_users"."authority_id" IS '用户角色ID';


--
-- Name: COLUMN "sys_users"."phone"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_users"."phone" IS '用户手机号';


--
-- Name: COLUMN "sys_users"."email"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_users"."email" IS '用户邮箱';


--
-- Name: COLUMN "sys_users"."enable"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_users"."enable" IS '用户是否被冻结 1正常 2冻结';


--
-- Name: COLUMN "sys_users"."origin_setting"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_users"."origin_setting" IS '配置';


--
-- Name: sys_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_users_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_users_id_seq" OWNED BY "public"."sys_users"."id";


--
-- Name: sys_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."sys_versions" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "version_name" character varying(255),
    "version_code" character varying(100),
    "description" character varying(500),
    "version_data" "text"
);


--
-- Name: COLUMN "sys_versions"."version_name"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_versions"."version_name" IS '版本名称';


--
-- Name: COLUMN "sys_versions"."version_code"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_versions"."version_code" IS '版本号';


--
-- Name: COLUMN "sys_versions"."description"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_versions"."description" IS '版本描述';


--
-- Name: COLUMN "sys_versions"."version_data"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN "public"."sys_versions"."version_data" IS '版本数据JSON';


--
-- Name: sys_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."sys_versions_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."sys_versions_id_seq" OWNED BY "public"."sys_versions"."id";


--
-- Name: tel_acquisition_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."tel_acquisition_tasks" (
    "id" bigint NOT NULL,
    "gva_task_id" bigint NOT NULL,
    "candidate_id" "text" NOT NULL,
    "idempotency_key" "text" NOT NULL,
    "state" "text" DEFAULT 'queued'::"text" NOT NULL,
    "stage" "text",
    "bytes_completed" bigint DEFAULT 0 NOT NULL,
    "bytes_total" bigint DEFAULT 0 NOT NULL,
    "result_json" "text",
    "error_code" "text",
    "error_message" "text",
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "started_at" timestamp with time zone,
    "completed_at" timestamp with time zone,
    CONSTRAINT "tel_acquisition_tasks_state_check" CHECK (("state" = ANY (ARRAY['queued'::"text", 'running'::"text", 'completed'::"text", 'failed'::"text", 'cancelled'::"text"])))
);


--
-- Name: tel_acquisition_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."tel_acquisition_tasks_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_acquisition_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."tel_acquisition_tasks_id_seq" OWNED BY "public"."tel_acquisition_tasks"."id";


--
-- Name: tel_dialogs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."tel_dialogs" (
    "id" bigint NOT NULL,
    "telegram_id" bigint NOT NULL,
    "access_hash" bigint,
    "name" "text" NOT NULL,
    "username" "text",
    "dialog_type" "text" NOT NULL,
    "member_count" integer,
    "is_channel" boolean DEFAULT false NOT NULL,
    "is_group" boolean DEFAULT false NOT NULL,
    "is_megagroup" boolean DEFAULT false NOT NULL,
    "is_broadcast" boolean DEFAULT false NOT NULL,
    "is_archived" boolean DEFAULT false NOT NULL,
    "is_muted" boolean DEFAULT false NOT NULL,
    "unread_count" integer DEFAULT 0 NOT NULL,
    "last_message_id" bigint,
    "last_message_at" timestamp with time zone,
    "last_read_message_id" bigint,
    "status" "text" DEFAULT 'normal'::"text" NOT NULL,
    "is_creator" boolean DEFAULT false NOT NULL,
    "created_at" timestamp with time zone NOT NULL,
    "updated_at" timestamp with time zone NOT NULL,
    "resource_compliance_status" "text" DEFAULT 'allowed'::"text" NOT NULL
);


--
-- Name: tel_dialogs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."tel_dialogs_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_dialogs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."tel_dialogs_id_seq" OWNED BY "public"."tel_dialogs"."id";


--
-- Name: tel_leave_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."tel_leave_history" (
    "id" bigint NOT NULL,
    "dialog_id" bigint,
    "telegram_id" bigint NOT NULL,
    "name" "text" NOT NULL,
    "left_at" timestamp with time zone NOT NULL,
    "result" "text" NOT NULL,
    "error_message" "text"
);


--
-- Name: tel_leave_history_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."tel_leave_history_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_leave_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."tel_leave_history_id_seq" OWNED BY "public"."tel_leave_history"."id";


--
-- Name: tel_resource_files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."tel_resource_files" (
    "id" bigint NOT NULL,
    "task_id" bigint NOT NULL,
    "telegram_chat_id" bigint NOT NULL,
    "telegram_message_id" bigint NOT NULL,
    "telegram_file_id" "text",
    "original_filename" "text" NOT NULL,
    "file_type" "text" NOT NULL,
    "file_size" bigint DEFAULT 0 NOT NULL,
    "mime_type" "text",
    "object_key" "text",
    "download_status" "text" DEFAULT 'queued'::"text" NOT NULL,
    "upload_status" "text" DEFAULT 'pending'::"text" NOT NULL,
    "bytes_downloaded" bigint DEFAULT 0 NOT NULL,
    "attempts" integer DEFAULT 0 NOT NULL,
    "error_message" "text",
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "completed_at" timestamp with time zone,
    CONSTRAINT "tel_resource_files_download_status_check" CHECK (("download_status" = ANY (ARRAY['queued'::"text", 'downloading'::"text", 'success'::"text", 'failed'::"text", 'cancelled'::"text", 'duplicate'::"text"]))),
    CONSTRAINT "tel_resource_files_upload_status_check" CHECK (("upload_status" = ANY (ARRAY['pending'::"text", 'uploading'::"text", 'success'::"text", 'failed'::"text", 'cancelled'::"text"])))
);


--
-- Name: tel_resource_files_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."tel_resource_files_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_resource_files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."tel_resource_files_id_seq" OWNED BY "public"."tel_resource_files"."id";


--
-- Name: tel_resource_sources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."tel_resource_sources" (
    "id" bigint NOT NULL,
    "file_id" bigint NOT NULL,
    "dialog_id" bigint,
    "telegram_chat_id" bigint NOT NULL,
    "telegram_message_id" bigint NOT NULL,
    "source_name" "text" NOT NULL,
    "source_username" "text",
    "message_text" "text",
    "compliance_status" "text" DEFAULT 'allowed'::"text" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


--
-- Name: tel_resource_sources_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."tel_resource_sources_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_resource_sources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."tel_resource_sources_id_seq" OWNED BY "public"."tel_resource_sources"."id";


--
-- Name: tel_resource_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."tel_resource_tasks" (
    "id" bigint NOT NULL,
    "keyword" "text" NOT NULL,
    "file_types" "text"[] DEFAULT '{}'::"text"[] NOT NULL,
    "source_dialog_ids" bigint[] DEFAULT '{}'::bigint[] NOT NULL,
    "max_downloads" integer NOT NULL,
    "status" "text" DEFAULT 'queued'::"text" NOT NULL,
    "files_found" integer DEFAULT 0 NOT NULL,
    "files_completed" integer DEFAULT 0 NOT NULL,
    "error_message" "text",
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "started_at" timestamp with time zone,
    "completed_at" timestamp with time zone,
    "cancel_requested" boolean DEFAULT false NOT NULL,
    "source_mode" "text" DEFAULT 'manual'::"text" NOT NULL,
    "source_strategy" "text" DEFAULT 'manual'::"text" NOT NULL,
    CONSTRAINT "tel_resource_tasks_max_downloads_check" CHECK ((("max_downloads" >= 1) AND ("max_downloads" <= 100))),
    CONSTRAINT "tel_resource_tasks_status_check" CHECK (("status" = ANY (ARRAY['queued'::"text", 'running'::"text", 'completed'::"text", 'failed'::"text", 'cancelled'::"text"])))
);


--
-- Name: tel_resource_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."tel_resource_tasks_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_resource_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."tel_resource_tasks_id_seq" OWNED BY "public"."tel_resource_tasks"."id";


--
-- Name: tel_scan_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."tel_scan_history" (
    "id" bigint NOT NULL,
    "started_at" timestamp with time zone NOT NULL,
    "finished_at" timestamp with time zone,
    "status" "text" NOT NULL,
    "dialogs_total" integer DEFAULT 0 NOT NULL,
    "dialogs_success" integer DEFAULT 0 NOT NULL,
    "dialogs_failed" integer DEFAULT 0 NOT NULL,
    "error_message" "text"
);


--
-- Name: tel_scan_history_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."tel_scan_history_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tel_scan_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."tel_scan_history_id_seq" OWNED BY "public"."tel_scan_history"."id";


--
-- Name: tel_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."tel_settings" (
    "key" "text" NOT NULL,
    "value" "text"
);


--
-- Name: xy_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_accounts" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "external_user_id" character varying(64),
    "device_id" character varying(128),
    "label" character varying(80) NOT NULL,
    "note" character varying(500),
    "status" character varying(32) DEFAULT 'unconfigured'::character varying NOT NULL,
    "enabled" boolean DEFAULT true NOT NULL
);


--
-- Name: xy_accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_accounts_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_accounts_id_seq" OWNED BY "public"."xy_accounts"."id";


--
-- Name: xy_agent_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_agent_config" (
    "id" bigint NOT NULL,
    "max_wall_time_seconds" bigint NOT NULL,
    "max_iterations" bigint NOT NULL,
    "max_model_calls" bigint NOT NULL,
    "max_tool_calls" bigint NOT NULL,
    "max_search_calls" bigint NOT NULL,
    "max_write_tool_calls" bigint NOT NULL,
    "max_output_tokens" bigint NOT NULL,
    "conversation_history_limit" bigint NOT NULL,
    "summary_trigger_tokens" bigint NOT NULL,
    "max_concurrent_tool_calls" bigint DEFAULT 3 NOT NULL
);


--
-- Name: xy_agent_config_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_agent_config_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_agent_config_id_seq" OWNED BY "public"."xy_agent_config"."id";


--
-- Name: xy_agent_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_agent_events" (
    "id" bigint NOT NULL,
    "owner_id" bigint NOT NULL,
    "session_id" bigint NOT NULL,
    "run_id" bigint,
    "turn_no" bigint,
    "step_no" bigint,
    "event_type" character varying(80) NOT NULL,
    "call_id" character varying(200),
    "parent_call_id" character varying(200),
    "payload" "text" NOT NULL,
    "created_at" timestamp with time zone NOT NULL
);


--
-- Name: xy_agent_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_agent_events_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_agent_events_id_seq" OWNED BY "public"."xy_agent_events"."id";


--
-- Name: xy_agent_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_agent_messages" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "session_id" bigint NOT NULL,
    "run_id" bigint,
    "sequence" bigint NOT NULL,
    "role" character varying(32) NOT NULL,
    "content" "text" NOT NULL,
    "directives" "text"
);


--
-- Name: xy_agent_messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_agent_messages_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_agent_messages_id_seq" OWNED BY "public"."xy_agent_messages"."id";


--
-- Name: xy_agent_run_tools; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_agent_run_tools" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "run_id" bigint NOT NULL,
    "call_id" character varying(200) NOT NULL,
    "mcp_server_code" character varying(64),
    "tool_name" character varying(120) NOT NULL,
    "arguments" "text",
    "status" character varying(32) NOT NULL,
    "duration_millis" bigint DEFAULT 0 NOT NULL,
    "error_code" character varying(64),
    "error_message" character varying(1000),
    "started_at" timestamp with time zone NOT NULL,
    "completed_at" timestamp with time zone,
    "result" "text"
);


--
-- Name: xy_agent_run_tools_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_agent_run_tools_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_run_tools_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_agent_run_tools_id_seq" OWNED BY "public"."xy_agent_run_tools"."id";


--
-- Name: xy_agent_runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_agent_runs" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "session_id" bigint NOT NULL,
    "idempotency_key" character varying(100),
    "model_id" bigint NOT NULL,
    "model_name" character varying(120) NOT NULL,
    "status" character varying(32) DEFAULT 'running'::character varying NOT NULL,
    "end_reason" character varying(64),
    "iterations" bigint DEFAULT 0 NOT NULL,
    "model_calls" bigint DEFAULT 0 NOT NULL,
    "tool_calls" bigint DEFAULT 0 NOT NULL,
    "input_tokens" bigint DEFAULT 0 NOT NULL,
    "output_tokens" bigint DEFAULT 0 NOT NULL,
    "wall_time_millis" bigint DEFAULT 0 NOT NULL,
    "config_snapshot" "text" NOT NULL,
    "error_message" character varying(2000),
    "started_at" timestamp with time zone NOT NULL,
    "completed_at" timestamp with time zone
);


--
-- Name: xy_agent_runs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_agent_runs_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_runs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_agent_runs_id_seq" OWNED BY "public"."xy_agent_runs"."id";


--
-- Name: xy_agent_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_agent_sessions" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "title" character varying(200) NOT NULL,
    "summary" "text",
    "last_model_id" bigint,
    "last_message_at" timestamp with time zone
);


--
-- Name: xy_agent_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_agent_sessions_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_agent_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_agent_sessions_id_seq" OWNED BY "public"."xy_agent_sessions"."id";


--
-- Name: xy_ai_models; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_ai_models" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "name" character varying(80) NOT NULL,
    "provider" character varying(40) NOT NULL,
    "base_url" character varying(500) NOT NULL,
    "api_key" "text",
    "model" character varying(120) NOT NULL,
    "enabled" boolean DEFAULT true NOT NULL,
    "is_default" boolean DEFAULT false NOT NULL,
    "supports_tools" boolean DEFAULT true NOT NULL,
    "supports_streaming" boolean DEFAULT true NOT NULL,
    "context_window" bigint DEFAULT 128000 NOT NULL,
    "max_output_tokens" bigint DEFAULT 2000 NOT NULL,
    "request_timeout_seconds" bigint DEFAULT 60 NOT NULL,
    "extra_config" "text"
);


--
-- Name: xy_ai_models_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_ai_models_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_ai_models_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_ai_models_id_seq" OWNED BY "public"."xy_ai_models"."id";


--
-- Name: xy_business_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_business_events" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "account_id" bigint,
    "product_id" bigint,
    "order_id" bigint,
    "type" character varying(32) NOT NULL,
    "status" character varying(16) NOT NULL,
    "summary" character varying(255) NOT NULL,
    "error" character varying(500)
);


--
-- Name: xy_business_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_business_events_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_business_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_business_events_id_seq" OWNED BY "public"."xy_business_events"."id";


--
-- Name: xy_conversation_takeovers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_conversation_takeovers" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "account_id" bigint NOT NULL,
    "session_id" character varying(255) NOT NULL,
    "active" boolean DEFAULT true NOT NULL
);


--
-- Name: xy_conversation_takeovers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_conversation_takeovers_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_conversation_takeovers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_conversation_takeovers_id_seq" OWNED BY "public"."xy_conversation_takeovers"."id";


--
-- Name: xy_credentials; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_credentials" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "account_id" bigint NOT NULL,
    "cookie_ciphertext" "text" NOT NULL,
    "token_ciphertext" "text",
    "token_expires_at" timestamp with time zone,
    "version" bigint DEFAULT 1 NOT NULL
);


--
-- Name: xy_credentials_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_credentials_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_credentials_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_credentials_id_seq" OWNED BY "public"."xy_credentials"."id";


--
-- Name: xy_delivery_resources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_delivery_resources" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "rule_id" bigint NOT NULL,
    "value_hash" character varying(64) NOT NULL,
    "value_ciphertext" "text" NOT NULL,
    "status" character varying(16) NOT NULL,
    "reserved_task_id" bigint,
    "consumed_at" timestamp with time zone
);


--
-- Name: xy_delivery_resources_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_delivery_resources_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_delivery_resources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_delivery_resources_id_seq" OWNED BY "public"."xy_delivery_resources"."id";


--
-- Name: xy_delivery_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_delivery_rules" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "product_id" bigint NOT NULL,
    "sku_key" character varying(128) NOT NULL,
    "sku_name" character varying(255),
    "mode" character varying(16) NOT NULL,
    "template" "text" NOT NULL,
    "enabled" boolean NOT NULL
);


--
-- Name: xy_delivery_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_delivery_rules_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_delivery_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_delivery_rules_id_seq" OWNED BY "public"."xy_delivery_rules"."id";


--
-- Name: xy_delivery_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_delivery_tasks" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "account_id" bigint NOT NULL,
    "order_id" bigint NOT NULL,
    "rule_id" bigint NOT NULL,
    "resource_id" bigint,
    "payload_ciphertext" "text" NOT NULL,
    "status" character varying(32) NOT NULL,
    "attempts" bigint DEFAULT 0 NOT NULL,
    "next_attempt_at" timestamp with time zone NOT NULL,
    "last_error" character varying(500),
    "sent_at" timestamp with time zone,
    "confirmed_at" timestamp with time zone
);


--
-- Name: xy_delivery_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_delivery_tasks_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_delivery_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_delivery_tasks_id_seq" OWNED BY "public"."xy_delivery_tasks"."id";


--
-- Name: xy_mcp_servers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_mcp_servers" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "name" character varying(80) NOT NULL,
    "code" character varying(64) NOT NULL,
    "base_url" character varying(500) NOT NULL,
    "auth_header" character varying(100),
    "auth_token" "text",
    "enabled" boolean DEFAULT true NOT NULL,
    "request_timeout_seconds" bigint DEFAULT 15 NOT NULL,
    "task_timeout_seconds" bigint DEFAULT 300 NOT NULL,
    "last_status" character varying(32) DEFAULT 'unknown'::character varying NOT NULL,
    "last_checked_at" timestamp with time zone,
    "last_error" character varying(1000)
);


--
-- Name: xy_mcp_servers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_mcp_servers_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_mcp_servers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_mcp_servers_id_seq" OWNED BY "public"."xy_mcp_servers"."id";


--
-- Name: xy_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_messages" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "account_id" bigint NOT NULL,
    "pnm_id" character varying(255) NOT NULL,
    "session_id" character varying(255),
    "content_type" bigint DEFAULT 0 NOT NULL,
    "content" "text",
    "sender_user_name" character varying(255),
    "sender_user_id" character varying(128),
    "sender_app_version" character varying(64),
    "sender_platform" character varying(64),
    "reminder_url" character varying(2048),
    "external_product_id" character varying(128),
    "external_order_id" character varying(128),
    "message_time" bigint DEFAULT 0 NOT NULL,
    "direction" character varying(16) DEFAULT 'inbound'::character varying NOT NULL,
    "source" character varying(32) DEFAULT 'remote'::character varying NOT NULL
);


--
-- Name: xy_messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_messages_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_messages_id_seq" OWNED BY "public"."xy_messages"."id";


--
-- Name: xy_order_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_order_events" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "account_id" bigint NOT NULL,
    "order_id" bigint NOT NULL,
    "pnm_id" character varying(255) NOT NULL,
    "type" character varying(32) NOT NULL,
    "event_time" bigint NOT NULL
);


--
-- Name: xy_order_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_order_events_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_order_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_order_events_id_seq" OWNED BY "public"."xy_order_events"."id";


--
-- Name: xy_orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_orders" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "account_id" bigint NOT NULL,
    "external_order_id" character varying(128) NOT NULL,
    "product_id" bigint,
    "external_product_id" character varying(128),
    "product_title" character varying(255),
    "buyer_user_id" character varying(128),
    "buyer_user_name" character varying(255),
    "session_id" character varying(255),
    "sku_key" character varying(128),
    "status" character varying(32) NOT NULL,
    "amount_cent" bigint DEFAULT 0 NOT NULL
);


--
-- Name: xy_orders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_orders_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_orders_id_seq" OWNED BY "public"."xy_orders"."id";


--
-- Name: xy_product_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_product_settings" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "product_id" bigint NOT NULL,
    "auto_reply_enabled" boolean DEFAULT false NOT NULL,
    "auto_delivery_enabled" boolean DEFAULT false NOT NULL,
    "auto_relist_enabled" boolean DEFAULT false NOT NULL,
    "reply_delay_seconds" bigint DEFAULT 0 NOT NULL,
    "relist_delay_seconds" bigint DEFAULT 60 NOT NULL,
    "relist_description" "text",
    "relist_image_urls" "text",
    "relist_address" character varying(255),
    "virtual_product" boolean DEFAULT false NOT NULL
);


--
-- Name: xy_product_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_product_settings_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_product_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_product_settings_id_seq" OWNED BY "public"."xy_product_settings"."id";


--
-- Name: xy_products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_products" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "account_id" bigint NOT NULL,
    "external_product_id" character varying(64) NOT NULL,
    "title" character varying(255) NOT NULL,
    "price_cent" bigint DEFAULT 0 NOT NULL,
    "status" character varying(32) DEFAULT 'on_sale'::character varying NOT NULL,
    "image_url" character varying(2048),
    "publish_snapshot" "text",
    "source_json" "text"
);


--
-- Name: xy_products_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_products_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_products_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_products_id_seq" OWNED BY "public"."xy_products"."id";


--
-- Name: xy_relist_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_relist_tasks" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "account_id" bigint NOT NULL,
    "order_id" bigint NOT NULL,
    "product_id" bigint NOT NULL,
    "status" character varying(32) NOT NULL,
    "attempts" bigint DEFAULT 0 NOT NULL,
    "next_attempt_at" timestamp with time zone NOT NULL,
    "new_external_product_id" character varying(128),
    "last_error" character varying(500)
);


--
-- Name: xy_relist_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_relist_tasks_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_relist_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_relist_tasks_id_seq" OWNED BY "public"."xy_relist_tasks"."id";


--
-- Name: xy_reply_contents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_reply_contents" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "rule_id" bigint NOT NULL,
    "text" "text" NOT NULL
);


--
-- Name: xy_reply_contents_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_reply_contents_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_reply_contents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_reply_contents_id_seq" OWNED BY "public"."xy_reply_contents"."id";


--
-- Name: xy_reply_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_reply_rules" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "product_id" bigint NOT NULL,
    "type" character varying(16) NOT NULL,
    "keyword" character varying(255) NOT NULL,
    "priority" bigint DEFAULT 0 NOT NULL,
    "enabled" boolean NOT NULL
);


--
-- Name: xy_reply_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_reply_rules_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_reply_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_reply_rules_id_seq" OWNED BY "public"."xy_reply_rules"."id";


--
-- Name: xy_reply_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_reply_tasks" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "account_id" bigint NOT NULL,
    "message_id" bigint NOT NULL,
    "product_id" bigint NOT NULL,
    "external_product_id" character varying(128) NOT NULL,
    "rule_id" bigint NOT NULL,
    "session_id" character varying(255) NOT NULL,
    "recipient_id" character varying(128) NOT NULL,
    "reply_text" "text" NOT NULL,
    "status" character varying(16) NOT NULL,
    "error" character varying(500),
    "scheduled_at" timestamp with time zone NOT NULL,
    "completed_at" timestamp with time zone
);


--
-- Name: xy_reply_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_reply_tasks_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_reply_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_reply_tasks_id_seq" OWNED BY "public"."xy_reply_tasks"."id";


--
-- Name: xy_resource_candidates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_resource_candidates" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "search_id" bigint NOT NULL,
    "mcp_server_code" character varying(64) NOT NULL,
    "external_id" character varying(1000) NOT NULL,
    "name" character varying(500) NOT NULL,
    "size" bigint DEFAULT 0 NOT NULL,
    "mime_type" character varying(200),
    "source" "text" NOT NULL,
    "preview" "text",
    "license" character varying(100) DEFAULT 'unknown'::character varying NOT NULL,
    "risk" character varying(1000),
    "expires_at" timestamp with time zone
);


--
-- Name: xy_resource_candidates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_resource_candidates_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_resource_candidates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_resource_candidates_id_seq" OWNED BY "public"."xy_resource_candidates"."id";


--
-- Name: xy_resource_searches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_resource_searches" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "query" character varying(500) NOT NULL,
    "status" character varying(32) DEFAULT 'running'::character varying NOT NULL,
    "directives" "text"
);


--
-- Name: xy_resource_searches_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_resource_searches_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_resource_searches_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_resource_searches_id_seq" OWNED BY "public"."xy_resource_searches"."id";


--
-- Name: xy_resource_sources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_resource_sources" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "resource_id" bigint NOT NULL,
    "mcp_server_code" character varying(64) NOT NULL,
    "external_id" character varying(1000) NOT NULL,
    "source_name" character varying(500),
    "source_reference" character varying(1000),
    "metadata" "text"
);


--
-- Name: xy_resource_sources_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_resource_sources_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_resource_sources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_resource_sources_id_seq" OWNED BY "public"."xy_resource_sources"."id";


--
-- Name: xy_resource_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_resource_tasks" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "agent_run_id" bigint,
    "candidate_id" bigint NOT NULL,
    "mcp_server_code" character varying(64) NOT NULL,
    "external_task_id" character varying(200),
    "idempotency_key" character varying(100) NOT NULL,
    "state" character varying(32) DEFAULT 'queued'::character varying NOT NULL,
    "stage" character varying(64),
    "progress_percent" bigint DEFAULT 0 NOT NULL,
    "progress_done" bigint DEFAULT 0 NOT NULL,
    "progress_total" bigint DEFAULT 0 NOT NULL,
    "progress_unit" character varying(32),
    "result" "text",
    "error_code" character varying(64),
    "error_message" character varying(2000),
    "started_at" timestamp with time zone,
    "completed_at" timestamp with time zone,
    "delivery_token_hash" character varying(64),
    "delivery_expires_at" timestamp with time zone,
    "upload_object_key" character varying(1000)
);


--
-- Name: xy_resource_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_resource_tasks_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_resource_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_resource_tasks_id_seq" OWNED BY "public"."xy_resource_tasks"."id";


--
-- Name: xy_resources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_resources" (
    "id" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "owner_id" bigint NOT NULL,
    "task_id" bigint NOT NULL,
    "name" character varying(500) NOT NULL,
    "size" bigint NOT NULL,
    "mime_type" character varying(200),
    "sha256" character varying(64),
    "bucket" character varying(200) NOT NULL,
    "object_key" character varying(1000) NOT NULL,
    "state" character varying(32) DEFAULT 'available'::character varying NOT NULL
);


--
-- Name: xy_resources_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE "public"."xy_resources_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: xy_resources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE "public"."xy_resources_id_seq" OWNED BY "public"."xy_resources"."id";


--
-- Name: xy_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE "public"."xy_settings" (
    "key" character varying(128) NOT NULL,
    "value" "text" NOT NULL
);


--
-- Name: casbin_rule id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."casbin_rule" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."casbin_rule_id_seq"'::"regclass");


--
-- Name: exa_attachment_category id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."exa_attachment_category" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."exa_attachment_category_id_seq"'::"regclass");


--
-- Name: exa_customers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."exa_customers" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."exa_customers_id_seq"'::"regclass");


--
-- Name: exa_file_chunks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."exa_file_chunks" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."exa_file_chunks_id_seq"'::"regclass");


--
-- Name: exa_file_upload_and_downloads id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."exa_file_upload_and_downloads" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."exa_file_upload_and_downloads_id_seq"'::"regclass");


--
-- Name: exa_files id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."exa_files" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."exa_files_id_seq"'::"regclass");


--
-- Name: gva_announcements_info id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."gva_announcements_info" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."gva_announcements_info_id_seq"'::"regclass");


--
-- Name: jwt_blacklists id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."jwt_blacklists" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."jwt_blacklists_id_seq"'::"regclass");


--
-- Name: sys_ai_workflow_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_ai_workflow_sessions" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_ai_workflow_sessions_id_seq"'::"regclass");


--
-- Name: sys_api_tokens id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_api_tokens" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_api_tokens_id_seq"'::"regclass");


--
-- Name: sys_apis id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_apis" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_apis_id_seq"'::"regclass");


--
-- Name: sys_authorities authority_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_authorities" ALTER COLUMN "authority_id" SET DEFAULT "nextval"('"public"."sys_authorities_authority_id_seq"'::"regclass");


--
-- Name: sys_auto_code_histories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_auto_code_histories" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_auto_code_histories_id_seq"'::"regclass");


--
-- Name: sys_auto_code_packages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_auto_code_packages" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_auto_code_packages_id_seq"'::"regclass");


--
-- Name: sys_base_menu_btns id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_base_menu_btns" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_base_menu_btns_id_seq"'::"regclass");


--
-- Name: sys_base_menu_parameters id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_base_menu_parameters" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_base_menu_parameters_id_seq"'::"regclass");


--
-- Name: sys_base_menus id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_base_menus" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_base_menus_id_seq"'::"regclass");


--
-- Name: sys_dictionaries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_dictionaries" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_dictionaries_id_seq"'::"regclass");


--
-- Name: sys_dictionary_details id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_dictionary_details" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_dictionary_details_id_seq"'::"regclass");


--
-- Name: sys_error id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_error" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_error_id_seq"'::"regclass");


--
-- Name: sys_export_template_condition id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_export_template_condition" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_export_template_condition_id_seq"'::"regclass");


--
-- Name: sys_export_template_join id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_export_template_join" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_export_template_join_id_seq"'::"regclass");


--
-- Name: sys_export_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_export_templates" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_export_templates_id_seq"'::"regclass");


--
-- Name: sys_ignore_apis id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_ignore_apis" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_ignore_apis_id_seq"'::"regclass");


--
-- Name: sys_login_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_login_logs" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_login_logs_id_seq"'::"regclass");


--
-- Name: sys_operation_records id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_operation_records" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_operation_records_id_seq"'::"regclass");


--
-- Name: sys_params id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_params" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_params_id_seq"'::"regclass");


--
-- Name: sys_users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_users" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_users_id_seq"'::"regclass");


--
-- Name: sys_versions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_versions" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sys_versions_id_seq"'::"regclass");


--
-- Name: tel_acquisition_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_acquisition_tasks" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."tel_acquisition_tasks_id_seq"'::"regclass");


--
-- Name: tel_dialogs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_dialogs" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."tel_dialogs_id_seq"'::"regclass");


--
-- Name: tel_leave_history id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_leave_history" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."tel_leave_history_id_seq"'::"regclass");


--
-- Name: tel_resource_files id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_resource_files" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."tel_resource_files_id_seq"'::"regclass");


--
-- Name: tel_resource_sources id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_resource_sources" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."tel_resource_sources_id_seq"'::"regclass");


--
-- Name: tel_resource_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_resource_tasks" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."tel_resource_tasks_id_seq"'::"regclass");


--
-- Name: tel_scan_history id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_scan_history" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."tel_scan_history_id_seq"'::"regclass");


--
-- Name: xy_accounts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_accounts" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_accounts_id_seq"'::"regclass");


--
-- Name: xy_agent_config id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_config" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_agent_config_id_seq"'::"regclass");


--
-- Name: xy_agent_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_events" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_agent_events_id_seq"'::"regclass");


--
-- Name: xy_agent_messages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_messages" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_agent_messages_id_seq"'::"regclass");


--
-- Name: xy_agent_run_tools id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_run_tools" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_agent_run_tools_id_seq"'::"regclass");


--
-- Name: xy_agent_runs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_runs" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_agent_runs_id_seq"'::"regclass");


--
-- Name: xy_agent_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_sessions" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_agent_sessions_id_seq"'::"regclass");


--
-- Name: xy_ai_models id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_ai_models" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_ai_models_id_seq"'::"regclass");


--
-- Name: xy_business_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_business_events" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_business_events_id_seq"'::"regclass");


--
-- Name: xy_conversation_takeovers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_conversation_takeovers" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_conversation_takeovers_id_seq"'::"regclass");


--
-- Name: xy_credentials id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_credentials" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_credentials_id_seq"'::"regclass");


--
-- Name: xy_delivery_resources id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_delivery_resources" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_delivery_resources_id_seq"'::"regclass");


--
-- Name: xy_delivery_rules id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_delivery_rules" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_delivery_rules_id_seq"'::"regclass");


--
-- Name: xy_delivery_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_delivery_tasks" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_delivery_tasks_id_seq"'::"regclass");


--
-- Name: xy_mcp_servers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_mcp_servers" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_mcp_servers_id_seq"'::"regclass");


--
-- Name: xy_messages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_messages" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_messages_id_seq"'::"regclass");


--
-- Name: xy_order_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_order_events" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_order_events_id_seq"'::"regclass");


--
-- Name: xy_orders id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_orders" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_orders_id_seq"'::"regclass");


--
-- Name: xy_product_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_product_settings" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_product_settings_id_seq"'::"regclass");


--
-- Name: xy_products id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_products" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_products_id_seq"'::"regclass");


--
-- Name: xy_relist_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_relist_tasks" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_relist_tasks_id_seq"'::"regclass");


--
-- Name: xy_reply_contents id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_reply_contents" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_reply_contents_id_seq"'::"regclass");


--
-- Name: xy_reply_rules id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_reply_rules" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_reply_rules_id_seq"'::"regclass");


--
-- Name: xy_reply_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_reply_tasks" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_reply_tasks_id_seq"'::"regclass");


--
-- Name: xy_resource_candidates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_resource_candidates" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_resource_candidates_id_seq"'::"regclass");


--
-- Name: xy_resource_searches id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_resource_searches" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_resource_searches_id_seq"'::"regclass");


--
-- Name: xy_resource_sources id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_resource_sources" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_resource_sources_id_seq"'::"regclass");


--
-- Name: xy_resource_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_resource_tasks" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_resource_tasks_id_seq"'::"regclass");


--
-- Name: xy_resources id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_resources" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."xy_resources_id_seq"'::"regclass");


--
-- Name: casbin_rule casbin_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."casbin_rule"
    ADD CONSTRAINT "casbin_rule_pkey" PRIMARY KEY ("id");


--
-- Name: exa_attachment_category exa_attachment_category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."exa_attachment_category"
    ADD CONSTRAINT "exa_attachment_category_pkey" PRIMARY KEY ("id");


--
-- Name: exa_customers exa_customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."exa_customers"
    ADD CONSTRAINT "exa_customers_pkey" PRIMARY KEY ("id");


--
-- Name: exa_file_chunks exa_file_chunks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."exa_file_chunks"
    ADD CONSTRAINT "exa_file_chunks_pkey" PRIMARY KEY ("id");


--
-- Name: exa_file_upload_and_downloads exa_file_upload_and_downloads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."exa_file_upload_and_downloads"
    ADD CONSTRAINT "exa_file_upload_and_downloads_pkey" PRIMARY KEY ("id");


--
-- Name: exa_files exa_files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."exa_files"
    ADD CONSTRAINT "exa_files_pkey" PRIMARY KEY ("id");


--
-- Name: gva_announcements_info gva_announcements_info_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."gva_announcements_info"
    ADD CONSTRAINT "gva_announcements_info_pkey" PRIMARY KEY ("id");


--
-- Name: jwt_blacklists jwt_blacklists_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."jwt_blacklists"
    ADD CONSTRAINT "jwt_blacklists_pkey" PRIMARY KEY ("id");


--
-- Name: sys_ai_workflow_sessions sys_ai_workflow_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_ai_workflow_sessions"
    ADD CONSTRAINT "sys_ai_workflow_sessions_pkey" PRIMARY KEY ("id");


--
-- Name: sys_api_tokens sys_api_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_api_tokens"
    ADD CONSTRAINT "sys_api_tokens_pkey" PRIMARY KEY ("id");


--
-- Name: sys_apis sys_apis_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_apis"
    ADD CONSTRAINT "sys_apis_pkey" PRIMARY KEY ("id");


--
-- Name: sys_authority_menus sys_authority_menus_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_authority_menus"
    ADD CONSTRAINT "sys_authority_menus_pkey" PRIMARY KEY ("sys_base_menu_id", "sys_authority_authority_id");


--
-- Name: sys_auto_code_histories sys_auto_code_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_auto_code_histories"
    ADD CONSTRAINT "sys_auto_code_histories_pkey" PRIMARY KEY ("id");


--
-- Name: sys_auto_code_packages sys_auto_code_packages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_auto_code_packages"
    ADD CONSTRAINT "sys_auto_code_packages_pkey" PRIMARY KEY ("id");


--
-- Name: sys_base_menu_btns sys_base_menu_btns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_base_menu_btns"
    ADD CONSTRAINT "sys_base_menu_btns_pkey" PRIMARY KEY ("id");


--
-- Name: sys_base_menu_parameters sys_base_menu_parameters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_base_menu_parameters"
    ADD CONSTRAINT "sys_base_menu_parameters_pkey" PRIMARY KEY ("id");


--
-- Name: sys_base_menus sys_base_menus_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_base_menus"
    ADD CONSTRAINT "sys_base_menus_pkey" PRIMARY KEY ("id");


--
-- Name: sys_data_authority_id sys_data_authority_id_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_data_authority_id"
    ADD CONSTRAINT "sys_data_authority_id_pkey" PRIMARY KEY ("sys_authority_authority_id", "data_authority_id_authority_id");


--
-- Name: sys_dictionaries sys_dictionaries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_dictionaries"
    ADD CONSTRAINT "sys_dictionaries_pkey" PRIMARY KEY ("id");


--
-- Name: sys_dictionary_details sys_dictionary_details_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_dictionary_details"
    ADD CONSTRAINT "sys_dictionary_details_pkey" PRIMARY KEY ("id");


--
-- Name: sys_error sys_error_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_error"
    ADD CONSTRAINT "sys_error_pkey" PRIMARY KEY ("id");


--
-- Name: sys_export_template_condition sys_export_template_condition_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_export_template_condition"
    ADD CONSTRAINT "sys_export_template_condition_pkey" PRIMARY KEY ("id");


--
-- Name: sys_export_template_join sys_export_template_join_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_export_template_join"
    ADD CONSTRAINT "sys_export_template_join_pkey" PRIMARY KEY ("id");


--
-- Name: sys_export_templates sys_export_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_export_templates"
    ADD CONSTRAINT "sys_export_templates_pkey" PRIMARY KEY ("id");


--
-- Name: sys_ignore_apis sys_ignore_apis_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_ignore_apis"
    ADD CONSTRAINT "sys_ignore_apis_pkey" PRIMARY KEY ("id");


--
-- Name: sys_login_logs sys_login_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_login_logs"
    ADD CONSTRAINT "sys_login_logs_pkey" PRIMARY KEY ("id");


--
-- Name: sys_operation_records sys_operation_records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_operation_records"
    ADD CONSTRAINT "sys_operation_records_pkey" PRIMARY KEY ("id");


--
-- Name: sys_params sys_params_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_params"
    ADD CONSTRAINT "sys_params_pkey" PRIMARY KEY ("id");


--
-- Name: sys_user_authority sys_user_authority_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_user_authority"
    ADD CONSTRAINT "sys_user_authority_pkey" PRIMARY KEY ("sys_user_id", "sys_authority_authority_id");


--
-- Name: sys_users sys_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_users"
    ADD CONSTRAINT "sys_users_pkey" PRIMARY KEY ("id");


--
-- Name: sys_versions sys_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_versions"
    ADD CONSTRAINT "sys_versions_pkey" PRIMARY KEY ("id");


--
-- Name: tel_acquisition_tasks tel_acquisition_tasks_gva_task_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_acquisition_tasks"
    ADD CONSTRAINT "tel_acquisition_tasks_gva_task_id_key" UNIQUE ("gva_task_id");


--
-- Name: tel_acquisition_tasks tel_acquisition_tasks_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_acquisition_tasks"
    ADD CONSTRAINT "tel_acquisition_tasks_idempotency_key_key" UNIQUE ("idempotency_key");


--
-- Name: tel_acquisition_tasks tel_acquisition_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_acquisition_tasks"
    ADD CONSTRAINT "tel_acquisition_tasks_pkey" PRIMARY KEY ("id");


--
-- Name: tel_dialogs tel_dialogs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_dialogs"
    ADD CONSTRAINT "tel_dialogs_pkey" PRIMARY KEY ("id");


--
-- Name: tel_dialogs tel_dialogs_telegram_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_dialogs"
    ADD CONSTRAINT "tel_dialogs_telegram_id_key" UNIQUE ("telegram_id");


--
-- Name: tel_leave_history tel_leave_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_leave_history"
    ADD CONSTRAINT "tel_leave_history_pkey" PRIMARY KEY ("id");


--
-- Name: tel_resource_files tel_resource_files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_resource_files"
    ADD CONSTRAINT "tel_resource_files_pkey" PRIMARY KEY ("id");


--
-- Name: tel_resource_files tel_resource_files_telegram_chat_id_telegram_message_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_resource_files"
    ADD CONSTRAINT "tel_resource_files_telegram_chat_id_telegram_message_id_key" UNIQUE ("telegram_chat_id", "telegram_message_id");


--
-- Name: tel_resource_sources tel_resource_sources_file_id_telegram_chat_id_telegram_mess_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_resource_sources"
    ADD CONSTRAINT "tel_resource_sources_file_id_telegram_chat_id_telegram_mess_key" UNIQUE ("file_id", "telegram_chat_id", "telegram_message_id");


--
-- Name: tel_resource_sources tel_resource_sources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_resource_sources"
    ADD CONSTRAINT "tel_resource_sources_pkey" PRIMARY KEY ("id");


--
-- Name: tel_resource_tasks tel_resource_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_resource_tasks"
    ADD CONSTRAINT "tel_resource_tasks_pkey" PRIMARY KEY ("id");


--
-- Name: tel_scan_history tel_scan_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_scan_history"
    ADD CONSTRAINT "tel_scan_history_pkey" PRIMARY KEY ("id");


--
-- Name: tel_settings tel_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_settings"
    ADD CONSTRAINT "tel_settings_pkey" PRIMARY KEY ("key");


--
-- Name: sys_authorities uni_sys_authorities_authority_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."sys_authorities"
    ADD CONSTRAINT "uni_sys_authorities_authority_id" PRIMARY KEY ("authority_id");


--
-- Name: xy_accounts xy_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_accounts"
    ADD CONSTRAINT "xy_accounts_pkey" PRIMARY KEY ("id");


--
-- Name: xy_agent_config xy_agent_config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_config"
    ADD CONSTRAINT "xy_agent_config_pkey" PRIMARY KEY ("id");


--
-- Name: xy_agent_events xy_agent_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_events"
    ADD CONSTRAINT "xy_agent_events_pkey" PRIMARY KEY ("id");


--
-- Name: xy_agent_messages xy_agent_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_messages"
    ADD CONSTRAINT "xy_agent_messages_pkey" PRIMARY KEY ("id");


--
-- Name: xy_agent_run_tools xy_agent_run_tools_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_run_tools"
    ADD CONSTRAINT "xy_agent_run_tools_pkey" PRIMARY KEY ("id");


--
-- Name: xy_agent_runs xy_agent_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_runs"
    ADD CONSTRAINT "xy_agent_runs_pkey" PRIMARY KEY ("id");


--
-- Name: xy_agent_sessions xy_agent_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_agent_sessions"
    ADD CONSTRAINT "xy_agent_sessions_pkey" PRIMARY KEY ("id");


--
-- Name: xy_ai_models xy_ai_models_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_ai_models"
    ADD CONSTRAINT "xy_ai_models_pkey" PRIMARY KEY ("id");


--
-- Name: xy_business_events xy_business_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_business_events"
    ADD CONSTRAINT "xy_business_events_pkey" PRIMARY KEY ("id");


--
-- Name: xy_conversation_takeovers xy_conversation_takeovers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_conversation_takeovers"
    ADD CONSTRAINT "xy_conversation_takeovers_pkey" PRIMARY KEY ("id");


--
-- Name: xy_credentials xy_credentials_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_credentials"
    ADD CONSTRAINT "xy_credentials_pkey" PRIMARY KEY ("id");


--
-- Name: xy_delivery_resources xy_delivery_resources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_delivery_resources"
    ADD CONSTRAINT "xy_delivery_resources_pkey" PRIMARY KEY ("id");


--
-- Name: xy_delivery_rules xy_delivery_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_delivery_rules"
    ADD CONSTRAINT "xy_delivery_rules_pkey" PRIMARY KEY ("id");


--
-- Name: xy_delivery_tasks xy_delivery_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_delivery_tasks"
    ADD CONSTRAINT "xy_delivery_tasks_pkey" PRIMARY KEY ("id");


--
-- Name: xy_mcp_servers xy_mcp_servers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_mcp_servers"
    ADD CONSTRAINT "xy_mcp_servers_pkey" PRIMARY KEY ("id");


--
-- Name: xy_messages xy_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_messages"
    ADD CONSTRAINT "xy_messages_pkey" PRIMARY KEY ("id");


--
-- Name: xy_order_events xy_order_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_order_events"
    ADD CONSTRAINT "xy_order_events_pkey" PRIMARY KEY ("id");


--
-- Name: xy_orders xy_orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_orders"
    ADD CONSTRAINT "xy_orders_pkey" PRIMARY KEY ("id");


--
-- Name: xy_product_settings xy_product_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_product_settings"
    ADD CONSTRAINT "xy_product_settings_pkey" PRIMARY KEY ("id");


--
-- Name: xy_products xy_products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_products"
    ADD CONSTRAINT "xy_products_pkey" PRIMARY KEY ("id");


--
-- Name: xy_relist_tasks xy_relist_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_relist_tasks"
    ADD CONSTRAINT "xy_relist_tasks_pkey" PRIMARY KEY ("id");


--
-- Name: xy_reply_contents xy_reply_contents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_reply_contents"
    ADD CONSTRAINT "xy_reply_contents_pkey" PRIMARY KEY ("id");


--
-- Name: xy_reply_rules xy_reply_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_reply_rules"
    ADD CONSTRAINT "xy_reply_rules_pkey" PRIMARY KEY ("id");


--
-- Name: xy_reply_tasks xy_reply_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_reply_tasks"
    ADD CONSTRAINT "xy_reply_tasks_pkey" PRIMARY KEY ("id");


--
-- Name: xy_resource_candidates xy_resource_candidates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_resource_candidates"
    ADD CONSTRAINT "xy_resource_candidates_pkey" PRIMARY KEY ("id");


--
-- Name: xy_resource_searches xy_resource_searches_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_resource_searches"
    ADD CONSTRAINT "xy_resource_searches_pkey" PRIMARY KEY ("id");


--
-- Name: xy_resource_sources xy_resource_sources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_resource_sources"
    ADD CONSTRAINT "xy_resource_sources_pkey" PRIMARY KEY ("id");


--
-- Name: xy_resource_tasks xy_resource_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_resource_tasks"
    ADD CONSTRAINT "xy_resource_tasks_pkey" PRIMARY KEY ("id");


--
-- Name: xy_resources xy_resources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_resources"
    ADD CONSTRAINT "xy_resources_pkey" PRIMARY KEY ("id");


--
-- Name: xy_settings xy_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."xy_settings"
    ADD CONSTRAINT "xy_settings_pkey" PRIMARY KEY ("key");


--
-- Name: idx_casbin_rule; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_casbin_rule" ON "public"."casbin_rule" USING "btree" ("ptype", "v0", "v1", "v2", "v3", "v4", "v5");


--
-- Name: idx_exa_attachment_category_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_exa_attachment_category_deleted_at" ON "public"."exa_attachment_category" USING "btree" ("deleted_at");


--
-- Name: idx_exa_customers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_exa_customers_deleted_at" ON "public"."exa_customers" USING "btree" ("deleted_at");


--
-- Name: idx_exa_file_chunks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_exa_file_chunks_deleted_at" ON "public"."exa_file_chunks" USING "btree" ("deleted_at");


--
-- Name: idx_exa_file_upload_and_downloads_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_exa_file_upload_and_downloads_deleted_at" ON "public"."exa_file_upload_and_downloads" USING "btree" ("deleted_at");


--
-- Name: idx_exa_files_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_exa_files_deleted_at" ON "public"."exa_files" USING "btree" ("deleted_at");


--
-- Name: idx_gva_announcements_info_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_gva_announcements_info_deleted_at" ON "public"."gva_announcements_info" USING "btree" ("deleted_at");


--
-- Name: idx_jwt_blacklists_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_jwt_blacklists_deleted_at" ON "public"."jwt_blacklists" USING "btree" ("deleted_at");


--
-- Name: idx_sys_ai_workflow_sessions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_ai_workflow_sessions_deleted_at" ON "public"."sys_ai_workflow_sessions" USING "btree" ("deleted_at");


--
-- Name: idx_sys_ai_workflow_sessions_tab; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_ai_workflow_sessions_tab" ON "public"."sys_ai_workflow_sessions" USING "btree" ("tab");


--
-- Name: idx_sys_ai_workflow_sessions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_ai_workflow_sessions_user_id" ON "public"."sys_ai_workflow_sessions" USING "btree" ("user_id");


--
-- Name: idx_sys_api_tokens_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_api_tokens_deleted_at" ON "public"."sys_api_tokens" USING "btree" ("deleted_at");


--
-- Name: idx_sys_apis_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_apis_deleted_at" ON "public"."sys_apis" USING "btree" ("deleted_at");


--
-- Name: idx_sys_auto_code_histories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_auto_code_histories_deleted_at" ON "public"."sys_auto_code_histories" USING "btree" ("deleted_at");


--
-- Name: idx_sys_auto_code_packages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_auto_code_packages_deleted_at" ON "public"."sys_auto_code_packages" USING "btree" ("deleted_at");


--
-- Name: idx_sys_base_menu_btns_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_base_menu_btns_deleted_at" ON "public"."sys_base_menu_btns" USING "btree" ("deleted_at");


--
-- Name: idx_sys_base_menu_parameters_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_base_menu_parameters_deleted_at" ON "public"."sys_base_menu_parameters" USING "btree" ("deleted_at");


--
-- Name: idx_sys_base_menus_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_base_menus_deleted_at" ON "public"."sys_base_menus" USING "btree" ("deleted_at");


--
-- Name: idx_sys_dictionaries_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_dictionaries_deleted_at" ON "public"."sys_dictionaries" USING "btree" ("deleted_at");


--
-- Name: idx_sys_dictionary_details_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_dictionary_details_deleted_at" ON "public"."sys_dictionary_details" USING "btree" ("deleted_at");


--
-- Name: idx_sys_error_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_error_deleted_at" ON "public"."sys_error" USING "btree" ("deleted_at");


--
-- Name: idx_sys_export_template_condition_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_export_template_condition_deleted_at" ON "public"."sys_export_template_condition" USING "btree" ("deleted_at");


--
-- Name: idx_sys_export_template_join_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_export_template_join_deleted_at" ON "public"."sys_export_template_join" USING "btree" ("deleted_at");


--
-- Name: idx_sys_export_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_export_templates_deleted_at" ON "public"."sys_export_templates" USING "btree" ("deleted_at");


--
-- Name: idx_sys_ignore_apis_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_ignore_apis_deleted_at" ON "public"."sys_ignore_apis" USING "btree" ("deleted_at");


--
-- Name: idx_sys_login_logs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_login_logs_deleted_at" ON "public"."sys_login_logs" USING "btree" ("deleted_at");


--
-- Name: idx_sys_operation_records_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_operation_records_deleted_at" ON "public"."sys_operation_records" USING "btree" ("deleted_at");


--
-- Name: idx_sys_params_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_params_deleted_at" ON "public"."sys_params" USING "btree" ("deleted_at");


--
-- Name: idx_sys_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_users_deleted_at" ON "public"."sys_users" USING "btree" ("deleted_at");


--
-- Name: idx_sys_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_users_username" ON "public"."sys_users" USING "btree" ("username");


--
-- Name: idx_sys_users_uuid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_users_uuid" ON "public"."sys_users" USING "btree" ("uuid");


--
-- Name: idx_sys_versions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_sys_versions_deleted_at" ON "public"."sys_versions" USING "btree" ("deleted_at");


--
-- Name: idx_tel_resource_files_filter; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_tel_resource_files_filter" ON "public"."tel_resource_files" USING "btree" ("download_status", "file_type");


--
-- Name: idx_tel_resource_files_task; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_tel_resource_files_task" ON "public"."tel_resource_files" USING "btree" ("task_id");


--
-- Name: idx_tel_resource_files_telegram_file; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_tel_resource_files_telegram_file" ON "public"."tel_resource_files" USING "btree" ("telegram_file_id") WHERE ("telegram_file_id" IS NOT NULL);


--
-- Name: idx_xy_accounts_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_accounts_deleted_at" ON "public"."xy_accounts" USING "btree" ("deleted_at");


--
-- Name: idx_xy_accounts_external_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_accounts_external_user_id" ON "public"."xy_accounts" USING "btree" ("external_user_id");


--
-- Name: idx_xy_accounts_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_accounts_owner" ON "public"."xy_accounts" USING "btree" ("owner_id");


--
-- Name: idx_xy_agent_events_call_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_events_call_id" ON "public"."xy_agent_events" USING "btree" ("call_id");


--
-- Name: idx_xy_agent_events_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_events_created_at" ON "public"."xy_agent_events" USING "btree" ("created_at");


--
-- Name: idx_xy_agent_events_owner_session; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_events_owner_session" ON "public"."xy_agent_events" USING "btree" ("owner_id", "session_id");


--
-- Name: idx_xy_agent_events_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_events_run_id" ON "public"."xy_agent_events" USING "btree" ("run_id");


--
-- Name: idx_xy_agent_events_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_events_session_id" ON "public"."xy_agent_events" USING "btree" ("session_id");


--
-- Name: idx_xy_agent_events_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_events_type" ON "public"."xy_agent_events" USING "btree" ("event_type");


--
-- Name: idx_xy_agent_messages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_messages_deleted_at" ON "public"."xy_agent_messages" USING "btree" ("deleted_at");


--
-- Name: idx_xy_agent_messages_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_messages_owner" ON "public"."xy_agent_messages" USING "btree" ("owner_id");


--
-- Name: idx_xy_agent_messages_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_messages_run_id" ON "public"."xy_agent_messages" USING "btree" ("run_id");


--
-- Name: idx_xy_agent_messages_session_sequence; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_agent_messages_session_sequence" ON "public"."xy_agent_messages" USING "btree" ("session_id", "sequence");


--
-- Name: idx_xy_agent_run_tools_call_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_agent_run_tools_call_id" ON "public"."xy_agent_run_tools" USING "btree" ("call_id");


--
-- Name: idx_xy_agent_run_tools_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_run_tools_deleted_at" ON "public"."xy_agent_run_tools" USING "btree" ("deleted_at");


--
-- Name: idx_xy_agent_run_tools_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_run_tools_run_id" ON "public"."xy_agent_run_tools" USING "btree" ("run_id");


--
-- Name: idx_xy_agent_runs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_runs_deleted_at" ON "public"."xy_agent_runs" USING "btree" ("deleted_at");


--
-- Name: idx_xy_agent_runs_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_runs_owner" ON "public"."xy_agent_runs" USING "btree" ("owner_id");


--
-- Name: idx_xy_agent_runs_owner_idempotency; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_agent_runs_owner_idempotency" ON "public"."xy_agent_runs" USING "btree" ("owner_id", "idempotency_key");


--
-- Name: idx_xy_agent_runs_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_runs_session_id" ON "public"."xy_agent_runs" USING "btree" ("session_id");


--
-- Name: idx_xy_agent_runs_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_runs_status" ON "public"."xy_agent_runs" USING "btree" ("status");


--
-- Name: idx_xy_agent_sessions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_sessions_deleted_at" ON "public"."xy_agent_sessions" USING "btree" ("deleted_at");


--
-- Name: idx_xy_agent_sessions_last_message_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_sessions_last_message_at" ON "public"."xy_agent_sessions" USING "btree" ("last_message_at");


--
-- Name: idx_xy_agent_sessions_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_agent_sessions_owner" ON "public"."xy_agent_sessions" USING "btree" ("owner_id");


--
-- Name: idx_xy_ai_models_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_ai_models_deleted_at" ON "public"."xy_ai_models" USING "btree" ("deleted_at");


--
-- Name: idx_xy_ai_models_is_default; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_ai_models_is_default" ON "public"."xy_ai_models" USING "btree" ("is_default");


--
-- Name: idx_xy_business_events_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_business_events_account_id" ON "public"."xy_business_events" USING "btree" ("account_id");


--
-- Name: idx_xy_business_events_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_business_events_deleted_at" ON "public"."xy_business_events" USING "btree" ("deleted_at");


--
-- Name: idx_xy_business_events_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_business_events_order_id" ON "public"."xy_business_events" USING "btree" ("order_id");


--
-- Name: idx_xy_business_events_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_business_events_owner" ON "public"."xy_business_events" USING "btree" ("owner_id");


--
-- Name: idx_xy_business_events_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_business_events_product_id" ON "public"."xy_business_events" USING "btree" ("product_id");


--
-- Name: idx_xy_business_events_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_business_events_status" ON "public"."xy_business_events" USING "btree" ("status");


--
-- Name: idx_xy_business_events_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_business_events_type" ON "public"."xy_business_events" USING "btree" ("type");


--
-- Name: idx_xy_candidate_external; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_candidate_external" ON "public"."xy_resource_candidates" USING "btree" ("search_id", "mcp_server_code", "external_id");


--
-- Name: idx_xy_conversation_takeovers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_conversation_takeovers_deleted_at" ON "public"."xy_conversation_takeovers" USING "btree" ("deleted_at");


--
-- Name: idx_xy_credentials_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_credentials_account_id" ON "public"."xy_credentials" USING "btree" ("account_id");


--
-- Name: idx_xy_credentials_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_credentials_deleted_at" ON "public"."xy_credentials" USING "btree" ("deleted_at");


--
-- Name: idx_xy_delivery_resource_rule_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_delivery_resource_rule_hash" ON "public"."xy_delivery_resources" USING "btree" ("rule_id", "value_hash");


--
-- Name: idx_xy_delivery_resources_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_resources_deleted_at" ON "public"."xy_delivery_resources" USING "btree" ("deleted_at");


--
-- Name: idx_xy_delivery_resources_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_resources_owner" ON "public"."xy_delivery_resources" USING "btree" ("owner_id");


--
-- Name: idx_xy_delivery_resources_reserved_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_resources_reserved_task_id" ON "public"."xy_delivery_resources" USING "btree" ("reserved_task_id");


--
-- Name: idx_xy_delivery_resources_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_resources_status" ON "public"."xy_delivery_resources" USING "btree" ("status");


--
-- Name: idx_xy_delivery_rule_product_sku; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_delivery_rule_product_sku" ON "public"."xy_delivery_rules" USING "btree" ("product_id", "sku_key");


--
-- Name: idx_xy_delivery_rules_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_rules_deleted_at" ON "public"."xy_delivery_rules" USING "btree" ("deleted_at");


--
-- Name: idx_xy_delivery_rules_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_rules_owner" ON "public"."xy_delivery_rules" USING "btree" ("owner_id");


--
-- Name: idx_xy_delivery_tasks_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_tasks_account_id" ON "public"."xy_delivery_tasks" USING "btree" ("account_id");


--
-- Name: idx_xy_delivery_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_tasks_deleted_at" ON "public"."xy_delivery_tasks" USING "btree" ("deleted_at");


--
-- Name: idx_xy_delivery_tasks_next_attempt_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_tasks_next_attempt_at" ON "public"."xy_delivery_tasks" USING "btree" ("next_attempt_at");


--
-- Name: idx_xy_delivery_tasks_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_delivery_tasks_order_id" ON "public"."xy_delivery_tasks" USING "btree" ("order_id");


--
-- Name: idx_xy_delivery_tasks_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_tasks_owner" ON "public"."xy_delivery_tasks" USING "btree" ("owner_id");


--
-- Name: idx_xy_delivery_tasks_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_tasks_resource_id" ON "public"."xy_delivery_tasks" USING "btree" ("resource_id");


--
-- Name: idx_xy_delivery_tasks_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_delivery_tasks_status" ON "public"."xy_delivery_tasks" USING "btree" ("status");


--
-- Name: idx_xy_mcp_servers_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_mcp_servers_code" ON "public"."xy_mcp_servers" USING "btree" ("code");


--
-- Name: idx_xy_mcp_servers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_mcp_servers_deleted_at" ON "public"."xy_mcp_servers" USING "btree" ("deleted_at");


--
-- Name: idx_xy_messages_account_pnm; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_messages_account_pnm" ON "public"."xy_messages" USING "btree" ("account_id", "pnm_id");


--
-- Name: idx_xy_messages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_messages_deleted_at" ON "public"."xy_messages" USING "btree" ("deleted_at");


--
-- Name: idx_xy_messages_external_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_messages_external_order_id" ON "public"."xy_messages" USING "btree" ("external_order_id");


--
-- Name: idx_xy_messages_external_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_messages_external_product_id" ON "public"."xy_messages" USING "btree" ("external_product_id");


--
-- Name: idx_xy_messages_message_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_messages_message_time" ON "public"."xy_messages" USING "btree" ("message_time");


--
-- Name: idx_xy_messages_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_messages_owner" ON "public"."xy_messages" USING "btree" ("owner_id");


--
-- Name: idx_xy_messages_session; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_messages_session" ON "public"."xy_messages" USING "btree" ("session_id");


--
-- Name: idx_xy_order_events_account_pnm; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_order_events_account_pnm" ON "public"."xy_order_events" USING "btree" ("account_id", "pnm_id");


--
-- Name: idx_xy_order_events_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_order_events_deleted_at" ON "public"."xy_order_events" USING "btree" ("deleted_at");


--
-- Name: idx_xy_order_events_event_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_order_events_event_time" ON "public"."xy_order_events" USING "btree" ("event_time");


--
-- Name: idx_xy_order_events_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_order_events_order_id" ON "public"."xy_order_events" USING "btree" ("order_id");


--
-- Name: idx_xy_order_events_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_order_events_owner" ON "public"."xy_order_events" USING "btree" ("owner_id");


--
-- Name: idx_xy_orders_account_external; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_orders_account_external" ON "public"."xy_orders" USING "btree" ("account_id", "external_order_id");


--
-- Name: idx_xy_orders_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_orders_deleted_at" ON "public"."xy_orders" USING "btree" ("deleted_at");


--
-- Name: idx_xy_orders_external_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_orders_external_product_id" ON "public"."xy_orders" USING "btree" ("external_product_id");


--
-- Name: idx_xy_orders_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_orders_owner" ON "public"."xy_orders" USING "btree" ("owner_id");


--
-- Name: idx_xy_orders_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_orders_product_id" ON "public"."xy_orders" USING "btree" ("product_id");


--
-- Name: idx_xy_orders_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_orders_session_id" ON "public"."xy_orders" USING "btree" ("session_id");


--
-- Name: idx_xy_orders_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_orders_status" ON "public"."xy_orders" USING "btree" ("status");


--
-- Name: idx_xy_product_settings_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_product_settings_deleted_at" ON "public"."xy_product_settings" USING "btree" ("deleted_at");


--
-- Name: idx_xy_product_settings_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_product_settings_product_id" ON "public"."xy_product_settings" USING "btree" ("product_id");


--
-- Name: idx_xy_products_account_external; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_products_account_external" ON "public"."xy_products" USING "btree" ("account_id", "external_product_id");


--
-- Name: idx_xy_products_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_products_deleted_at" ON "public"."xy_products" USING "btree" ("deleted_at");


--
-- Name: idx_xy_products_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_products_owner" ON "public"."xy_products" USING "btree" ("owner_id");


--
-- Name: idx_xy_relist_tasks_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_relist_tasks_account_id" ON "public"."xy_relist_tasks" USING "btree" ("account_id");


--
-- Name: idx_xy_relist_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_relist_tasks_deleted_at" ON "public"."xy_relist_tasks" USING "btree" ("deleted_at");


--
-- Name: idx_xy_relist_tasks_next_attempt_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_relist_tasks_next_attempt_at" ON "public"."xy_relist_tasks" USING "btree" ("next_attempt_at");


--
-- Name: idx_xy_relist_tasks_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_relist_tasks_order_id" ON "public"."xy_relist_tasks" USING "btree" ("order_id");


--
-- Name: idx_xy_relist_tasks_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_relist_tasks_owner" ON "public"."xy_relist_tasks" USING "btree" ("owner_id");


--
-- Name: idx_xy_relist_tasks_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_relist_tasks_product_id" ON "public"."xy_relist_tasks" USING "btree" ("product_id");


--
-- Name: idx_xy_relist_tasks_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_relist_tasks_status" ON "public"."xy_relist_tasks" USING "btree" ("status");


--
-- Name: idx_xy_reply_contents_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_reply_contents_deleted_at" ON "public"."xy_reply_contents" USING "btree" ("deleted_at");


--
-- Name: idx_xy_reply_contents_rule_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_reply_contents_rule_id" ON "public"."xy_reply_contents" USING "btree" ("rule_id");


--
-- Name: idx_xy_reply_rule_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_reply_rule_unique" ON "public"."xy_reply_rules" USING "btree" ("product_id", "type", "keyword");


--
-- Name: idx_xy_reply_rules_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_reply_rules_deleted_at" ON "public"."xy_reply_rules" USING "btree" ("deleted_at");


--
-- Name: idx_xy_reply_rules_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_reply_rules_owner" ON "public"."xy_reply_rules" USING "btree" ("owner_id");


--
-- Name: idx_xy_reply_tasks_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_reply_tasks_account_id" ON "public"."xy_reply_tasks" USING "btree" ("account_id");


--
-- Name: idx_xy_reply_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_reply_tasks_deleted_at" ON "public"."xy_reply_tasks" USING "btree" ("deleted_at");


--
-- Name: idx_xy_reply_tasks_external_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_reply_tasks_external_product_id" ON "public"."xy_reply_tasks" USING "btree" ("external_product_id");


--
-- Name: idx_xy_reply_tasks_message_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_reply_tasks_message_id" ON "public"."xy_reply_tasks" USING "btree" ("message_id");


--
-- Name: idx_xy_reply_tasks_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_reply_tasks_owner" ON "public"."xy_reply_tasks" USING "btree" ("owner_id");


--
-- Name: idx_xy_reply_tasks_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_reply_tasks_product_id" ON "public"."xy_reply_tasks" USING "btree" ("product_id");


--
-- Name: idx_xy_reply_tasks_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_reply_tasks_session_id" ON "public"."xy_reply_tasks" USING "btree" ("session_id");


--
-- Name: idx_xy_reply_tasks_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_reply_tasks_status" ON "public"."xy_reply_tasks" USING "btree" ("status");


--
-- Name: idx_xy_resource_candidates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_candidates_deleted_at" ON "public"."xy_resource_candidates" USING "btree" ("deleted_at");


--
-- Name: idx_xy_resource_candidates_search_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_candidates_search_id" ON "public"."xy_resource_candidates" USING "btree" ("search_id");


--
-- Name: idx_xy_resource_searches_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_searches_deleted_at" ON "public"."xy_resource_searches" USING "btree" ("deleted_at");


--
-- Name: idx_xy_resource_searches_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_searches_owner" ON "public"."xy_resource_searches" USING "btree" ("owner_id");


--
-- Name: idx_xy_resource_source_external; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_resource_source_external" ON "public"."xy_resource_sources" USING "btree" ("resource_id", "mcp_server_code", "external_id");


--
-- Name: idx_xy_resource_sources_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_sources_deleted_at" ON "public"."xy_resource_sources" USING "btree" ("deleted_at");


--
-- Name: idx_xy_resource_sources_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_sources_resource_id" ON "public"."xy_resource_sources" USING "btree" ("resource_id");


--
-- Name: idx_xy_resource_task_idempotency; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_resource_task_idempotency" ON "public"."xy_resource_tasks" USING "btree" ("owner_id", "idempotency_key");


--
-- Name: idx_xy_resource_tasks_agent_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_tasks_agent_run_id" ON "public"."xy_resource_tasks" USING "btree" ("agent_run_id");


--
-- Name: idx_xy_resource_tasks_candidate_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_tasks_candidate_id" ON "public"."xy_resource_tasks" USING "btree" ("candidate_id");


--
-- Name: idx_xy_resource_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_tasks_deleted_at" ON "public"."xy_resource_tasks" USING "btree" ("deleted_at");


--
-- Name: idx_xy_resource_tasks_external_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_tasks_external_task_id" ON "public"."xy_resource_tasks" USING "btree" ("external_task_id");


--
-- Name: idx_xy_resource_tasks_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_tasks_owner" ON "public"."xy_resource_tasks" USING "btree" ("owner_id");


--
-- Name: idx_xy_resource_tasks_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resource_tasks_state" ON "public"."xy_resource_tasks" USING "btree" ("state");


--
-- Name: idx_xy_resources_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resources_deleted_at" ON "public"."xy_resources" USING "btree" ("deleted_at");


--
-- Name: idx_xy_resources_object_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_resources_object_key" ON "public"."xy_resources" USING "btree" ("object_key");


--
-- Name: idx_xy_resources_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resources_owner" ON "public"."xy_resources" USING "btree" ("owner_id");


--
-- Name: idx_xy_resources_sha256; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resources_sha256" ON "public"."xy_resources" USING "btree" ("sha256");


--
-- Name: idx_xy_resources_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_resources_task_id" ON "public"."xy_resources" USING "btree" ("task_id");


--
-- Name: idx_xy_takeovers_account_session; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX "idx_xy_takeovers_account_session" ON "public"."xy_conversation_takeovers" USING "btree" ("account_id", "session_id");


--
-- Name: idx_xy_takeovers_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idx_xy_takeovers_owner" ON "public"."xy_conversation_takeovers" USING "btree" ("owner_id");


--
-- Name: tel_resource_files tel_resource_files_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_resource_files"
    ADD CONSTRAINT "tel_resource_files_task_id_fkey" FOREIGN KEY ("task_id") REFERENCES "public"."tel_resource_tasks"("id") ON DELETE CASCADE;


--
-- Name: tel_resource_sources tel_resource_sources_dialog_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_resource_sources"
    ADD CONSTRAINT "tel_resource_sources_dialog_id_fkey" FOREIGN KEY ("dialog_id") REFERENCES "public"."tel_dialogs"("id") ON DELETE SET NULL;


--
-- Name: tel_resource_sources tel_resource_sources_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY "public"."tel_resource_sources"
    ADD CONSTRAINT "tel_resource_sources_file_id_fkey" FOREIGN KEY ("file_id") REFERENCES "public"."tel_resource_files"("id") ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--


