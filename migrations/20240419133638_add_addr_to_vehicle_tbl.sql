-- +goose Up
-- +goose StatementBegin
SET search_path = trips_api, public;

ALTER TABLE vehicles ADD COLUMN owner_address bytea
    CONSTRAINT vehicles_owner_address_check CHECK (length(owner_address) = 20);

ALTER TABLE trips ADD COLUMN owner_address bytea
    CONSTRAINT trip_owner_address CHECK (length(owner_address) = 20);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SET search_path = trips_api, public;

ALTER TABLE vehicles DROP COLUMN owner_address;

-- +goose StatementEnd
