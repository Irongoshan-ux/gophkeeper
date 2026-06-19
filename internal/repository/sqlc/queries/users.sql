-- name: CreateUser :one
INSERT INTO users (login, password_hash, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByLogin :one
SELECT * FROM users WHERE login = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;
