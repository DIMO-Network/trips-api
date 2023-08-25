-- +goose Up
-- +goose StatementBegin
SET search_path = trips_api, public;

ALTER TABLE trips
    ADD COLUMN start_position POINT NOT NULL;

ALTER TABLE trips
    ADD COLUMN end_position POINT NOT NULL;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SET search_path = trips_api, public;

ALTER TABLE trips
    DROP COLUMN start_position;

ALTER TABLE trips
    DROP COLUMN end_position;

-- +goose StatementEnd