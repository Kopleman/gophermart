-- name: GetAllUsers :many
SELECT id, login, password_hash, created_at, updated_at, deleted_at FROM users ORDER BY created_at ASC;

-- name: GetUserByLogin :one
SELECT id, login, password_hash, created_at, updated_at, deleted_at FROM users WHERE login=$1 LIMIT 1;

-- name: GetUserById :one
SELECT id, login, password_hash, created_at, updated_at, deleted_at FROM users WHERE id=$1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (login, password_hash, created_at)
VALUES ($1, $2,  now())
RETURNING id, login, password_hash, created_at, updated_at, deleted_at;