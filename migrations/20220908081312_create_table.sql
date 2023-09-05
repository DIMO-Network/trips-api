-- +goose Up
-- +goose StatementBegin
SET search_path = trips_api, public;

CREATE TABLE fulltrips 
(
    trip_id TEXT PRIMARY KEY, 
    device_id TEXT, 
    trip_start timestamptz, 
    trip_end timestamptz
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = trips_api, public;

DROP TABLE fulltrips;
-- +goose StatementEnd