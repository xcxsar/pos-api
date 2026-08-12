-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (gen_random_uuid(), now(), now(), $1, $2)
RETURNING id, created_at, updated_at, email;

-- name: GetUserByID :one
SELECT id, created_at, updated_at, email FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserEmail :one
UPDATE users SET email = $1, updated_at = now() WHERE id = $2
RETURNING id, created_at, updated_at, email;

-- name: UpdateUserPassword :one
UPDATE users SET hashed_password = $1, updated_at = now() WHERE id = $2
RETURNING id, created_at, updated_at, email;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;