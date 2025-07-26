-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $2,
    $1
)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;