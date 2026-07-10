-- name: CreateURL :one
INSERT INTO urls (original_url, short_code)
VALUES ($1, $2)
RETURNING id, original_url, short_code, created_at;

-- name: GetURL :one
SELECT id, original_url, short_code, created_at
FROM urls 
WHERE short_code = $1;

-- name: GetURLStats :one
SELECT click_count, created_at
FROM urls
WHERE short_code = $1;