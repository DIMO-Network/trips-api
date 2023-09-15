-- +goose Up
-- +goose StatementBegin
SET search_path = trips_api, public;

ALTER TABLE trips
    ADD COLUMN tx_hash BYTEA;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = trips_api, public;

ALTER TABLE trips
    DROP COLUMN tx_hash;
-- +goose StatementEnd