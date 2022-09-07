-- +goose Up
-- +goose StatementBegin
CREATE TABLE fulltrips (trip_id TEXT PRIMARY KEY, device_id TEXT, trip_start timestamptz, trip_end timestamptz);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE fulltrips;
-- +goose StatementEnd
