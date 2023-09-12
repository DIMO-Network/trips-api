-- +goose Up
-- +goose StatementBegin
SET search_path = trips_api, public;

ALTER TABLE trips
    RENAME COLUMN "start" TO start_time;

ALTER TABLE trips
    RENAME COLUMN "end" TO end_time;

ALTER TABLE trips
    ADD COLUMN start_position POINT;

ALTER TABLE trips
    ADD COLUMN end_position POINT;

ALTER TABLE trips
    ALTER COLUMN start_time SET NOT NULL;

ALTER TABLE trips
    ALTER COLUMN start_position SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SET search_path = trips_api, public;

ALTER TABLE trips
    ALTER COLUMN start_position DROP NOT NULL;

ALTER TABLE trips
    ALTER COLUMN start_time DROP NOT NULL;

ALTER TABLE trips
    DROP COLUMN start_position;

ALTER TABLE trips
    DROP COLUMN end_position;


ALTER TABLE trips
    RENAME COLUMN start_time TO "start";

ALTER TABLE trips
    RENAME COLUMN end_time TO "end";

-- +goose StatementEnd
