-- +goose Up
ALTER TABLE urls
ADD COLUMN expires_at TIMESTAMPTZ;

CREATE INDEX idx_urls_expired ON urls(expires_at) WHERE expires_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_urls_expired;

ALTER TABLE urls
DROP COLUMN expires_at;