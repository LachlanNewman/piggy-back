-- +goose Up
CREATE TABLE IF NOT EXISTS user_locations (
    user_id    TEXT             NOT NULL UNIQUE,
    lat        DOUBLE PRECISION NOT NULL,
    lng        DOUBLE PRECISION NOT NULL,
    updated_at TIMESTAMPTZ      NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS user_locations;
