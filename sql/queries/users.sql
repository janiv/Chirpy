-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING id, created_at, updated_at, email;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUserEmailAndPassword :one
UPDATE users SET hashed_password = $1, email = $2, updated_at = $3 WHERE id = $4 RETURNING *;

-- name: Reset :exec
DELETE FROM users;