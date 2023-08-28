-- +goose Up
-- +goose StatementBegin

CREATE TABLE vehicles 
(
    token_id NUMERIC(78, 0) PRIMARY KEY, 
    user_device_id CHAR(27), 
    encryption_key TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = trips_api, public;

DROP TABLE vehicles;
-- +goose StatementEnd