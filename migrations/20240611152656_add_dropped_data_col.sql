-- +goose Up
-- +goose StatementBegin
SET search_path = trips_api, public;
ALTER TABLE trips
    ADD COLUMN dropped_data BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SET search_path = trips_api, public;
ALTER TABLE trips
    DROP COLUMN dropped_data;

-- +goose StatementEnd
