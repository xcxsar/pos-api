-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, now(), now(), $2, now() + interval '60 days', null)
RETURNING *;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = now(), updated_at = now() WHERE token = $1;

-- name: GetRefreshToken :one
SELECT * from refresh_tokens WHERE token = $1;

-- name: GetUserFromRefreshToken :one
SELECT user_id FROM refresh_tokens WHERE token = $1;