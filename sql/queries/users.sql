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
SELECT id, created_at, updated_at, email, hashed_password, is_chirpy_red FROM users
WHERE email=$1;

-- name: ClearUsers :exec
DELETE FROM users;

-- name: UpdateUser :one
UPDATE users
SET hashed_password=$2, email=$3
WHERE id=$1
RETURNING *;

-- name: UpdateChirpyRedStatus :one
UPDATE users
SET is_chirpy_red = @status
WHERE id = @user_id
RETURNING *;