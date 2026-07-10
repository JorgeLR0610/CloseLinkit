-- +goose Up
CREATE TABLE urls(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_url TEXT NOT NULL,
    short_code VARCHAR(7) NOT NULL CONSTRAINT urls_short_code_unique UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    click_count INTEGER NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE urls;