-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES(
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: Reset :exec
DELETE FROM users;