-- +goose Up
-- Truncate existing rows — they have no auth_subject (dev/staging only, see design.md migration plan)
TRUNCATE users;

ALTER TABLE users
    ADD COLUMN auth_subject     TEXT    NOT NULL UNIQUE,
    ADD COLUMN profile_complete BOOLEAN NOT NULL DEFAULT false,
    ALTER COLUMN date_of_birth  DROP NOT NULL,
    ALTER COLUMN weight         DROP NOT NULL,
    ALTER COLUMN gender         DROP NOT NULL;

-- +goose Down
DELETE FROM users WHERE date_of_birth IS NULL OR weight IS NULL OR gender IS NULL;
ALTER TABLE users
    DROP COLUMN IF EXISTS profile_complete,
    DROP COLUMN IF EXISTS auth_subject,
    ALTER COLUMN date_of_birth SET NOT NULL,
    ALTER COLUMN weight        SET NOT NULL,
    ALTER COLUMN gender        SET NOT NULL;
