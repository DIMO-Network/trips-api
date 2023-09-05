-- +goose Up
-- +goose StatementBegin
CREATE TABLE vehicles(
    token_id int CONSTRAINT vehicles_pkey PRIMARY KEY, 
    user_device_id varchar NOT NULL CONSTRAINT vehicles_user_device_id_key UNIQUE
);

ALTER TABLE trips ADD COLUMN vehicle_token_id int NOT NULL CONSTRAINT trips_vehicle_token_id_fkey REFERENCES vehicles(token_id);
ALTER TABLE trips DROP COLUMN user_device_id;

ALTER TABLE trips ADD COLUMN encryption_key bytea;
ALTER TABLE trips ADD COLUMN bundlr_id varchar;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trips DROP COLUMN bundlr_id;
ALTER TABLE trips DROP COLUMN encryption_key;
ALTER TABLE trips DROP COLUMN vehicle_token_id;

ALTER TABLE trips ADD COLUMN user_device_id text NOT NULL;

DROP TABLE vehicles;
-- +goose StatementEnd