-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps ORDER BY created_at;

-- name: GetChirpsByUserID :many
SELECT * FROM chirps WHERE user_id = $1 ORDER BY created_at;

-- name: GetChirpByID :one
SELECT * FROM chirps WHERE id = $1 LIMIT 1;

-- name: DeleteChirpByID :exec
DELETE FROM chirps WHERE id = $1 AND user_id=$2;