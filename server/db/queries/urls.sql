-- name: CreateURL :one
INSERT INTO urls (original_url, short_code, user_id)
VALUES ($1, $2, $3)
RETURNING short_code;

-- name: GetURL :one
SELECT original_url
FROM urls 
WHERE short_code = $1;

-- name: GetURLStats :one
SELECT click_count, created_at
FROM urls
WHERE short_code = $1;

-- name: IncrementClickCount :exec
UPDATE urls
SET click_count = click_count + 1
WHERE short_code = $1;

-- name: GetURLsByUserID :many
SELECT *
FROM urls
WHERE user_id = $1
ORDER BY created_at DESC;
