-- +goose Up
DROP TABLE IF EXISTS jwt_blacklists;

-- +goose Down
CREATE SEQUENCE jwt_blacklists_id_seq;
CREATE TABLE jwt_blacklists (
    id bigint NOT NULL DEFAULT nextval('jwt_blacklists_id_seq'::regclass),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    jwt text,
    CONSTRAINT jwt_blacklists_pkey PRIMARY KEY (id)
);
ALTER SEQUENCE jwt_blacklists_id_seq OWNED BY jwt_blacklists.id;
CREATE INDEX idx_jwt_blacklists_deleted_at ON jwt_blacklists USING btree (deleted_at);
