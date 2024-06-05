-- +goose Up
-- +goose StatementBegin
CREATE INDEX trips_vehicle_token_id_idx ON trips (vehicle_token_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX trips_vehicle_token_id_idx;
-- +goose StatementEnd
