-- +goose Up
CREATE TABLE users(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    email TEXT UNIQUE NOT NULL,
    email_verified_at TIMESTAMPTZ,
    hashed_password TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_users_email ON users (LOWER(email));

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE users;