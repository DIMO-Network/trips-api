-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS trips_api;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA trips_api CASCADE;
-- +goose StatementEnd
