-- name: CreateRefreshToken :one
INSERT INTO tokens (token, user_id, expires_at, created_at, updated_at, revoked_at)
VALUES (@token, @user_id, @expires_at, NOW(), NOW(), NULL)
RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at;

-- name: GetUserFromRefreshToken :one
SELECT token, user_id, expires_at, tokens.created_at, tokens.updated_at, revoked_at
FROM tokens
INNER JOIN users ON users.id = tokens.user_id
WHERE token=$1;
 
-- name: RevokeToken :exec
UPDATE tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE token = @token; 