-- name: CreateURL :one
INSERT INTO urls (original_url, short_code)
VALUES ($1, $2)
RETURNING id, original_url, short_code, created_at;

-- name: GetURL :one
SELECT original_url
FROM urls 
WHERE short_code = $1;

-- name: GetURLStats :one
SELECT original_url, click_count, created_at
FROM urls
WHERE short_code = $1;

-- name: IncrementClickCount :exec
UPDATE urls
SET click_count = click_count + 1
WHERE short_code = $1;
