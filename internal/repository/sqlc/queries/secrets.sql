-- name: CreateSecret :one
INSERT INTO secrets (user_id, type, name, encrypted_data, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetSecret :one
SELECT * FROM secrets
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListSecrets :many
SELECT * FROM secrets
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY updated_at DESC;

-- name: UpdateSecret :one
UPDATE secrets
SET name = $3, encrypted_data = $4, type = $5, version = $6, updated_at = $7
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteSecret :one
UPDATE secrets
SET deleted_at = $3, updated_at = $3, version = $4
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: ListSecretsSince :many
SELECT * FROM secrets
WHERE user_id = $1 AND updated_at > $2
ORDER BY updated_at ASC;
