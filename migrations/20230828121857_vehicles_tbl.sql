-- +goose Up
-- +goose StatementBegin

CREATE TABLE vehicles 
(
    token_id NUMERIC(78, 0) PRIMARY KEY, 
    user_device_id CHAR(27) NOT NULL, 
    encryption_key BYTEA NOT NULL
);

ALTER TABLE trips
ADD COLUMN vehicle_token_id numeric(78, 0) 
    CONSTRAINT vehicle_token_id_fkey REFERENCES vehicles(token_id);

ALTER TABLE trips
ADD COLUMN nonce bytea NOT NULL;

ALTER TABLE trips
ADD COLUMN bundlr_id text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = trips_api, public;

DROP TABLE vehicles;

ALTER TABLE trips
DROP COLUMN vehicle_token_id;

ALTER TABLE trips
DROP COLUMN nonce;

ALTER TABLE trips
DROP COLUMN bundlr_id;
-- +goose StatementEnd