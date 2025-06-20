// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: refresh_tokens.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createRefreshToken = `-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, user_id)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING token, created_at, updated_at, expires_at, revoked_at, user_id
`

type CreateRefreshTokenParams struct {
	Token     string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
	UserID    uuid.UUID
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, createRefreshToken,
		arg.Token,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.ExpiresAt,
		arg.UserID,
	)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiresAt,
		&i.RevokedAt,
		&i.UserID,
	)
	return i, err
}

const getRefreshTokenByToken = `-- name: GetRefreshTokenByToken :one
SELECT token, created_at, updated_at, expires_at, revoked_at, user_id FROM refresh_tokens WHERE token = $1
`

func (q *Queries) GetRefreshTokenByToken(ctx context.Context, token string) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, getRefreshTokenByToken, token)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiresAt,
		&i.RevokedAt,
		&i.UserID,
	)
	return i, err
}

const getRefreshTokenByUserID = `-- name: GetRefreshTokenByUserID :one
SELECT token, created_at, updated_at, expires_at, revoked_at, user_id FROM refresh_tokens WHERE user_id = $1
`

func (q *Queries) GetRefreshTokenByUserID(ctx context.Context, userID uuid.UUID) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, getRefreshTokenByUserID, userID)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiresAt,
		&i.RevokedAt,
		&i.UserID,
	)
	return i, err
}

const getUserFromRefreshToken = `-- name: GetUserFromRefreshToken :one
SELECT user_id FROM refresh_tokens WHERE token = $1
`

func (q *Queries) GetUserFromRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getUserFromRefreshToken, token)
	var user_id uuid.UUID
	err := row.Scan(&user_id)
	return user_id, err
}

const updateRefreshTokenRevoke = `-- name: UpdateRefreshTokenRevoke :exec
UPDATE refresh_tokens SET updated_at = $1, revoked_at = $1 WHERE token = $2
`

type UpdateRefreshTokenRevokeParams struct {
	UpdatedAt time.Time
	Token     string
}

func (q *Queries) UpdateRefreshTokenRevoke(ctx context.Context, arg UpdateRefreshTokenRevokeParams) error {
	_, err := q.db.ExecContext(ctx, updateRefreshTokenRevoke, arg.UpdatedAt, arg.Token)
	return err
}

const updateRefreshTokenUpdateTime = `-- name: UpdateRefreshTokenUpdateTime :exec
UPDATE refresh_tokens SET updated_at = $1 WHERE token = $2
`

type UpdateRefreshTokenUpdateTimeParams struct {
	UpdatedAt time.Time
	Token     string
}

func (q *Queries) UpdateRefreshTokenUpdateTime(ctx context.Context, arg UpdateRefreshTokenUpdateTimeParams) error {
	_, err := q.db.ExecContext(ctx, updateRefreshTokenUpdateTime, arg.UpdatedAt, arg.Token)
	return err
}
