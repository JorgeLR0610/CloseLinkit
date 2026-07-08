-- name: CreateURL :one
INSERT INTO urls (original_url, short_code, created_at, expires_at, updated_at)
VALUES ($1, $2, now(), $3, now())
RETURNING *;

-- name: GetURL :one
SELECT id, original_url, short_code, created_at, updated_at, expires_at
FROM urls 
WHERE short_code = $1;

-- name: updateOriginalURL :one
UPDATE urls
SET original_url = $1, updated_at = $2
WHERE short_code = $3
RETURNING original_url, short_code, created_at, updated_at, expires_at;

-- name: deleteURL :exec
DELETE FROM urls
WHERE short_code = $1;

-- name: getURLStats :one
SELECT * FROM urls
WHERE short_code = $1;