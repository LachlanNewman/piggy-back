-- +goose Up
ALTER TABLE ride_requests
    ADD COLUMN status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted')),
    ALTER COLUMN pickup_lat  DROP NOT NULL,
    ALTER COLUMN pickup_lng  DROP NOT NULL,
    ALTER COLUMN dropoff_lat DROP NOT NULL,
    ALTER COLUMN dropoff_lng DROP NOT NULL;

-- +goose Down
ALTER TABLE ride_requests
    DROP COLUMN status,
    ALTER COLUMN pickup_lat  SET NOT NULL,
    ALTER COLUMN pickup_lng  SET NOT NULL,
    ALTER COLUMN dropoff_lat SET NOT NULL,
    ALTER COLUMN dropoff_lng SET NOT NULL;
