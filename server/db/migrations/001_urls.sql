-- +goose Up
CREATE TABLE urls(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_url VARCHAR UNIQUE NOT NULL,
    short_code VARCHAR UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    click_count INTEGER NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE urls;