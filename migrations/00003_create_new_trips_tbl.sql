-- +goose Up
-- +goose StatementBegin
SET search_path = trips_api, public;

DROP TABLE fulltrips;

CREATE TABLE trips 
(
    vehicle_token_id numeric(78, 0) NOT NULL,
    "start" timestamptz NOT NULL,
    "end" timestamptz NOT NULL,
    start_hex BIGINT NOT NULL,
    end_hex BIGINT NOT NULL,
    bunldr_id text NOT NULL,
    encryption_key text NOT NULL,
    trip_token_id  numeric(78, 0)
        CONSTRAINT trip_token_id_key UNIQUE,
    PRIMARY KEY(vehicle_token_id, "start")
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SET search_path = trips_api, public;

DROP TABLE trips;

-- +goose StatementEnd