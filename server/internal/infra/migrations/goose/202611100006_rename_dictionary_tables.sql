-- +goose Up
ALTER TABLE s_dictionaries RENAME TO s_dict;
ALTER TABLE s_dictionary_details RENAME TO s_dict_detail;

ALTER SEQUENCE s_dictionaries_id_seq RENAME TO s_dict_id_seq;
ALTER SEQUENCE s_dictionary_details_id_seq RENAME TO s_dict_detail_id_seq;

ALTER INDEX s_dictionaries_pkey RENAME TO s_dict_pkey;
ALTER INDEX s_dictionary_details_pkey RENAME TO s_dict_detail_pkey;
ALTER INDEX idx_s_dictionaries_deleted_at RENAME TO idx_s_dict_deleted_at;
ALTER INDEX idx_s_dictionary_details_deleted_at RENAME TO idx_s_dict_detail_deleted_at;

-- +goose Down
ALTER INDEX idx_s_dict_detail_deleted_at RENAME TO idx_s_dictionary_details_deleted_at;
ALTER INDEX idx_s_dict_deleted_at RENAME TO idx_s_dictionaries_deleted_at;
ALTER INDEX s_dict_detail_pkey RENAME TO s_dictionary_details_pkey;
ALTER INDEX s_dict_pkey RENAME TO s_dictionaries_pkey;

ALTER SEQUENCE s_dict_detail_id_seq RENAME TO s_dictionary_details_id_seq;
ALTER SEQUENCE s_dict_id_seq RENAME TO s_dictionaries_id_seq;

ALTER TABLE s_dict_detail RENAME TO s_dictionary_details;
ALTER TABLE s_dict RENAME TO s_dictionaries;
