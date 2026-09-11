-- +goose Up
ALTER TABLE urls 
ADD COLUMN user_id UUID,
ADD CONSTRAINT user_fk
    FOREIGN KEY(user_id)
    REFERENCES users(id)
    ON DELETE SET NULL;

CREATE INDEX idx_urls_user_id ON urls(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_urls_user_id;

ALTER TABLE urls
DROP COLUMN user_id;