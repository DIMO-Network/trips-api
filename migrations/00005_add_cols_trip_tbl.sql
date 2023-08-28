-- +goose Up
-- +goose StatementBegin

ALTER TABLE trips
ADD COLUMN vehicle_token_id numeric(78, 0) 
    CONSTRAINT vehicle_token_id_fkey REFERENCES vehicles(token_id);

ALTER TABLE trips
ADD COLUMN nonce numeric(20) NOT NULL;

ALTER TABLE trips
ADD COLUMN bundlr_id text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trips
DROP COLUMN vehicle_token_id;

ALTER TABLE trips
DROP COLUMN nonce;

ALTER TABLE trips
DROP COLUMN bundlr_id;
-- +goose StatementEnd
