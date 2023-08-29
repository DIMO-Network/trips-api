-- +goose Up
-- +goose StatementBegin
ALTER TABLE fulltrips RENAME TO trips;

ALTER TABLE trips RENAME COLUMN trip_id TO id;
ALTER TABLE trips RENAME COLUMN device_id TO user_device_id;
ALTER TABLE trips RENAME COLUMN trip_start TO "start";
ALTER TABLE trips RENAME COLUMN trip_end TO "end";

ALTER TABLE trips ALTER COLUMN user_device_id SET NOT NULL;
ALTER TABLE trips ALTER COLUMN "start" SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trips ALTER COLUMN user_device_id DROP NOT NULL;
ALTER TABLE trips ALTER COLUMN "start" DROP NOT NULL;  

ALTER TABLE trips RENAME COLUMN id TO trip_id;
ALTER TABLE trips RENAME COLUMN user_device_id TO device_id;
ALTER TABLE trips RENAME COLUMN "start" TO trip_start;
ALTER TABLE trips RENAME COLUMN "end" TO trip_end;

ALTER TABLE trips RENAME TO fulltrips;
-- +goose StatementEnd
