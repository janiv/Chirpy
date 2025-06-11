-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, user_id)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetRefreshTokenByUserID :one
SELECT * FROM refresh_tokens WHERE user_id = $1;

-- name: GetUserFromRefreshToken :one
SELECT user_id FROM refresh_tokens WHERE token = $1;

-- name: GetRefreshTokenByToken :one
SELECT * FROM refresh_tokens WHERE token = $1;

-- name: UpdateRefreshTokenUpdateTime :exec
UPDATE refresh_tokens SET updated_at = $1 WHERE token = $2;

-- name: UpdateRefreshTokenRevoke :exec
UPDATE refresh_tokens SET updated_at = $1, revoked_at = $1 WHERE token = $2;