-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE email=$1;

-- name: ClearUsers :exec
DELETE FROM users;

-- name: UpdateUser :one
UPDATE users
SET hashed_password=$2, email=$3
WHERE id=$1
RETURNING *;