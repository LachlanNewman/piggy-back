-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE gender_enum AS ENUM ('male', 'female', 'unknown');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    first_name    TEXT         NOT NULL,
    last_name     TEXT         NOT NULL,
    email         TEXT         NOT NULL UNIQUE,
    date_of_birth DATE         NOT NULL,
    weight        NUMERIC(5,2) NOT NULL,
    gender        gender_enum  NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS gender_enum;
