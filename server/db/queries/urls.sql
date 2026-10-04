-- name: CreateURL :one
INSERT INTO urls (original_url, short_code, user_id, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING short_code, expires_at;

-- name: GetURL :one
SELECT original_url
FROM urls 
WHERE short_code = $1
AND (expires_at is NULL OR expires_at > NOW());

-- name: GetURLStats :one
SELECT click_count, created_at
FROM urls
WHERE short_code = $1;

-- name: IncrementClickCount :exec
UPDATE urls
SET click_count = click_count + 1
WHERE short_code = $1;

-- name: GetURLsByUserID :many
SELECT original_url, short_code, created_at, click_count
FROM urls
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteExpiredURLs :exec
DELETE FROM urls
WHERE expires_at <= NOW();

-- name: ClaimURLsByShortCodes :many
UPDATE urls
SET user_id = $1, expires_at = NULL
WHERE short_code = ANY(@short_codes::text[]) AND user_id IS NULL
RETURNING short_code;

-- name: DeleteURLByShortCode :execrows
DELETE FROM urls
WHERE short_code = $1 AND user_id = $2;
