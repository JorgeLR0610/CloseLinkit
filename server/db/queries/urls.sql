-- name: CreateURL :one
INSERT INTO urls (original_url, short_code, expires_at)
VALUES ($1, $2, $3)
RETURNING original_url, short_code, created_at, expires_at;

-- name: GetURL :one
SELECT id, original_url, short_code, created_at, expires_at
FROM urls 
WHERE short_code = $1;

-- name: GetURLStats :one
SELECT click_count, created_at, expires_at
FROM urls
WHERE short_code = $1;