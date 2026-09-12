-- name: CreateUser :one
INSERT INTO users (email, hashed_password)
VALUES ($1, $2)
RETURNING id, created_at, updated_at, email;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE LOWER(email) = LOWER($1);

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = $1,
    updated_at = NOW()
WHERE id = $2;

-- name: MarkEmailVerified :exec
UPDATE users
SET email_verified_at = NOW(),
    updated_at = NOW()
WHERE id = $1;
