-- +goose Up
ALTER TABLE ride_requests
    ADD COLUMN driver_id  TEXT        NOT NULL DEFAULT '',
    ADD COLUMN expires_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE ride_requests DROP CONSTRAINT IF EXISTS ride_requests_status_check;
ALTER TABLE ride_requests
    ADD CONSTRAINT ride_requests_status_check
    CHECK (status IN ('pending', 'accepted', 'declined', 'expired'));

-- Treat existing rows as immediately expired with no driver
UPDATE ride_requests SET expires_at = now() WHERE expires_at = now();

-- +goose Down
ALTER TABLE ride_requests
    DROP COLUMN IF EXISTS driver_id,
    DROP COLUMN IF EXISTS expires_at;

ALTER TABLE ride_requests DROP CONSTRAINT IF EXISTS ride_requests_status_check;
ALTER TABLE ride_requests
    ADD CONSTRAINT ride_requests_status_check
    CHECK (status IN ('pending', 'accepted'));
